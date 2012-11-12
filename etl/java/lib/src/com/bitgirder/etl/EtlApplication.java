package com.bitgirder.etl;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.AbstractPulse;
import com.bitgirder.process.ProcessRpcClient;
import com.bitgirder.process.ProcessRpcServer;
import com.bitgirder.process.ProcessExit;

import com.bitgirder.process.management.ProcessFactory;
import com.bitgirder.process.management.ProcessManager;
import com.bitgirder.process.management.ProcessControl;

import com.bitgirder.concurrent.Duration;

import com.bitgirder.event.EventBehavior;
import com.bitgirder.event.EventTopic;
import com.bitgirder.event.EventManager;

import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleIdentifiedName;

import java.util.Map;
import java.util.Set;

public
final
class EtlApplication
extends AbstractVoidProcess
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Map< MingleIdentifier, EtlProcessorGroup > groups;
    private final Map< MingleIdentifiedName, ProcessorContext > pCtxByName;
    private final Map< MingleIdentifier, ProcessorGroupContext > procGrpsById;

    private final Map< FeedGroupKey, Set< MingleIdentifiedName > > idleFeeds =
        Lang.newMap();

    private final Duration feedPulse;

    private
    Map< MingleIdentifier, ProcessorGroupContext >
    buildProcessorGroupsById()
    {
        Map< MingleIdentifier, ProcessorGroupContext > res = Lang.newMap();

        for ( Map.Entry< MingleIdentifier, EtlProcessorGroup > e :
                groups.entrySet() )
        {
            res.put( e.getKey(), new ProcessorGroupContext( e.getValue() ) );
        }

        return res;
    }

    private
    Map< MingleIdentifiedName, ProcessorContext >
    buildProcessorContextsByName()
    {
        Map< MingleIdentifiedName, ProcessorContext > res = Lang.newMap();

        for ( Map.Entry< MingleIdentifier, EtlProcessorGroup > e :
                groups.entrySet() )
        {
            ProcessorGroupContext grpCtx = 
                state.get( procGrpsById, e.getKey(), "procGrpsById" );

            for ( Map.Entry< MingleIdentifiedName, ProcessFactory< ? > > e2 :
                    e.getValue().getProcessors().entrySet() )
            {
                ProcessorContext pCtx =
                    new ProcessorContext( 
                        e.getKey(), e.getValue(), e2.getValue(), grpCtx );
                
                res.put( e2.getKey(), pCtx );
            }
        }

        return res;
    }

    private
    static
    ProcessManager
    createProcessManager( Builder b )
    {
        inputs.notNull( b.appName, "appName" );

        return ProcessManager.create( EventTopic.create( b.appName ) );
    }

    private
    EtlApplication( Builder b )
    {
        super( 
            createProcessManager( b ),
            EventBehavior.create( inputs.notNull( b.evMgr, "evMgr" ) ),
            ProcessRpcClient.create(),
            new ProcessRpcServer.Builder().build()
        );

        this.groups = Lang.unmodifiableCopy( b.groups );
        inputs.isFalse( groups.isEmpty(), "No groups supplied" );

        // order is important
        procGrpsById = buildProcessorGroupsById();
        pCtxByName = buildProcessorContextsByName();

        this.feedPulse = b.feedPulse;
    }

    private
    final
    static
    class ProcessorGroupContext
    {
        private final EtlProcessorGroup grp;

        private AbstractProcess< ? > stateMgr;
        private AbstractProcess< ? > lister;

        private
        ProcessorGroupContext( EtlProcessorGroup grp )
        {
            this.grp = grp;
        }
    }

    private
    static
    class ProcessorContext
    {
        private final MingleIdentifier grpId;
        private final EtlProcessorGroup grp;
        private final ProcessFactory< ? > procFact;
        private final ProcessorGroupContext grpCtx;

        private AbstractProcess< ? > activeProc;

        private
        ProcessorContext( MingleIdentifier grpId,
                          EtlProcessorGroup grp,
                          ProcessFactory< ? > procFact,
                          ProcessorGroupContext grpCtx )
        {
            this.grpId = grpId;
            this.grp = grp;
            this.procFact = procFact;
            this.grpCtx = grpCtx;
        }
    }

    @Override
    protected
    void
    childExited( AbstractProcess< ? > proc,
                 ProcessExit< ? > exit )
    {
        if ( ! exit.isOk() ) fail( exit.getThrowable() );
    }

    private
    ProcessorContext
    processorContext( MingleIdentifiedName nm )
    {
        return state.get( pCtxByName, nm, "pCtxByName" );
    }

    private
    ProcessorGroupContext
    groupContext( MingleIdentifiedName nm )
    {
        return processorContext( nm ).grpCtx;
    } 

    private
    ProcessorGroupContext
    groupContext( MingleIdentifier grpId )
    {
        return state.get( procGrpsById, grpId, "procGrpsById" );
    }

    private
    AbstractProcess< ? >
    spawnAndManage( ProcessControl< ? > ctl,
                    MingleIdentifier grpId,
                    CharSequence procIdStr )
    {
        MingleIdentifier procId = MingleIdentifier.create( procIdStr );

        CharSequence id =
            grpId.getExternalForm() + "/" + procId.getExternalForm();

        AbstractProcess< ? > proc = 
            behavior( ProcessManager.class ).manageAndProxy( id, ctl );
 
        spawn( proc );
        return proc;
    }

    private
    void
    initProcessorGroups()
    {
        for ( Map.Entry< MingleIdentifier, ProcessorGroupContext > e :
                procGrpsById.entrySet() )
        {
            ProcessorGroupContext ctx = e.getValue();

            ctx.stateMgr = 
                spawnAndManage( 
                    ctx.grp.getStateManager(), e.getKey(), "state-manager" );
            
            ctx.lister =
                spawnAndManage( ctx.grp.getLister(), e.getKey(), "lister" );
        }
    }

    private
    final
    static
    class FeedGroupKey
    {
        private final MingleIdentifier grpId;
        private final Object pos;

        private
        FeedGroupKey( MingleIdentifier grpId,
                      Object pos )
        {
            this.grpId = grpId;
            this.pos = pos;
        }

        public
        int
        hashCode()
        {
            return grpId.hashCode() | ( pos == null ? 0 : pos.hashCode() );
        }

        public
        boolean
        equals( Object other )
        {
            if ( other == this ) return true;
            else if ( other instanceof FeedGroupKey )
            {
                FeedGroupKey k = (FeedGroupKey) other;

                if ( grpId.equals( k.grpId ) )
                {
                    return pos == null ? k.pos == null : pos.equals( k.pos );
                }
                else return false;
            }
            else return false;
        }
    }

    private
    final
    class FeedPulse
    extends AbstractPulse
    {
        private FeedGroupKey k;

        private
        FeedPulse( FeedGroupKey k )
        {
            super( feedPulse, self() );

            this.k = k;
        }

        private
        void
        processNextFile()
        {
//            BoundMingleServiceCall.Builder b =
//                mgCallBuilder( SVC_ID_FILESERVICE ).
//                    setParameter( ID_FILE_GROUP, k.grpId );
//            
//            if ( k.pos != null ) b.setParameter( k.pos );
//
//            b.setResponseHandler(
//                new ObjectReceiver< MingleValue >
//                setParameter( 
//            svcCli.clientFor( SVC_ID_FILESERVICE ).beginRpc(
//                NS_LOGGING,
//                SVC,
//            LogFileListApi.GetSuccessorFile req =
//                k.pos == null
//                    ? LogFileListApi.GetSuccessorFile.create( k.grpId )
//                    : LogFileListApi.GetSuccessorFile.create( k.grpId, k.pos );
//
//            beginRpc(
//                groupContext( k.grpId ).lister,
//                req,
//                new DefaultRpcHandler() {
//                    @Override protected void rpcSucceeded() {
//                        throw new UnsupportedOperationException(
//                            "Unimplemented" );
//                    }
//                }
//            );
//
//            code( "Sent file list req" );
            throw new UnsupportedOperationException( "Unimplemented" );
        }

        @Override protected void beginPulse() { processNextFile(); }
    }

    private
    void
    feedGroupsReady( Map< FeedGroupKey, Set< MingleIdentifiedName > > grps )
    {
        idleFeeds.putAll( grps );

        for ( FeedGroupKey k : grps.keySet() ) new FeedPulse( k ).start();
    }

    private
    void
    addToFeedGroup( 
        final MingleIdentifiedName nm,
        final ProcessorContext ctx,
        final Map< FeedGroupKey, Set< MingleIdentifiedName > > feedGroups,
        final Set< MingleIdentifiedName > waitSet )
    {
        beginRpc(
            groupContext( nm ).stateMgr,
            EtlProcessors.createGetProcessorFeedPosition( nm ),
            new DefaultRpcHandler() {
                @Override protected void rpcSucceeded( Object pos )
                {
                    FeedGroupKey key = new FeedGroupKey( ctx.grpId, pos );

                    Set< MingleIdentifiedName > s = feedGroups.get( key );
                    if ( s == null ) feedGroups.put( key, s = Lang.newSet() );
                    s.add( nm );

                    state.remove( waitSet, nm, "waitSet" );
                    if ( waitSet.isEmpty() ) feedGroupsReady( feedGroups );
                }
            }
        );
    }

    private
    void
    beginFeeds()
    {
        Map< FeedGroupKey, Set< MingleIdentifiedName > > feedGroups =
            Lang.newMap();

        Set< MingleIdentifiedName > waitSet = 
            Lang.newSet( pCtxByName.keySet() );

        for ( Map.Entry< MingleIdentifiedName, ProcessorContext > e :
                pCtxByName.entrySet() )
        {
            addToFeedGroup( e.getKey(), e.getValue(), feedGroups, waitSet );
        }
    }

    private
    abstract
    class AbstractInitRpcHandler
    extends DefaultRpcHandler
    {
        final MingleIdentifiedName nm;
        final Set< MingleIdentifiedName > initWaits;

        private
        AbstractInitRpcHandler( MingleIdentifiedName nm,
                                Set< MingleIdentifiedName > initWaits )
        {
            this.nm = nm;
            this.initWaits = initWaits;
        }

        final
        void
        initDone()
        {
            state.remove( initWaits, nm, "initWaits" );

            if ( initWaits.isEmpty() ) beginFeeds();
        }
    }
 
    private
    final
    class SetInitStateHandler
    extends AbstractInitRpcHandler
    {
        private
        SetInitStateHandler( MingleIdentifiedName nm,
                             Set< MingleIdentifiedName > initWaits )
        {
            super( nm, initWaits );
        }

        @Override protected void rpcSucceeded() { initDone(); }
    }

    private
    final
    class GetInitStateHandler
    extends AbstractInitRpcHandler
    {
        private
        GetInitStateHandler( MingleIdentifiedName nm,
                             Set< MingleIdentifiedName > initWaits )
        {
            super( nm, initWaits );
        }

        @Override
        protected
        void
        rpcSucceeded( Object resp )
        {
            if ( resp == null ) initDone();
            else 
            {
                beginRpc(
                    state.get( pCtxByName, nm, "pCtxByName" ).activeProc,
                    EtlProcessors.createSetProcessorState( nm, resp ),
                    new AbstractInitRpcHandler( nm, initWaits ) {
                        @Override protected void rpcSucceeded() { initDone(); }
                    }
                );
            }
        }
    }

    private
    void
    beginInit( MingleIdentifiedName nm,
               ProcessorGroupContext grpCtx,
               Set< MingleIdentifiedName > initWaits )
    {
        beginRpc(
            groupContext( nm ).stateMgr,
            EtlProcessors.createGetProcessorState( nm ),
            new GetInitStateHandler( nm, initWaits ) 
        );
    }

    private
    void
    initRecordProcessors()
        throws Exception
    {
        Set< MingleIdentifiedName > initWaits = Lang.newSet();

        for ( Map.Entry< MingleIdentifiedName, ProcessorContext > e :
                pCtxByName.entrySet() )
        {
            ProcessorContext pCtx = e.getValue();
            spawn( pCtx.activeProc = pCtx.procFact.newProcess() );

            initWaits.add( e.getKey() );
            beginInit( e.getKey(), e.getValue().grpCtx, initWaits );
        }
    }

    protected
    void
    startImpl()
        throws Exception
    {
        initProcessorGroups();
        initRecordProcessors();
    }

    public
    final
    static
    class Builder
    {
        private MingleIdentifier appName;
        private EventManager evMgr;
        private Duration feedPulse = Duration.fromMinutes( 5 );

        private Map< MingleIdentifier, EtlProcessorGroup > groups =
            Lang.newMap();

        public
        Builder
        setApplicationName( MingleIdentifier appName )
        {
            this.appName = inputs.notNull( appName, "appName" );
            return this;
        }

        public
        Builder
        setEventManager( EventManager evMgr )
        {
            this.evMgr = inputs.notNull( evMgr, "evMgr" );
            return this;
        }

        public
        Builder
        setFeedPulse( Duration feedPulse )
        {
            this.feedPulse = inputs.notNull( feedPulse, "feedPulse" );
            return this;
        }

        private
        MingleIdentifier
        makeName( CharSequence name )
        {
            return MingleIdentifier.create( inputs.notNull( name, "name" ) );
        }

        public
        Builder
        addGroup( MingleIdentifier name,
                  EtlProcessorGroup grp )
        {
            Lang.putUnique(
                groups,
                inputs.notNull( name, "name" ),
                inputs.notNull( grp, "grp" )
            );

            return this;
        }

        public
        Builder
        addGroup( CharSequence name,
                  EtlProcessorGroup grp )
        {
            return addGroup( makeName( name ), grp );
        }

        private
        void
        checkIdIntegrity()
        {
            Map< MingleIdentifiedName, MingleIdentifier > seen = Lang.newMap();

            for ( Map.Entry< MingleIdentifier, EtlProcessorGroup > e :
                    groups.entrySet() )
            {
                for ( MingleIdentifiedName nm : 
                        e.getValue().getProcessors().keySet() )
                {
                    MingleIdentifier cur = e.getKey();
                    MingleIdentifier prev = seen.get( nm );

                    if ( prev == null ) seen.put( nm, cur );
                    else 
                    {
                        inputs.fail( 
                            "Processor groups", cur, "and", prev, 
                            "both contain a processor with id", nm );
                    }
                }
            }
        }

        public 
        EtlApplication 
        build() 
        { 
            checkIdIntegrity();
            return new EtlApplication( this ); 
        }
    }
}
