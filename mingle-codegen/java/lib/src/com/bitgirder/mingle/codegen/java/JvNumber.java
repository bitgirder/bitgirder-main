package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class JvNumber
implements JvExpression
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final CharSequence lit;

    JvNumber( CharSequence lit ) { this.lit = state.notNull( lit, "lit" ); }

    public void validate() {}
}
