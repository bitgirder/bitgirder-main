package com.bitgirder.io;

import java.io.IOException;

public
interface CharReader
{
    public
    int
    peek()
        throws IOException;

    public
    int
    read()
        throws IOException;
}
