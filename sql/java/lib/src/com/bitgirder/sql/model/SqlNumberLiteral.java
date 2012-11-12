package com.bitgirder.sql.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class SqlNumberLiteral
implements SqlExpression
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Number num;

    public 
    SqlNumberLiteral( Number num ) 
    { 
        this.num = inputs.notNull( num, "num" );
    }

    public Number number() { return num; }
}
