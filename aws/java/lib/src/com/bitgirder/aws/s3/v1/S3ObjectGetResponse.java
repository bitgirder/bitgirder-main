package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.http.HttpHeaders;

public
final
class S3ObjectGetResponse
extends S3ObjectResponse< S3GetObjectResponseInfo >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Object bodyObj;

    S3ObjectGetResponse( Object bodyObj,
                         S3GetObjectResponseInfo info,
                         HttpHeaders hdrs )
    {
        super( info, hdrs );

        this.bodyObj = state.notNull( bodyObj, "bodyObj" );
    }

    public Object getBodyObject() { return bodyObj; }
}
