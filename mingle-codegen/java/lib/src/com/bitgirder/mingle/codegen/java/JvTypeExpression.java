package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;

final
class JvTypeExpression
implements JvType
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    JvType type;
    final List< JvType > args = Lang.newList();
    JvType next;
    
    public
    void
    validate()
    {
        state.notNull( type, "type" );
        for ( JvType a : state.noneNull( args, "args" ) ) a.validate();
        if ( next != null ) next.validate();
    }

    static
    JvTypeExpression
    dotTypeOf( JvType left,
               JvType right )
    {
        state.notNull( left, "left" );
        state.notNull( right, "right" );

        JvTypeExpression res = new JvTypeExpression();
        res.type = left;
        res.next = right;
        
        return res;
    }

    static
    JvTypeExpression
    withParams( JvType... typs )
    {
        state.noneNull( typs, "typs" );
        state.isTrue( typs.length > 1 );

        JvTypeExpression res = new JvTypeExpression();
        res.type = typs[ 0 ];

        for ( int i = 1, e = typs.length; i < e; ++i )
        {
            res.args.add( typs[ i ] );
        }

        return res;
    }
}
