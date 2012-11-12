package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.http.HttpMethod;

public
abstract
class S3BucketRequest< L extends S3BucketLocation >
extends S3Request< L >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    S3BucketRequest( Builder< L, ? > b,
                     HttpMethod m )
    {
        super( state.notNull( b, "b" ), m );
    }

    CharSequence 
    getHttpResource() 
    { 
        return "/" + S3Bucket.fromString( location().bucket() ).asString(); 
    }

    public
    static
    abstract
    class Builder< L extends S3BucketLocation, B extends Builder< L, B > >
    extends S3Request.Builder< L, B >
    {}
}
