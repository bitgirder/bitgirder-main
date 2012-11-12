package com.bitgirder.etl;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.ObjectReceiver;

import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.ProcessExit;
import com.bitgirder.process.ProcessRpcClient;

import com.bitgirder.mingle.model.MingleIdentifiedName;

import java.util.Map;

public
abstract
class AbstractEtlProcessorTest< P extends AbstractProcess< ? >, 
                                R extends EtlTestReactor >
extends AbstractVoidProcess
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private R reactor;
    private P testProc;

    protected AbstractEtlProcessorTest() { super( ProcessRpcClient.create() ); }

    // overridable
    protected int getFeedLength() { return 1000; }
    protected int getBatchSize() { return 100; }
    protected boolean sendShutdownOnComplete() { return true; }
    protected boolean expectRecordSetAbort() { return false; }

    protected final R reactor() { return reactor; }

    protected
    abstract
    void
    beginAssert( P testProc,
                 Object lastState,
                 Runnable onComp );

    @Override
    protected
    final
    void
    childExited( AbstractProcess< ? > proc,
                 ProcessExit< ? > exit )
    {
        if ( ! exit.isOk() ) fail( exit.getThrowable() );
        state.isFalse( hasChildren() );
        exit();
    }

    // Will be called after it is legal to call reactor()
    protected
    abstract
    P
    createTestProcessor()
        throws Exception;
    
    protected
    abstract
    EtlTestRecordGenerator
    createGenerator()
        throws Exception;

    protected
    abstract
    R
    createReactor()
        throws Exception;
    
    protected
    class EventHandlerImpl
    extends EtlTestProcessorFeed.AbstractEventHandler
    {
        @Override
        public
        final
        void
        completeTest( Map< AbstractProcess< ? >, MingleIdentifiedName > procs,
                      final Runnable onComplete,
                      EtlTestProcessorFeed f )
        {
            state.equalInt( 1, procs.size() );

            // there's only one proc in there and it was the result of
            // createTestProcessor()
            @SuppressWarnings( "unchecked" )
            final P p = (P) procs.keySet().iterator().next();

            MingleIdentifiedName nm = procs.get( p );

            f.getProcessorState( nm,
                new ObjectReceiver< Object >() {
                    public void receive( Object obj ) {
                        beginAssert( p, obj, onComplete );
                    }
                }
            );
        }
    }

    protected
    EventHandlerImpl
    createEventHandler()
    {
        return new EventHandlerImpl();
    }

    protected Object getInitialState() throws Exception { return null; }

    private
    EtlTestProcessorFeed
    buildTestProcessorFeed()
        throws Exception
    {
        EtlTestProcessorFeed.Builder b =
            new EtlTestProcessorFeed.Builder().
                setGenerator( createGenerator() ).
                setReactor( reactor ).
                setFeedLength( getFeedLength() ).
                setBatchSize( getBatchSize() ).
                setExpectRecordSetAbort( expectRecordSetAbort() ).
                setSendShutdownOnComplete( sendShutdownOnComplete() ).
                setEventHandler( createEventHandler() );
        
        AbstractProcess< ? > testProc = createTestProcessor();
        Object init = getInitialState();

        if ( init == null ) b.addProcessor( testProc );
        else b.addProcessor( testProc, init );

        return b.build();
    }

    protected
    final
    void
    startImpl()
        throws Exception
    {
        reactor = createReactor(); // see note at createTestProcessor()

        spawn( buildTestProcessorFeed() );
    }
}
