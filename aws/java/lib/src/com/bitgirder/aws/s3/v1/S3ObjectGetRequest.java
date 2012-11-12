package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.FileWrapper;
import com.bitgirder.io.IoProcessor;

import com.bitgirder.http.HttpMethod;

import java.nio.ByteBuffer;

public
final
class S3ObjectGetRequest
extends S3ObjectRequest
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private 
    S3ObjectGetRequest( Builder b ) 
    { 
        super( b, HttpMethod.GET ); 

        b.assertBodyHandlerSet();
    }

    public
    final
    static
    class Builder
    extends S3ObjectRequest.Builder< Builder >
    {
        public
        Builder
        setReceiveToByteBuffer()
        {
            return setByteBufferBodyHandler();
        }

        public
        Builder
        setReceiveToFile( FileWrapper recvTo,
                          IoProcessor ioProc )
        {
            return setFileReceiveBodyHandler( recvTo, ioProc );
        }

        public
        S3ObjectGetRequest
        build()
        {
            return new S3ObjectGetRequest( this );
        }
    }
}
