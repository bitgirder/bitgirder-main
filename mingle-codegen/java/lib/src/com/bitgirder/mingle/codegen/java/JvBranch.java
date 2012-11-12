package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;

final
class JvBranch
implements JvExpression
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();
    
    JvExpression test;
    final List< JvExpression > onTrue = Lang.newList();
    final List< JvExpression > onFalse = Lang.newList();

    public
    void
    validate()
    {
        state.notNull( test, "test" );

        state.isFalse( onTrue.isEmpty(), "If branch is empty" );

        for ( JvExpression e : state.noneNull( onTrue, "onTrue" ) ) 
        {
            e.validate();
        }

        for ( JvExpression e : state.noneNull( onFalse, "onFalse" ) ) 
        {
            e.validate();
        }
    }
}
