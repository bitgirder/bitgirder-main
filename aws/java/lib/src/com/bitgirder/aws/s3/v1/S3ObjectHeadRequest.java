package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.http.HttpMethod;

public
final
class S3ObjectHeadRequest
extends S3ObjectRequest
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private S3ObjectHeadRequest( Builder b ) { super( b, HttpMethod.HEAD ); }

    public
    final
    static
    class Builder
    extends S3ObjectRequest.Builder< Builder >
    {
        public 
        S3ObjectHeadRequest 
        build() 
        { 
            return new S3ObjectHeadRequest( this );
        }
    }
}
