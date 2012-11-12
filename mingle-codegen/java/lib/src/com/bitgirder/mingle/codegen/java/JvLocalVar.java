package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class JvLocalVar
implements JvExpression
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    JvId name;
    JvType type;
    JvExpression assign;

    public
    void
    validate()
    {
        state.notNull( name, "name" );
        state.notNull( type, "type" ).validate();
        if ( assign != null ) assign.validate();
    }
}
