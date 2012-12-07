package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.io.IOException;

public
final
class MingleBinaryException
extends IOException
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    MingleBinaryException( String msg ) { super( msg ); }

    MingleBinaryException( String msg,
                           Throwable cause )
    {
        super( msg, cause );
    }
}
