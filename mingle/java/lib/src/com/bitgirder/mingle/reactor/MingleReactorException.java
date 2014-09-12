package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class MingleReactorException
extends Exception
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public MingleReactorException( String msg ) { super( msg ); }
}
