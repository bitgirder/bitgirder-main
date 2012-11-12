package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.nio.ByteBuffer;
import java.nio.ByteOrder;

abstract
class BinIoProcessor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final ByteBuffer bb;

    BinIoProcessor( ByteOrder bo )
    {
        byte[] arr = new byte[ 8 ];
        bb = ByteBuffer.wrap( arr ).order( inputs.notNull( bo, "bo" ) );
    }

    final
    ByteBuffer
    flipBuffer()
    {
        bb.flip();
        return bb;
    }

    final 
    ByteBuffer 
    resetBuffer( int len )
    {
        bb.position( 0 );
        bb.limit( len );

        return bb;
    }
}
