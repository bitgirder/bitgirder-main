package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.AbstractDelegateProcessor;
import com.bitgirder.io.ByteBufferAccumulator;

import com.bitgirder.http.HttpHeaders;

import com.bitgirder.xml.Xpaths;
import com.bitgirder.xml.XmlIo;

import java.nio.ByteBuffer;

import org.w3c.dom.Document;

final
class S3RemoteExceptionReceiver
extends AbstractDelegateProcessor< ByteBuffer, ByteBufferAccumulator >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final S3Request req;
    private final HttpHeaders respHdrs;

    S3RemoteExceptionReceiver( S3Request req,
                               HttpHeaders respHdrs )
    {
        super( ByteBufferAccumulator.create( 256 ) );

        this.req = state.notNull( req, "req" );
        this.respHdrs = state.notNull( respHdrs, "respHdrs" );
    }

    private
    Document
    parseDoc()
        throws Exception
    {
        return XmlIo.parseDocument( delegate().getBuffers() );
    }

    private
    S3AmazonRequestIds
    getAmazonIds( Document doc )
    {
        S3AmazonRequestIds.Builder b = new S3AmazonRequestIds.Builder();

        CharSequence reqId = respHdrs.getFirst( "x-amz-request-id" );
        if ( reqId != null ) b.setRequestId( reqId.toString() );

        CharSequence id2 = respHdrs.getFirst( "x-amz-id-2" );
        if ( id2 != null ) b.setId2( id2.toString() );

        return b.build();
    }

    private
    < B extends S3RemoteException.AbstractBuilder< B > >
    B
    init( B b,
          S3AmazonRequestIds amzIds )
    {
        return b.setAmazonRequestIds( amzIds );
    }

    private
    NoSuchS3ObjectException
    exceptionForNoSuchKey( S3AmazonRequestIds amzIds,
                           S3Request req )
    {
        S3ObjectLocation loc = ( (S3ObjectRequest) req ).location();

        return
            init( new NoSuchS3ObjectException.Builder(), amzIds ).
                setKey( loc.key() ).
                setBucket( loc.bucket() ).
                build();
    }

    private
    GenericS3RemoteException
    genericRemoteException( String code,
                            S3AmazonRequestIds amzIds,
                            Document doc )
        throws Exception
    {
        GenericS3RemoteException.Builder b =
            init( new GenericS3RemoteException.Builder(), amzIds );

        b.setCode( code );

        String msg = Xpaths.evaluate( "/Error/Message/text()", doc );
        if ( msg != null ) b.setMessage( msg );

        b.setErrorXml( XmlIo.toByteBuffer( doc ) );

        return b.build();
    }
 
    private
    Exception
    exceptionFor( String code,
                  S3AmazonRequestIds amzIds,
                  Document doc,
                  S3Request req )
        throws Exception
    {
        // tests are done in such a way that code may be null
        if ( "NoSuchKey".equals( code ) ) 
        {
            return exceptionForNoSuchKey( amzIds, req );
        }
        else return genericRemoteException( code, amzIds, doc );
    }

    public
    Exception
    getException()
        throws Exception
    {
        Document doc = parseDoc();

        String code = Xpaths.evaluate( "/Error/Code/text()", doc );
        S3AmazonRequestIds amzIds = getAmazonIds( doc );

        return exceptionFor( code, amzIds, doc, req );
    }
}
