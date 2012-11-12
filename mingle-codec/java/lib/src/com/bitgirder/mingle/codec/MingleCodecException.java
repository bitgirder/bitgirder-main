package com.bitgirder.mingle.codec;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
class MingleCodecException
extends Exception
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    MingleCodecException() {}

    public MingleCodecException( String msg ) { super( msg ); }

    public
    MingleCodecException( String msg,
                          Throwable cause )
    {
        super( msg, cause );
    }
}
