package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.http.HttpHeaders;

public
abstract
class S3ObjectResponse< R extends S3ObjectResponseInfo >
extends S3BucketResponse< R >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    S3ObjectResponse( R info,
                      HttpHeaders hdrs )
    {
        super( info, hdrs );
    }
}
