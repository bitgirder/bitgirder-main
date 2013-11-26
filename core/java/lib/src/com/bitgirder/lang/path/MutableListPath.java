package com.bitgirder.lang.path;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class MutableListPath< E >
extends ListPath< E >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private int idx;

    MutableListPath( ObjectPath< E > parent,
                     int idx )
    {
        super( parent, "parent" );
        setIndex( idx );
    }

    public
    MutableListPath< E >
    setIndex( int idx )
    {
        this.idx = inputs.nonnegativeI( idx, "idx" );
        return this;
    }

    public int getIndex() { return idx; }

    public 
    MutableListPath< E >
    increment()
    {
        return setIndex( getIndex() + 1 );
    }

    public
    MutableListPath< E >
    createCopy()
    {
        return new MutableListPath< E >( getParent(), idx );
    }
}
