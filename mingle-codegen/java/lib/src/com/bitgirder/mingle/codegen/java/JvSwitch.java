package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;

final
class JvSwitch
implements JvExpression
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    JvExpression target;
    final List< JvCase > cases = Lang.newList();
    final List< JvExpression > defl = Lang.newList();

    public
    void
    validate()
    {
        state.notNull( target, "target" );
        for ( JvCase c : state.noneNull( cases, "cases" ) ) c.validate();
        for ( JvExpression e : state.noneNull( defl, "defl" ) ) e.validate();
    }
}
