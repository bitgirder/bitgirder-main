package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.http.HttpHeaders;

public
abstract
class S3BucketResponse< R extends S3BucketResponseInfo >
extends S3Response< R >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    S3BucketResponse( R info,
                      HttpHeaders hdrs )
    {
        super( info, hdrs );
    }
}
