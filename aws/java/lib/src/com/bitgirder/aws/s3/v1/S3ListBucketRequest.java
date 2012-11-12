package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.Charsets;

import com.bitgirder.http.HttpMethod;
import com.bitgirder.http.HttpQueryStringBuilder;

import com.amazonaws.s3.doc._2006_03_01.ListBucketResult;

public
final
class S3ListBucketRequest
extends S3BucketRequest< S3BucketLocation >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();
    
    private S3ObjectKey prefix;
    private Integer maxKeys;
    private S3ObjectKey marker;
    private CharSequence delim;

    private
    S3ListBucketRequest( Builder b )
    {
        super( b, HttpMethod.GET );

        this.prefix = b.prefix;
        this.maxKeys = b.maxKeys;
        this.marker = b.marker;
        this.delim = b.delim;
    }

    private
    void
    putOptKey( HttpQueryStringBuilder qsb,
               String param,
               S3ObjectKey key )
    {
        if ( key != null ) qsb.putEncoded( param, key.asString( false ) );
    }

    @Override
    CharSequence
    getHttpResource()
    {
        CharSequence res = super.getHttpResource(); // start with base uri

        HttpQueryStringBuilder qsb = 
            HttpQueryStringBuilder.create( Charsets.UTF_8 );

        putOptKey( qsb, "prefix", prefix );
        putOptKey( qsb, "marker", marker );
        if ( maxKeys != null ) qsb.putEncoded( "max-keys", maxKeys.toString() );
        if ( delim != null ) qsb.put( "delimiter", delim );

        if ( qsb.map().isEmpty() ) return res;
        else
        {
            StringBuilder sb = new StringBuilder( res );
            return qsb.appendTo( sb, true );
        } 
    }

    @Override 
    CharSequence getResourceToSign() { return super.getHttpResource(); } 

    public
    final
    static
    class Builder
    extends S3BucketRequest.Builder< S3BucketLocation, Builder >
    {
        private S3ObjectKey prefix;
        private S3ObjectKey marker;
        private CharSequence delim;
        private Integer maxKeys;

        public
        Builder
        setPrefix( S3ObjectKey prefix )
        {
            this.prefix = inputs.notNull( prefix, "prefix" );
            return this;
        }

        public
        Builder
        setMarker( S3ObjectKey marker )
        {
            this.marker = inputs.notNull( marker, "marker" );
            return this;
        }

        public
        Builder
        setMaxKeys( int maxKeys )
        {
            this.maxKeys = inputs.positiveI( maxKeys, "maxKeys" );
            return this;
        }

        public
        Builder
        setDelimiter( CharSequence delim )
        {
            this.delim = inputs.notNull( delim, "delim" );
            return this;
        }

        public
        Builder
        setBindXmlResponse()
        {
            return setReceiveBoundXml( ListBucketResult.class );
        }

        public 
        S3ListBucketRequest 
        build() 
        { 
            if ( ! isBodyHandlerSet() ) setBindXmlResponse();
            return new S3ListBucketRequest( this );
        }
    }
}
