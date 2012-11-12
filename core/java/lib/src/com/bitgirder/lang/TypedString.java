package com.bitgirder.lang;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
abstract
class TypedString< S extends TypedString< S > >
extends AbstractTypedString< S >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    // This constructor is built to facilitate the brief subclass idiom:
    //
    //      public Foo( String fooStr ) { super( fooStr, "fooStr" ); }
    //
    protected
    TypedString( CharSequence s,
                 String paramName )
    {
        super( inputs.notNull( s, paramName ).toString(), true );
    }

    protected TypedString( CharSequence s ) { this( s, "s" ); }
    
    final
    boolean
    isEqualString( String s1,
                   String s2 )
    {
        return s1.equals( s2 );
    }

    final
    int
    compareString( String s1,
                   String s2 )
    {
        return s1.compareTo( s2 );
    }
}
