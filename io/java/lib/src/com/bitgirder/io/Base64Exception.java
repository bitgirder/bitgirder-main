package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.io.IOException;

public
final
class Base64Exception
extends IOException
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    Base64Exception( String msg ) { super( msg ); }
}
