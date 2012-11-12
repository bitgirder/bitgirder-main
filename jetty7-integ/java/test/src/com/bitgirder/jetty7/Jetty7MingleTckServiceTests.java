package com.bitgirder.jetty7;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.ObjectReceiver;

import com.bitgirder.process.ProcessActivity;
import com.bitgirder.process.ProcessRpcServer;
import com.bitgirder.process.AbstractProcess;

import com.bitgirder.mingle.tck.v1.MingleTckServiceTests;
import com.bitgirder.mingle.tck.v1.MingleTckTests;
import com.bitgirder.mingle.tck.v1.AbstractTckServiceTestBackend;

import com.bitgirder.mingle.http.MingleHttpTesting;
import com.bitgirder.mingle.http.MingleHttpCodecContext;
import com.bitgirder.mingle.http.MingleHttpCodecFactory;
import com.bitgirder.mingle.http.MingleHttpCodecFactorySelector;
import com.bitgirder.mingle.http.MingleHttpTckServiceTests;
import com.bitgirder.mingle.http.MingleHttpServiceImplTests;
import com.bitgirder.mingle.http.MingleHttpTesting;
import com.bitgirder.mingle.http.MingleHttpCodecContext;

import com.bitgirder.mingle.http.test.BitGirderHttpMingleTests;

import com.bitgirder.mingle.service.MingleRpcClient;

import com.bitgirder.mingle.codec.MingleCodec;

import com.bitgirder.mingle.bind.MingleBindTests;
import com.bitgirder.mingle.bind.MingleBinder;

import com.bitgirder.mingle.model.MingleIdentifiedName;

import com.bitgirder.mglib.MgLibTesting;

import com.bitgirder.test.Test;
import com.bitgirder.test.TestFactory;
import com.bitgirder.test.TestRuntime;

import com.bitgirder.testing.Testing;

import java.util.List;
import java.util.Map;

@Test
final
class Jetty7MingleTckServiceTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static Object KEY_BACKEND =
        Jetty7MingleTckServiceTests.class.getName() + ".backend";

    private final TestRuntime rt;
    private final BackendImpl backend;

    private
    Jetty7MingleTckServiceTests( TestRuntime rt )
    {
        this.rt = rt;

        this.backend =
            Testing.expectObject( rt, KEY_BACKEND, BackendImpl.class );
    }

    private
    AbstractProcess< ? >
    getBitGirderHttpClientFactory()
    {
        return MingleHttpTesting.expectBitGirderHttpClientFactory( rt );
    }

    private
    MingleHttpTesting.ServerLocation
    getBitGirderServerLocation( boolean isSsl )
    {
        return BitGirderHttpMingleTests.getServerLocation( rt, isSsl );
    }

    private
    MingleHttpTckServiceTests.Builder
    initServiceTestsBuilder( MingleHttpCodecContext codecCtx )
    {
        return
            new MingleHttpTckServiceTests.Builder().
                setMingleBinder( MgLibTesting.expectDefaultBinder( rt ) ).
                setCodecContext( codecCtx );
    }

    private
    void
    addClientTests( MingleHttpCodecContext codecCtx,
                    List< Object > l )
    {
        String lblBase = "ClientTests/ctype=" + codecCtx.contentType() + ",";

        l.add(
            initServiceTestsBuilder( codecCtx ).
            setLabel( lblBase + "cli=jetty7,srv=bitgirder,ssl=false" ).
            setClientFactory( backend ).
            setLocation( getBitGirderServerLocation( false ) ).
            build()
        );

        l.add(
            initServiceTestsBuilder( codecCtx ).
            setLabel( lblBase + "cli=jetty7,srv=jetty7" ).
            setClientFactory( backend ).
            setLocation( backend.getTestLocation() ).
            build()
        );

        l.add(
            initServiceTestsBuilder( codecCtx ).
            setLabel( lblBase + "cli=bitgirder,srv=jetty7" ).
            setClientFactory( getBitGirderHttpClientFactory() ).
            setLocation( backend.getTestLocation() ).
            build()
        );
    }

    private
    void
    addClientTests( List< Object > l )
    {
        for ( MingleHttpCodecContext codecCtx :
                MingleHttpTesting.expectClientCodecContexts( rt ) )
        {
            addClientTests( codecCtx, l );
        }
    }

    @TestFactory
    private
    List< ? >
    createTckTests()
    {
        List< Object > res = Lang.newList();

        addClientTests( res );

        res.add(
            new MingleHttpServiceImplTests.Builder().
                setLabel( "ServiceImplTests" ).
                setLocation( backend.getTestLocation() ).
                setRuntime( rt ).
                build()
        );

        return res;
    }

    // Spawns a jetty server which will serve mingle reqs and holds rpc requests
    // and blocks its RuntimeInitializer completion until the jetty server has
    // started and chosen a port
    private
    final
    static
    class BackendImpl
    extends AbstractTckServiceTestBackend
    implements Jetty7HttpMingleServer.EventHandler
    {
        private final MingleHttpCodecFactory codecFact;
        private final Testing.RuntimeInitializerContext ctx;

        private Jetty7HttpMingleServer srv;
        private final Jetty7HttpClient httpCli = Jetty7HttpClient.create();

        private int port;
        private final String host = "localhost";
        private final String uri = "/service";

        private
        BackendImpl( MingleBinder mb,
                     MingleHttpCodecFactory codecFact,
                     Testing.RuntimeInitializerContext ctx )
        {
            super( mb );

            this.codecFact = codecFact;
            this.ctx = ctx;
        }

        private
        MingleHttpTesting.ServerLocation
        getTestLocation()
        {
            return 
                new MingleHttpTesting.ServerLocation( host, port, uri, false );
        }

        private
        final
        class RpcClientFactoryImpl
        implements MingleTckTests.MingleRpcClientFactory
        {
            private final MingleHttpTesting.ServerLocation testLoc;
            private final MingleHttpCodecContext codecCtx;

            private 
            RpcClientFactoryImpl( MingleHttpTesting.ServerLocation testLoc,
                                  MingleHttpCodecContext codecCtx ) 
            { 
                this.testLoc = testLoc;
                this.codecCtx = codecCtx;
            }

            public
            MingleRpcClient
            createRpcClient( ProcessActivity.Context ctx )
            {
                return
                    new Jetty7HttpMingleRpcClient.Builder().
                        setActivityContext( ctx ).
                        setHost( testLoc.host() ).
                        setPort( testLoc.port() ).
                        setUri( testLoc.uri() ).
                        setCodecContext( codecCtx ).
                        setHttpClient( httpCli ).
                        build();
            }
        }

        @ProcessRpcServer.Responder
        private
        MingleTckTests.MingleRpcClientFactory
        handle( MingleHttpTckServiceTests.CreateRpcClientFactory req )
        {
            return 
                new RpcClientFactoryImpl( 
                    req.testLocation(), req.codecContext() );
        }

        private
        void
        setServerLocation( final Runnable onComp )
        {
            Testing.awaitTestObject(
                ctx,
                MingleHttpTesting.KEY_SERVER_LOCATIONS,
                Object.class,
                new ObjectReceiver< Object >() {
                    public void receive( Object obj )
                    {
                        Map< MingleIdentifiedName, 
                             MingleHttpTesting.ServerLocation > m =
                            Lang.castUnchecked( obj );
 
                        m.put( 
                            MingleIdentifiedName.
                                create( "bitgirder:jetty7@v1/server/plain" ),
                            getTestLocation()
                        );

                        submit( onComp );
                    }
                }
            );
        }

        public
        void
        httpStarted( int port )
        {
            this.port = port;

            setServerLocation( new AbstractTask() {
                protected void runImpl()
                {
                    resumeRpcRequests();
                    ctx.complete();
                }
            });
        }

        @Override
        protected
        void
        completeBackendInit()
        {
            holdRpcRequests();

            spawnStoppable( httpCli );

            srv =
                new Jetty7HttpMingleServer.Builder().
                    setMingleConnector( 
                        new Jetty7MingleConnector.Builder().
                            setMingleEndpoint( endpoint() ).
                            setCodecFactory( codecFact ).
                            build()
                    ).
                    setEventHandler( this ).
                    build();
 
            spawnStoppable( srv );
        }
    }

    private
    static
    void
    awaitBinder( final MingleHttpCodecFactory codecFact,
                 final Testing.RuntimeInitializerContext ctx )
    {
        Testing.awaitTestObject(
            ctx,
            MingleBindTests.KEY_BINDER,
            MingleBinder.class,
            new ObjectReceiver< MingleBinder >() {
                public void receive( final MingleBinder mb ) 
                {
                    ctx.spawnAndSetStoppable( 
                        new BackendImpl( mb, codecFact, ctx ), KEY_BACKEND );
                }
            }
        );
    }

    @Testing.RuntimeInitializer
    private
    static
    void
    init( final Testing.RuntimeInitializerContext ctx )
    {
        Testing.awaitTestObject(
            ctx,
            MingleHttpTesting.KEY_SERVER_CODEC_FACTORY,
            MingleHttpCodecFactory.class,
            new ObjectReceiver< MingleHttpCodecFactory >() {
                public void receive( MingleHttpCodecFactory codecFact ) {
                    awaitBinder( codecFact, ctx );
                }
            }
        );
    }
}
