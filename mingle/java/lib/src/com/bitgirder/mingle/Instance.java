package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class Instance
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private Instance() {}

    static
    RuntimeException
    unimplemented( Object inst,
                   String method )
    {
        return state.createFailf( "%s does not implement %s()", 
            inst.getClass().getName(), method );
    }
}
