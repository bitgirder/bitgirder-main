package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.ProtocolProcessor;
import com.bitgirder.io.ProtocolProcessors;
import com.bitgirder.io.FileWrapper;
import com.bitgirder.io.FileReceive;
import com.bitgirder.io.IoProcessor;
import com.bitgirder.io.ByteBufferAccumulator;

import com.bitgirder.process.ProcessActivity;

import com.bitgirder.http.HttpMethod;
import com.bitgirder.http.HttpDate;
import com.bitgirder.http.HttpRequestMessage;
import com.bitgirder.http.HttpResponseMessage;
import com.bitgirder.http.HttpContentMd5;
import com.bitgirder.http.HttpHeaders;

import com.bitgirder.xml.XmlDocumentProcessor;

import com.bitgirder.xml.bind.XmlBindingContext;

import java.nio.ByteBuffer;

public
abstract
class S3Request< L extends S3Location >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

//    private final static boolean DEFAULT_IS_SSL = true;

    private final static XmlBindingContext xbCtx =
        XmlBindingContext.create( "com.amazonaws.s3.doc._2006_03_01" );

//    private final boolean isSsl;
    private final L loc;
    private final HttpMethod method;

    // Will let callers set later
    private final HttpDate date = HttpDate.now();
    
    private final BodyHandler< ? > bh; // maybe null

    S3Request( Builder< L, ? > b,
               HttpMethod method )
    {
        state.notNull( b, "b" );

        this.loc = inputs.notNull( b.loc, "loc" );
        this.method = state.notNull( method, "method" );
        this.bh = b.bh;
    }

    final boolean isSsl() { return loc.useSsl(); }
    final L location() { return loc; }
    final HttpMethod getHttpMethod() { return method; }
    final HttpDate getDate() { return date; }
    final BodyHandler< ? > getBodyHandler() { return bh; }

    abstract
    CharSequence
    getHttpResource();

    // overridable as needed, but subclasses should chain overrides
    void 
    init( ProcessActivity.Context ctx ) 
        throws Exception 
    {
        if ( bh != null ) bh.init( ctx.getActivityContext() );
    }

    // overriders should chain calls to super.addHeaders first
    void addHeaders( HttpRequestMessage.Builder b ) {}
    
    CharSequence getResourceToSign() { return getHttpResource(); }
    ProtocolProcessor< ByteBuffer > getBody() { return null; }
    CharSequence getContentType() { return null; }
    HttpContentMd5 getContentMd5() { return null; }
    HttpHeaders getMetaHeaders() { return null; }
 
    static
    interface BodyHandler< V >
    {
        public
        void
        init( ProcessActivity.Context ctx )
            throws Exception;

        public
        ProtocolProcessor< ByteBuffer >
        getBodyReceiver( HttpResponseMessage msg )
            throws Exception;
        
        public
        V
        getCompletionObject()
            throws Exception;
    }

    private
    final
    static
    class ByteBufferBodyHandler
    implements BodyHandler< ByteBuffer >
    {
        private ByteBuffer body;

        public void init( ProcessActivity.Context ctx ) {}

        public
        ProtocolProcessor< ByteBuffer >
        getBodyReceiver( HttpResponseMessage msg )
        {
            body = ByteBuffer.allocate( (int) msg.h().expectContentLength() );
            return ProtocolProcessors.createBufferReceive( body );
        }

        public ByteBuffer getCompletionObject() { return body; }
    }

    private
    final
    static
    class ByteBufferAccumulatorBodyHandler
    implements BodyHandler< Iterable< ByteBuffer > >
    {
        private final ByteBufferAccumulator acc =
            ByteBufferAccumulator.create( 8192 );
        
        public void init( ProcessActivity.Context ctx ) {}

        public
        ProtocolProcessor< ByteBuffer >
        getBodyReceiver( HttpResponseMessage msg )
        {
            return acc;
        }

        public
        Iterable< ByteBuffer >
        getCompletionObject() 
        { 
            return acc.getBuffers();
        }
    }

    private
    final
    static 
    class FileReceiveBodyHandler
    implements BodyHandler< FileWrapper >
    {
        private final FileWrapper recvTo;
        private final IoProcessor ioProc;

        private FileReceive recv;

        private
        FileReceiveBodyHandler( FileWrapper recvTo,
                                IoProcessor ioProc )
        {
            this.recvTo = recvTo;
            this.ioProc = ioProc;
        }

        public
        void
        init( ProcessActivity.Context ctx )
            throws Exception
        {
            recv =
                new FileReceive.Builder().
                    setClient( ioProc.createClient( ctx ) ).
                    setFile( recvTo ).
                    build();
        }

        public
        ProtocolProcessor< ByteBuffer >
        getBodyReceiver( HttpResponseMessage msg )
        {
            return recv;
        }

        public FileWrapper getCompletionObject() { return recvTo; }
    }

    private
    final
    static
    class XmlBindResponseHandler< V >
    implements BodyHandler< V >
    {
        private final Class< V > cls;

        private final XmlDocumentProcessor proc = XmlDocumentProcessor.create();

        private XmlBindResponseHandler( Class< V > cls ) { this.cls = cls; }

        public void init( ProcessActivity.Context ctx ) {}

        public
        ProtocolProcessor< ByteBuffer >
        getBodyReceiver( HttpResponseMessage msg )
        {
            return proc;
        }

        public
        V
        getCompletionObject()
        {
            return xbCtx.fromDocument( proc.getDocument(), cls );
        }
    }

    public
    abstract
    static
    class Builder< L extends S3Location, B extends Builder< L, B > >
    {
//        private boolean isSsl = DEFAULT_IS_SSL;
        private L loc;
        private BodyHandler< ? > bh;

        final
        B
        castThis()
        {
            @SuppressWarnings( "unchecked" )
            B res = (B) this;

            return res;
        }

        final boolean isBodyHandlerSet() { return bh != null; }

        final
        B
        setLocation( L loc )
        {
            this.loc = inputs.notNull( loc, "loc" );
            return castThis();
        }

        final
        B
        setBodyHandler( BodyHandler< ? > bh )
        {
            state.isTrue( this.bh == null, "A body handler is already set" );
            this.bh = state.notNull( bh, "bh" );
            
            return castThis();
        }

        final
        B
        setByteBufferBodyHandler()
        {
            return setBodyHandler( new ByteBufferBodyHandler() );
        }

        final
        B
        setAccumulateBody()
        {
            return setBodyHandler( new ByteBufferAccumulatorBodyHandler() );
        }

        // does input checking on behalf of public frontends
        final
        B
        setFileReceiveBodyHandler( FileWrapper recvTo,
                                   IoProcessor ioProc )
        {
            inputs.notNull( recvTo, "recvTo" );
            inputs.notNull( ioProc, "ioProc" );

            return
                setBodyHandler( new FileReceiveBodyHandler( recvTo, ioProc ) );
        }

        final
        < V >
        B
        setReceiveBoundXml( Class< V > cls )
        {
            inputs.notNull( cls, "cls" );
            return setBodyHandler( new XmlBindResponseHandler< V >( cls ) );
        }

        final
        void
        assertBodyHandlerSet()
        {
            inputs.isFalse( bh == null, "No body handler is set" );
        }
    }
}
