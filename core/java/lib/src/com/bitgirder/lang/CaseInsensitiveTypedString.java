package com.bitgirder.lang;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
abstract
class CaseInsensitiveTypedString< T extends CaseInsensitiveTypedString >
extends AbstractTypedString< T >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    protected
    CaseInsensitiveTypedString( CharSequence cs,
                                String paramName )
    {
        super( inputs.notNull( cs, paramName ).toString(), false );
    }

    protected
    CaseInsensitiveTypedString( CharSequence cs )
    {
        this( cs, "cs" ); 
    }

    final
    boolean
    isEqualString( String s1,
                   String s2 )
    {
        return s1.equalsIgnoreCase( s2 );
    }

    final
    int
    compareString( String s1,
                   String s2 )
    {
        return s1.compareToIgnoreCase( s2 );
    }
}
