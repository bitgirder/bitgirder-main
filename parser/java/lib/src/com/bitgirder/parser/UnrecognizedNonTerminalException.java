package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class UnrecognizedNonTerminalException
extends RuntimeException
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    UnrecognizedNonTerminalException( String msg ) { super( msg ); }
}
