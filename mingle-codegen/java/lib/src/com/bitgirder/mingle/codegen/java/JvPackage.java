package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.TypedString;

import com.bitgirder.mingle.model.QualifiedTypeName;

final
class JvPackage
extends TypedString< JvPackage >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    JvPackage( CharSequence pkg ) { super( pkg ); }
}
