package com.bitgirder.lang;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class OverloadException
extends RuntimeException
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    public OverloadException() { super(); }

    // can add other constructors as needed
}
