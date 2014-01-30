package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class MingleNumberFormatException
extends NumberFormatException
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    MingleNumberFormatException( String msg ) { super( msg ); }
}
