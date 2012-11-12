package com.bitgirder.mingle.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.TypedString;

public
final
class IdentifiableText
extends TypedString< IdentifiableText >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    IdentifiableText( CharSequence cs ) { super( cs ); }

    public
    boolean
    equalsString( CharSequence other )
    {
        if ( other == null ) return false;
        else return toString().equals( other.toString() );
    }
}
