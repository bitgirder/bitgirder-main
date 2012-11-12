package com.bitgirder.mingle.service;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.ProcessRpcClient;
import com.bitgirder.process.ProcessRpcServer;
import com.bitgirder.process.ProcessExit;
import com.bitgirder.process.ProcessBehavior;
import com.bitgirder.process.Stoppable;

import com.bitgirder.process.management.ProcessManager;

import com.bitgirder.event.EventManager;
import com.bitgirder.event.EventBehavior;
import com.bitgirder.event.EventTopic;

import com.bitgirder.concurrent.Duration;

import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleNamespace;
import com.bitgirder.mingle.model.MingleException;
import com.bitgirder.mingle.model.MingleServiceRequest;
import com.bitgirder.mingle.model.MingleServiceResponse;

import java.util.Map;
import java.util.List;
import java.util.Queue;
import java.util.ArrayDeque;
import java.util.Set;

public
final
class MingleServiceEndpoint
extends AbstractVoidProcess
implements Stoppable
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static Object REFUSE_ROUTE = new Object();

    private final static ManagedRouteOptions DEFAULT_MANAGED_ROUTE_OPTIONS =
        ManagedRouteOptions.withBackLog( 100 );

    private final Map< MingleNamespace, Map< MingleIdentifier, Object > >
        routes;

    private final Map< Object, Map.Entry< ?, Object > > managedRoutes;
    private final Map< Object, ManagedRouteOptions > managedRouteOpts;

    private final EventTopic< ? > lifecycleTopic;
    private final AbstractProcess< ? > manager;
 
    private final Duration defaultRequestTimeout;

    private
    static
    Map< MingleNamespace, Map< MingleIdentifier, Object > >
    buildRouteMap( Map< MingleNamespace, Map< MingleIdentifier, Object > > m )
    {
        Map< MingleNamespace, Map< MingleIdentifier, Object > > res = 
            Lang.newMap();
        
        for ( Map.Entry< MingleNamespace, Map< MingleIdentifier, Object > > e :
                m.entrySet() )
        {
            res.put( e.getKey(), Lang.copyOf( e.getValue() ) );
        }

        return res;
    }

    // returns an unmodifiable map, although the entries it contains as values
    // are themselves mutable
    private
    static
    Map< Object, Map.Entry< ?, Object > >
    buildManagedRoutes( 
        Map< MingleNamespace, Map< MingleIdentifier, Object > > routes )
    {
        Map< Object, Map.Entry< ?, Object > > res = Lang.newMap();
        
        for ( Map.Entry< MingleNamespace, Map< MingleIdentifier, Object > > e : 
                routes.entrySet() )
        {
            for ( Map.Entry< MingleIdentifier, Object > e2 : 
                    e.getValue().entrySet() )
            {
                Object val = e2.getValue();
                
                if ( val instanceof ManagedTargetContext )
                {
                    ManagedTargetContext mtc = (ManagedTargetContext) val;
                    Lang.putUnique( res, mtc.id, e2 );
                }
            }
        }

        return Lang.unmodifiableMap( res );
    }

    private
    MingleServiceEndpoint( Builder b,
                           AbstractVoidProcess.Builder procBldr )
    {
        super( procBldr );

        this.routes = buildRouteMap( b.routes );

        this.managedRoutes = buildManagedRoutes( this.routes );
        this.managedRouteOpts = Lang.unmodifiableCopy( b.managedRouteOpts );
        this.lifecycleTopic = b.lifecycleTopic;
        this.manager = b.manager;

        this.defaultRequestTimeout = b.defaultRequestTimeout;
    }

    @Override
    protected
    void
    childExited( AbstractProcess< ? > child,
                 ProcessExit< ? > exit )
    {
        warnIfFailed( child, exit );
        if ( ! hasChildren() ) exit();
    }

    public void stop() { behavior( ProcessRpcServer.class ).stop(); }

    private
    final
    class LifecycleEventReceiver
    extends ProcessManager.AbstractManagerEventReceiver
    {
        private
        LifecycleEventReceiver()
        {
            super( 
                manager.behavior( ProcessManager.class ),
                MingleServiceEndpoint.this.getActivityContext()
            );
        }

        @Override
        protected
        void
        processStarted( ProcessManager.ProcessStarted ps )
        {
            Map.Entry< ?, Object > route = managedRoutes.get( ps.getId() );
            if ( route != null ) setActiveTarget( ps.getProcess(), route );
        }

        @Override
        protected
        void
        processExited( ProcessManager.ProcessExited ps )
        {
            Object id = ps.getId();

            Map.Entry< ?, Object > route = managedRoutes.get( id );

            if ( route != null ) 
            {
                Object val = route.getValue();

                // the else branch covers the only other case: that we received
                // a process exit without knowing about its start, either
                // because it failed to start or because it started before we
                // began receiving events
                if ( val instanceof AbstractProcess )
                {
                    ManagedRouteOptions opts = managedRouteOpts.get( id );
                    route.setValue( new ManagedTargetContext( id, opts ) );
                }
                else state.isTrue( val instanceof ManagedTargetContext );
            }
        }

        @Override
        protected
        void
        restartDeclined( ProcessManager.RestartDeclined rd )
        {
            Object id = rd.getId();

            Map.Entry< ?, Object > route = managedRoutes.get( id );

            if ( route != null )
            {
                Object prev = route.getValue();
                route.setValue( REFUSE_ROUTE );

                if ( prev instanceof ManagedTargetContext )
                {
                    failAll( (ManagedTargetContext) prev );
                }
            }
        }

        @Override
        protected
        void
        managerStopped( ProcessManager.ManagerStopped ms )
        {
            for ( Map.Entry< ?, Object > e : managedRoutes.values() )
            {
                Object prev = e.getValue();
                e.setValue( REFUSE_ROUTE );

                if ( prev instanceof ManagedTargetContext )
                {
                    failAll( (ManagedTargetContext) prev );
                }
            }
        }
    }

    private
    void
    failRequest( 
        MingleServiceCallContext cc,
        ProcessRpcServer.ResponderContext< MingleServiceResponse > ctx )
    { 
        ctx.respond( 
            MingleServiceResponse.createFailure(
                MingleServices.getInternalServiceException() ) );
    }

    private
    void
    resumeHeldRequests( ManagedTargetContext mtc,
                        AbstractProcess< ? > dest )
    {
        for ( ManagedTargetContext.HeldRequest hr : mtc.heldReqs )
        {
            beginRequest( dest, hr.cc, hr.ctx );
        }
    }

    private
    void
    failAll( ManagedTargetContext mtc )
    {
        for ( ManagedTargetContext.HeldRequest hr : mtc.heldReqs )
        {
            failRequest( hr.cc, hr.ctx );
        }
    }

    private
    void
    setActiveTarget( AbstractProcess< ? > dest,
                     Map.Entry< ?, Object > e )
    {
        Object val = e.getValue();

        if ( val instanceof ManagedTargetContext )
        {
            ManagedTargetContext mtc = (ManagedTargetContext) val;
            
            e.setValue( dest );
            resumeHeldRequests( mtc, dest );
        }
    }

    private
    void
    getActiveManagedTargets()
    {
        beginRpc( 
            manager, 
            new ProcessManager.GetActiveTargets( managedRoutes.keySet() ),
            new DefaultRpcHandler() {
                @Override protected void rpcSucceeded( Object resp )
                {
                    @SuppressWarnings( "unchecked" )
                    Map< Object, AbstractProcess< ? > > m =
                        (Map< Object, AbstractProcess< ? > >) resp;

                    for ( Map.Entry< Object, AbstractProcess< ? > > e : 
                            m.entrySet() )
                    {
                        Map.Entry< ?, Object > e2 =
                            state.get( 
                                managedRoutes, e.getKey(), "managedRoutes" );

                        setActiveTarget( e.getValue(), e2 );
                    }
                }
            }
        );
    }

    protected 
    void 
    startImpl() 
    { 
        if ( lifecycleTopic != null ) 
        {
            behavior( EventBehavior.class ).
                subscribe( lifecycleTopic, new LifecycleEventReceiver() );
            
            getActiveManagedTargets();
        }
    }

    private
    Object
    findRouteDestination( MingleServiceRequest req )
        throws NoSuchNamespaceException,
               NoSuchServiceException,
               InternalServiceException
    {
        MingleNamespace ns = req.getNamespace();
        MingleIdentifier svc = req.getService();

        Map< MingleIdentifier, Object > procs = routes.get( ns );

        if ( procs == null ) throw new NoSuchNamespaceException( ns );
        else
        {
            Object res = procs.get( svc );

            if ( res == null ) throw new NoSuchServiceException( svc );
            else return res;
        }
    }

    private
    Object
    getRouteDestination( 
        MingleServiceRequest req,
        ProcessRpcServer.ResponderContext< MingleServiceResponse > ctx )
            throws Exception
    {
        try { return findRouteDestination( req ); }
        catch ( Exception ex )
        {
            MingleException me = MingleServices.asServiceException( ex );

            if ( me == null ) throw ex;
            else
            {
                ctx.respond( MingleServiceResponse.createFailure( me ) );
                return null;
            }
        }
    }

    private
    final
    static
    class RpcHandler
    extends ProcessRpcClient.AbstractResponseHandler
    {
        private final 
            ProcessRpcServer.ResponderContext< MingleServiceResponse > ctx;
        
        private
        RpcHandler(
            ProcessRpcServer.ResponderContext< MingleServiceResponse > ctx )
        {
            this.ctx = ctx;
        }

        @Override protected void rpcFailed( Throwable th ) { ctx.fail( th ); }

        @Override
        protected 
        void 
        rpcSucceeded( Object resp ) 
        {
            ctx.respond( (MingleServiceResponse) resp );
        }
    }

    private
    ProcessRpcClient.Call
    createRpcCall( 
        AbstractProcess< ? > dest,
        MingleServiceCallContext cc,
        ProcessRpcServer.ResponderContext< MingleServiceResponse > ctx )
    { 
        ProcessRpcClient.Call.Builder b =
            new ProcessRpcClient.Call.Builder().
                setDestination( dest ).
                setRequest( cc ).
                setResponseHandler( new RpcHandler( ctx ) );
        
        if ( defaultRequestTimeout != null ) 
        {
            b.setTimeout( defaultRequestTimeout );
        }

        return b.build();
    }

    private
    void
    beginRequest( 
        AbstractProcess< ? > dest,
        MingleServiceCallContext cc,
        ProcessRpcServer.ResponderContext< MingleServiceResponse > ctx )
    { 
        behavior( ProcessRpcClient.class ).
            beginRpc( createRpcCall( dest, cc, ctx ) );
    }

    private
    void
    processRequest(
        ManagedTargetContext mtc,
        MingleServiceCallContext cc,
        ProcessRpcServer.ResponderContext< MingleServiceResponse > ctx )
            throws InternalServiceException
    {
        if ( ! mtc.hold( cc, ctx ) ) throw new InternalServiceException();
    }

    @ProcessRpcServer.Responder
    private
    void
    route( MingleServiceCallContext cc,
           ProcessRpcServer.ResponderContext< MingleServiceResponse > ctx )
        throws Exception
    {
        Object dest = getRouteDestination( cc.getRequest(), ctx );

        if ( dest != null )
        {
            if ( dest instanceof AbstractProcess ) 
            {
                beginRequest( (AbstractProcess< ? >) dest, cc, ctx );
            }
            else if ( dest instanceof ManagedTargetContext )
            {
                processRequest( (ManagedTargetContext) dest, cc, ctx );
            }
            else if ( dest == REFUSE_ROUTE ) failRequest( cc, ctx );
            else state.fail( "Unexpected route dest:", dest );
        }
    }

    @ProcessRpcServer.Responder
    private
    void
    route( MingleServiceRequest req,
           ProcessRpcServer.ResponderContext< MingleServiceResponse > ctx )
        throws Exception
    {
        route( MingleServiceCallContext.create( req ), ctx );
    }

    public
    final
    static
    class ManagedRouteOptions
    {
        private final int backLog;

        private
        ManagedRouteOptions( int backLog )
        {
            this.backLog = backLog;
        }

        public
        static
        ManagedRouteOptions
        withBackLog( int backLog )
        {
            inputs.positiveI( backLog, "backLog" );

            return new ManagedRouteOptions( backLog );
        }
    }

    private
    final
    static
    class ManagedTargetContext
    {
        private final Object id;

        private final int backLog;

        private final Queue< HeldRequest > heldReqs;

        private 
        ManagedTargetContext( Object id,
                              ManagedRouteOptions opts ) 
        { 
            this.id = id; 

            if ( opts == null ) opts = DEFAULT_MANAGED_ROUTE_OPTIONS;
            this.backLog = opts.backLog;
            
            heldReqs = new ArrayDeque< HeldRequest >( backLog );
        }

        private
        final
        static
        class HeldRequest
        {
            private final MingleServiceCallContext cc;

            private final 
                ProcessRpcServer.ResponderContext< MingleServiceResponse > ctx;
            
            private
            HeldRequest(
                MingleServiceCallContext cc,
                ProcessRpcServer.ResponderContext< MingleServiceResponse > ctx )
            {
                this.cc = cc;
                this.ctx = ctx;
            }
        }

        private
        boolean
        hold( MingleServiceCallContext cc,
              ProcessRpcServer.ResponderContext< MingleServiceResponse > ctx )
        {
            if ( heldReqs.size() < backLog )
            {
                heldReqs.add( new HeldRequest( cc, ctx ) );
                return true;
            }
            else return false;
        }
    }

    public
    final
    static
    class Builder
    {
        private final 
            Map< MingleNamespace, Map< MingleIdentifier, Object > > routes =
                Lang.newMap();
    
        private Duration defaultRequestTimeout;
        private EventManager evMgr;
        private EventTopic< ? > lifecycleTopic;
        private AbstractProcess< ? > manager;

        private final Set< Object > managedRouteIds = Lang.newSet();

        private final Map< Object, ManagedRouteOptions > managedRouteOpts =
            Lang.newMap();

        private
        Builder
        addRouteObject( MingleNamespace ns,
                        MingleIdentifier svc,
                        Object obj )
        {
            Map< MingleIdentifier, Object > procs = routes.get( ns );

            if ( procs == null )
            {
                procs = Lang.newMap();
                routes.put( ns, procs );
            }

            Lang.putUnique( procs, svc, obj );

            return this;
        }

        public
        Builder
        addRoute( MingleNamespace ns,
                  MingleIdentifier svc,
                  AbstractProcess< ? > proc )
        {
            inputs.notNull( ns, "ns" );
            inputs.notNull( svc, "svc" );
            inputs.notNull( proc, "proc" );

            return addRouteObject( ns, svc, proc );
        }
 
        public
        Builder
        addRoute( CharSequence nsStr,
                  CharSequence svcStr,
                  AbstractProcess< ? > proc )
        {
            MingleNamespace ns =
                MingleNamespace.create( inputs.notNull( nsStr, "nsStr" ) );

            MingleIdentifier svc =
                MingleIdentifier.create( inputs.notNull( svcStr, "svcStr" ) );

            return addRoute( ns, svc, proc );
        }

        // null-checks params which may have come from a public method
        private
        Builder
        doAddManagedRoute( MingleNamespace ns,
                           MingleIdentifier svc,
                           Object id,
                           ManagedRouteOptions opts )
        {
            inputs.notNull( ns, "ns" );
            inputs.notNull( svc, "svc" );
            inputs.notNull( id, "id" );

            state.isTrue( 
                managedRouteIds.add( id ),
                "A managed route has already been added with id:", id );

            if ( opts != null ) managedRouteOpts.put( id, opts );

            ManagedTargetContext mtc = new ManagedTargetContext( id, opts );

            return addRouteObject( ns, svc, mtc );
        }

        public
        Builder
        addManagedRoute( MingleNamespace ns,
                         MingleIdentifier svc,
                         Object id )
        {
            return doAddManagedRoute( ns, svc, id, null );
        }

        public
        Builder
        addManagedRoute( MingleNamespace ns,
                         MingleIdentifier svc,
                         Object id,
                         ManagedRouteOptions opts )
        {
            return 
                doAddManagedRoute( 
                    ns, svc, id, inputs.notNull( opts, "opts" ) );
        }

        public
        Builder
        addManagedRoute( CharSequence ns,
                         CharSequence svc,
                         Object id )
        {
            return
                addManagedRoute(
                    MingleNamespace.create( inputs.notNull( ns, "ns" ) ),
                    MingleIdentifier.create( inputs.notNull( svc, "svc" ) ),
                    id
                );
        }

        public
        Builder
        addManagedRoute( CharSequence ns,
                         CharSequence svc,
                         Object id,
                         ManagedRouteOptions opts )
        {
            return
                addManagedRoute(
                    MingleNamespace.create( inputs.notNull( ns, "ns" ) ),
                    MingleIdentifier.create( inputs.notNull( svc, "svc" ) ),
                    id,
                    opts
                );
        }

        public
        Builder
        setDefaultRequestTimeout( Duration defaultRequestTimeout )
        {
            this.defaultRequestTimeout = 
                inputs.notNull( 
                    defaultRequestTimeout, "defaultRequestTimeout" );

            return this;
        }

        public
        Builder
        setManagerSubscription( EventManager evMgr,
                                EventTopic< ? > lifecycleTopic,
                                AbstractProcess< ? > manager )
        {
            state.isTrue( 
                this.evMgr == null && 
                this.lifecycleTopic == null &&
                this.manager == null,
                "subscription info is already set" );
 
            this.evMgr = inputs.notNull( evMgr, "evMgr" );

            this.lifecycleTopic = 
                inputs.notNull( lifecycleTopic, "lifecycleTopic" );
 
            this.manager = inputs.notNull( manager, "manager" );

            inputs.isTrue( 
                manager.hasBehavior( ProcessManager.class ),
                "Manager process does not include ProcessManager behavior" );

            return this;
        }

        public
        MingleServiceEndpoint
        build()
        {
            AbstractVoidProcess.Builder b = 
                new AbstractVoidProcess.Builder() {};

            b.mixin( ProcessRpcClient.create() );
            b.mixin( ProcessRpcServer.createStandard() );

            if ( ! managedRouteIds.isEmpty() )
            {
                state.isFalse( 
                    evMgr == null || lifecycleTopic == null,
                    "Endpoint has a managed route but no manager " +
                    "subscription" );

                b.mixin( EventBehavior.create( evMgr ) );
            }

            return new MingleServiceEndpoint( this, b );
        }
    }
}
