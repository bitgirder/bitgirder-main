package com.bitgirder.io;

import java.nio.ByteBuffer;

public
interface OctetDigest< D >
{
    // Implementations should change the position of src to reflect how much was
    // digested
    public
    void
    update( ByteBuffer src )
        throws Exception;

    public
    D
    digest()
        throws Exception;
}
