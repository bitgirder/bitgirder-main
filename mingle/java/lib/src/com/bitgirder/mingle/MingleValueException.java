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

    MingleValueException( String err,
                          ObjectPath< MingleIdentifier > loc )
    {
        super();

        this.err = state.notNull( err, "err" );
        this.loc = state.notNull( loc, "loc" );
    }

    public final String error() { return err; }
    public final ObjectPath< MingleIdentifier > location() { return loc; }

    @Override
    public
    final
    String
    getMessage()
    {
        StringBuilder sb = Mingle.appendIdPath( loc, new StringBuilder() );
        sb.append( ": " );
        sb.append( err );

        return sb.toString();
    }
}
