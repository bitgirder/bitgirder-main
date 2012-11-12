package com.bitgirder.sql;

import java.sql.Connection;

public
interface ConnectionUser< V >
{
    public
    V
    useConnection( Connection conn )
        throws Exception;
}
