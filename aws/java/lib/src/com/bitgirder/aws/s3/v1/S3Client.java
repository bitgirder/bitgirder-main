package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.StandardThread;

import com.bitgirder.io.ProtocolProcessor;
import com.bitgirder.io.ByteBufferAccumulator;

import com.bitgirder.net.SelectorManager;
import com.bitgirder.net.SelectableChannelManager;
import com.bitgirder.net.NetProtocolTransportFactory;
import com.bitgirder.net.Networking;

import com.bitgirder.net.ssl.SslIoExecutors;

import com.bitgirder.http.HttpClient;
import com.bitgirder.http.HttpClientConnection;
import com.bitgirder.http.HttpProtocolProcessors;
import com.bitgirder.http.HttpMethod;
import com.bitgirder.http.HttpOperations;
import com.bitgirder.http.HttpMessages;
import com.bitgirder.http.HttpRequestMessage;
import com.bitgirder.http.HttpResponseMessage;

import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.ProcessRpcServer;
import com.bitgirder.process.Stoppable;

import java.lang.reflect.Field;

import java.net.InetSocketAddress;

import java.nio.ByteBuffer;

public
final
class S3Client
extends AbstractVoidProcess
implements Stoppable
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static InetSocketAddress DEFAULT_HTTP_ADDRESS =
        InetSocketAddress.createUnresolved( "s3.amazonaws.com", 80 );

    private final static InetSocketAddress DEFAULT_HTTPS_ADDRESS =
        InetSocketAddress.createUnresolved( "s3.amazonaws.com", 443 );

    private final static int DEFAULT_MAX_CONCURRENT_CONNS = 5;

    private final SelectorManager selMgr;
    private InetSocketAddress httpAddr;
    private InetSocketAddress httpsAddr;
    private final NetProtocolTransportFactory sslTransportFact;
    private final S3HttpRequestFactory reqFact;
    private final int maxConcurrentConns;

    private HttpClient httpCli;

    private
    S3Client( Builder b )
    {
        super( 
            ProcessRpcServer.createStandard(),
            SelectableChannelManager.create()
        );

        this.selMgr = inputs.notNull( b.selMgr, "selMgr" );
        this.httpAddr = b.httpAddr;
        this.httpsAddr = b.httpsAddr;

        this.sslTransportFact = 
            inputs.notNull( b.sslTransportFact, "sslTransportFact" );
        
        this.reqFact = inputs.notNull( b.reqFact, "reqFact" );
        this.maxConcurrentConns = b.maxConcurrentConns;
    }

    public void stop() { exit(); }

    private
    boolean
    addrsResolved()
    {
        return ! ( httpAddr.isUnresolved() || httpsAddr.isUnresolved() );
    }

    private
    void
    resolveAddrs()
    {
        holdRpcRequests();

        new StandardThread( "s3-cli-addr-resolver-%1$d" ) 
        {
            public void run() 
            {
                try 
                { 
                    httpAddr = Networking.ensureResolved( httpAddr );
                    httpsAddr = Networking.ensureResolved( httpsAddr );

                    submit(
                        new AbstractTask() {
                            protected void runImpl() { resumeRpcRequests(); }
                        }
                    );
                }
                catch ( Throwable th ) { fail( th ); }
            }
        }.
        start();
    }

    protected
    void
    startImpl()
    {
        httpCli = 
            new HttpClient.Builder().
                setNetworking( selMgr ).
                setActivityContext( getActivityContext() ).
                setSslTransportFactory( sslTransportFact ).
                setMaxConcurrentConnections( maxConcurrentConns ).
                build();
 
        if ( ! addrsResolved() ) resolveAddrs();
    }

    private
    InetSocketAddress
    getSocketAddress( S3Request req )
    {
        return req.isSsl() ? httpsAddr : httpAddr;
    }

    private
    abstract
    class S3RequestResponder< Q extends S3Request< ? >, R extends S3Response >
    extends ProcessRpcServer.AbstractAsyncResponder< R >
    {
        private final Class< Q > reqCls;
        private final Class< R > respCls;

        private Q req;

        private
        S3RequestResponder( Class< Q > reqCls,
                            Class< R > respCls )
        {
            this.reqCls = reqCls;
            this.respCls = respCls;
        }

        final
        Q
        request()
        {
            if ( req == null )
            {
                try
                {
                    Field f = getClass().getDeclaredField( "req" );
                    f.setAccessible( true );
                    req = reqCls.cast( f.get( this ) );
                }
                catch ( Throwable th ) 
                { 
                    throw new RuntimeException( 
                        "Couldn't find or access request field", th );
                }
            }

            return req;
        }

        abstract
        R
        makeResponse( HttpResponseMessage msg,
                      Object bodyObj )
            throws Exception;
        
        // overridable
        boolean
        isSuccess( HttpResponseMessage msg )
        {
            return msg.statusCodeValue() == 200;
        }

        // overridable
        Exception
        getHttpException( HttpResponseMessage msg )
        {
            return new S3HttpException( msg.getStatus() );
        }

        private
        final
        class RequestOp
        extends HttpOperations.AbstractRequest
        {
            private final InetSocketAddress addr;
            
            private Runnable onComp;

            private
            RequestOp( InetSocketAddress addr )
            {
                super( getActivityContext() );

                this.addr = addr;
            }

            protected
            ProtocolProcessor< ByteBuffer >
            createSend()
                throws Exception
            {
                String host = addr.getHostName();
                int port = addr.getPort();

                HttpRequestMessage msg = 
                    reqFact.httpMessageFor( request(), host, port );

                HttpProtocolProcessors.RequestSendBuilder b = sendBuilder();
                b.setMessage( msg );

                ProtocolProcessor< ByteBuffer > body = request().getBody();
                if ( body != null ) b.setBody( body );

                return b.build();
            }

            private
            boolean
            responseIsError()
            {
                return response().statusCodeValue() >= 400;
            }

            private
            boolean
            isXmlErrorResponse()
            {
                CharSequence ctyp = response().h().getContentTypeString();

                return 
                    ctyp != null && ctyp.toString().equals( "application/xml" );
            }

            private
            ProtocolProcessor< ByteBuffer >
            createXmlErrorReceiverNoBody()
            {
                onComp =
                    new AbstractTask() {
                        protected void runImpl() {
                            fail( getHttpException( response() ) );
                        }
                    };
                
                return null;
            }

            private
            ProtocolProcessor< ByteBuffer >
            createXmlErrorReceiverWithBody()
            {
                final S3RemoteExceptionReceiver proc = 
                    new S3RemoteExceptionReceiver( request(), response().h() );

                onComp = 
                    new AbstractTask() {
                        protected void runImpl() throws Exception {
                            fail( proc.getException() );
                        }
                    };

                return asBodyReceiver( proc );
            }

            private
            ProtocolProcessor< ByteBuffer >
            createXmlErrorReceiver()
            {
                if ( request().getHttpMethod().equals( HttpMethod.HEAD ) )
                {
                    return createXmlErrorReceiverNoBody();
                }
                else return createXmlErrorReceiverWithBody();
            }

            private
            ProtocolProcessor< ByteBuffer >
            createErrorResponseReceiver()
            {
                if ( isXmlErrorResponse() ) return createXmlErrorReceiver();
                else throw new UnsupportedOperationException( "Unimplemented" );
            }

            private
            ProtocolProcessor< ByteBuffer >
            createSuccessReceiver()
                throws Exception
            {
                final S3Request.BodyHandler< ? > bh = req.getBodyHandler();

                onComp =
                    new AbstractTask() {
                        protected void runImpl() throws Exception 
                        { 
                            Object bodyObj = 
                                bh == null ? null : bh.getCompletionObject();

                            respond( makeResponse( response(), bodyObj ) ); 
                        }
                    };

                ProtocolProcessor< ByteBuffer > res = 
                    bh == null ? null : bh.getBodyReceiver( response() );
                
                return res == null ? null : asBodyReceiver( res );
            }

            @Override
            protected
            ProtocolProcessor< ByteBuffer >
            getResponseBodyReceiver()
                throws Exception
            {
                if ( isSuccess( response() ) ) return createSuccessReceiver();
                else return createErrorResponseReceiver();
            }

            @Override 
            protected 
            void 
            requestComplete( HttpClientConnection conn ) 
            { 
                conn.close(); 
                onComp.run();
            }
        }
 
        protected
        final
        void
        startImpl()
            throws Exception
        {
            request().init( getActivityContext() );

            final InetSocketAddress addr = getSocketAddress( request() );

            httpCli.connect( 
                addr,
                request().isSsl(),
                new HttpClient.AbstractConnectHandler( getActivityContext() ) 
                {
                    @Override 
                    protected 
                    void 
                    connectSucceededImpl( HttpClientConnection conn )
                    {
                        conn.begin( new RequestOp( addr ) );
                    }
                }
            );
        }
    }

    private
    abstract
    class S3ObjectRequestResponder< Q extends S3ObjectRequest,
                                    R extends S3ObjectResponse< ? > >
    extends S3RequestResponder< Q, R >
    {
        private
        S3ObjectRequestResponder( Class< Q > reqCls,
                                  Class< R > respCls )
        {
            super( reqCls, respCls );
        }

        final
        < B extends S3ObjectResponseInfo.AbstractBuilder< B > >
        B
        initObjInfo( HttpResponseMessage msg,
                     B b )
        {
            return S3Methods.initObjectResponseInfo( msg.h(), b, request() );
        }
    }

    @ProcessRpcServer.Responder
    private
    final
    class S3ListBucketResponder
    extends S3RequestResponder< S3ListBucketRequest, S3ListBucketResponse >
    {
        private S3ListBucketRequest req;

        private
        S3ListBucketResponder()
        {
            super( S3ListBucketRequest.class, S3ListBucketResponse.class );
        }

        S3ListBucketResponse
        makeResponse( HttpResponseMessage msg,
                      Object bodyObj )
        {
            S3BucketResponseInfo.Builder b = 
                new S3BucketResponseInfo.Builder();

            S3Methods.initBucketResponseInfo( msg.h(), b, req );

            return new S3ListBucketResponse( bodyObj, b.build(), msg.h() );
        }
    }

    @ProcessRpcServer.Responder
    private
    final
    class S3ObjectHeadResponder
    extends S3ObjectRequestResponder< S3ObjectHeadRequest, 
                                      S3ObjectHeadResponse >
    {
        private S3ObjectHeadRequest req;

        private
        S3ObjectHeadResponder()
        {
            super( S3ObjectHeadRequest.class, S3ObjectHeadResponse.class );
        }

        @Override
        Exception
        getHttpException( HttpResponseMessage msg )
        {
            if ( msg.statusCodeValue() == 404 )
            {
                return
                    new NoSuchS3ObjectException.Builder().
                        setKey( req.location().key() ).
                        setBucket( req.location().bucket() ).
                        build();
            }
            else return super.getHttpException( msg );
        }

        S3ObjectHeadResponse
        makeResponse( HttpResponseMessage msg,
                      Object bodyObj )
        {
            S3HeadObjectResponseInfo info =
                initObjInfo( msg, new S3HeadObjectResponseInfo.Builder() ).
                build();

            return new S3ObjectHeadResponse( info, msg.h() );
        }
    }

    @ProcessRpcServer.Responder
    private
    final
    class S3ObjectGetResponder
    extends S3ObjectRequestResponder< S3ObjectGetRequest, S3ObjectGetResponse >
    {
        private S3ObjectGetRequest req;

        private
        S3ObjectGetResponder()
        {
            super( S3ObjectGetRequest.class, S3ObjectGetResponse.class );
        }

        S3ObjectGetResponse
        makeResponse( HttpResponseMessage msg,
                      Object bodyObj )
        {
            S3GetObjectResponseInfo info =
                initObjInfo( msg, new S3GetObjectResponseInfo.Builder() ).
                build();

            return new S3ObjectGetResponse( bodyObj, info, msg.h() );
        }
    }

    @ProcessRpcServer.Responder
    private
    final
    class S3ObjectPutResponder
    extends S3ObjectRequestResponder< S3ObjectPutRequest, S3ObjectPutResponse >
    {
        private S3ObjectPutRequest req;
 
        private
        S3ObjectPutResponder()
        {
            super( S3ObjectPutRequest.class, S3ObjectPutResponse.class );
        }

        S3ObjectPutResponse
        makeResponse( HttpResponseMessage msg,
                      Object bodyObj )
        {
            S3PutObjectResponseInfo info =
                initObjInfo( msg, new S3PutObjectResponseInfo.Builder() ).
                setEtag( msg.h().expectEtagString().toString() ).
                build();

            return new S3ObjectPutResponse( info, msg.h() );
        }
    }

    @ProcessRpcServer.Responder
    private
    final
    class S3ObjectDeleteResponder
    extends S3ObjectRequestResponder< S3ObjectDeleteRequest, 
                                      S3ObjectDeleteResponse >
    {
        private S3ObjectDeleteRequest req;

        private
        S3ObjectDeleteResponder()
        {
            super( S3ObjectDeleteRequest.class, S3ObjectDeleteResponse.class );
        }

        @Override
        boolean
        isSuccess( HttpResponseMessage msg )
        {
            return msg.statusCodeValue() == 204;
        }

        S3ObjectDeleteResponse
        makeResponse( HttpResponseMessage msg,
                      Object bodyObj )
        {
            S3DeleteObjectResponseInfo info =
                initObjInfo( msg, new S3DeleteObjectResponseInfo.Builder() ).
                build();

            return new S3ObjectDeleteResponse( info, msg.h() );
        }
    }

    public
    final
    static
    class Builder
    {
        private SelectorManager selMgr;
        private InetSocketAddress httpAddr = DEFAULT_HTTP_ADDRESS;
        private InetSocketAddress httpsAddr = DEFAULT_HTTPS_ADDRESS;
        private NetProtocolTransportFactory sslTransportFact;
        private S3HttpRequestFactory reqFact;
        private int maxConcurrentConns = DEFAULT_MAX_CONCURRENT_CONNS;

        public
        Builder
        setNetworking( SelectorManager selMgr )
        {
            this.selMgr = inputs.notNull( selMgr, "selMgr" );
            return this;
        }

        public
        Builder
        setHttpAddress( InetSocketAddress httpAddr )
        {
            this.httpAddr = inputs.notNull( httpAddr, "httpAddr" );
            return this;
        }

        public
        Builder
        setHttpsAddress( InetSocketAddress httpsAddr )
        {
            this.httpsAddr = inputs.notNull( httpsAddr, "httpsAddr" );
            return this;
        }

        public
        Builder
        setSslTransportFactory( NetProtocolTransportFactory sslTransportFact )
        {
            this.sslTransportFact = 
                inputs.notNull( sslTransportFact, "sslTransportFact" );

            return this;
        }

        public
        Builder
        setUseDefaultSslTransport()
        {
            return setSslTransportFactory( 
                SslIoExecutors.createDefaultTransportFactory( true ) 
            );
        }

        public
        Builder
        setRequestFactory( S3HttpRequestFactory reqFact )
        {
            this.reqFact = inputs.notNull( reqFact, "reqFact" );
            return this;
        }

        public
        Builder
        setMaxConcurrentConnections( int maxConcurrentConns )
        {
            this.maxConcurrentConns =
                inputs.positiveI( maxConcurrentConns, "maxConcurrentConns" );
            
            return this;
        }

        public 
        S3Client 
        build() 
        { 
            if ( sslTransportFact == null ) setUseDefaultSslTransport();

            return new S3Client( this ); 
        }
    }
}
