package com.bitgirder.mingle.bind;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

abstract
class BoundStructureBuilder< B extends BoundStructureBuilder< B > >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    protected final B castThis() { return Lang.< B >castUnchecked( this ); }
}
