package com.bitgirder.etl;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.ObjectReceiver;
import com.bitgirder.lang.Lang;

import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.ProcessExit;
import com.bitgirder.process.ProcessRpcClient;

import com.bitgirder.mingle.model.MingleIdentifiedName;
import com.bitgirder.mingle.model.MingleIdentifiedNameGenerator;

import java.util.Map;
import java.util.Set;

public
final
class EtlTestProcessorFeed
extends AbstractVoidProcess
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static MingleIdentifiedNameGenerator NM_GEN =
        MingleIdentifiedNameGenerator.
            forPrefix( "bitgirder:etl@v1/etlTestProcessorFeed" );

    private final static EventHandler DEFAULT_EVENT_HANDLER =
        new AbstractEventHandler() {};

    private final Map< AbstractProcess< ? >, MingleIdentifiedName > procs;
    private final Map< AbstractProcess< ? >, Object > initStates;
    private final EtlTestRecordGenerator< ? > g;
    private final EtlTestReactor reactor;
    private final EventHandler eh;
    private final long feedLength;
    private final boolean expectRecordSetAbort;
    private final boolean sendShutdownOnComplete;
    private final int batchSize;

    private final Set< AbstractProcess< ? > > activeProcs = Lang.newSet();

    private
    EtlTestProcessorFeed( Builder b )
    {
        super( b.mixin( ProcessRpcClient.create() ) );

        this.procs = Lang.unmodifiableCopy( b.procs );
        inputs.isFalse( b.procs.isEmpty(), "Need at least one processor" );

        this.initStates = Lang.unmodifiableCopy( b.initStates );

        this.g = inputs.notNull( b.g, "g" );
        this.reactor = inputs.notNull( b.reactor, "reactor" );
        this.eh = inputs.notNull( b.eh, "eh" );
        this.feedLength = inputs.nonnegativeL( b.feedLength, "feedLength" );
        this.expectRecordSetAbort = b.expectRecordSetAbort;
        this.sendShutdownOnComplete = b.sendShutdownOnComplete;
        this.batchSize = b.batchSize;
    }

    public 
    Map< AbstractProcess< ? >, MingleIdentifiedName > 
    getProcessors()
    {
        return procs;
    }

    public long getFeedLength() { return feedLength; }

    final
    void
    getProcessorState( MingleIdentifiedName procId,
                       final ObjectReceiver< Object > recv )
    {
        inputs.notNull( recv, "recv" );

        EtlProcessors.GetProcessorState req =
            EtlProcessors.createGetProcessorState( procId );

        beginRpc( reactor.getStateManager(), req, 
            new DefaultRpcHandler() 
            {
                @Override 
                protected void rpcSucceeded( Object resp ) throws Exception {
                    recv.receive( resp );
                }
            }
        );
    }

    private
    void
    stopReactor()
        throws Exception
    {
        reactor.stopTestProcesses( getActivityContext() );
    }

//    private
//    void
//    getLastState()
//    {
//        getProcessorState(
//            new ObjectReceiver< Object >() {
//                public void receive( Object o ) throws Exception
//                {
//                    lastState = o;
//
//                    eh.completeTest(
//                        lastState,
//                        new AbstractTask() {
//                            protected void runImpl() throws Exception {
//                                stopReactor();
//                            }
//                        },
//                        EtlTestProcessorFeed.this
//                    );
//                }
//            }
//        );
//    }

    private
    void
    completeTest()
        throws Exception
    {
        eh.completeTest(
            procs,
            new AbstractTask() {
                protected void runImpl() throws Exception { stopReactor(); }
            },
            EtlTestProcessorFeed.this
        );
    }

    @Override
    protected
    final
    void
    childExited( AbstractProcess< ? > proc,
                 ProcessExit< ? > exit )
        throws Exception
    {
        if ( ! exit.isOk() ) fail( exit.getThrowable() );

        if ( activeProcs.remove( proc ) && activeProcs.isEmpty() ) 
        {
            completeTest();
        }

        if ( ! hasChildren() ) exit();
    }

    final
    void
    sendShutdown( boolean isUrgent )
    {
        for ( AbstractProcess< ? > p : activeProcs )
        {
            beginRpc( 
                p, 
                EtlProcessors.getShutdownRequest( isUrgent ),
                new DefaultRpcHandler() {}
            );
        }
    }

    private 
    void 
    procsDone() 
    { 
        if ( sendShutdownOnComplete ) sendShutdown( false ); 
    }

    private
    void
    advance( long remain )
        throws Exception
    {
        if ( remain > 0 ) sendNext( remain ); else procsDone();
    }

    private
    EtlRecordSet
    nextRecordSet( long remain )
        throws Exception
    {
        int len = (int) Math.min( batchSize, remain );

        Object[] data = new Object[ len ];

        long off = feedLength - remain;
        for ( int i = 0; i < len; ++i ) data[ i ] = g.next( off + i );

        return EtlRecordSets.create( data, len == remain );
    }
 
    private
    final
    class RecordSendHandler
    extends DefaultRpcHandler
    {
        private final long remain;
        
        private int waitCount = activeProcs.size();

        private RecordSendHandler( long remain ) { this.remain = remain; }

        private
        void
        advanceConditional()
        {
            try { if ( --waitCount == 0 ) advance( remain ); }
            catch ( Throwable th ) { fail( th ); }
        }

        @Override 
        public
        void 
        rpcSucceeded( Object procState,
                      ProcessRpcClient.Call call ) 
        {
            if ( procState != null )
            {
                MingleIdentifiedName procId =
                    state.get( procs, call.getDestination(), "procs" );

                EtlProcessors.SetProcessorState req =
                    EtlProcessors.createSetProcessorState( procId, procState );
    
                beginRpc( reactor.getStateManager(), req,
                    new DefaultRpcHandler() 
                    {
                        @Override
                        protected void rpcSucceeded() { advanceConditional(); }
                    }
                );
            }
            else advanceConditional();
        }

        @Override
        protected
        void
        rpcFailed( Throwable th )
        {
            if ( ! ( th instanceof EtlRecordSetAbortedException &&
                     expectRecordSetAbort ) )
            {
                fail( th );
            }
        }
    }

    private
    void
    sendNext( long remain )
        throws Exception
    {
        EtlRecordSet rs = nextRecordSet( remain );
        RecordSendHandler h = new RecordSendHandler( remain - rs.size() );

        for ( AbstractProcess< ? > p : activeProcs ) beginRpc( p, rs, h );

        eh.recordsSent( rs, this );
    }

    private void startFeed() throws Exception { advance( feedLength ); }

    private
    final
    class SetInitStateHandler
    extends DefaultRpcHandler
    {
        private int waitCount;

        private
        SetInitStateHandler( int waitCount ) 
        { 
            this.waitCount = waitCount; 
        }

        @Override
        protected
        void
        rpcSucceeded()
            throws Exception
        {
            if ( --waitCount == 0 ) startFeed();
        }
    }

    private
    Map< AbstractProcess< ? >, EtlProcessors.SetProcessorState >
    getInitialStateRequests()
    {
        Map< AbstractProcess< ? >, EtlProcessors.SetProcessorState > res =
            Lang.newMap();

        for ( Map.Entry< AbstractProcess< ? >, MingleIdentifiedName > e :
                procs.entrySet() )
        {
            Object obj = initStates.get( e.getKey() );

            if ( obj != null )
            {
                res.put( e.getKey(),
                    EtlProcessors.createSetProcessorState( e.getValue(), obj )
                );
            }
        }

        return res;
    }

    private
    void
    sendInitialStateReqs( 
        Map< AbstractProcess< ? >, EtlProcessors.SetProcessorState > m )
    {
        SetInitStateHandler h = new SetInitStateHandler( m.size() );

        for ( Map.Entry< AbstractProcess< ? >, 
                         EtlProcessors.SetProcessorState > e : m.entrySet() )
        {
            beginRpc( e.getKey(), e.getValue(), h );
        }
    }
    
    private
    void
    testProcessesReady()
        throws Exception
    {
        for ( AbstractProcess< ? > p : procs.keySet() ) 
        {
            spawn( p );
            activeProcs.add( p );
        }

        Map< AbstractProcess< ? >, EtlProcessors.SetProcessorState > m =
            getInitialStateRequests();

        if ( m.isEmpty() ) startFeed(); else sendInitialStateReqs( m );
    }

    protected
    final
    void
    startImpl()
        throws Exception
    {
        reactor.startTestProcesses(
            getActivityContext(),
            new AbstractTask() {
                protected void runImpl() throws Exception { 
                    testProcessesReady(); 
                }
            }
        );
    }

    public
    static
    interface EventHandler
    {
        public
        void
        recordsSent( EtlRecordSet rs,
                     EtlTestProcessorFeed f )
            throws Exception;
        
        public
        void
        completeTest( Map< AbstractProcess< ? >, MingleIdentifiedName > procs,
                      Runnable onComp,
                      EtlTestProcessorFeed f )
            throws Exception;
    }

    public
    static
    abstract
    class AbstractEventHandler
    implements EventHandler
    {
        public
        void
        recordsSent( EtlRecordSet rs,
                     EtlTestProcessorFeed f )
            throws Exception
        {}
        
        public
        void
        completeTest( Map< AbstractProcess< ? >, MingleIdentifiedName > procs,
                      Runnable onComp,
                      EtlTestProcessorFeed f )
            throws Exception
        {
            onComp.run();
        }
    }

    public
    final
    static
    class Builder
    extends AbstractProcess.Builder< Builder >
    {
        private final Map< AbstractProcess< ? >, MingleIdentifiedName > procs =
            Lang.newMap();

        private final Map< AbstractProcess< ? >, Object > initStates =
            Lang.newMap();

        private EtlTestRecordGenerator< ? > g;
        private long feedLength;
        private int batchSize = 1000;
        private EtlTestReactor reactor;
        private boolean sendShutdownOnComplete = true;
        private boolean expectRecordSetAbort;
        private EventHandler eh = DEFAULT_EVENT_HANDLER;

        private
        Builder
        doAddProcessor( AbstractProcess< ? > proc,
                        MingleIdentifiedName name,
                        Object initState )
        {
            Lang.putUnique(
                procs,
                inputs.notNull( proc, "proc" ),
                inputs.notNull( name, "name" )
            );

            if ( initState != null ) initStates.put( proc, initState );

            return this;
        }

        public
        Builder
        addProcessor( AbstractProcess< ? > proc,
                      MingleIdentifiedName name )
        {
            return doAddProcessor( proc, name, null );
        }

        public
        Builder
        addProcessor( AbstractProcess< ? > proc,
                      MingleIdentifiedName name,
                      Object initState )
        {
            inputs.notNull( initState, "initState" );
            return doAddProcessor( proc, name, initState );
        }

        public
        Builder
        addProcessor( AbstractProcess< ? > proc )
        {
            return addProcessor( proc, NM_GEN.next() );
        }

        public
        Builder
        addProcessor( AbstractProcess< ? > proc,
                      Object initState )
        {
            return addProcessor( proc, NM_GEN.next(), initState );
        }

        public
        Builder
        setGenerator( EtlTestRecordGenerator< ? > g )
        {
            this.g = inputs.notNull( g, "g" );
            return this;
        }

        public
        Builder
        setFeedLength( long feedLength )
        {
            this.feedLength = inputs.nonnegativeL( feedLength, "feedLength" );
            return this;
        }

        public
        Builder
        setBatchSize( int batchSize )
        {
            this.batchSize = inputs.positiveI( batchSize, "batchSize" );
            return this;
        }

        public
        Builder
        setReactor( EtlTestReactor reactor )
        {
            this.reactor = inputs.notNull( reactor, "reactor" );
            return this;
        }

        public
        Builder
        setSendShutdownOnComplete( boolean sendShutdownOnComplete )
        {
            this.sendShutdownOnComplete = sendShutdownOnComplete;
            return this;
        }

        public
        Builder
        setExpectRecordSetAbort( boolean expectRecordSetAbort )
        {
            this.expectRecordSetAbort = expectRecordSetAbort;
            return this;
        }

        public
        Builder
        setEventHandler( EventHandler eh )
        {
            this.eh = inputs.notNull( eh, "eh" );
            return this;
        }

        public 
        EtlTestProcessorFeed 
        build() 
        { 
            return new EtlTestProcessorFeed( this ); 
        }
    }
}
