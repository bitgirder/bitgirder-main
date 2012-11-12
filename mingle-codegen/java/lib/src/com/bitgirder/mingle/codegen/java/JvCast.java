package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class JvCast
implements JvExpression
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final JvType target;
    final JvExpression e;

    JvCast( JvType target,
            JvExpression e )
    {
        this.target = state.notNull( target, "target" );
        this.e = state.notNull( e, "e" );
    }

    public void validate() {}
}
