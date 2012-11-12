package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.TypedString;

final
class JvPrimitiveType
extends TypedString< JvPrimitiveType >
implements JvType
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final static JvPrimitiveType VOID = 
        new JvPrimitiveType( "void", JvQname.create( "java.lang", "Void" ) );

    final static JvPrimitiveType LONG = 
        new JvPrimitiveType( "long", JvQname.create( "java.lang", "Long" ) );

    final static JvPrimitiveType INT = 
        new JvPrimitiveType( "int", JvQname.create( "java.lang", "Integer" ) );

    final static JvPrimitiveType DOUBLE =
        new JvPrimitiveType( 
            "double", JvQname.create( "java.lang", "Double" ) );

    final static JvPrimitiveType FLOAT =
        new JvPrimitiveType( "float", JvQname.create( "java.lang", "Float" ) );

    final static JvPrimitiveType BOOLEAN = 
        new JvPrimitiveType( 
            "boolean", JvQname.create( "java.lang", "Boolean" ) );

    final JvType boxed;

    private 
    JvPrimitiveType( CharSequence s,
                     JvType boxed ) 
    { 
        super( s ); 
        this.boxed = boxed;
    }

    public void validate() {}
}
