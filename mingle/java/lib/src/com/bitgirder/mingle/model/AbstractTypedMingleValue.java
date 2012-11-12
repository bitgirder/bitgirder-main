package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

abstract
class AbstractTypedMingleValue
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final AtomicTypeReference typeRef;

    AbstractTypedMingleValue( MingleTypedValueBuilder< ?, ? > b )
    {
        state.notNull( b, "b" );

        this.typeRef = inputs.notNull( b.typeRef, "typeRef" );
    }

    public 
    final 
    AtomicTypeReference 
    getType() 
    { 
        return typeRef; 
    }
}
