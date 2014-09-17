package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;

public
final
class MingleUnrecognizedFieldException
extends MingleValueException
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    public
    MingleUnrecognizedFieldException( MingleIdentifier fld,
                                      ObjectPath< MingleIdentifier > loc )
    {
        super(
            "unrecognized field: " + 
                inputs.notNull( fld, "fld" ).getExternalForm(),
            loc
        );
    }
}
