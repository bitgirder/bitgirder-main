package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class JvTypeExtendParameter
implements JvType
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final JvType type;
    final JvType sprTyp;

    JvTypeExtendParameter( JvType type,
                           JvType sprTyp )
    {
        this.type = state.notNull( type, "type" );
        this.sprTyp = state.notNull( sprTyp, "sprTyp" );
    }

    public void validate() {}
}
