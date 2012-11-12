package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Inspectable;
import com.bitgirder.lang.Inspector;

final
class TerminalInstanceMatcher< T >
implements TerminalMatcher< T >,
           Inspectable
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final Class< ? extends T > cls;

    private 
    TerminalInstanceMatcher( Class< ? extends T > cls ) 
    { 
        this.cls = cls; 
    }

    public boolean isMatch( T inst ) { return cls.isInstance( inst ); }

    public Inspector accept( Inspector i ) { return i.add( "cls", cls ); }

    static
    < T >
    TerminalInstanceMatcher< T >
    forClass( Class< ? extends T > cls )
    { 
        return new TerminalInstanceMatcher< T >( inputs.notNull( cls, "cls" ) );
    }
}
