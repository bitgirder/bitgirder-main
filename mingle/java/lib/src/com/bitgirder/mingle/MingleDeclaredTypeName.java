package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.TypedString;

public
final
class MingleDeclaredTypeName
extends TypedString< MingleDeclaredTypeName >
implements AtomicTypeReference.Name
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    MingleDeclaredTypeName( CharSequence s ) { super( s ); }

    public CharSequence getExternalForm() { return toString(); }

    public
    static
    MingleDeclaredTypeName
    create( CharSequence cs )
    {
        throw new UnsupportedOperationException( "Unimplemented" );
    }

    public
    static
    MingleDeclaredTypeName
    parse( CharSequence cs )
        throws MingleSyntaxException
    {
        throw new UnsupportedOperationException( "Unimplemented" );
    }
}
