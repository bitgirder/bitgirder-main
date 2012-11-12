package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class StructDefinition
extends StructureDefinition
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private StructDefinition( Builder b ) { super( b ); }

    public
    final
    static
    class Builder
    extends StructureDefinition.Builder< StructDefinition, Builder >
    {
        public
        StructDefinition
        build()
        {
            return new StructDefinition( this );
        }
    }
}
