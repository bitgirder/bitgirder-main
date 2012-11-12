package com.bitgirder.lang;

import com.bitgirder.validation.Inputs;

import java.util.concurrent.atomic.AtomicInteger;

public
class StandardThread
extends Thread
{
    private final static Inputs inputs = new Inputs();

    private final static AtomicInteger id = new AtomicInteger();

    private
    static
    String
    makeName( String fmt )
    {
        inputs.notNull( fmt, "fmt" );
        return String.format( fmt, id.getAndIncrement() );
    }

    private StandardThread() {} // disallow without name

    protected StandardThread( String fmt ) { super( makeName( fmt ) ); }

    public
    StandardThread( Runnable r,
                    String fmt )
    {
        super( inputs.notNull( r, "r" ), makeName( fmt ) );
    }
}
