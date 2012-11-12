package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.io.ProtocolProcessor;
import com.bitgirder.io.ProtocolProcessors;
import com.bitgirder.io.IoUtils;
import com.bitgirder.io.IoProcessor;
import com.bitgirder.io.FileWrapper;
import com.bitgirder.io.FileSend;

import com.bitgirder.process.ProcessActivity;

import com.bitgirder.http.HttpMethod;
import com.bitgirder.http.HttpRequestMessage;
import com.bitgirder.http.HttpContentMd5;
import com.bitgirder.http.HttpHeaders;
import com.bitgirder.http.HttpHeaderName;

import java.nio.ByteBuffer;

public
final
class S3ObjectPutRequest
extends S3ObjectRequest
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final BodyProducer bodyProd;
    private CharSequence ctype;
    private HttpContentMd5 md5;
    private final HttpHeaders metaHdrs;

    private
    S3ObjectPutRequest( Builder b )
    {
        super( b, HttpMethod.PUT );

        this.bodyProd = inputs.notNull( b.bodyProd, "bodyProd" );
        this.ctype = b.ctype;
        this.md5 = b.md5;
        this.metaHdrs = b.metaHdrs.build();
    }

    @Override
    void
    init( ProcessActivity.Context ctx )
        throws Exception
    {
        super.init( ctx );
        bodyProd.init( ctx );
    }

    @Override
    void
    addHeaders( HttpRequestMessage.Builder b )
    {
        super.addHeaders( b );

        b.h().setContentLength( bodyProd.getContentLength() );
    }

    @Override 
    ProtocolProcessor< ByteBuffer > getBody() { return bodyProd.getBody(); }

    @Override CharSequence getContentType() { return ctype; }
    @Override HttpContentMd5 getContentMd5() { return md5; }
    @Override HttpHeaders getMetaHeaders() { return metaHdrs; }

    public
    static
    interface BodyProducer
    {
        public
        void
        init( ProcessActivity.Context ctx )
            throws Exception;

        public
        ProtocolProcessor< ByteBuffer >
        getBody();
        
        public
        long
        getContentLength();
    }

    private
    final
    static
    class ByteBufferProducer
    implements BodyProducer
    {
        private final ByteBuffer bb;

        private ByteBufferProducer( ByteBuffer bb ) { this.bb = bb; }

        public void init( ProcessActivity.Context ctx ) {}

        public
        ProtocolProcessor< ByteBuffer >
        getBody()
        {
            return ProtocolProcessors.createBufferSend( bb );
        }

        public long getContentLength() { return bb.remaining(); }
    }

    private
    final
    static
    class FileSendProducer
    implements BodyProducer
    {
        private final FileWrapper fw;
        private final IoProcessor ioProc;

        private FileSend fs;
        private long clen;

        private
        FileSendProducer( FileWrapper fw,
                          IoProcessor ioProc )
        {
            this.fw = fw;
            this.ioProc = ioProc;
        }

        public
        void
        init( ProcessActivity.Context ctx )
            throws Exception
        {
            fs = 
                new FileSend.Builder().
                    setFile( fw ).
                    setClient( ioProc.createClient( ctx ) ).
                    build();

            clen = fw.getSize().getByteCount();
        }

        public ProtocolProcessor< ByteBuffer > getBody() { return fs; }
        public long getContentLength() { return clen; }
    }

    public
    final
    static
    class Builder
    extends S3ObjectRequest.Builder< Builder >
    {
        private BodyProducer bodyProd;
        private CharSequence ctype;
        private HttpContentMd5 md5;
        private HttpHeaders.Builder metaHdrs = HttpHeaders.newBuilder();

        public
        Builder
        setBody( BodyProducer bodyProd )
        {
            this.bodyProd = inputs.notNull( bodyProd, "bodyProd" );
            return castThis();
        }

        public
        Builder
        setBody( ByteBuffer data )
        {
            inputs.notNull( data, "data" );

            return setBody( new ByteBufferProducer( data ) );
        }

        public
        Builder
        setBody( FileWrapper fw,
                 IoProcessor ioProc )
        {
            return
                setBody(
                    new FileSendProducer(
                        inputs.notNull( fw, "fw" ),
                        inputs.notNull( ioProc, "ioProc" )
                    )
                );
        }

        public
        Builder
        setContentType( CharSequence ctype )
        {
            this.ctype = inputs.notNull( ctype, "ctype" );
            return this;
        }

        public
        Builder
        setContentMd5( HttpContentMd5 md5 )
        {
            this.md5 = inputs.notNull( md5, "md5" );
            return this;
        }

        public
        Builder
        setBase64ContentMd5( CharSequence md5 )
        {
            inputs.notNull( md5, "md5" );
            return setContentMd5( HttpContentMd5.forBase64String( md5 ) );
        }

        public
        Builder
        setContentMd5( ByteBuffer md5 )
        {
            inputs.notNull( md5, "md5" );
            return setContentMd5( HttpContentMd5.forBinaryMd5( md5 ) );
        }

        public
        Builder
        setContentMd5( byte[] md5 )
        {
            inputs.notNull( md5, "md5" );
            return setContentMd5( HttpContentMd5.forBinaryMd5( md5 ) );
        }

        private
        HttpHeaderName
        getMetaHeaderName( CharSequence nm )
        {
            String lcNm = nm.toString().toLowerCase();

            if ( lcNm.startsWith( "x-amz-meta-" ) )
            {
                return HttpHeaderName.forString( lcNm );
            }
            else
            {
                throw inputs.createFail(
                    "Meta headers should start with 'x-amz-meta-' " +
                    "(case-insensitive); got: '" + lcNm + "'"
                );
            }
        }

        public
        Builder
        addMetaField( CharSequence nm,
                      CharSequence val )
        {
            inputs.notNull( nm, "nm" );
            inputs.notNull( val, "val" );

            metaHdrs.addField( getMetaHeaderName( nm ), val );

            return this;
        }

        public
        Builder
        setMetaField( CharSequence nm,
                      CharSequence val )
        {
            inputs.notNull( nm, "nm" );
            inputs.notNull( val, "val" );

            metaHdrs.setField( getMetaHeaderName( nm ), val );

            return this;
        }

        public 
        S3ObjectPutRequest 
        build() 
        { 
            return new S3ObjectPutRequest( this ); 
        }
    }
}
