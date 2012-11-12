package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class JvReturn
implements JvExpression
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final JvExpression e;

    JvReturn( JvExpression e ) { this.e = inputs.notNull( e, "e" ); }

    public void validate() {}
}
