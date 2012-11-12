package com.bitgirder.mingle.codec;

import java.nio.ByteBuffer;

public
interface MingleCodecDetector
{
    public
    Boolean
    update( ByteBuffer bb )
        throws Exception;
}
