package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class JvCompilationUnit
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    JvPackage pkg;
    JvDeclaredType decl;

    void
    validate()
    {
        state.notNull( pkg, "pkg" );
        state.notNull( decl, "decl" ).validate();
    }
}
