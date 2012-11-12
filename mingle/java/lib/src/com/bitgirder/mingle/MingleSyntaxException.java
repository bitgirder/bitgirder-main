package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class MingleSyntaxException
extends Exception
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final String msg;
    private final int col;

    MingleSyntaxException( String msg,
                           int col )
    {
        super( 
            String.format( "[%d]: %s",
                state.nonnegativeI( col, "col" ),
                state.notNull( msg, "msg" )
            )
        );

        this.msg = msg;
        this.col = col;
    }

    public String getError() { return msg; }
    public int getColumn() { return col; }
}
