package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;

final
class JvCase
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    JvExpression label;
    final List< JvExpression > body = Lang.newList();

    void
    validate()
    {
        state.notNull( label, "label" );
        for ( JvExpression e : state.noneNull( body, "body" ) ) e.validate();
    }

    static
    JvCase
    create( JvExpression... args )
    {
        state.noneNull( args, "args" );
        state.isTrue( args.length > 0 );

        JvCase res = new JvCase();
        res.label = args[ 0 ];

        if ( args.length > 1 )
        {
            res.body.addAll( Lang.asList( args ).subList( 1, args.length ) );
        }
        
        return res;
    }
}
