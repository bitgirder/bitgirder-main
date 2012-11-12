package com.bitgirder.sql.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public 
final
class SqlAliasExpression
implements SqlExpression
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final SqlExpression expr;
    private final SqlId id;

    public 
    SqlAliasExpression( SqlExpression expr,
                        SqlId id )
    {
        this.expr = inputs.notNull( expr, "expr" );
        this.id = inputs.notNull( id, "id" );
    }

    public SqlExpression expression() { return expr; }
    public SqlId id() { return id; }
}
