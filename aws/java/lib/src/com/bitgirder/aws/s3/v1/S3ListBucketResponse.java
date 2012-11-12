package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.http.HttpHeaders;

public
final
class S3ListBucketResponse
extends S3BucketResponse< S3BucketResponseInfo >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Object resObj;

    S3ListBucketResponse( Object resObj,
                          S3BucketResponseInfo info,
                          HttpHeaders hdrs )
    {
        super( info, hdrs );
        this.resObj = inputs.notNull( resObj, "resObj" );
    }

    public Object getResultObject() { return resObj; }
}
