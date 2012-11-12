package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;

final
class JvCatch
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    JvType type;
    JvId id;
    final List< JvExpression > body = Lang.newList();

    void
    validate()
    {
        state.notNull( type, "type" );
        state.notNull( id, "id" );
        for ( JvExpression e : state.noneNull( body, "body" ) ) e.validate();
    }
}
