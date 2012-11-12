package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;

final
class JvTry
implements JvExpression
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final List< JvExpression > body = Lang.newList();
    final List< JvCatch > catches = Lang.newList();

    public
    void
    validate()
    {
        state.isFalse( body.isEmpty(), "No body for try" );
        for ( JvExpression e : state.noneNull( body, "body" ) ) e.validate();

        state.isFalse( catches.isEmpty(), "No catches supplied" );
        for ( JvCatch c : state.noneNull( catches, "catches" ) ) c.validate();
    }
}
