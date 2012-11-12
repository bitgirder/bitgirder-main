package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.TypedString;

import com.bitgirder.mingle.model.QualifiedTypeName;
import com.bitgirder.mingle.model.MingleTypeName;

import java.util.List;

final
class JvTypeName
extends TypedString< JvTypeName >
implements JvExpression,
           JvType
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    JvTypeName( CharSequence nm ) { super( nm, "nm" ); }

    public void validate() {}

    static
    JvTypeName
    ofQname( QualifiedTypeName qn )
    {
        state.notNull( qn, "qn" );

        List< MingleTypeName > nm = qn.getName();
        state.equalInt( 1, nm.size() );

        return new JvTypeName( nm.get( nm.size() - 1 ).getExternalForm() );
    }
}
