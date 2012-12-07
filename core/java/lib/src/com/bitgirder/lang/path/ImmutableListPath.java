package com.bitgirder.lang.path;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class ImmutableListPath< E >
extends ListPath< E >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final int indx;

    private
    ImmutableListPath( int indx,
                       ObjectPath< E > parent,
                       String paramName )
    {
        super( parent, paramName );

        this.indx = indx;
    }

    public int getIndex() { return indx; }

    public
    ImmutableListPath< E >
    next()
    {
        return new ImmutableListPath< E >( indx + 1, getParent(), null );
    }

    public
    static
    < E >
    ImmutableListPath< E >
    start( ObjectPath< E > parent,
           int idx )
    {
        inputs.nonnegativeI( idx, "idx" );
        return new ImmutableListPath< E >( idx, parent, "parent" );
    }

    public
    static
    < E >
    ImmutableListPath< E >
    start( ObjectPath< E > parent )
    {
        return start( parent, 0 );
    }
}
