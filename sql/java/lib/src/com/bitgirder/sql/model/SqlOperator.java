package com.bitgirder.sql.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.TypedString;

public
final
class SqlOperator
extends TypedString< SqlOperator >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public final static SqlOperator EQUALS = new SqlOperator( "=" );
    public final static SqlOperator LESS_THAN = new SqlOperator( "<" );
    public final static SqlOperator GREATER_THAN = new SqlOperator( ">" );
    public final static SqlOperator PLUS = new SqlOperator( "+" );
    public final static SqlOperator MINUS = new SqlOperator( "-" );
    public final static SqlOperator AND = new SqlOperator( "and" );
    public final static SqlOperator OR = new SqlOperator( "or" );
    
    public SqlOperator( CharSequence s ) { super( s, "s" ); }
}
