package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class JvAnnotation
implements JvExpression
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final static JvAnnotation OVERRIDE = 
        JvAnnotation.create( JvQname.create( "java.lang", "Override" ) );

    JvType type;

    static
    JvAnnotation
    create( JvType type )
    {
        JvAnnotation res = new JvAnnotation();
        res.type = type;

        return res;
    }

    public
    void
    validate()
    {
        state.notNull( type, "type" );
    }
}
