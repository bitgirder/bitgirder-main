package com.bitgirder.sql.mysql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import javax.sql.DataSource;

import com.mysql.jdbc.jdbc2.optional.MysqlDataSource;

public
final
class MysqlDataSources
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private MysqlDataSources() {}

    public
    static
    MysqlDataSource
    forUrl( CharSequence url )
    {
        inputs.notNull( url, "url" );

        MysqlDataSource res = new MysqlDataSource();
        res.setUrl( url.toString() );

        return res;
    }
}
