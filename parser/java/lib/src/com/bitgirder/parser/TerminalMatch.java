package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Inspector;

public
final
class TerminalMatch< T >
implements ProductionMatch< T >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final T terminal;

    private TerminalMatch( T terminal ) { this.terminal = terminal; }

    public T getTerminal() { return terminal; }

    public
    Inspector
    accept( Inspector i )
    {
        return i.add( "terminal", terminal ); 
    }

    static
    < T >
    TerminalMatch< T >
    forTerminal( T terminal )
    {
        inputs.notNull( terminal, "terminal" );
        return new TerminalMatch< T >( terminal );
    }
}
