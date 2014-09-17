package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;

public
abstract
class MingleValueException
extends RuntimeException
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final String err;
    private final ObjectPath< MingleIdentifier > loc;

    protected
    MingleValueException( String err,
                          ObjectPath< MingleIdentifier > loc )
    {
        super();

        this.err = state.notNull( err, "err" );

        this.loc = loc == null ? 
            ObjectPath.< MingleIdentifier >getRoot() : loc;
    }

    public final String error() { return err; }
    public final ObjectPath< MingleIdentifier > location() { return loc; }

    @Override
    public
    final
    String
    getMessage()
    {
        return Mingle.formatError( loc, err );
    }
}
