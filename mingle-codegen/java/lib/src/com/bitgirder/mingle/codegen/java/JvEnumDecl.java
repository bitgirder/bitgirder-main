package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;

final
class JvEnumDecl
extends JvDeclaredType
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    JvVisibility vis;
    final List< JvExpression > constants = Lang.newList();
    final List< JvField > fields = Lang.newList();
    final List< JvDeclaredType > nestedTypes = Lang.newList();

    void
    validate()
    {
        state.notNull( vis, "vis" );

        for ( JvExpression e : state.noneNull( constants, "constants" ) )
        {
            e.validate();
        }

        for ( JvField f : state.noneNull( fields, "fields" ) ) f.validate();

        for ( JvDeclaredType t : state.notNull( nestedTypes, "nestedTypes" ) )
        {
            t.validate();
        }
    }
}
