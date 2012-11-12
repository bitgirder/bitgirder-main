package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.http.HttpHeaders;

public
final
class S3ObjectHeadResponse
extends S3ObjectResponse< S3HeadObjectResponseInfo >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    S3ObjectHeadResponse( S3HeadObjectResponseInfo info,
                          HttpHeaders hdrs )
    {
        super( info, hdrs );
    }
}
