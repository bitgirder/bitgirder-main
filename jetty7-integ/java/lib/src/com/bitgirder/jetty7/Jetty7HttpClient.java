package com.bitgirder.jetty7;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.http.HttpMethod;

import com.bitgirder.process.ProcessRpcServer;

import com.bitgirder.io.IoUtils;

import java.nio.ByteBuffer;

import java.io.IOException;

import java.util.Map;
import java.util.List;

import java.util.concurrent.TimeoutException;

import org.eclipse.jetty.io.Buffer;
import org.eclipse.jetty.io.ByteArrayBuffer;

import org.eclipse.jetty.io.nio.IndirectNIOBuffer;
import org.eclipse.jetty.io.nio.DirectNIOBuffer;

import org.eclipse.jetty.http.HttpFields;

import org.eclipse.jetty.client.HttpClient;
import org.eclipse.jetty.client.ContentExchange;
import org.eclipse.jetty.client.Address;

public
final
class Jetty7HttpClient
extends Jetty7LifeCycleManager< HttpClient >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public final static int DEFAULT_HTTP_PORT = 80;

    // referenced and exposed here so clients can directly access them without
    // having to also import or reference the com.bitgirder.http.HttpMethod
    // class itself
    public final static HttpMethod METHOD_OPTIONS = HttpMethod.OPTIONS;
    public final static HttpMethod METHOD_GET = HttpMethod.GET;
    public final static HttpMethod METHOD_HEAD = HttpMethod.HEAD;
    public final static HttpMethod METHOD_POST = HttpMethod.POST;
    public final static HttpMethod METHOD_PUT = HttpMethod.PUT;
    public final static HttpMethod METHOD_DELETE = HttpMethod.DELETE;
    public final static HttpMethod METHOD_TRACE = HttpMethod.TRACE;
    public final static HttpMethod METHOD_CONNECT = HttpMethod.CONNECT;

    private 
    Jetty7HttpClient( HttpClient cli ) 
    { 
        super( 
            new Builder< HttpClient, Builder< HttpClient, ? > >() {}.
                setLifeCycle( cli ).
                mixin( ProcessRpcServer.createStandard() )
        );
    }

    public HttpClient getClient() { return lifeCycle(); }

    public
    final
    static
    class BufferRequest
    {
        private final HttpMethod method;
        private final String host;
        private final int port;
        private final String uri;
        private final String scheme;
        private final Map< String, List< String > > headers;
        private final ByteBuffer body;

        private
        BufferRequest( Builder b )
        {
            this.method = inputs.notNull( b.method, "method" );
            this.uri = inputs.notNull( b.uri, "uri" );
            this.scheme = b.scheme;
            this.port = b.port;
            this.host = inputs.notNull( b.host, "host" );
            this.headers = Lang.unmodifiableDeepListMapCopy( b.headers );
            this.body = b.body;
        }

        public
        final
        static
        class Builder
        {
            private HttpMethod method;
            private String host;
            private int port = DEFAULT_HTTP_PORT;
            private String uri;
            private String scheme;
            private final Map< String, List< String > > headers = Lang.newMap();
            private ByteBuffer body;

            public
            Builder
            setMethod( HttpMethod method )
            {
                this.method = inputs.notNull( method, "method" );
                return this;
            }

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
            setScheme( String scheme )
            {
                this.scheme = inputs.notNull( scheme, "scheme" );
                return this;
            }

            public
            Builder
            setHeader( String hdr,
                       String value )
            {
                headers.put(
                    inputs.notNull( hdr, "hdr" ),
                    Lang.singletonList( inputs.notNull( value, "value" ) ) );
                
                return this;
            }

            public
            Builder
            addHeader( String hdr,
                       String value )
            {
                inputs.notNull( hdr, "hdr" );
                inputs.notNull( value, "value" );

                Lang.putAppend( headers, hdr, value );

                return this;
            }

            public
            Builder
            setBody( ByteBuffer body )
            {
                this.body = inputs.notNull( body, "body" );
                return this;
            }
            
            public BufferRequest build() { return new BufferRequest( this ); }
        }
    }

    public
    final
    static
    class BufferResponse
    {
        private final int status;
        private final HttpFields fields;
        private final ByteBuffer body;

        private
        BufferResponse( int status,
                        HttpFields fields,
                        ByteBuffer body )
        {
            this.status = status;
            this.fields = fields;
            this.body = body;
        }

        public int getStatus() { return status; }
        public HttpFields getFields() { return fields; }
        public ByteBuffer getBody() { return body; }

        @Override
        public
        String
        toString()
        {
            return Strings.inspect( this, true,
                "status", status,
                "fields", fields,
                "body", body 
            ).toString();
        }
    }

    private
    final
    class BufferExchange
    extends ContentExchange
    {
        private final BufferRequest req;
        private final ProcessRpcServer.ResponderContext< BufferResponse > ctx;
    
        private
        BufferExchange( 
            BufferRequest req,
            ProcessRpcServer.ResponderContext< BufferResponse > ctx )
        {
            super( true );

            this.req = req;
            this.ctx = ctx;
        }

        private
        void
        doSetScheme()
        {
            if ( req.scheme != null )
            {
                setScheme( new ByteArrayBuffer( req.scheme ) );
            }
        }
    
        private
        Buffer
        createRequestContent()
            throws Exception 
        {
            ByteBuffer bb = req.body;
    
            if ( bb.isDirect() ) return new DirectNIOBuffer( bb, true );
            else
            {
                // IndirectNIOBuffer attempts to access ByteBuffer.array() which
                // fails if bb is read only, so we have to make a copy in that
                // case.
                if ( bb.isReadOnly() )
                {
                    ByteBuffer copy = ByteBuffer.allocate( bb.remaining() );
                    copy.put( bb );
                    copy.flip();
    
                    return new IndirectNIOBuffer( copy, true );
                }
                else return new IndirectNIOBuffer( bb, true );
            }
        }

        private
        void
        setHeaders()
        {
            for ( Map.Entry< String, List < String > > e : 
                    req.headers.entrySet() )
            {
                for ( String s : e.getValue() )
                {
                    addRequestHeader( e.getKey(), s );
                }
            }
        }

        private
        void
        initProperties()
            throws Exception
        {
            setAddress( new Address( req.host, req.port ) );
            setURI( req.uri );
            doSetScheme();
            setRequestContent( createRequestContent() );
            setMethod( req.method.toString() );
            setHeaders();
        }

        private void failExchange( Throwable th ) { ctx.fail( th ); }

        @Override
        protected 
        void 
        onConnectionFailed( Throwable th ) 
        { 
            failExchange( th ); 
        }
    
        @Override 
        protected void onException( Throwable th ) { failExchange( th ); }
    
        @Override
        protected void onExpire() { failExchange( new TimeoutException() ); }
    
        @Override
        protected
        void
        onResponseComplete()
            throws IOException
        {
            byte[] bodyArr = getResponseContentBytes();

            ByteBuffer body = 
                bodyArr == null 
                    ? IoUtils.emptyByteBuffer() : ByteBuffer.wrap( bodyArr );

            BufferResponse resp =
                new BufferResponse(
                    getResponseStatus(), getResponseFields(), body );

            ctx.respond( resp );
        }
        
        private
        void
        begin()
            throws Exception
        {
            initProperties();
            getClient().send( this );
        }
    }

    @ProcessRpcServer.Responder
    private
    void
    handle( BufferRequest req,
            ProcessRpcServer.ResponderContext< BufferResponse > ctx )
        throws Exception
    {
        new BufferExchange( req, ctx ).begin();
    }

    public
    static
    Jetty7HttpClient
    create()
    {
        HttpClient cli = new HttpClient();
        cli.setConnectorType( HttpClient.CONNECTOR_SELECT_CHANNEL );

        return new Jetty7HttpClient( cli );
    }
}
