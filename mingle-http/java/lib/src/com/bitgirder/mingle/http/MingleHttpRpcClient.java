package com.bitgirder.mingle.http;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.process.ProcessActivity;
import com.bitgirder.process.ProcessFailureTarget;

import com.bitgirder.net.Networking;
import com.bitgirder.net.SelectorManager;
import com.bitgirder.net.NetProtocolTransportFactory;

import com.bitgirder.io.ProtocolProcessor;
import com.bitgirder.io.GzipProcessor;
import com.bitgirder.io.IoUtils;
import com.bitgirder.io.DataSize;
import com.bitgirder.io.Charsets;
import com.bitgirder.io.ByteBufferAccumulator;

import com.bitgirder.http.HttpClient;
import com.bitgirder.http.HttpClientConnection;
import com.bitgirder.http.HttpConstants;
import com.bitgirder.http.HttpStatusCode;
import com.bitgirder.http.HttpOperations;
import com.bitgirder.http.HttpRequestMessage;
import com.bitgirder.http.HttpProtocolProcessors;
import com.bitgirder.http.HttpHeaders;

import com.bitgirder.mingle.model.MingleServiceRequest;
import com.bitgirder.mingle.model.MingleServiceResponse;
import com.bitgirder.mingle.model.MingleStruct;

import com.bitgirder.mingle.service.MingleServices;

import com.bitgirder.mingle.codec.MingleCodec;
import com.bitgirder.mingle.codec.MingleDecoder;
import com.bitgirder.mingle.codec.MingleCodecs;

import com.bitgirder.mingle.service.MingleRpcClient;

import java.net.InetSocketAddress;

import java.nio.ByteBuffer;

public
final
class MingleHttpRpcClient
extends ProcessActivity
implements MingleRpcClient
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static Reactor DEFAULT_REACTOR = new AbstractReactor() {};
    
    private final HttpClient cli;
    private final InetSocketAddress addr;
    private final CharSequence uri;
    private final boolean isSsl;
    private final MingleHttpCodecContext codecCtx;
    private final boolean useGzip;
    private final DataSize chunkSize = DataSize.ofBytes( 2048 );
    private final Reactor reactor;

    private
    HttpClient
    buildHttpClient( Builder b )
    {
        HttpClient.Builder cliBldr = 
            new HttpClient.Builder().
                setNetworking( inputs.notNull( b.networking, "networking" ) ).
                setActivityContext( getActivityContext() );
        
        if ( b.sslTransFact != null )
        {
            cliBldr.setSslTransportFactory( b.sslTransFact );
        }

        return cliBldr.build();
    }

    private
    MingleHttpRpcClient( Builder b )
    {
        super( b );

        cli = buildHttpClient( b );
        
        this.addr = inputs.notNull( b.addr, "addr" );
        this.uri = inputs.notNull( b.uri, "uri" );
        this.isSsl = b.sslTransFact != null;
        this.codecCtx = inputs.notNull( b.codecCtx, "codecCtx" );
        this.useGzip = b.useGzip;
        this.reactor = b.reactor;
    }

    private
    final
    class RpcContext
    implements ProcessFailureTarget
    {
        private final MingleServiceRequest req;
        private final MingleRpcClient.Handler h;

        private
        RpcContext( MingleServiceRequest req,
                    MingleRpcClient.Handler h )
        {
            this.req = req;
            this.h = h;
        }

        public
        void
        fail( Throwable th )
        {
            h.rpcFailed( th, req, MingleHttpRpcClient.this );
        }

        private
        void
        succeed( MingleServiceResponse resp )
        {
            h.rpcSucceeded( resp, req, MingleHttpRpcClient.this );
        }
    }

    private
    final
    class RequestOp
    extends HttpOperations.AbstractRequest
    {
        private final RpcContext rpcCtx;

        private ByteBufferAccumulator errAcc;
        private MingleDecoder< MingleStruct > respDec;

        private
        RequestOp( RpcContext rpcCtx )
        {
            super( getActivityContext( rpcCtx ) );

            this.rpcCtx = rpcCtx;
        }
        
        private
        HttpRequestMessage
        createRequestMessage()
            throws Exception
        {
            HttpRequestMessage.Builder b =
                requestBuilder().
                    setMethod( "POST" ).
                    setRequestUri( uri ).
                    h().setHost( addr.getHostName(), addr.getPort() ).
                    h().setConnection( "Close" ).
                    h().setContentType( codecCtx.contentType() ).
                    h().setField( 
                        HttpConstants.NAME_TRANSFER_ENCODING, "chunked" );
 
            if ( useGzip )
            {
                b.h().setContentEncoding( "gzip" );
                b.h().setAcceptEncoding( "gzip,identity" );
            }

            reactor.addRequestHeaders( b.h(), rpcCtx.req );

            return b.build();
        }

        private
        ProtocolProcessor< ByteBuffer >
        createBodySendProcessor()
            throws Exception
        {
            ProtocolProcessor< ByteBuffer > proc =
                MingleCodecs.createSendProcessor( 
                    codecCtx.codec(), 
                    MingleServices.asMingleStruct( rpcCtx.req ) 
                );

            if ( useGzip ) 
            {
                proc = 
                    GzipProcessor.create( proc, getActivityContext( rpcCtx ) );
            }

            return createChunkedSend( proc, chunkSize );
        }

        protected
        ProtocolProcessor< ByteBuffer >
        createSend()
            throws Exception
        {
            return 
                sendBuilder().
                setMessage( createRequestMessage() ).
                setBody( createBodySendProcessor() ).
                build();
        }

        private
        ProtocolProcessor< ByteBuffer >
        getSuccessReceiver()
            throws Exception
        {
//            code( "resp:", inspect( response() ) );
            respDec = codecCtx.codec().createDecoder( MingleStruct.class );

            return 
                asBodyReceiver( 
                    MingleCodecs.createReceiveProcessor( respDec ) );
        }

        private
        ProtocolProcessor< ByteBuffer >
        getFailureReceiver()
        {
            return 
                asBodyReceiver( errAcc = ByteBufferAccumulator.create( 256 ) );
        }

        @Override
        protected
        ProtocolProcessor< ByteBuffer >
        getResponseBodyReceiver()
            throws Exception
        {
            if ( response().getStatus().getCode() == HttpStatusCode.OK )
            {
                return getSuccessReceiver();
            }
            else return getFailureReceiver();
        }

        private
        void
        successComplete()
            throws Exception
        {
            rpcCtx.succeed( 
                MingleServices.asServiceResponse( respDec.getResult() ) );
        }

        private
        CharSequence
        getErrorAccMessage()
            throws Exception
        {
            // Assuming err is utf-8 for now
            CharSequence str =
                IoUtils.asString( 
                    errAcc.getBuffers(), Charsets.UTF_8.newDecoder() );
            
            return str.toString().trim();
        }

        private
        void
        errorComplete()
            throws Exception
        {
            rpcCtx.fail(
                state.createFail( 
                    "Got non-success http response, status:",
                    response().getStatus().getCode(),
                    "; message:",
                    getErrorAccMessage()
                )
            );
        }

        @Override
        protected
        void
        requestComplete( HttpClientConnection conn )
            throws Exception
        {
            if ( errAcc == null ) successComplete();
            else errorComplete();
            
            conn.close();
        }
    }

    private
    final
    class HandlerImpl
    extends HttpClient.AbstractConnectHandler
    {
        private final RpcContext rpcCtx;

        private
        HandlerImpl( RpcContext rpcCtx )
        {
            super( rpcCtx );
            this.rpcCtx = rpcCtx;
        }

        @Override
        protected
        void
        connectSucceededImpl( HttpClientConnection conn )
        {
            conn.begin( new RequestOp( rpcCtx ) );
        }
    }

    public
    void
    beginRpc( MingleServiceRequest req,
              Handler h )
    {
        inputs.notNull( req, "req" );
        inputs.notNull( h, "h" );
 
        cli.connect( addr, isSsl, new HandlerImpl( new RpcContext( req, h ) ) );
    }

    public
    static
    interface Reactor
    {
        public
        void
        addRequestHeaders( HttpHeaders.Builder< ? > b,
                           MingleServiceRequest req )
            throws Exception;
    }

    public
    static
    abstract
    class AbstractReactor
    implements Reactor
    {
        public
        void
        addRequestHeaders( HttpHeaders.Builder< ? > b,
                           MingleServiceRequest req )
            throws Exception
        {}
    }

    public
    final
    static
    class Builder
    extends ProcessActivity.Builder< Builder >
    {
        private SelectorManager networking;
        private InetSocketAddress addr;
        private CharSequence uri;
        private NetProtocolTransportFactory sslTransFact;
        private MingleHttpCodecContext codecCtx;
        private boolean useGzip;
        private Reactor reactor = DEFAULT_REACTOR;

        public
        Builder
        setNetworking( SelectorManager networking )
        {
            this.networking = inputs.notNull( networking, "networking" );
            return this;
        }

        public
        Builder
        setAddress( InetSocketAddress addr )
        {
            this.addr = inputs.notNull( addr, "addr" );
            return this;
        }

        public
        Builder
        setAddress( CharSequence host,
                    int port )
        {
            inputs.notNull( host, "host" );

            return setAddress( Networking.createSocketAddress( host, port ) );
        }

        public
        Builder
        setUri( CharSequence uri )
        {
            this.uri = inputs.notNull( uri, "uri" );
            return this;
        }

        public
        Builder
        setCodecContext( MingleHttpCodecContext codecCtx )
        {
            this.codecCtx = inputs.notNull( codecCtx, "codecCtx" );
            return this;
        }

        public
        Builder
        setSslTransportFactory( NetProtocolTransportFactory sslTransFact )
        {
            this.sslTransFact = inputs.notNull( sslTransFact, "sslTransFact" );
            return this;
        }

        public
        Builder
        setUseGzip( boolean useGzip )
        {
            this.useGzip = useGzip;
            return this;
        }

        public
        Builder
        setReactor( Reactor reactor )
        {
            this.reactor = inputs.notNull( reactor, "reactor" );
            return this;
        }
        
        public
        MingleHttpRpcClient
        build()
        {
            return new MingleHttpRpcClient( this );
        }
    }
}
