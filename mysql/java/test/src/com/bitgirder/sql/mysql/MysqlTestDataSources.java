package com.bitgirder.sql.mysql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.IoUtils;

import com.bitgirder.sql.AbstractSqlTestContext;

import java.net.URL;

import java.util.Properties;

import javax.sql.DataSource;

import com.mysql.jdbc.jdbc2.optional.MysqlDataSource;

public
final
class MysqlTestDataSources
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static String RSRC_NAME = "mysql-testing.properties";

    // lazily initialized and guarded by ensurePropsLoad()
    private static Properties props;
    private static URL rsrcUrl;

    private MysqlTestDataSources() {}

    private
    static
    synchronized
    void
    ensurePropsLoad()
        throws Exception
    {
        if ( props == null )
        {
            rsrcUrl = IoUtils.expectSingleResource( RSRC_NAME );
            props = IoUtils.loadProperties( rsrcUrl );
        }
    }

    private
    static
    String
    expectProperty( String prop )
        throws Exception
    {
        ensurePropsLoad();

        String url = props.getProperty( prop );

        if ( url == null )
        {
            throw state.createFail( rsrcUrl, "has no entry for", prop );
        }
        else return url;
    }

    public
    static
    MysqlDataSource
    getDataSource( String urlProp )
        throws Exception
    {
        inputs.notNull( urlProp, "urlProp" );

        return MysqlDataSources.forUrl( expectProperty( urlProp ) );
    }

    public
    static
    MysqlDataSource
    getDefaultDataSource()
        throws Exception
    {
        return getDataSource( "user1.jdbcUrl" );
    }

    private
    final
    static
    class SqlTestContextImpl
    extends AbstractSqlTestContext
    {
        private SqlTestContextImpl() { super( "mysql" ); }

        public
        DataSource
        getDataSource()
            throws Exception
        {
            return getDefaultDataSource();
        }
    }
}
