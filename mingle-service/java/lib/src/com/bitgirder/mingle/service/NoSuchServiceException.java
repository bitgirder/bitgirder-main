package com.bitgirder.mingle.service;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.model.MingleIdentifier;

public
final
class NoSuchServiceException
extends Exception
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleIdentifier svc;

    public
    NoSuchServiceException( MingleIdentifier svc )
    {
        super( inputs.notNull( svc, "svc" ).getExternalForm().toString() );
        this.svc = svc;
    }

    public MingleIdentifier getService() { return svc; }
}
