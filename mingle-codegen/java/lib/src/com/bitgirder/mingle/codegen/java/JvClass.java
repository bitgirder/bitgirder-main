package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;

final
class JvClass
extends JvDeclaredType
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    JvModifiers mods = new JvModifiers();
    JvType sprTyp;
    final List< JvField > fields = Lang.newList();
    final List< JvConstructor > constructors = Lang.newList();
    final List< JvMethod > methods = Lang.newList();
    final List< JvDeclaredType > nestedTypes = Lang.newList();
    final List< JvType > implemented = Lang.newList();

    public
    void
    validate()
    {
        super.validate();
        mods.validate();

        for ( JvField f : state.noneNull( fields, "fields" ) ) f.validate();

        for ( JvConstructor c : state.noneNull( constructors, "constructors" ) )
        {
            c.validate();
        }

        for ( JvMethod m : state.noneNull( methods, "methods" ) ) m.validate();

        for ( JvDeclaredType t : state.notNull( nestedTypes, "nestedTypes" ) )
        {
            t.validate();
        }

        if ( sprTyp != null ) sprTyp.validate();

        for ( JvType t : state.notNull( implemented, "implemented" ) )
        {
            t.validate();
        }
    }
}
