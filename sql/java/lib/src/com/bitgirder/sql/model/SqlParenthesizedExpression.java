package com.bitgirder.sql.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class SqlParenthesizedExpression
implements SqlExpression
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final SqlExpression expr;

    public
    SqlParenthesizedExpression( SqlExpression expr )
    {
        this.expr = inputs.notNull( expr, "expr" );
    }

    public SqlExpression expression() { return expr; }
}
