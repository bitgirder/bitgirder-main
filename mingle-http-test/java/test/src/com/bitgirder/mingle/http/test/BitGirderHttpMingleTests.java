package com.bitgirder.mingle.http.test;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;
import com.bitgirder.lang.ObjectReceiver;

import com.bitgirder.mingle.http.MingleHttpCodecFactory;
import com.bitgirder.mingle.http.MingleHttpCodecFactorySelector;
import com.bitgirder.mingle.http.MingleHttpResponder;
import com.bitgirder.mingle.http.MingleHttpServiceImplTests;
import com.bitgirder.mingle.http.MingleHttpTesting;
import com.bitgirder.mingle.http.MingleHttpTckServiceTests;
import com.bitgirder.mingle.http.MingleHttpCodecContext;

import com.bitgirder.mingle.tck.v1.AbstractTckServiceTestBackend;

import com.bitgirder.mingle.bind.MingleBindTests;
import com.bitgirder.mingle.bind.MingleBinder;

import com.bitgirder.mingle.model.MingleIdentifiedName;

import com.bitgirder.mglib.MgLibTesting;

import com.bitgirder.net.SelectorManager;
import com.bitgirder.net.NetTests;
import com.bitgirder.net.NetProtocolTransportFactory;
import com.bitgirder.net.ConnectionTransportFactory;

import com.bitgirder.net.ssl.NetSslTests;
import com.bitgirder.net.ssl.SslIoExecutorTransportFactory;

import com.bitgirder.http.HttpServer;
import com.bitgirder.http.HttpService;
import com.bitgirder.http.HttpResponderFactories;
import com.bitgirder.http.DefaultHttpService;

import com.bitgirder.test.Test;
import com.bitgirder.test.TestRuntime;
import com.bitgirder.test.TestFactory;

import com.bitgirder.testing.Testing;

import java.util.List;
import java.util.Map;

import java.net.InetSocketAddress;

@Test
public
final
class BitGirderHttpMingleTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static String SVC_URI = "/service";

    private final TestRuntime rt;

    private BitGirderHttpMingleTests( TestRuntime rt ) { this.rt = rt; }

    public
    static
    Object
    getServerLocationKey( boolean isSsl )
    {
        return 
            BitGirderHttpMingleTests.class.getName() + ".server,isSsl=" + isSsl;
    }

    public
    static
    MingleHttpTesting.ServerLocation
    getServerLocation( TestRuntime rt,
                       boolean isSsl )
    {
        inputs.notNull( rt, "rt" );

        return
            Testing.expectObject(
                rt,
                getServerLocationKey( isSsl ),
                MingleHttpTesting.ServerLocation.class
            );
    }

    private
    String
    createLabel( boolean isSsl,
                 boolean useGzip,
                 MingleHttpCodecContext codecCtx )
    {
        return 
            "ClientTests/" +
            Strings.crossJoin( "=", ",",
                "cli", "bitgirder",
                "srv", "bitgirder",
                "codecCtx", codecCtx.contentType(),
                "isSsl", isSsl,
                "useGzip", useGzip
            );
    }

    private
    MingleHttpTckServiceTests
    createClientTests( boolean isSsl,
                       boolean useGzip,
                       MingleHttpCodecContext codecCtx )
    {
        String lbl = createLabel( isSsl, useGzip, codecCtx );

        return
            new MingleHttpTckServiceTests.Builder().
                setLabel( lbl ).
                setClientFactory( 
                    MingleHttpTesting.expectBitGirderHttpClientFactory( rt ) ).
                setUseGzip( useGzip ).
                setLocation( getServerLocation( rt, isSsl ) ).
                setCodecContext( codecCtx ).
                setMingleBinder( MgLibTesting.expectDefaultBinder( rt ) ).
                build();
    }

    private
    MingleHttpServiceImplTests
    createServiceImplTests( boolean isSsl )
    {
        return
            new MingleHttpServiceImplTests.Builder().
                setLabel( "BitGirderServiceImplTests/ssl=" + isSsl ).
                setLocation( getServerLocation( rt, isSsl ) ).
                setRuntime( rt ).
                build();
    }

    @TestFactory
    private
    List< ? >
    createTests( TestRuntime rt )
    {
        List< Object > res = Lang.newList();

        List< MingleHttpCodecContext > l =
            MingleHttpTesting.expectClientCodecContexts( rt );

        for ( MingleHttpCodecContext codecCtx : l )
        for ( int i = 0; i < 2; ++i )
        {
            for ( int j = 0; j < 2; ++j )
            {
                res.add( createClientTests( i == 0, j == 0, codecCtx ) );
            }

            res.add( createServiceImplTests( i == 0 ) );
        }

        return res;
    }

    private
    final
    static
    class InitContext
    {
        private Testing.RuntimeInitializerContext ctx;

        private MingleBinder mb;
        private SelectorManager selMgr;
        private SslIoExecutorTransportFactory sslTransFact;
        private MingleHttpCodecFactory codecFact;

        private Map< MingleIdentifiedName, MingleHttpTesting.ServerLocation >
            srvLocs;

        private int initWait = 5;

        private
        void
        initServers()
        {
            ctx.spawnStoppable( new BackendImpl( this ) );
        }

        private
        < V >
        ObjectReceiver< V >
        wrap( final ObjectReceiver< V > recv )
        {
            return
                new ObjectReceiver< V >() {
                    public void receive( V obj ) throws Exception 
                    {
                        recv.receive( obj );
                        if ( --initWait == 0 ) initServers();
                    }
                };
        }

        private
        < V >
        void
        awaitTestObject( Object key,
                         Class< V > cls,
                         ObjectReceiver< V > recv )
        {
            Testing.awaitTestObject( ctx, key, cls, wrap( recv ) );
        }
    }

    private
    final
    static
    class BackendImpl
    extends AbstractTckServiceTestBackend
    {
        private final InitContext initCtx;

        private int startWait = 2;

        private 
        BackendImpl( InitContext initCtx ) 
        { 
            super( initCtx.mb );

            this.initCtx = initCtx; 
        }

        private
        HttpService
        createHttpService()
        {
            MingleHttpResponder resp =
                new MingleHttpResponder.Builder().
                    setEndpoint( endpoint() ).
                    setCodecFactory( initCtx.codecFact ).
                    build();

            return
                new DefaultHttpService.Builder().
                    setResponderFactory(
                        HttpResponderFactories.forUri( SVC_URI, resp )
                    ).
                    build();
        }

        private
        NetProtocolTransportFactory
        getTransportFactory( boolean isSsl )
        {
            return isSsl
                ? initCtx.sslTransFact
                : ConnectionTransportFactory.create();
        }

        private
        HttpServer
        createServer( HttpService svc,
                      boolean isSsl )
        {
            return
                new HttpServer.Builder().
                    setNetworking( initCtx.selMgr ).
                    setChooseEphemeralListenAddress().
                    setService( svc ).
                    setTransportFactory( getTransportFactory( isSsl ) ).
                    build();
        }

        private
        void
        putServerLocation( MingleHttpTesting.ServerLocation loc )
        {
            initCtx.srvLocs.put(
                MingleIdentifiedName.create(
                    "bitgirder:mingle:http:test@v1/bgServer/" +
                        ( loc.isSsl() ? "ssl" : "plain" )
                ),
                loc
            );
        }

        private
        void
        setTestLocation( boolean isSsl,
                         InetSocketAddress addr )
        {
            MingleHttpTesting.ServerLocation loc =
                new MingleHttpTesting.ServerLocation(
                    addr.getHostName(),
                    addr.getPort(),
                    SVC_URI,
                    isSsl
                );
            
            initCtx.ctx.setObject( getServerLocationKey( isSsl ), loc );
            putServerLocation( loc );

            if ( --startWait == 0 ) initCtx.ctx.complete();
        }

        private
        void
        startServer( HttpService svc,
                     final boolean isSsl )
        {
            HttpServer srv = createServer( svc, isSsl );
            spawnStoppable( srv );

            beginRpc( srv, new HttpServer.GetLocalAddress(), 
                new DefaultRpcHandler() {
                    @Override protected void rpcSucceeded( Object resp ) {
                        setTestLocation( isSsl, (InetSocketAddress) resp );
                    }
                }
            );
        }

        @Override
        protected
        void
        completeBackendInit()
        {
            HttpService svc = createHttpService();

            startServer( svc, true );
            startServer( svc, false );
        }
    }

    @Testing.RuntimeInitializer
    private
    static
    void
    init( Testing.RuntimeInitializerContext ctx )
    {
        final InitContext initCtx = new InitContext();
        initCtx.ctx = ctx;

        initCtx.awaitTestObject( 
            NetTests.KEY_SELECTOR_MANAGER,
            SelectorManager.class,
            new ObjectReceiver< SelectorManager >() {
                public void receive( SelectorManager selMgr ) {
                    initCtx.selMgr = selMgr;
                }
            }
        );

        initCtx.awaitTestObject(
            MingleBindTests.KEY_BINDER,
            MingleBinder.class,
            new ObjectReceiver< MingleBinder >() {
                public void receive( MingleBinder mb ) {
                    initCtx.mb = mb;
                }
            }
        );

        initCtx.awaitTestObject(
            MingleHttpTesting.KEY_SERVER_CODEC_FACTORY,
            MingleHttpCodecFactory.class,
            new ObjectReceiver< MingleHttpCodecFactory >() {
                public void receive( MingleHttpCodecFactory codecFact ) {
                    initCtx.codecFact = codecFact;
                }
            }
        );

        Testing.awaitTestObject(
            ctx,
            MingleHttpTesting.KEY_SERVER_LOCATIONS,
            Object.class,
            initCtx.wrap(
                new ObjectReceiver< Object >() {
                    public void receive( Object srvLocs ) {
                        initCtx.srvLocs = Lang.castUnchecked( srvLocs );
                    }
                }
            )
        );

        NetSslTests.awaitTransportFactory(
            initCtx.ctx,
            false,
            initCtx.wrap(
                new ObjectReceiver< SslIoExecutorTransportFactory >() {
                    public void receive( SslIoExecutorTransportFactory f )
                    {
                        initCtx.sslTransFact = f;
                    }
                }
            )
        );
    }

    // To test extras
    //
    //  - bg failure in process rpc to endpoint that comes back as mingle
    //  internal failure (internal failure created by MingleHttpResponder)
    //
    //  - failure in serializing mingle service response comes back as http 500
    //  error
}
