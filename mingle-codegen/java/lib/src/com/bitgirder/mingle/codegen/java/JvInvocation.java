package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;

abstract
class JvInvocation
implements JvExpression
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    JvExpression target;
    final List< JvExpression > params = Lang.newList();

    public
    final
    void
    validate()
    {
        state.notNull( target, "target" );
        for ( JvExpression e : state.notNull( params, "params" ) ) e.validate();
    }

    static
    < I extends JvInvocation >
    I
    initFromVarargs( I res,
                     JvExpression... args )
    {
        state.noneNull( args, "args" );

        res.target = args[ 0 ];

        for ( int i = 1, e = args.length; i < e; ++i ) 
        {
            res.params.add( args[ i ] );
        }

        return res;
    }
}
