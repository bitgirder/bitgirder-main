package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.TypedString;

final
class JvLiteral
extends TypedString< JvLiteral >
implements JvExpression
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final static JvLiteral NULL = new JvLiteral( "null" );

    private JvLiteral( String lit ) { super( lit ); }

    public void validate() {}
}
