package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.http.HttpMethod;

import com.bitgirder.io.Charsets;

import java.nio.ByteBuffer;

public
abstract
class S3ObjectRequest
extends S3BucketRequest< S3ObjectLocation >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    S3ObjectRequest( Builder< ? > b,
                     HttpMethod m )
    {
        super( inputs.notNull( b, "b" ), m );
    }

    @Override
    CharSequence
    getHttpResource()
    {
        return 
            new StringBuilder().
                append( super.getHttpResource() ).
                append( S3ObjectKey.encodeAndCreate( location().key() ) );
    }

    public
    static
    abstract
    class Builder< B extends Builder< B > >
    extends S3BucketRequest.Builder< S3ObjectLocation, B >
    {}
}
