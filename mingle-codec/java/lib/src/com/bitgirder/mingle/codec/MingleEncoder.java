package com.bitgirder.mingle.codec;

import java.io.IOException;

import java.nio.ByteBuffer;

public
interface MingleEncoder
extends MingleCoder
{
    // returns true if this serializer is complete, false if it needs more
    // buffer space to continue. Implementations may write no data and still
    // return false when the input buffer does not allow the implementation to
    // make progress. In that case, further calls should be made with buffers
    // having more room available.
    public
    boolean
    writeTo( ByteBuffer buf )
        throws Exception;
}
