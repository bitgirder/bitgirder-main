package com.bitgirder.mingle.service;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class InternalServiceException
extends Exception
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public InternalServiceException() { super(); }
}
