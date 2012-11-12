package com.bitgirder.sql;

import javax.sql.DataSource;

public
interface SqlTestContext
{
    public
    CharSequence
    getLabel();

    public
    DataSource
    getDataSource()
        throws Exception;
}
