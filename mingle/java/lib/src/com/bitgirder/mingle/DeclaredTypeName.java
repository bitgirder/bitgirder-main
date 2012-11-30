package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.TypedString;

public
final
class DeclaredTypeName
extends TypedString< DeclaredTypeName >
implements TypeName
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    DeclaredTypeName( CharSequence s ) { super( s ); }

    public CharSequence getExternalForm() { return toString(); }

    public
    static
    DeclaredTypeName
    create( CharSequence cs )
    {
        inputs.notNull( cs, "cs" );
        return MingleParser.createDeclaredTypeName( cs );
    }

    public
    static
    DeclaredTypeName
    parse( CharSequence cs )
        throws MingleSyntaxException
    {
        inputs.notNull( cs, "cs" );
        return MingleParser.parseDeclaredTypeName( cs );
    }
}
