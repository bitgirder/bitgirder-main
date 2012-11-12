package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.model.FieldDefinition;

final
class JvField
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final JvModifiers mods = new JvModifiers();
    JvType type;
    JvId name;
    JvExpression assign;
    FieldDefinition mgField;
    FieldGeneratorParameters fgParams;

    void
    validate()
    {
        state.notNull( type, "type" );
        state.notNull( name, "name" );
        mods.validate();
        if ( assign != null ) assign.validate();
    }

    JvField
    copyOf()
    {
        JvField res = new JvField();

        res.mods.setFrom( mods );
        res.type = type;
        res.name = name;
        res.assign = assign;
        res.mgField = mgField;

        return res;
    }

    static
    JvField
    createConstField()
    {
        JvField res = new JvField();

        res.mods.vis = JvVisibility.PRIVATE;
        res.mods.isFinal = true;
        res.mods.isStatic = true;

        return res;
    }
}
