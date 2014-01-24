package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class MingleValueReactorException
extends Exception
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public MingleValueReactorException( String msg ) { super( msg ); }
}
