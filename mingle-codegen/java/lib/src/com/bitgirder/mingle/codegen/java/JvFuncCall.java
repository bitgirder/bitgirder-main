package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final 
class JvFuncCall 
extends JvInvocation
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    static
    JvFuncCall
    create( JvExpression... args )
    {
        return initFromVarargs( new JvFuncCall(), args );
    } 
}
