package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.http.HttpHeaders;

public
abstract
class S3Response< R extends S3ResponseInfo >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final R info;
    private final HttpHeaders hdrs;

    S3Response( R info,
                HttpHeaders hdrs )
    {
        this.info = inputs.notNull( info, "info" );
        this.hdrs = inputs.notNull( hdrs, "hdrs" );
    }

    public final R info() { return info; }
    public final HttpHeaders headers() { return hdrs; }
    public final HttpHeaders h() { return headers(); }
}
