package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.http.HttpMethod;

public
final
class S3ObjectDeleteRequest
extends S3ObjectRequest
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private
    S3ObjectDeleteRequest( Builder b )
    {
        super( b, HttpMethod.DELETE ); 
    }

    public 
    static 
    class Builder 
    extends S3ObjectRequest.Builder< Builder > 
    {
        public
        S3ObjectDeleteRequest
        build()
        {
            return new S3ObjectDeleteRequest( this );
        }
    }
}
