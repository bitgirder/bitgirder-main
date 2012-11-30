package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;

public
final
class MingleValidationException
extends RuntimeException
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final String err;

    MingleValidationException( String err )
    {
        super( err );
        this.err = err;
    }

    public String getError() { return err; }
}
