package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.Inspector;
import com.bitgirder.lang.Inspectable;

import java.util.Set;

// Could generalize this beyond chars as needed
final
class TerminalSetMatcher< T >
implements TerminalMatcher< T >,
           Inspectable
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final Set< T > terminals;
    private final boolean isComplement;

    private
    TerminalSetMatcher( Set< T > terminals,
                        boolean isComplement )
    {
        this.terminals = terminals;
        this.isComplement = isComplement;
    }

    public 
    boolean 
    isMatch( T terminal )
    {
        boolean isIn = terminals.contains( terminal );
        return isComplement ? ! isIn : isIn;
    }
    
    @Override
    public
    Inspector
    accept( Inspector i )
    {
        return i.add( "terminals", terminals ).
                 add( "isComplement", isComplement );
    }

    static
    < T >
    TerminalSetMatcher< T >
    forSet( Set< T > set,
            boolean isComplement )
    {
        inputs.noneNull( set, "set" );

        return 
            new TerminalSetMatcher< T >( 
                Lang.unmodifiableSet( set ), isComplement );
    }
}
