package com.bitgirder.mingle.codec;

import java.nio.ByteBuffer;

public
interface MingleCodecDetection
{
    // Returns true when detection is complete; no further calls should be made
    // after a return value of true is encoutered
    public
    boolean
    update( ByteBuffer bb )
        throws Exception;

    // Should throw an exception if not enough input was available to make a
    // detection (aka, if this method is called before a true value is returned
    // by update()), should throw NoSuchMingleCodecException if it was
    // determined definitively that no codecs match, and should throw a
    // detection exception if more than one codec matches
    public
    MingleCodec
    getResult()
        throws MingleCodecException;
}
