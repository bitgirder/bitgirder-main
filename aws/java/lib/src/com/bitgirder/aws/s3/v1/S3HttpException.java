package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.http.HttpResponseStatus;

public
final
class S3HttpException
extends Exception
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final HttpResponseStatus status;

    private
    static
    String
    makeMessage( HttpResponseStatus status )
    {
        CharSequence msg = status.getReasonPhrase();
        return msg == null ? null : msg.toString();
    }

    S3HttpException( HttpResponseStatus status )
    {
        super( makeMessage( inputs.notNull( status, "status" ) ) );
 
        this.status = status;
    }

    public HttpResponseStatus getStatus() { return status; }
}
