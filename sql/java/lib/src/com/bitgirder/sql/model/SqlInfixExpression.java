package com.bitgirder.sql.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class SqlInfixExpression
implements SqlExpression
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final SqlExpression left;
    private final SqlOperator op;
    private final SqlExpression right;

    public
    SqlInfixExpression( SqlExpression left,
                        SqlOperator op,
                        SqlExpression right )
    {
        this.left = inputs.notNull( left, "left" );
        this.op = inputs.notNull( op, "op" );
        this.right = inputs.notNull( right, "right" );
    }

    public SqlExpression left() { return left; }
    public SqlOperator operator() { return op; }
    public SqlExpression right() { return right; }
}
