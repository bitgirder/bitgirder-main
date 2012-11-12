package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class ExceptionDefinition
extends StructureDefinition
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private ExceptionDefinition( Builder b ) { super( b ); }

    public
    final
    static
    class Builder
    extends StructureDefinition.Builder< ExceptionDefinition, Builder >
    {
        public
        ExceptionDefinition
        build()
        {
            return new ExceptionDefinition( this );
        }
    }
}
