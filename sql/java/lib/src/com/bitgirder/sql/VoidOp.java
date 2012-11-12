package com.bitgirder.sql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.sql.Connection;

public
abstract
class VoidOp
implements ConnectionUser< Void >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    protected
    abstract
    void
    execute( Connection conn )
        throws Exception;
    
    public
    final
    Void
    useConnection( Connection conn )
        throws Exception
    {
        execute( conn );
        return null;
    }
}
