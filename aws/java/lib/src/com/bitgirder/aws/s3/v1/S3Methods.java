package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.http.HttpHeaders;
import com.bitgirder.http.HttpHeaderName;

import java.util.List;
import java.util.Map;

final
class S3Methods
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private S3Methods() {}

    static
    boolean
    isAmzMetaHeaderName( CharSequence str )
    {
        inputs.notNull( str, "str" );
                
        return 
            str.toString().regionMatches( 
                true, 0, S3Constants.X_AMZ_META_STRING, 0, 
                S3Constants.X_AMZ_META_STRING.length() );
    }

    private
    static
    String
    metaKeyFor( HttpHeaderName nm )
    {
        String pref = S3Constants.X_AMZ_META_STRING;

        String s = nm.toString().toLowerCase();

        return s.startsWith( pref ) ? s.substring( pref.length() ) : null;
    }

    private
    static
    void
    setMetaData( HttpHeaders hdrs,
                 S3ResponseInfo.AbstractBuilder< ? > b )
    {
        List< S3ObjectMetaData > res = Lang.newList();

        for ( Map.Entry< HttpHeaderName, List< CharSequence > > e :
                hdrs.entrySet() )
        {
            String metaKey = metaKeyFor( e.getKey() );

            if ( metaKey != null )
            {
                for ( CharSequence val : e.getValue() )
                {
                    res.add( 
                        new S3ObjectMetaData.Builder().
                            setKey( metaKey ).
                            setValue( val.toString() ).
                            build()
                    );
                }
            }
        }

        b.setMeta( res );
    }

    static
    < B extends S3ResponseInfo.AbstractBuilder< B > >
    B
    initResponseInfo( HttpHeaders hdrs,
                      B b )
    {
        state.notNull( hdrs, "hdrs" );
        state.notNull( b, "b" );
        
        b.setAmazonRequestId(
            hdrs.expectOne( S3Constants.HDR_AMZ_REQ_ID ).toString() );
    
        b.setAmazonId2( hdrs.expectOne( S3Constants.HDR_AMZ_ID_2 ).toString() );
        
        setMetaData( hdrs, b );

        return b;
    }

    static
    < B extends S3BucketResponseInfo.AbstractBuilder< B > >
    B
    initBucketResponseInfo( HttpHeaders hdrs,
                            B b,
                            S3BucketRequest< ? > req )
    {
        initResponseInfo( hdrs, b ); // null-checks hdrs, b
        state.notNull( req, "req" );
                    
        return b.setBucket( req.location().bucket() );
    }

    static
    < B extends S3ObjectResponseInfo.AbstractBuilder< B > >
    B
    initObjectResponseInfo( HttpHeaders hdrs,
                            B b,
                            S3ObjectRequest req )
    {
        initBucketResponseInfo( hdrs, b, req ); // does null checks
        return b.setKey( req.location().key() );
    }
}
