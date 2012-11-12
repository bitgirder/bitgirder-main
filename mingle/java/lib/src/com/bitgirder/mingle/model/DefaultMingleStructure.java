package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

abstract
class DefaultMingleStructure< S extends MingleStructure >
extends AbstractTypedMingleValue
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleSymbolMap fields;

    DefaultMingleStructure( MingleStructureBuilder< ?, S > b )
    {
        super( b );
        this.fields = b.symBld.build();
    }

    public final MingleSymbolMap getFields() { return fields; }
}
