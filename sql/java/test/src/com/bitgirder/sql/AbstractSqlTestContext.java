package com.bitgirder.sql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
abstract
class AbstractSqlTestContext
implements SqlTestContext
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final CharSequence lbl;

    protected
    AbstractSqlTestContext( CharSequence lbl )
    {
        this.lbl = inputs.notNull( lbl, "lbl" );
    }

    public final CharSequence getLabel() { return lbl; }
}
