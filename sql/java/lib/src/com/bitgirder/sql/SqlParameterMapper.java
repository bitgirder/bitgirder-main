package com.bitgirder.sql;

import java.sql.PreparedStatement;

public
interface SqlParameterMapper< I, M >
{
    public
    boolean
    setParameters( I obj,
                   M mapperObj,
                   PreparedStatement ps,
                   int offset )
        throws Exception;
}
