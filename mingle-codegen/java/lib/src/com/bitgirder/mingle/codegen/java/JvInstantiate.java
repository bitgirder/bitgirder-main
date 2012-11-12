package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final 
class JvInstantiate 
extends JvInvocation 
{
    static
    JvInstantiate
    create( JvExpression... args )
    {
        return initFromVarargs( new JvInstantiate(), args );
    }
}
