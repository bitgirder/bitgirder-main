package com.bitgirder.sql.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class SqlBoundParamTarget
implements SqlExpression
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final static SqlBoundParamTarget INSTANCE = new SqlBoundParamTarget();

    private SqlBoundParamTarget() {}
}
