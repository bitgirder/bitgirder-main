package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.io.Charsets;
import com.bitgirder.io.Base64Encoder;

import com.bitgirder.aws.AwsAccessKeyId;

import com.bitgirder.http.HttpRequestMessage;
import com.bitgirder.http.HttpHeaders;
import com.bitgirder.http.HttpHeaderName;
import com.bitgirder.http.HttpMethod;
import com.bitgirder.http.HttpContentMd5;

import com.bitgirder.crypto.CryptoUtils;

import java.util.List;
import java.util.Set;
import java.util.Map;
import java.util.SortedMap;

import java.nio.ByteBuffer;

import javax.crypto.Mac;
import javax.crypto.SecretKey;

public
final
class S3HttpRequestFactories
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static Base64Encoder enc = new Base64Encoder();

    private S3HttpRequestFactories() {}

    private
    final
    static
    class HttpRequestFactoryImpl
    implements S3HttpRequestFactory
    {
        private final AwsAccessKeyId accessKeyId;
        private final Mac mac;

        private
        HttpRequestFactoryImpl( AwsAccessKeyId accessKeyId,
                                Mac mac )
        {
            this.accessKeyId = accessKeyId;
            this.mac = mac;
        }

        private
        void
        setMethod( S3Request req,
                   HttpRequestMessage.Builder b,
                   StringBuilder toSign )
        {
            HttpMethod m = req.getHttpMethod();
            b.setMethod( m );
            toSign.append( m ).append( "\n" );
        }

        private
        void
        setContentMd5( S3Request req,
                       HttpRequestMessage.Builder b,
                       StringBuilder toSign )
        {
            HttpContentMd5 md5 = req.getContentMd5();

            if ( md5 != null )
            {
                b.h().setContentMd5( md5 );
                toSign.append( md5.getBase64Signature() );
            }

            toSign.append( "\n" );
        }

        private
        void
        setContentType( S3Request req,
                        HttpRequestMessage.Builder b,
                        StringBuilder toSign )
        {
            CharSequence ctype = req.getContentType();

            if ( ctype != null )
            {
                b.h().setContentType( ctype );
                toSign.append( ctype );
            }

            toSign.append( "\n" );
        }

        private
        void
        setDate( S3Request req,
                 HttpRequestMessage.Builder b,
                 StringBuilder toSign )
        {
            String dateStr = req.getDate().toRfc1123String();
            b.h().setDate( dateStr );
            toSign.append( dateStr ).append( "\n" );
        }

        // Not attempting to parse and fold multiline headers yet
        private
        SortedMap< String, CharSequence >
        sortAndFoldAmzHeaders( 
            Set< Map.Entry< HttpHeaderName, List< CharSequence > > > s )
        {
            SortedMap< String, CharSequence > res = Lang.newSortedMap();

            for ( Map.Entry< HttpHeaderName, List< CharSequence > > e : s )
            {
                res.put(
                    e.getKey().toString().toLowerCase(),
                    Strings.join( ",", e.getValue() )
                );
            }

            return res;
        }

        private
        void
        setAmzHeaders( S3Request req,
                       HttpRequestMessage.Builder b,
                       StringBuilder toSign )
        {
            HttpHeaders h = req.getMetaHeaders();

            if ( h != null )
            {
                SortedMap< String, CharSequence > folded =
                    sortAndFoldAmzHeaders( h.entrySet() );

                if ( ! folded.isEmpty() )
                {
                    for ( Map.Entry< String, CharSequence > e :
                            folded.entrySet() )
                    {
                        toSign.append( e.getKey() ).
                               append( ':' ).
                               append( e.getValue() ).
                               append( '\n' );
                        
                        b.h().setField( e.getKey(), e.getValue() );
                    }
                }
            }
        }

        private
        void
        setResource( S3Request req,
                     HttpRequestMessage.Builder b,
                     StringBuilder toSign )
        {
            b.setRequestUri( req.getHttpResource() );
            toSign.append( req.getResourceToSign() );
        }

        private
        void
        setSignature( HttpRequestMessage.Builder b,
                      StringBuilder toSign )
            throws Exception
        {
//            code( "toSign:", toSign );

            ByteBuffer bb = Charsets.UTF_8.asByteBuffer( toSign );
            ByteBuffer sig = CryptoUtils.sign( bb, mac );

            StringBuilder hdrVal = 
                new StringBuilder().
                    append( "AWS " ).
                    append( accessKeyId ).
                    append( ':' ).
                    append( enc.encode( sig ) );
            
            b.h().setField( "Authorization", hdrVal );
        }

        public
        HttpRequestMessage
        httpMessageFor( S3Request req,
                        String host,
                        int port )
            throws Exception
        {
            HttpRequestMessage.Builder b = new HttpRequestMessage.Builder();
            StringBuilder toSign = new StringBuilder();

            setMethod( req, b, toSign );
            setContentMd5( req, b, toSign );
            setContentType( req, b, toSign );
            setDate( req, b, toSign );
            setAmzHeaders( req, b, toSign );
            setResource( req, b, toSign );
            
            b.h().setHost( host, port );
            b.h().setField( "Connection", "Close" );
            setSignature( b, toSign );
            req.addHeaders( b );

            return b.build();
        }
    }

    public
    static
    S3HttpRequestFactory
    create( AwsAccessKeyId accessKeyId,
            SecretKey key )
    {
        inputs.notNull( accessKeyId, "accessKeyId" );
        inputs.notNull( key, "key" );

        Mac mac = CryptoUtils.expectMac( key, "HmacSha1" );

        return new HttpRequestFactoryImpl( accessKeyId, mac );
    }
}
