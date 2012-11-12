package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.io.IOException;

public
final
class CountingCharReader
implements CharReader
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final CharReader cr;

    private long pos;

    public
    CountingCharReader( CharReader cr )
    {
        this.cr = inputs.notNull( cr, "cr" );
    }

    public long position() { return pos; }

    public int peek() throws IOException { return cr.peek(); }

    public
    int
    read()
        throws IOException
    {
        int res = cr.read();
        if ( res >= 0 ) ++pos;

        return res;
    }
}
