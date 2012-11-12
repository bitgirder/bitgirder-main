package com.bitgirder.sql.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class SqlStatementWriters
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private SqlStatementWriters() {}

    public
    static
    SqlStatementWriter
    createDefaultWriter()
    {
        return new SqlStatementWriter(); 
    }

    public
    static
    CharSequence
    writeExpression( SqlExpression expr,
                     SqlStatementWriter w )
    {
        inputs.notNull( expr, "expr" );
        inputs.notNull( w, "w" );

        StringBuilder sb = new StringBuilder();
        w.writeExpression( expr, sb );

        return sb;
    }

    public
    static
    CharSequence
    writeExpression( SqlExpression expr )
    {
        return writeExpression( expr, createDefaultWriter() );
    }
}
