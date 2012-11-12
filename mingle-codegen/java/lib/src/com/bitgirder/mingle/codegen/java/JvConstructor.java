package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;

final
class JvConstructor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    JvVisibility vis = JvVisibility.PACKAGE;
    List< JvParam > params = Lang.newList();
    List< JvExpression > body = Lang.newList();

    void
    validate()
    {
        state.notNull( vis, "vis" );
        for ( JvParam p : state.noneNull( params, "params" ) ) p.validate();
        for ( JvExpression e : state.noneNull( body, "body" ) ) e.validate();
    }
}
