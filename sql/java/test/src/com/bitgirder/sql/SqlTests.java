package com.bitgirder.sql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.reflect.ReflectUtils;

import com.bitgirder.io.IoUtils;

import com.bitgirder.test.Test;

import java.lang.reflect.Constructor;

import java.util.List;
import java.util.Properties;

import java.net.URL;

import java.sql.SQLException;

@Test
public
final
class SqlTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static String RSRC_NAME = 
        "com/bitgirder/sql/sql-test-context.properties";

    private final static String PROP_CLASS_NAME = "className";

    // guarded by sync and read-only once initialized
    private static List< SqlTestContext > sqlCtxs;

    private SqlTests() {}

    @Test
    private
    void
    testIsDuplicateKeyException()
    {
        state.isFalse(
            Sql.isDuplicateKeyException( new SQLException( "", "1" ) ) );

        state.isTrue(
            Sql.isDuplicateKeyException(
                new SQLException( "", Sql.SQL_STATE_DUP_KEY ) ) );
    }

    private
    static
    void
    initCtx( List< SqlTestContext > l,
             URL u )
        throws Exception
    {
        Properties props = IoUtils.loadProperties( u );

        String clsNm = 
            state.getProperty( props, PROP_CLASS_NAME, u.toString() );

        Class< ? > cls = Class.forName( clsNm );

        SqlTestContext ctx = (SqlTestContext) ReflectUtils.newInstance( cls );

        l.add( ctx );
    }

    private
    static
    synchronized
    List< SqlTestContext >
    getSqlTestContexts()
        throws Exception
    {
        if ( sqlCtxs == null )
        {
            sqlCtxs = Lang.newList();
 
            for ( URL u : IoUtils.getResources( RSRC_NAME ) )
            {
                try { initCtx( sqlCtxs, u ); }
                catch ( Throwable th ) 
                {
                    throw new Exception( "Couldn't init ctx from: " + u, th );
                }
            }

            sqlCtxs = Lang.unmodifiableList( sqlCtxs );
        }
 
        return sqlCtxs;
    }

    public
    static
    < T >
    List< T >
    createSqlSuite( Class< T > cls )
        throws Exception
    {
        inputs.notNull( cls, "cls" );

        List< T > res = Lang.newList();

        Constructor< T > cons = 
            cls.getDeclaredConstructor( SqlTestContext.class );

        cons.setAccessible( true );
        
        for ( SqlTestContext sqlCtx : getSqlTestContexts() )
        {
            res.add( ReflectUtils.invoke( cons, sqlCtx ) );
        }

        return res;
    }
}
