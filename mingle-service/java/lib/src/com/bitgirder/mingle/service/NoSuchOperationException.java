package com.bitgirder.mingle.service;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.model.MingleIdentifier;

public
final
class NoSuchOperationException
extends RuntimeException
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleIdentifier op;

    public
    NoSuchOperationException( MingleIdentifier op )
    {
        super( inputs.notNull( op, "op" ).getExternalForm().toString() );
        this.op = op;
    }

    public MingleIdentifier getOperation() { return op; }
}
