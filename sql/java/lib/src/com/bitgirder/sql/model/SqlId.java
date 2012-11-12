package com.bitgirder.sql.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.TypedString;

public
final
class SqlId
extends TypedString< SqlId >
implements SqlExpression
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public SqlId( CharSequence id ) { super( id, "id" ); }
}
