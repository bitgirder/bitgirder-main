package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

abstract
class TypedMingleValue
implements MingleValue
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final AtomicTypeReference typ;

    TypedMingleValue( AtomicTypeReference typ )
    {
        this.typ = inputs.notNull( typ, "typ" );
    }

    public final AtomicTypeReference getType() { return typ; }
}
