package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class StructGeneratorParameters
extends StructureGeneratorParameters
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();
    
    private StructGeneratorParameters( Builder b ) { super( b ); }

    final
    static
    class Builder
    extends StructureGeneratorParameters.Builder< Builder >
    {
        StructGeneratorParameters
        build()
        {
            return new StructGeneratorParameters( this );
        }
    }
}
