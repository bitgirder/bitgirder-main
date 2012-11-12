package com.bitgirder.mingle.bind;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
abstract
class BoundStruct
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    protected BoundStruct() {}

    public
    static
    abstract
    class AbstractBuilder< B extends AbstractBuilder< B > >
    extends BoundStructureBuilder< B >
    {}

    protected
    static
    abstract
    class AbstractBindImplementation
    extends com.bitgirder.mingle.bind.AbstractBindImplementation
    {
        protected AbstractBindImplementation() {};
    }
}
