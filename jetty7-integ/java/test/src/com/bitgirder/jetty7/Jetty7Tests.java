package com.bitgirder.jetty7;

import static com.bitgirder.mingle.service.TestServiceConstants.*;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Completion;
import com.bitgirder.lang.Lang;

import com.bitgirder.concurrent.Duration;

import com.bitgirder.io.Charsets;

import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.ProcessRpcServer;
import com.bitgirder.process.ProcessRpcClient;
import com.bitgirder.process.ProcessExit;

import com.bitgirder.mingle.service.MingleServices;
import com.bitgirder.mingle.service.MingleServiceCallContext;
import com.bitgirder.mingle.service.MingleServiceTests;
import com.bitgirder.mingle.service.MingleServiceEndpoint;
import com.bitgirder.mingle.service.NativeTestService;
import com.bitgirder.mingle.service.MingleRpcClient;
import com.bitgirder.mingle.service.AbstractMingleRpcClientHandler;
import com.bitgirder.mingle.service.AbstractMingleService;
import com.bitgirder.mingle.service.TestServiceConstants;
import com.bitgirder.mingle.service.InternalServiceException;

import com.bitgirder.mingle.model.MingleServiceRequest;
import com.bitgirder.mingle.model.MingleServiceResponse;
import com.bitgirder.mingle.model.MingleString;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleNull;
import com.bitgirder.mingle.model.MingleIdentifierFormat;
import com.bitgirder.mingle.model.ModelTestInstances;

import com.bitgirder.mingle.codec.MingleCodec;
import com.bitgirder.mingle.codec.MingleCodecs;

import com.bitgirder.mingle.http.MingleHttpCodecContext;
import com.bitgirder.mingle.http.MingleHttpCodecFactory;
import com.bitgirder.mingle.http.MingleHttpTesting;

import com.bitgirder.mingle.http.json.JsonMingleHttpCodecFactory;

import com.bitgirder.test.Test;
import com.bitgirder.test.AbstractLabeledTestObject;
import com.bitgirder.test.TestFactory;
import com.bitgirder.test.TestRuntime;

import java.nio.ByteBuffer;

import java.util.Map;
import java.util.List;

import javax.servlet.http.HttpServletRequest;

final
class Jetty7Tests
extends AbstractLabeledTestObject
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final TestRuntime rt;

    private
    Jetty7Tests( TestRuntime rt ) 
    { 
        super( Jetty7Tests.class.getSimpleName() );
        this.rt = rt; 
    }

    private
    abstract
    class AbstractJetty7MingleTest
    extends AbstractVoidProcess
    implements Jetty7HttpMingleServer.EventHandler
    {
        private final String host = "localhost";

        private MingleServiceEndpoint ep;
        private Jetty7MingleConnector mgConn;
        private Jetty7HttpMingleServer srv;
        private Jetty7HttpClient httpCli;
        private Jetty7HttpMingleRpcClient mgRpcCli;

        private int port;

        private
        AbstractJetty7MingleTest()
        {
            super( ProcessRpcClient.create() ); 
        }

        final MingleRpcClient mgRpcCli() { return mgRpcCli; }
        final Jetty7HttpClient httpCli() { return httpCli; }
        final Jetty7MingleConnector mgConn() { return mgConn; }
        final String host() { return host; }

        final void setEndpoint( MingleServiceEndpoint ep ) { this.ep = ep; }

        final 
        int 
        port() 
        { 
            state.isTrue( port > 0, "port not yet determined" );
            return port; 
        }

        final
        Jetty7HttpClient.BufferRequest.Builder
        createBufferRequestBuilder()
        { 
            // the setScheme() isn't needed for the test technically, but we
            // include it to get basic coverage of its explicity handling by
            // Jetty7HttpClient
            return
                new Jetty7HttpClient.BufferRequest.Builder().
                    setMethod( Jetty7HttpClient.METHOD_POST ).
                    setHost( host() ).
                    setPort( port() ).
                    setScheme( "http" ).
                    setUri( mgConn().getServicePath() );
        }

        void
        childExitedImpl( AbstractProcess< ? > proc,
                         ProcessExit< ? > exit )
            throws Exception
        {}

        @Override
        protected
        final
        void
        childExited( AbstractProcess< ? > proc,
                     ProcessExit< ? > exit )
            throws Exception
        {
            childExitedImpl( proc, exit );
            if ( ! hasChildren() ) exit();
        }

        // must call setEndpoint() before calling onReady
        abstract
        void
        startTestProcesses( Runnable onReady )
            throws Exception;

        final
        void
        sendStop( AbstractProcess< ? > proc )
        {
            behavior( ProcessRpcClient.class ).
                sendAsync( proc, new ProcessRpcServer.Stop() );
        }

        abstract
        void
        stopEndpoint( MingleServiceEndpoint ep );

        private
        void
        stopTestProcesses()
        {
            httpCli.stop();
            srv.stop();

            stopEndpoint( ep );
        }

        final void testDone() { stopTestProcesses(); }

        abstract
        void
        startTest()
            throws Exception;

        Jetty7HttpMingleRpcClient.Builder
        completeBuild( Jetty7HttpMingleRpcClient.Builder b )
        {
            return b;
        }

        private
        Jetty7HttpMingleRpcClient
        buildMgRpcCli()
        {
            MingleHttpCodecContext codecCtx =
                JsonMingleHttpCodecFactory.
                    getInstance().
                    getDefaultCodecContext();

            return
                completeBuild(
                    new Jetty7HttpMingleRpcClient.Builder().
                        setActivityContext( getActivityContext() ).
                        setHost( host() ).
                        setPort( port ).
                        setUri( "/service" ).
                        setCodecContext( codecCtx ).
                        setHttpClient( httpCli )
                ).
                build();
        }
 
        public
        void
        httpStarted( final int port )
        {
            this.port = port;

            submit(
                new AbstractTask() {
                    protected void runImpl() throws Exception
                    { 
                        mgRpcCli = buildMgRpcCli();
                        startTest();
                    }
                }
            );
        }

        void completeBuild( Jetty7MingleConnector.Builder b ) {}

        private
        void
        spawnServer()
        {
            MingleHttpCodecFactory codecFact =
                MingleHttpTesting.expectServerCodecFactory( rt );

            Jetty7MingleConnector.Builder b = 
                new Jetty7MingleConnector.Builder().
                    setCodecFactory( codecFact ).
                    setMingleEndpoint( ep );
            
            completeBuild( b );
            mgConn = b.build();

            srv =
                new Jetty7HttpMingleServer.Builder().
                    setMingleConnector( mgConn ).
                    setEventHandler( this ).
                    build();
            
            spawn( srv );
        }

        protected
        final
        void
        startImpl()
            throws Exception
        {
            httpCli = Jetty7HttpClient.create();
            spawn( httpCli );

            startTestProcesses(
                new Runnable() { public void run() { spawnServer(); } } );
        }

        abstract
        class RpcHandler
        extends AbstractMingleRpcClientHandler
        {
            @Override protected void rpcFailed( Throwable th ) { fail( th ); }
        }
    }

    private
    abstract
    class AbstractJetty7MingleServiceTestContextTest
    extends AbstractJetty7MingleTest
    implements Jetty7HttpMingleServer.EventHandler
    {
        private final NativeTestService svc = new NativeTestService();

        final
        MingleServiceRequest.Builder
        createBaseRequestBuilder()
        {
            return
                new MingleServiceRequest.Builder().
                    setNamespace( TEST_NS ).
                    setService( TEST_SVC );
        }

        void
        startTestProcesses( final Runnable onComp )
            throws Exception
        {
            spawn( svc );

            MingleServiceEndpoint ep =
                new MingleServiceEndpoint.Builder().
                    addRoute(
                        TestServiceConstants.TEST_NS,
                        TestServiceConstants.TEST_SVC,
                        svc
                    ).
                    build();
 
            spawn( ep );
            setEndpoint( ep );
            onComp.run();
        }

        void
        stopEndpoint( MingleServiceEndpoint ep )
        {
            sendStop( ep );
            svc.stop();
        }

        abstract
        class SingleCallHandler
        implements MingleRpcClient.Handler
        {
            public
            void
            rpcFailed( Throwable th,
                       MingleServiceRequest mgReq,
                       MingleRpcClient cli )
            {
                fail( th );
            }

            abstract
            void
            checkResponse( MingleServiceResponse mgResp )
                throws Exception;
    
            public
            void
            rpcSucceeded( MingleServiceResponse mgResp,
                          MingleServiceRequest mgReq,
                          MingleRpcClient rpcCli )
            {
                try
                {
                    checkResponse( mgResp );
                    testDone();
                }
                catch ( Exception ex ) { fail( ex ); }
            }
        }
    }

    @Test
    private
    final
    class Jetty7HttpTimeoutTest
    extends AbstractJetty7MingleServiceTestContextTest
    {
        private final Duration contTimeout = Duration.fromMillis( 100 );
        private final Duration echoDelay = Duration.fromMillis( 3000 );

        private final MingleValue echoValue =
            ModelTestInstances.TEST_STRUCT1_INST1;

        @Override
        void
        completeBuild( Jetty7MingleConnector.Builder b )
        {
            b.setContinuationTimeout( contTimeout );
        }

        void
        startTest()
        {
            MingleServiceRequest.Builder b =
                createBaseRequestBuilder().
                setOperation( OP_DO_DELAYED_ECHO );
 
            b.params().setInt64( ID_DELAY_MILLIS, echoDelay.asMillis() );
            b.params().set( ID_ECHO_VALUE, echoValue );

            mgRpcCli().beginRpc( b.build(), new SingleCallHandler() {
                protected void checkResponse( MingleServiceResponse mgResp ) 
                {
                    state.equal(
                        MingleServices.TYPE_REF_INTERNAL_SERVICE_EXCEPTION,
                        mgResp.getException().getType() );
                }
            });
        }
    }

    private final static class TestCrashException extends Exception {}

    // Test to ensure that calls made to a service which crashed (and which did
    // not recover or restart) still timeout at the jetty continuation layer and
    // are received by callers as InternalServiceExceptions. This is slightly
    // different from Jetty7HttpTimeoutTest in that this test ensures that the
    // timeouts handled by MingleServiceEndpoint itself are communicated to the
    // caller correctly and in the expected amount of time without the Jetty
    // continuation being invoked. Jetty7HttpTimeoutTest does the opposite and
    // ensures that even if the service endpoint does not time a request out,
    // Jetty7MingleConnector will.
    @Test
    private
    final
    class CallsToCrashedServicesFailCompletelyTest
    extends AbstractJetty7MingleTest
    {
        private TestService svc;
        
        private int waitCount = 2; // 1 call before crash; one after

        private long call2Started;

        private
        final
        class HandlerImpl
        extends AbstractMingleRpcClientHandler
        {
            @Override
            public
            void
            rpcSucceeded( MingleServiceResponse resp )
            {
                state.isFalse( resp.isOk() );

                state.equal(
                    MingleServices.TYPE_REF_INTERNAL_SERVICE_EXCEPTION,
                    resp.getException().getType()
                );

                if ( --waitCount == 0 ) 
                {
                    long elapsedMs = 
                        System.currentTimeMillis() - call2Started;

                    state.isTrue( elapsedMs < 5000 );

                    testDone();
                }
            }

            @Override public void rpcFailed( Throwable th ) { fail( th ); }
        }

        private
        void
        makeFailerCall()
            throws Exception
        {
            mgRpcCli().beginRpc(
                new MingleServiceRequest.Builder().
                    setNamespace( "jetty7:test@v1" ).
                    setService( "test-svc" ).
                    setOperation( "failer" ).
                    build(),
                new HandlerImpl()
            );
        }

        private
        final
        class TestService
        extends AbstractMingleService
        {
            @MingleServices.Operation
            private 
            void 
            failer( MingleServiceCallContext callCtx,
                    ProcessRpcServer.ResponderContext< Object > ignored ) 
            { 
                fail( new TestCrashException() ); 
            }
        }

        void
        startTestProcesses( Runnable onComp )
            throws Exception
        {
            spawn( svc = new TestService() );

            MingleServiceEndpoint ep =
                new MingleServiceEndpoint.Builder().
                    addRoute( "jetty7:test@v1", "test-svc", svc ).
                    setDefaultRequestTimeout( Duration.fromSeconds( 3 ) ).
                    build();
 
            spawn( ep );

            setEndpoint( ep );
            onComp.run();
        }
 
        void stopEndpoint( MingleServiceEndpoint ep ) { sendStop( ep ); }
        
        void startTest() throws Exception { makeFailerCall(); }

        void
        childExitedImpl( AbstractProcess< ? > child,
                         ProcessExit< ? > exit )
            throws Exception
        {
            if ( child instanceof TestService ) 
            {
                state.isFalse( exit.isOk() );

                call2Started = System.currentTimeMillis();
                makeFailerCall();
            }
        }
    }

    @Test
    private
    final
    class HeadersPassedAsControlParamTest
    extends AbstractJetty7MingleTest
    {
        private final String expctHeaderName = "test-header";
        private final String expctHeaderVal = "test-header-val";

        private TestService svc;

        private
        final
        class TestService
        extends AbstractMingleService
        {
            public void stop() { exit(); }

            @MingleServices.Operation
            private
            MingleServiceResponse
            getHeaderVal( MingleServiceCallContext callCtx )
                throws Exception
            {
                HttpServletRequest req = (HttpServletRequest)
                    callCtx.attachments().get(
                        new MingleServices.ControlName( "servlet-request" ) );

                MingleValue mv = MingleNull.getInstance();

                if ( req != null ) 
                {
                    mv = 
                        MingleModels.
                            asMingleString( req.getHeader( expctHeaderName ) );
                }

                return createSuccessResponse( callCtx, mv );
            }
        }

        void
        startTestProcesses( Runnable onComp )
            throws Exception
        {
            spawn( svc = new TestService() );

            MingleServiceEndpoint ep =
                new MingleServiceEndpoint.Builder().
                    addRoute( "some:ns@v1", "some-service", svc ).
                    build();
            
            spawn( ep );
            setEndpoint( ep );
            onComp.run();
        }

        void
        stopEndpoint( MingleServiceEndpoint ep )
        {
            sendStop( ep );
            svc.stop();
        }

        @Override
        Jetty7HttpMingleRpcClient.Builder
        completeBuild( Jetty7HttpMingleRpcClient.Builder b )
        {
            Map< String, List< String > > hdrs = Lang.newMap();
            hdrs.put( expctHeaderName, Lang.singletonList( expctHeaderVal ) );

            return b.setHeaders( hdrs );
        }

        void
        startTest()
            throws Exception
        {
            mgRpcCli().beginRpc(
                new MingleServiceRequest.Builder().
                    setNamespace( "some:ns@v1" ).
                    setService( "some-service" ).
                    setOperation( "get-header-val" ).
                    build(),
                new RpcHandler() 
                {
                    @Override 
                    protected void rpcSucceeded( MingleServiceResponse mgResp )
                    {
                        MingleString ms = (MingleString) mgResp.getResult();
                        state.equalString( expctHeaderVal, ms );

                        testDone();
                    }
                }
            );
        }
    }

    @TestFactory
    private
    static
    List< ? > 
    createTest( TestRuntime rt )
    {
        return Lang.singletonList( new Jetty7Tests( rt ) );
    }
}
