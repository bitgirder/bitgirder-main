package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.TypedString;

final
class JvId
extends TypedString< JvId >
implements JvExpression
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final static JvId THIS = new JvId( "this" );
    final static JvId SUPER = new JvId( "super" );
    final static JvId CLASS = new JvId( "class" );

    JvId( CharSequence id ) { super( id, "id" ); }

    public void validate() {}
}
