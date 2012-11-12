package com.bitgirder.mingle.service;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class AuthenticationMissingException
extends RuntimeException
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public AuthenticationMissingException() {}
    public AuthenticationMissingException( String msg ) { super( msg ); }
}
