package com.bitgirder.jetty7;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.mingle.service.MingleRpcClient;

import com.bitgirder.mingle.model.MingleServiceRequest;
import com.bitgirder.mingle.model.MingleServiceResponse;
import com.bitgirder.mingle.model.MingleStruct;

import com.bitgirder.mingle.service.MingleServices;

import com.bitgirder.mingle.codec.MingleCodecs;

import com.bitgirder.mingle.http.MingleHttpCodecContext;

import com.bitgirder.process.ProcessRpcClient;
import com.bitgirder.process.ProcessActivity;

import java.nio.ByteBuffer;

import java.util.Map;
import java.util.List;

public
final
class Jetty7HttpMingleRpcClient
extends ProcessActivity
implements MingleRpcClient
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final String host;
    private final int port;
    private final String uri;
    private final MingleHttpCodecContext codecCtx;
    private final Jetty7HttpClient httpCli;
    private final Map< String, List< String > > headers;

    private
    Jetty7HttpMingleRpcClient( Builder b )
    {
        super( b );
        b.context().expectBehavior( ProcessRpcClient.class );

        this.host = inputs.notNull( b.host, "host" );
        this.port = inputs.positiveI( b.port, "port" );
        this.uri = inputs.notNull( b.uri, "uri" );
        this.codecCtx = inputs.notNull( b.codecCtx, "codecCtx" );
        this.httpCli = inputs.notNull( b.httpCli, "httpCli" );
        this.headers = Lang.unmodifiableDeepListMapCopy( b.headers );
    }

    private
    void
    failMgRpc( Handler h,
               Throwable th,
               MingleServiceRequest req )
    {
        h.rpcFailed( th, req, this );
    }

    private
    final
    class RpcHandlerImpl
    extends ProcessRpcClient.AbstractResponseHandler
    {
        private final MingleServiceRequest req;
        private final Handler h;

        private
        RpcHandlerImpl( MingleServiceRequest req,
                        Handler h )
        {
            this.req = req;
            this.h = h;
        }

        private
        void
        failMgRpc( Throwable th )
        {
            Jetty7HttpMingleRpcClient.this.failMgRpc( h, th, req );
        }

        @Override
        protected void rpcFailed( Throwable th ) { failMgRpc( th ); }

        private
        MingleServiceResponse
        asServiceResponse( ByteBuffer body )
            throws Exception
        {
            return
                MingleServices.asServiceResponse( 
                    MingleCodecs.fromByteBuffer( 
                        codecCtx.codec(), body, MingleStruct.class ) );
        }

        @Override
        protected
        void
        rpcSucceeded( Object respObj )
        {
            Jetty7HttpClient.BufferResponse resp = 
                (Jetty7HttpClient.BufferResponse) respObj;

            try
            {
                state.equalInt( 200, resp.getStatus() );
 
                MingleServiceResponse mgResp = 
                    asServiceResponse( resp.getBody() );
 
                h.rpcSucceeded( 
                    mgResp, req, Jetty7HttpMingleRpcClient.this );
            }
            catch ( Exception ex ) { failMgRpc( ex ); }
        }
    }

    private
    void
    copyHeaders( Jetty7HttpClient.BufferRequest.Builder b )
    {
        for ( Map.Entry< String, List< String > > e : headers.entrySet() )
        {
            for ( String val : e.getValue() )
            {
                b.addHeader( e.getKey(), val );
            }
        }
    }

    private
    Jetty7HttpClient.BufferRequest
    createBufferRequest( MingleServiceRequest req )
        throws Exception
    {
        Jetty7HttpClient.BufferRequest.Builder b =
            new Jetty7HttpClient.BufferRequest.Builder().
                setHost( host ).
                setPort( port ).
                setUri( uri ).
                setMethod( Jetty7HttpClient.METHOD_POST ).
                addHeader( "Content-Type", codecCtx.contentType().toString() ).
                setBody( 
                    MingleCodecs.toByteBuffer( 
                        codecCtx.codec(), 
                        MingleServices.asMingleStruct( req ) 
                    ) 
                );
 
        copyHeaders( b );

        return b.build();
    }

    public
    void
    beginRpc( MingleServiceRequest req,
              Handler h )
    {
        inputs.notNull( req, "req" );
        inputs.notNull( h, "h" );

        try
        {
            behavior( ProcessRpcClient.class ).
                beginRpc( 
                    httpCli, 
                    createBufferRequest( req ), 
                    new RpcHandlerImpl( req, h ) 
                );
        }
        catch ( Throwable th ) { failMgRpc( h, th, req ); }
    }

    public
    final
    static
    class Builder
    extends ProcessActivity.Builder< Builder >
    {
        private String host;
        private int port;
        private String uri;
        private MingleHttpCodecContext codecCtx;
        private Jetty7HttpClient httpCli;
        private Map< String, List< String > > headers = Lang.emptyMap();

        public
        Builder
        setHost( String host )
        {
            this.host = inputs.notNull( host, "host" );
            return this;
        }

        public
        Builder
        setPort( int port )
        {
            this.port = inputs.positiveI( port, "port" );
            return this;
        }

        public
        Builder
        setUri( String uri )
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
        setHttpClient( Jetty7HttpClient httpCli )
        {
            this.httpCli = inputs.notNull( httpCli, "httpCli" );
            return this;
        }

        public
        Builder
        setHeaders( Map< String, List< String > > headers )
        {
            this.headers = inputs.noneNull( headers, "headers" );

            for ( Map.Entry< String, List< String > > e : headers.entrySet() )
            {
                inputs.noneNull( 
                    e.getValue(), "headers[ " + e.getKey() + " ]" );
            }

            return this;
        }

        public
        Jetty7HttpMingleRpcClient
        build()
        {
            return new Jetty7HttpMingleRpcClient( this );
        }
    }
}
