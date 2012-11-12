package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class MingleStructBuilder
extends MingleStructureBuilder< MingleStructBuilder, MingleStruct >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    MingleStructBuilder() {}

    public MingleStruct build() { return new DefaultMingleStruct( this ); }
}
