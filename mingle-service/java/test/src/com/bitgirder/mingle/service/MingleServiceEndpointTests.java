package com.bitgirder.mingle.service;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Range;

import com.bitgirder.process.ExecutorProcess;
import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.ProcessExit;
import com.bitgirder.process.Processes;
import com.bitgirder.process.ProcessActivity;
import com.bitgirder.process.ProcessRpcClient;
import com.bitgirder.process.ProcessRpcServer;

import static com.bitgirder.process.ProcessRpcServer.ResponderContext;

import com.bitgirder.process.management.ProcessManager;
import com.bitgirder.process.management.ProcessManagement;
import com.bitgirder.process.management.ProcessManagementTests;
import com.bitgirder.process.management.AbstractProcessControl;

import com.bitgirder.concurrent.Duration;

import com.bitgirder.event.EventManager;
import com.bitgirder.event.EventBehavior;
import com.bitgirder.event.EventTopic;

import com.bitgirder.mingle.model.MingleNamespace;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleString;
import com.bitgirder.mingle.model.MingleException;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleServiceRequest;
import com.bitgirder.mingle.model.MingleServiceResponse;
import com.bitgirder.mingle.model.MingleSymbolMapAccessor;
import com.bitgirder.mingle.model.MingleTypeReference;

import com.bitgirder.test.Test;

@Test
final
class MingleServiceEndpointTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private
    abstract
    class AbstractServiceEndpointTest
    extends AbstractVoidProcess
    {
        private MingleServiceEndpoint ep;
        private TestService svc;
        private MingleRpcClient mgCli;

        private
        AbstractServiceEndpointTest()
        {
            super( ProcessRpcClient.create() );
        }

        final TestService testService() { return svc; }

        final
        MingleServiceRequest.Builder
        createRequestBuilder()
        {
            return 
                new MingleServiceRequest.Builder().
                    setNamespace( "test:ns@v1" ).
                    setService( "test-svc" );
        } 

        final
        void
        beginRpc( MingleServiceRequest req,
                  MingleRpcClient.Handler h )
        {
            state.notNull( mgCli ).beginRpc( req, h );
        }

        abstract
        class RpcHandlerImpl
        extends AbstractMingleRpcClientHandler
        {
            void rpcSucceeded( MingleValue mv ) {}

            void
            rpcFailed( MingleException me )
            {
                state.fail( MingleModels.inspect( me ) );
            }

            @Override
            protected
            void
            rpcSucceeded( MingleServiceResponse resp )
            {
                if ( resp.isOk() ) rpcSucceeded( resp.getResult() );
                else rpcFailed( resp.getException() );
            }

            @Override protected void rpcFailed( Throwable th ) { fail( th ); }
        }

        void
        childExitedImpl( AbstractProcess< ? > child,
                         ProcessExit< ? > exit )
            throws Exception
        {}

        @Override
        protected
        final
        void
        childExited( AbstractProcess< ? > child,
                     ProcessExit< ? > exit )
            throws Exception
        {
            childExitedImpl( child, exit );

            if ( ! exit.isOk() ) fail( exit.getThrowable() );
            if ( ! hasChildren() ) exit();
        }

        final
        void
        testDone()
        {
            ProcessRpcClient cli = behavior( ProcessRpcClient.class );

            Processes.sendStop( ep, cli );
            Processes.sendStop( svc, cli );
        }

        abstract
        void
        startTest();

        final
        class TestService
        extends AbstractMingleService
        {
            @MingleServices.Operation
            private
            void
            longRunner( 
                final MingleServiceCallContext callCtx,
                final ResponderContext< MingleServiceResponse > ctx )
            {
                submit(
                    new AbstractTask() {
                        protected void runImpl() 
                        { 
                            ctx.respond(
                                createSuccessResponse( callCtx,
                                    callCtx.getParameters().
                                        expectMingleValue( "echo-value" )
                                )
                            );
                        }
                    },
                    Duration.fromMillis( 
                        callCtx.getParameters().expectLong( "delay-millis" ) )
                );
            }

            void
            exitAbruptly()
            {
                submit( 
                    new AbstractTask() { 
                        protected void runImpl() { exit(); } 
                    }
                );
            }
        }

        private
        AbstractProcess< ? >
        spawnTestService()
        {
            spawn( svc = new TestService() );

            return svc;
        }

        void completeBuild( MingleServiceEndpoint.Builder b ) {}

        private
        void
        spawnEndpoint()
        {
            MingleServiceEndpoint.Builder b =
                new MingleServiceEndpoint.Builder().
                    addRoute( "test:ns@v1", "test-svc", spawnTestService() );
 
            completeBuild( b );

            ep = b.build();
            spawn( ep );
        }

        protected
        void
        startImpl()
            throws Exception
        {
            spawnEndpoint();

            mgCli = 
                MingleServices.createRpcClient( 
                    ep, behavior( ProcessRpcClient.class ) );

            startTest();
        }
    }

    private
    abstract
    class AbstractMissingRouteTest
    extends AbstractServiceEndpointTest
    {
        private final MingleTypeReference exTyp;
        private final CharSequence expctKey;
        private final CharSequence expctVal;

        private
        AbstractMissingRouteTest( MingleTypeReference exTyp,
                                  CharSequence expctKey,
                                  CharSequence expctVal )
        {
            this.exTyp = exTyp;
            this.expctKey = expctKey;
            this.expctVal = expctVal;
        }

        abstract
        MingleServiceRequest
        buildRequest();

        private
        final
        class Handler
        extends RpcHandlerImpl
        {
            @Override void rpcSucceeded( MingleValue mv ) { state.fail(); }

            @Override
            void 
            rpcFailed( MingleException me ) 
            {
                state.equal( exTyp, me.getType() );

                state.equalString(
                    expctVal,
                    MingleSymbolMapAccessor.create( me ).
                        expectString( expctKey )
                );

                testDone();
            }
        }

        final
        void
        startTest()
        {
            beginRpc( buildRequest(), new Handler() );
        }
    }

    @Test
    private
    final
    class NoSuchNamespaceExceptionTest
    extends AbstractMissingRouteTest
    {
        private
        NoSuchNamespaceExceptionTest()
        {
            super(
                MingleServices.TYPE_REF_NO_SUCH_NAMESPACE_EXCEPTION,
                "namespace",
                "no:such:ns@v1"
            );
        }

        MingleServiceRequest
        buildRequest()
        {
            return
                createRequestBuilder().
                    setNamespace( "no:such:ns@v1" ).
                    setOperation( "ignored" ).
                    build();
        }
    }

    @Test
    private
    final
    class NoSuchServiceExceptionTest
    extends AbstractMissingRouteTest
    {
        private
        NoSuchServiceExceptionTest()
        {
            super(
                MingleServices.TYPE_REF_NO_SUCH_SERVICE_EXCEPTION,
                "service",
                "not-a-service-here"
            );
        }

        MingleServiceRequest
        buildRequest()
        {
            return
                createRequestBuilder().
                    setService( "not-a-service-here" ).
                    setOperation( "blah" ).
                    build();
        }
    }

    private
    abstract
    class AbstractRequestTimeoutTest
    extends AbstractServiceEndpointTest
    {
        final long epTimeoutMillis = 2000;
        
        final
        void
        completeBuild( MingleServiceEndpoint.Builder b )
        {
            Duration to = Duration.fromMillis( epTimeoutMillis );
            b.setDefaultRequestTimeout( to );
        }

        private
        final
        class TimeoutExpector
        extends AbstractMingleRpcClientHandler
        {
            private final long start = System.currentTimeMillis();

            @Override
            protected
            void
            rpcFailed( Throwable th )
            {
                try
                {
                    long elapsed = System.currentTimeMillis() - start;

                    Range< Long > r = 
                        Range.closed( epTimeoutMillis, epTimeoutMillis * 2 );
                    
                    code( "elapsed:", elapsed, "; r:", r );
                    state.isTrue( r.includes( elapsed ) );
 
                    testDone();
                }
                catch ( Throwable th2 ) { fail( th2 ); }
            }

            @Override
            protected
            void
            rpcSucceeded( MingleServiceResponse resp )
            {
                fail(
                    state.createFail( 
                        "Expected failure but got success:", 
                        MingleModels.inspect( resp ) ) );
            }
        }

        final
        void
        startTimeoutCall()
        {
            beginRpc(
                createRequestBuilder().
                    setOperation( "long-runner" ).
                    p().setInt64( "delay-millis", epTimeoutMillis * 8 ).
                    p().setString( "echo-value", "hello" ).
                    build(),
                new TimeoutExpector()
            );
        }
    }

    @Test( expected = IllegalStateException.class,
           expectedPattern =
            "^\\QProcess has not been started (no context set): com.bitgirder.mingle.service.MingleServiceEndpointTests$ErrorOnNonNullButUnspawnedServiceTest$ServiceImpl@\\E\\p{XDigit}+$" )
    private
    final
    class ErrorOnNonNullButUnspawnedServiceTest
    extends AbstractVoidProcess
    {
        private
        ErrorOnNonNullButUnspawnedServiceTest()
        {
            super( ProcessRpcClient.create() );
        }

        private
        final
        class ServiceImpl
        extends AbstractVoidProcess
        {
            private
            ServiceImpl()
            {
                super( ProcessRpcServer.createStandard() );
            }

            protected
            void
            startImpl()
            {
                state.fail( "Not intended to be started" ); 
            }
        }

        @Override
        protected
        void
        childExited( AbstractProcess< ? > proc,
                     ProcessExit< ? > exit )
        {
            if ( ! exit.isOk() ) fail( exit.getThrowable() );
            if ( ! hasChildren() ) exit();
        }

        private
        final
        class RpcHandler
        extends AbstractMingleRpcClientHandler
        {
            private final MingleServiceEndpoint ep;

            private RpcHandler( MingleServiceEndpoint ep ) { this.ep = ep; }

            @Override
            protected
            void
            rpcFailed( Throwable th )
            {
                fail( th );
            }

            @Override
            protected
            void
            rpcSucceeded( MingleServiceResponse resp )
            {
                code( "resp:", MingleModels.inspect( resp ) );
            }
        }

        protected
        void
        startImpl()
        {
            MingleServiceEndpoint ep =
                new MingleServiceEndpoint.Builder().
                    addRoute( "ns1@v1", "svc1", new ServiceImpl() ).
                    build();
            
            spawn( ep );

            MingleServices.createRpcClient( ep, this ).beginRpc(
                new MingleServiceRequest.Builder().
                    setNamespace( "ns1@v1" ).
                    setService( "svc1" ).
                    setOperation( "ignored" ).
                    build(),
                new RpcHandler( ep )
            );
        }
    }

    // tests timeout when the target service continues to run and process the
    // request (just takes too long to do so)
    @Test
    private
    final
    class RemoteExecutionTimesOutTest
    extends AbstractRequestTimeoutTest
    {
        void startTest() { startTimeoutCall(); }
    }

    // Tests timeout when the target service has exited before the request
    // arrives but the endpoint does not know not to stop routing requests
    // towards the now-exited process.
    @Test
    private
    final
    class RemoteProcessCrashedTimesOutTest
    extends AbstractRequestTimeoutTest
    {
        @Override
        void
        childExitedImpl( AbstractProcess< ? > child,
                         ProcessExit< ? > exit )
        {
            if ( child instanceof TestService ) startTimeoutCall();
        }

        void startTest() { testService().exitAbruptly(); }
    }

    private
    abstract
    class AbstractManagedRouteTest
    extends AbstractVoidProcess
    {
        // stored as a local final field so we can provide access from outside
        // the process thread
        private final EventManager evMgr;

        private MingleServiceEndpoint ep;

        private
        AbstractManagedRouteTest()
        {
            super( 
                ProcessRpcClient.create(), 
                ProcessManager.create(),
                ProcessRpcServer.createStandard(),
                EventBehavior.create( EventManager.create() ) 
            );

            this.evMgr = behavior( EventBehavior.class ).getEventManager();
        }

        void
        childExitedImpl( AbstractProcess< ? > child,
                         ProcessExit< ? > exit )
            throws Exception
        {}

        void childrenDone() { exit(); }

        @Override
        protected
        final
        void
        childExited( AbstractProcess< ? > child,
                     ProcessExit< ? > exit )
            throws Exception
        {
            childExitedImpl( child, exit );

            if ( ! exit.isOk() ) fail( exit.getThrowable() );
            if ( ! hasChildren() ) childrenDone();
        }

        final
        class TestService
        extends AbstractMingleService
        {
            @Override
            protected
            void
            initService()
                throws Exception
            {
                super.initService();

                AbstractManagedRouteTest.this.testServiceStarted( this );
            }

            @MingleServices.Operation
            private 
            MingleServiceResponse
            getString( MingleServiceCallContext ctx ) 
            { 
                return 
                    createSuccessResponse( 
                        ctx, MingleModels.asMingleString( "hello" ) );
            }

            void
            stop()
            {
                submit(
                    new AbstractTask() { protected void runImpl() { exit(); } }
                );
            }
        }

        final
        class TestServiceControl
        extends AbstractProcessControl< TestService >
        {
            TestServiceControl( int starts )
            {
                super( new ProcessManagementTests.FixedStartTracker( starts ) );
            }

            protected
            TestService
            newProcessImpl()
            {
                return new TestService(); 
            }

            public
            void
            stopProcess( TestService proc )
            {
                Processes.sendStop( proc, behavior( ProcessRpcClient.class ) );
            }
        }

        final
        void
        beginRpc( String svcId,
                  RpcHandler h )
        {
            MingleRpcClient mgCli =
                MingleServices.
                    createRpcClient( ep, behavior( ProcessRpcClient.class ) );

            mgCli.beginRpc(
                new MingleServiceRequest.Builder().
                       setNamespace( "test:ns@v1" ).
                       setService( svcId ).
                       setOperation( "get-string" ).
                       build(),
                h
            );
        }

        void testServiceStarted( final TestService ts ) {}

        final 
        ProcessManager 
        procMan() 
        { 
            return behavior( ProcessManager.class );
        }

        final 
        void
        stopEndpoint()
        {
            Processes.sendStop( ep, behavior( ProcessRpcClient.class ) );
        }

        final EventManager eventMgr() { return evMgr; }

        final 
        EventTopic< ? >
        lcTopic()
        {
            return ProcessManagement.TOPIC_PROCESS_LIFECYCLE;
        }

        abstract
        MingleServiceEndpoint
        buildEndpoint()
            throws Exception;

        void startTest() throws Exception {}

        protected
        final
        void
        startImpl()
            throws Exception
        {
            ep = buildEndpoint();
            spawn( ep );

            startTest();
        }

        abstract
        class RpcHandler
        extends AbstractMingleRpcClientHandler
        {
            void rpcSucceeded( String s ) {}

            final
            boolean
            isInternalServiceException( MingleException me )
            {
                return
                    me.getType().equals(
                        MingleServices.TYPE_REF_INTERNAL_SERVICE_EXCEPTION );
            }

            void
            rpcFailed( MingleException me )
            {
                state.fail( MingleModels.inspect( me ) );
            }

            @Override
            protected
            void
            rpcSucceeded( MingleServiceResponse resp )
            {
                if ( resp.isOk() )
                {
                    rpcSucceeded( 
                        ( (MingleString) resp.getResult() ).toString() );
                }
                else rpcFailed( resp.getException() );
            }

            @Override protected void rpcFailed( Throwable th ) { fail( th ); }
        }
    }

    @Test
    private
    final
    class ManagedRouteTargetFailureBehaviorsTest
    extends AbstractManagedRouteTest
    {
        // number of test services started
        private int starts;

        private TestService activeSvc;

        // number of rpcs begun
        private int calls;

        private TestService svc2;

        private
        final
        class Svc1RpcHandler
        extends RpcHandler
        {
            private final long start = System.currentTimeMillis();

            @Override
            protected
            void
            rpcFailed( MingleException me )
            {
                if ( isInternalServiceException( me ) )
                {
                    state.equalInt( 3, calls );

                    // This is a pretty loose upper bound and mainly accounts
                    // for heavy queue traffic when this is run as part of a
                    // loaded test run. Running uncontested this request should
                    // complete in a few millis. All we really care is that it
                    // is much less than the endpoint's default request timeout
                    state.isTrue( 
                        System.currentTimeMillis() - start <= 3000 );

                    stopEndpoint();
                }
                else super.rpcFailed( me );
            }

            @Override
            protected
            void
            rpcSucceeded( String str )
            {
                state.equal( "hello", str );
                state.isTrue( calls < 3 );

                activeSvc.stop();
                activeSvc = null;

                if ( starts == 2 )
                {
                    Duration stopWait = Duration.fromMillis( 300 );
    
                    submit( 
                        new AbstractTask() {
                            protected void runImpl() { startNextSvc1Rpc(); } 
                        },
                        stopWait
                    );
                }
            }
        }

        private
        void
        startNextSvc1Rpc()
        {
            beginRpc( "svc1", new Svc1RpcHandler() );
            ++calls;
        }

        @Override
        void
        testServiceStarted( final TestService ts )
        {
            submit(
                new AbstractTask() {
                    protected void runImpl() 
                    {
                        if ( ts != svc2 )
                        {
                            activeSvc = ts;
    
                            state.isTrue( ++starts <= 2 );
                            startNextSvc1Rpc();
                        }
                    }
                }
            );
        }

        // just check that an unmanaged route still functions as expected
        // alongside other managed routes
        private
        void
        beginSvc2Call()
        {
            beginRpc( 
                "svc2",
                new RpcHandler()
                    {
                        @Override protected void rpcSucceeded( String resp ) 
                        { 
                            state.equal( "hello", resp );
                            svc2.stop(); 
                        }
                    }
            );
        }

        MingleServiceEndpoint
        buildEndpoint()
        {
            procMan().manage( "svc1-id", new TestServiceControl( 2 ) );

            svc2 = new TestService();
            spawn( svc2 );

            return
                new MingleServiceEndpoint.Builder().
                    setDefaultRequestTimeout( Duration.fromSeconds( 15 ) ).
                    setManagerSubscription( eventMgr(), lcTopic(), this ).
                    addManagedRoute( "test:ns@v1", "svc1", "svc1-id" ).
                    addRoute( "test:ns@v1", "svc2", svc2 ).
                    build();
        }

        @Override void startTest() { beginSvc2Call(); }
    }

    @Test
    private
    final
    class ReqsFailFastWhenRouteManagerStops
    extends AbstractManagedRouteTest
    {
        private ManagerEncloser me;

        private
        void
        startRpc()
        {
            final long start = System.currentTimeMillis();

            beginRpc( "svc1", new RpcHandler() {

                @Override 
                protected void rpcSucceeded( String s ) {
                    fail( state.createFail( "Got success:", s ) );
                }

                @Override
                protected void rpcFailed( MingleException me ) 
                {
                    state.isTrue( isInternalServiceException( me ) );

                    state.isTrue( System.currentTimeMillis() - start <= 3000 );
                    exit();
                }
            });
        }

        @Override
        void
        childExitedImpl( AbstractProcess< ? > child,
                         ProcessExit< ? > exit )
        {
            if ( child == me && exit.isOk() ) startRpc();
        }

        @Override
        void
        testServiceStarted( TestService ts )
        {
            submit(
                new AbstractTask() {
                    protected void runImpl()
                    {
                        Processes.
                            sendStop( me, behavior( ProcessRpcClient.class ) );
                    }
                }
            );
        }

        // We use this innner class instead of using our own manager since we
        // want to be able to stop the manager and then stick around to send rpc
        // requests to the endpoint in front of the (by then exited) managed
        // targets. 
        private
        final
        class ManagerEncloser
        extends AbstractVoidProcess
        {
            private
            ManagerEncloser()
            {
                super(
                    ProcessRpcServer.createStandard(),
                    ProcessManager.create(),
                    EventBehavior.create( eventMgr() )
                );
            }

            protected
            void
            startImpl()
            {
                behavior( ProcessManager.class ).
                    manage( "svc1", new TestServiceControl( 1 ) );
            }
        }

        MingleServiceEndpoint
        buildEndpoint()
        {
            me = new ManagerEncloser();
            spawn( me );

            return
                new MingleServiceEndpoint.Builder().
                    setDefaultRequestTimeout( Duration.fromSeconds( 30 ) ).
                    addManagedRoute( "test:ns@v1", "svc1", "svc1" ).
                    setManagerSubscription( eventMgr(), lcTopic(), me ).
                    build();
        }
    }

    @Test
    private
    final
    class ManagedRouteBackLogBoundsTest
    extends AbstractManagedRouteTest
    {
        private final int backLog = 5;

        private int remain = backLog;

        private
        final
        class BackLogHandler
        extends RpcHandler
        {
            private final int idx;
            private final long start = System.currentTimeMillis();

            private BackLogHandler( int idx ) { this.idx = idx; }

            @Override
            protected
            void
            rpcFailed( Throwable th )
            {
                if ( idx == backLog )
                {
                    state.isTrue( th instanceof InternalServiceException );
                    state.isTrue( System.currentTimeMillis() - start <= 3000 );

                    // now start the managed targets and expect the backlogged
                    // requests to complete
                    procMan().manage( "svc1", new TestServiceControl( 1 ) );
                }
                else super.rpcFailed( th );
            }

            @Override
            protected
            void
            rpcSucceeded( String resp )
            {
                state.isFalse( idx == backLog );

                state.equal( "hello", resp );

                if ( --remain == 0 ) 
                {
                    stopEndpoint();
                    exit();
                }
            }
        }

        @Override void childrenDone() { shutdownDone(); }

        void
        startTest()
        {
            for ( int i = 0; i <= backLog; ++i )
            {
                beginRpc( "svc1", new BackLogHandler( i ) );
            }
        }
            
        MingleServiceEndpoint
        buildEndpoint()
        {
            MingleServiceEndpoint.ManagedRouteOptions opts =
                MingleServiceEndpoint.ManagedRouteOptions.
                    withBackLog( backLog );

            return
                new MingleServiceEndpoint.Builder().
                    addManagedRoute( "test:ns@v1", "svc1", "svc1", opts ).
                    setManagerSubscription( eventMgr(), lcTopic(), this ).
                    build();
        }
    }
}
