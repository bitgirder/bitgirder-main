package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.io.InputStream;
import java.io.IOException;

public
final
class CountingInputStream
extends InputStream
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final InputStream is;

    private long pos;

    public 
    CountingInputStream( InputStream is )
    {
        this.is = inputs.notNull( is, "is" );
    }

    // If any InputStream method on this instance throws an exception, the
    // return value of this method is undefined. This method returns -1 after a
    // successful call to close()
    public long position() { return pos; }

    @Override
    public 
    int 
    read() 
        throws IOException
    {
        int res = is.read();
        ++pos;

        return res;
    }

    @Override
    public
    int
    read( byte[] arr )
        throws IOException
    {
        inputs.notNull( arr, "arr" );
        return read( arr, 0, arr.length );
    }

    @Override
    public
    int
    read( byte[] arr,
          int off,
          int len )
        throws IOException
    {
        int res = is.read( arr, off, len );
        if ( res >= 0 ) pos += res;

        return res;
    }

    public
    void
    close()
        throws IOException
    {
        is.close();
        pos = -1L;
    }
}
