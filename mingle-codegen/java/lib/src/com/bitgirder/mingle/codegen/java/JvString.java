package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class JvString
implements JvExpression
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final CharSequence str;

    JvString( CharSequence str ) { this.str = state.notNull( str, "str" ); }

    public void validate() {}
}
