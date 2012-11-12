package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class JvParExpression
implements JvExpression
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final JvExpression e;

    JvParExpression( JvExpression e ) { this.e = state.notNull( e, "e" ); }

    public void validate() {}
}
