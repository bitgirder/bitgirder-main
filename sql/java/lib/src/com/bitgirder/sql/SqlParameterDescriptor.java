package com.bitgirder.sql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class SqlParameterDescriptor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final String name;
    private final int indx;
    private final SqlType sqlTyp;

    SqlParameterDescriptor( String name,
                            int indx,
                            SqlType sqlTyp )
    {
        this.name = state.notNull( name, "name" );
        this.indx = state.positiveI( indx, "indx" );
        this.sqlTyp = state.notNull( sqlTyp, "sqlTyp" );
    }

    public String getName() { return name; }
    public int getIndex() { return indx; }
    public SqlType getSqlType() { return sqlTyp; }
}
