package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;

final
class JvMethod
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final JvModifiers mods = new JvModifiers();
    JvType retType;
    JvId name;
    final List< JvParam > params = Lang.newList();
    final List< JvExpression > body = Lang.newList();
    final List< JvAnnotation > anns = Lang.newList();
    final List< JvType > thrown = Lang.newList();

    void
    validate()
    {
        mods.validate();
        state.notNull( name, "name" );
        for ( JvParam p : state.noneNull( params, "params" ) ) p.validate();
        state.notNull( retType, "retType" );
        for ( JvExpression e : state.noneNull( body, "body" ) ) e.validate();
        for ( JvAnnotation a : state.noneNull( anns, "anns" ) ) a.validate();
        for ( JvType t : state.noneNull( thrown, "thrown" ) ) t.validate();
    }

    static
    JvMethod
    namedAccessorFor( JvField f,
                      JvId methName )
    {
        JvMethod res = new JvMethod();
        res.mods.vis = JvVisibility.PUBLIC;
        res.mods.isFinal = true;
        res.retType = f.type;
        res.name = methName;

        res.body.add( new JvReturn( f.name ) );

        return res;
    }

    static
    JvMethod
    namedAccessorFor( JvField f )
    {
        return namedAccessorFor( f, f.name );
    }
}
