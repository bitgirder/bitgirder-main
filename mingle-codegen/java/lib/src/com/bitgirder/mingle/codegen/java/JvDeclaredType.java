package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;

abstract
class JvDeclaredType
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    JvTypeName name;
    final List< JvType > typeParams = Lang.newList();

    void
    validate()
    {
        state.notNull( name, "name" );

        for ( JvType t : state.noneNull( typeParams, "typeParams" ) )
        {
            t.validate();
        }
    }
}
