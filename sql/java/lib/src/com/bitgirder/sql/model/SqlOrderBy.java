package com.bitgirder.sql.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class SqlOrderBy
implements SqlExpression
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public static enum Order { ASC, DESC; }

    private final SqlExpression expr;
    private final Order ordr; // may be null

    private
    SqlOrderBy( SqlExpression expr,
                Order ordr )
    {
        this.expr = inputs.notNull( expr, "expr" );
        this.ordr = ordr;
    }

    public SqlExpression expression() { return expr; }
    public Order order() { return ordr; }

    public
    static
    SqlOrderBy
    create( SqlExpression expr,
            Order ordr )
    {
        return new SqlOrderBy( expr, inputs.notNull( ordr, "ordr" ) );
    }

    public
    static
    SqlOrderBy
    create( SqlExpression expr )
    {
        return new SqlOrderBy( expr, null );
    }
}
