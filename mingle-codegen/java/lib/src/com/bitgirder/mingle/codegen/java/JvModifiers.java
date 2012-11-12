package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class JvModifiers
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    JvVisibility vis = JvVisibility.PACKAGE;
    boolean isFinal = true;
    boolean isStatic = false;
    boolean isAbstract = false;

    void
    validate()
    {
        state.notNull( vis, "visiblity" );
    }

    void
    setFrom( JvModifiers other )
    {
        vis = other.vis;
        isFinal = other.isFinal;
        isStatic = other.isStatic;
        isAbstract = other.isAbstract;
    }

    void
    setAbstract()
    {
        isAbstract = true;
        isFinal = false;
    }
}
