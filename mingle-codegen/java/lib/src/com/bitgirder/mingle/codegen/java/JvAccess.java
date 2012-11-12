package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class JvAccess
implements JvExpression
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final JvExpression left;
    final JvExpression right;

    JvAccess( JvExpression left,
              JvExpression right )
    {
        this.left = state.notNull( left, "left" );
        this.right = state.notNull( right, "right" );
    }

    public void validate() {}
}
