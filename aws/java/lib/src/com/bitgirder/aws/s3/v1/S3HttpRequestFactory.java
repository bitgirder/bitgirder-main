package com.bitgirder.aws.s3.v1;

import com.bitgirder.http.HttpRequestMessage;

public
interface S3HttpRequestFactory
{
    public
    HttpRequestMessage
    httpMessageFor( S3Request req,
                    String host,
                    int port )
        throws Exception;
}
