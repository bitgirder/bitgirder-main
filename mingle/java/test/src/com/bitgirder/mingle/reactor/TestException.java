package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.MingleValueException;
import com.bitgirder.mingle.MingleIdentifier;

import com.bitgirder.lang.path.ObjectPath;

public
final
class TestException
extends MingleValueException
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    TestException( String msg,
                   ObjectPath< MingleIdentifier > path )
    {
        super( msg, path );
    }
}
