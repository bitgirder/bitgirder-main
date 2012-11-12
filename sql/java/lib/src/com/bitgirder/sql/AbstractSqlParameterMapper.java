package com.bitgirder.sql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.sql.PreparedStatement;

import java.util.Iterator;

public
abstract
class AbstractSqlParameterMapper< I, O >
implements SqlParameterMapper< I, O >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final SqlParameterGroupDescriptor grp;

    protected
    AbstractSqlParameterMapper( SqlParameterGroupDescriptor grp )
    {
        this.grp = inputs.notNull( grp, "grp" );
    }

    protected final SqlParameterGroupDescriptor parameters() { return grp; }

    protected
    abstract
    boolean
    setParameter( I obj,
                  O mapperObj,
                  SqlParameterDescriptor param,
                  PreparedStatement ps,
                  int indx )
        throws Exception;

    public
    final
    boolean
    setParameters( I obj,
                   O mapperObj,
                   PreparedStatement ps,
                   int offset )
        throws Exception
    {
        boolean res = true;
        Iterator< SqlParameterDescriptor > it = grp.getParameters().iterator();

        while ( it.hasNext() && res )
        {
            SqlParameterDescriptor param = it.next();

            res =
                setParameter( 
                    obj, 
                    mapperObj, 
                    param,
                    ps, 
                    offset + param.getIndex() 
                );
        }

        return res;
    }
}
