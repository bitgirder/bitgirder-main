package com.bitgirder.mingle.codec;

import java.nio.ByteBuffer;

public
interface MingleDecoder< E >
extends MingleCoder
{
    public
    boolean
    readFrom( ByteBuffer bb,
              boolean endOfInput )
        throws Exception;

    // should throw IllegalStateException if called more than once or before a
    // call to readFrom() returns true
    public
    E
    getResult()
        throws Exception;
}
