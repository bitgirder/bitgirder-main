package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;

public
final
class MingleValueCastException
extends MingleValueException
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    MingleValueCastException( String msg,
                              ObjectPath< MingleIdentifier > loc )
    {
        super( msg, loc );
    }
}
