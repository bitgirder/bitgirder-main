package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.http.HttpHeaders;

public
final
class S3ObjectPutResponse
extends S3ObjectResponse< S3PutObjectResponseInfo >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    S3ObjectPutResponse( S3PutObjectResponseInfo info,
                         HttpHeaders hdrs )
    {
        super( info, hdrs );
    }
}
