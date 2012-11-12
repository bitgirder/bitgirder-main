package com.bitgirder.sql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;

public
final
class SqlParameterGroupDescriptor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final List< SqlParameterDescriptor > params;

    private
    SqlParameterGroupDescriptor( List< SqlParameterDescriptor > params )
    {
        this.params = params;
    }

    public List< SqlParameterDescriptor > getParameters() { return params; }

    static
    SqlParameterGroupDescriptor
    createUnsafe( List< SqlParameterDescriptor > params )
    {
        state.notNull( params, "params" );
        state.isFalse( params.isEmpty(), "params is empty" );

        return 
            new SqlParameterGroupDescriptor( Lang.unmodifiableList( params ) );
    }
}
