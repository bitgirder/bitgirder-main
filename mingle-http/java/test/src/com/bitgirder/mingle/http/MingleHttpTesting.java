package com.bitgirder.mingle.http;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.ObjectReceiver;

import com.bitgirder.lang.reflect.ReflectUtils;

import com.bitgirder.io.IoUtils;

import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.ProcessRpcServer;
import com.bitgirder.process.ProcessActivity;

import com.bitgirder.http.HttpHeaders;

import com.bitgirder.mingle.model.MingleServiceRequest;
import com.bitgirder.mingle.model.MingleIdentifiedName;

import com.bitgirder.mingle.service.MingleRpcClient;

import com.bitgirder.mingle.tck.v1.MingleTckTests;

import com.bitgirder.net.NetTests;

import com.bitgirder.net.ssl.NetSslTests;

import com.bitgirder.test.TestRuntime;

import com.bitgirder.testing.Testing;

import java.util.List;

import java.util.concurrent.ConcurrentMap;

public
final
class MingleHttpTesting
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static Object KEY_BITGIRDER_HTTP_CLIENT_FACTORY =
        MingleHttpTesting.class.getName() + ".bitGirderHttpClientFactory";

    private final static Object KEY_CLIENT_CODEC_CONTEXTS =
        MingleHttpTesting.class.getName() + ".clientCodecContexts";

    public final static Object KEY_SERVER_CODEC_FACTORY =
        MingleHttpTesting.class.getName() + ".serverCodecFactory";

    public final static Object KEY_SERVER_LOCATIONS =
        MingleHttpTesting.class.getName() + ".serverLocations";

    private MingleHttpTesting() {}

    public
    static
    List< MingleHttpCodecContext >
    expectClientCodecContexts( TestRuntime rt )
    {
        inputs.notNull( rt, "rt" );

        return 
            Lang.castUnchecked(
                Testing.expectObject(
                    rt, KEY_CLIENT_CODEC_CONTEXTS, List.class ) );
    }

    public
    static
    MingleHttpCodecFactory
    expectServerCodecFactory( TestRuntime rt )
    {
        return 
            Testing.expectObject(
                inputs.notNull( rt, "rt" ),
                KEY_SERVER_CODEC_FACTORY,
                MingleHttpCodecFactory.class
            );
    }

    public
    final
    static
    class ServerLocation
    {
        private final String host;
        private final int port;
        private final String uri;
        private final boolean isSsl;

        public
        ServerLocation( String host,
                        Integer port,
                        String uri,
                        boolean isSsl )
        {
            this.host = inputs.notNull( host, "host" );
            this.port = inputs.notNull( port, "port" );
            this.uri = inputs.notNull( uri, "uri" );
            this.isSsl = isSsl;
        }

        public String host() { return host; }
        public int port() { return port; }
        public String uri() { return uri; }
        public boolean isSsl() { return isSsl; }
    }

    public
    static
    AbstractProcess< ? >
    expectBitGirderHttpClientFactory( TestRuntime rt )
    {
        return
            Testing.expectObject(
                inputs.notNull( rt, "rt" ),
                KEY_BITGIRDER_HTTP_CLIENT_FACTORY,
                AbstractProcess.class
            );
    }

    private
    final
    static
    class DebugFieldsReactor
    extends MingleHttpRpcClient.AbstractReactor
    {
        private final ProcessActivity.Context pCtx;

        private
        DebugFieldsReactor( ProcessActivity.Context pCtx )
        {
            this.pCtx = pCtx;
        }

        @Override
        public
        void
        addRequestHeaders( HttpHeaders.Builder< ? > b,
                           MingleServiceRequest req )
        {
            b.setField( "x-cli-pid", pCtx.getPid().toString() );
        }
    }

    private
    final
    static
    class BitGirderHttpClientFactory
    extends AbstractVoidProcess
    {
        private final TestRuntime rt;

        private
        BitGirderHttpClientFactory( TestRuntime rt )
        {
            super( ProcessRpcServer.createStandard() );

            this.rt = rt;
        }

        private
        final
        class RpcClientFactoryImpl
        implements MingleTckTests.MingleRpcClientFactory
        {
            private final MingleHttpTckServiceTests.CreateRpcClientFactory req;

            private
            RpcClientFactoryImpl(
                MingleHttpTckServiceTests.CreateRpcClientFactory req )
            {
                this.req = req;
            }

            private
            void
            setSslTransport( MingleHttpRpcClient.Builder b )
            {
                b.setSslTransportFactory(
                    NetSslTests.createTransportFactory( true, rt ) );
            }

            public
            MingleRpcClient
            createRpcClient( ProcessActivity.Context ctx )
            {
                ServerLocation loc = req.testLocation();

                MingleHttpRpcClient.Builder b =
                    new MingleHttpRpcClient.Builder().
                        setActivityContext( ctx ).
                        setNetworking( NetTests.expectSelectorManager( rt ) ).
                        setAddress( loc.host(), loc.port() ).
                        setUri( loc.uri() ).
                        setReactor( new DebugFieldsReactor( ctx ) ).
                        setCodecContext( req.codecContext() );
                
                if ( loc.isSsl() ) setSslTransport( b );
                if ( req.useGzip() ) b.setUseGzip( true );
                
                return b.build();
            }
        }

        @ProcessRpcServer.Responder
        private
        MingleTckTests.MingleRpcClientFactory
        handle( MingleHttpTckServiceTests.CreateRpcClientFactory req )
        {
            return new RpcClientFactoryImpl( req );
        }

        protected void startImpl() {}
    }

    @Testing.RuntimeInitializer
    private
    static
    void
    initHttpClientFactory( Testing.RuntimeInitializerContext ctx )
    {
        ctx.spawnAndSetRpcStoppable( 
            new BitGirderHttpClientFactory( ctx.getRuntime() ),
            KEY_BITGIRDER_HTTP_CLIENT_FACTORY
        );

        ctx.complete();
    }

    private
    static
    void
    initClientContext( String initClsNm,
                       List< MingleHttpCodecContext > l )
        throws Exception
    {
        MingleHttpCodecFactories.ClientCodecInitializer init = 
            (MingleHttpCodecFactories.ClientCodecInitializer)
                ReflectUtils.newInstance( Class.forName( initClsNm ) );

        l.addAll( init.getClientCodecContexts() );
    }

    private
    static
    List< MingleHttpCodecContext >
    loadTestClientCodecContexts()
        throws Exception
    {
        final List< MingleHttpCodecContext > res = Lang.newList();

        IoUtils.visitResourceLines(
            "mingle-http-client-codec-init.txt",
            "utf-8",
            new IoUtils.AbstractResourceLineVisitor()
            {
                public void visitLine( String line ) throws Exception {
                    initClientContext( line, res );
                }
            }
        );

        return res;
    }

    @Testing.RuntimeInitializer
    private
    static
    void
    initClientCodecContexts( Testing.RuntimeInitializerContext ctx )
    {
        Testing.submitInitTask( ctx, new Testing.AbstractInitTask( ctx ) {
            protected void runImpl() throws Exception 
            {
                context().setObject(
                    KEY_CLIENT_CODEC_CONTEXTS,
                    loadTestClientCodecContexts()
                );

                context().complete();
            }
        });
    }

    private
    static
    MingleHttpCodecFactory
    loadServerCodecFactory()
        throws Exception
    {
        final MingleHttpCodecFactorySelector.Builder b =
            new MingleHttpCodecFactorySelector.Builder();
        
        IoUtils.visitResourceLines(
            "mingle-http-codec-selector-init.txt",
            "utf-8",
            new IoUtils.AbstractResourceLineVisitor() {
                public void visitLine( String line ) throws Exception 
                {
                    MingleHttpCodecFactories.FactorySelectorInitializer i =
                        (MingleHttpCodecFactories.FactorySelectorInitializer)
                            ReflectUtils.newInstance( Class.forName( line ) );
                    
                    i.addFactories( b );
                }
            }
        );

        return b.build();
    }

    @Testing.RuntimeInitializer
    private
    static
    void
    initServerCodecFactory( Testing.RuntimeInitializerContext ctx )
    {
        Testing.submitInitTask( ctx, new Testing.AbstractInitTask( ctx ) {
            protected void runImpl() throws Exception
            {
                context().setObject(
                    KEY_SERVER_CODEC_FACTORY, loadServerCodecFactory() );
                
                context().complete();
            }
        });
    }

    public
    static
    ConcurrentMap< MingleIdentifiedName, ServerLocation >
    expectServerLocations( TestRuntime rt )
    {
        return 
            Lang.castUnchecked(
                Testing.expectObject( rt, KEY_SERVER_LOCATIONS, Object.class )
            );
    }

    @Testing.RuntimeInitializer
    private
    static
    void
    initServerLocations( Testing.RuntimeInitializerContext ctx )
    {
        ConcurrentMap< MingleIdentifiedName, ServerLocation > m = 
            Lang.newConcurrentMap();

        ctx.setObject( KEY_SERVER_LOCATIONS, m );

        ctx.complete();
    }
}
