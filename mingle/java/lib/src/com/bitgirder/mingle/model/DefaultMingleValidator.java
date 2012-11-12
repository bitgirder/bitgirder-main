package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;

class DefaultMingleValidator
implements MingleValidator
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final ObjectPath< MingleIdentifier > path;

    DefaultMingleValidator( ObjectPath< MingleIdentifier > path )
    {
        this.path = state.notNull( path, "path" );
    }

    public final ObjectPath< MingleIdentifier > getPath() { return path; }

    public
    final
    void
    isTrue( boolean val,
            Object... msg )
    {
        MingleValidation.isFalse( ! val, path, msg );
    }

    public
    final
    void
    isFalse( boolean val,
             Object... msg )
    {
        MingleValidation.isFalse( val, path, msg );
    }
}
