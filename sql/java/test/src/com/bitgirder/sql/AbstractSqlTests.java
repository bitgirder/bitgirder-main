package com.bitgirder.sql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.ObjectReceiver;
import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.lang.reflect.ReflectUtils;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.test.Test;
import com.bitgirder.test.InvocationFactory;
import com.bitgirder.test.LabeledTestObject;
import com.bitgirder.test.TestCall;
import com.bitgirder.test.Before;
import com.bitgirder.test.AbstractLabeledTestObject;

import java.util.Map;
import java.util.Set;
import java.util.List;
import java.util.Collection;
import java.util.AbstractList;
import java.util.Iterator;

import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.Date;
import java.sql.Time;
import java.sql.Timestamp;

import java.lang.reflect.Method;

import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Retention;
import java.lang.annotation.Target;
import java.lang.annotation.ElementType;

import java.nio.ByteBuffer;

import javax.sql.DataSource;

public
abstract
class AbstractSqlTests
extends AbstractLabeledTestObject
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    // guarded by 'synchronized' on initTable()
    private final static Set< String > INIT_STATUS = Lang.newSet();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final SqlTestContext sqlCtx;

    protected
    AbstractSqlTests( SqlTestContext sqlCtx )
    { 
        super( inputs.notNull( sqlCtx, "sqlCtx" ).getLabel() );

        this.sqlCtx = sqlCtx;
    }

    protected final SqlTestContext sqlTestContext() { return sqlCtx; }

    protected 
    final 
    DataSource 
    dataSource() 
        throws Exception
    { 
        return sqlCtx.getDataSource(); 
    }

    protected
    final
    void
    useConnection( ConnectionUser< ? > u )
        throws Exception
    {
        Sql.useConnection( u, sqlCtx.getDataSource() );
    }

    // see doInit() below
    private
    static
    interface LockTask
    {
        public
        void
        execute()
            throws Exception;
    }

    // Because there are multiple subclasses that may all enter this method
    // around the same time, we use basic class-level synchronization to ensure
    // that only the first actually executes anything
    private
    static
    synchronized
    void
    doInit( String lockNm,
            LockTask t )
        throws Exception
    {
        if ( ! INIT_STATUS.contains( lockNm ) )
        {
            t.execute();
            INIT_STATUS.add( lockNm );
        }
    }

    private
    static
    void
    initTable( final Connection conn,
               final String nm,
               final String colDefs )
        throws Exception
    {
        doInit( nm,
            new LockTask() {
                public void execute() throws Exception 
                {
                    Sql.executeUpdate( conn, "drop table if exists " + nm );
            
                    Sql.executeUpdate( conn,
                        "create table " + nm + " ( " + colDefs + " " + " ) " +
                        "engine=innodb, charset=utf8 "
                    );
                }
            }
        );
    }

    // should create:
    //  
    //  table: sql_impl1
    //  cols:
    //      uid: char( 36 ) primary key
    //      data1: varchar( 255 ) not null
    //      data2: integer default null
    //
    //  view: sql_impl1_view1
    //      uid: char( 36 ) primary key
    //      data1: varchar( 255 ) not null
    //
    private
    void
    initTableSqlImpl1( final Connection conn )
        throws Exception
    {
        initTable(
            conn,
            "sql_impl1",
            "uid char( 36 ) charset ascii primary key, " +
            "data1 varchar( 255 ) not null, " +
            "data2 integer default null "
        );

        doInit( "sql_impl1_view1",
            new LockTask() {
                public void execute() throws Exception
                {
                    Sql.executeUpdate( conn, 
                        "drop view if exists sql_impl1_view1" );

                    Sql.executeUpdate( conn,
                        "create view sql_impl1_view1 as " +
                        "select uid, data1 from sql_impl1" );
                }
            }
        );
    }

    // should create:
    //
    //  table: sql_impl2
    //  cols:
    //      uid: char( 36 ) not null
    //      data1: integer not null
    //  indices:
    //      unique( uid, data1 )
    //
    private
    void
    initTableSqlImpl2( Connection conn )
        throws Exception
    {
        initTable(
            conn,
            "sql_impl2",
            "uid char( 36 ) charset ascii not null, " +
            "data1 integer not null, " +
            "unique uid_data1_uniq( uid, data1 ) "
        );
    }

    // should create:
    //
    //  table: sql_impl3
    //  cols:
    //      test_id: char( 36 ) not null
    //      long1: bigint( 64 ) signed default null
    //      str1: varchar( 255 ) default null
    //      double1: double precision default null
    //      bool1: boolean default null
    //      blob1: blob( 255 ) default null
    //      date1: date default null
    //      datetime1: datetime default null
    //      timestamp1: timestamp default null
    //      time1: time default 0
    //
    private
    void
    initTableSqlImpl3( Connection conn )
        throws Exception
    {
        initTable(
            conn,
            "sql_impl3",
            "test_id char( 36 ) charset ascii not null, " +
            "long1 bigint( 64 ) signed default null, " +
            "str1 varchar( 255 ) default null, " +
            "double1 double precision default null, " +
            "bool1 boolean default null, " +
            "blob1 blob( 255 ) default null, " +
            "date1 date default null, " +
            "datetime1 datetime default null, " +
            "timestamp1 timestamp default '1970-01-01 00:00:01', " +
            "time1 time default 0" 
        );
    }

    @Before
    private
    void
    dbInit()
        throws Exception
    {
        Connection conn = sqlCtx.getDataSource().getConnection();

        try
        {
            initTableSqlImpl1( conn ); 
            initTableSqlImpl2( conn ); 
            initTableSqlImpl3( conn ); 
        }
        finally { conn.close(); }
    }

    protected
    static
    abstract
    class BatchImpl< V >
    extends AbstractList< V >
    {
        private final int batchLen;
        private int i;

        protected
        BatchImpl( int batchLen )
        {
            this.batchLen = inputs.positiveI( batchLen, "batchLen" );
        }

        public abstract V get( int i );
        public final int size() { return batchLen; }
    }

    protected
    static
    int
    getInt( Map< String, Object > m,
            String key )
    {
        inputs.notNull( m, "m" );
        inputs.notNull( key, "key" );

        return ( (Number) state.get( m, key, "m" ) ).intValue();
    }
 
    protected static String makeStringData( int i ) { return "sqlCtxing" + i; }
    protected static long makeIntData( int i ) { return i; }
    protected static long makeLongData( int i ) { return (long) i; }

    protected 
    static 
    double 
    makeDoubleData( int i ) 
    { 
        return (double) i + 0.5d; 
    }

    protected static boolean makeBooleanData( int i ) { return i % 2 == 0; }

    protected
    static
    byte[]
    makeBlobData( int i )
    {
        byte[] arr = new byte[ 100 ];
        for ( int j = 0, e = arr.length; j < e; ++j ) arr[ j ] = (byte) i;

        return arr;
    }

    // returns something in "01", "02", ... , "10"
    private
    static
    String
    mod10Str( int i )
    {
        return String.format( "%1$02d", 1 + ( i % 10 ) );
    }

    protected
    static
    Date
    makeDateData( int i )
    {
        return Date.valueOf( "2001-01-" + mod10Str( i ) );
    }

    protected
    static
    Timestamp
    makeTimestampData( int i )
    {
        String s = mod10Str( i );
        return Timestamp.valueOf( "2001-01-" + s + " 12:00:" + s );
    }

    protected
    static
    Time
    makeTimeData( int i )
    {
        return Time.valueOf( "12:00:" + mod10Str( i ) );
    }

    private
    static
    void
    addSqlImpl3Row( String testId,
                    int i,
                    PreparedStatement st )
        throws Exception
    {
        st.clearParameters();

        st.setString( 1, testId );
        st.setLong( 2, makeLongData( i ) );
        st.setString( 3, makeStringData( i ) );
        st.setDouble( 4, makeDoubleData( i ) );
        st.setBoolean( 5, makeBooleanData( i ) );
        st.setBytes( 6, makeBlobData( i ) );
        st.setDate( 7, makeDateData( i ) );
        st.setTimestamp( 8, makeTimestampData( i ) );
        st.setTimestamp( 9, makeTimestampData( i ) );
        st.setTime( 10, makeTimeData( i ) );

        st.addBatch();
    }
 
    protected
    static
    void
    assertSqlImpl3Row( Map< String, Object > m,
                       String testId,
                       int i )
    {
        inputs.notNull( m, "m" );
        inputs.notNull( testId, "testId" );
        inputs.nonnegativeI( i, "i" );

        state.equalInt( 10, m.size() );

        state.equal( testId, state.get( m, "test_id", "m" ) );
        state.equal( makeLongData( i ), state.get( m, "long1", "m" ) );
        state.equal( makeStringData( i ), state.get( m, "str1", "m" ) );
        state.equal( makeDoubleData( i ), state.get( m, "double1", "m" ) );
        state.equal( makeBooleanData( i ), state.get( m, "bool1", "m" ) );

        state.equal( 
            ByteBuffer.wrap( makeBlobData( i ) ),
            ByteBuffer.wrap( (byte[]) state.get( m, "blob1", "m" ) ) );

        state.equal( 
            makeTimestampData( i ), state.get( m, "timestamp1", "m" ) );

        state.equal( makeTimestampData( i ), state.get( m, "datetime1", "m" ) );
        state.equal( makeDateData( i ), state.get( m, "date1", "m" ) );
        state.equal( makeTimeData( i ), state.get( m, "time1", "m" ) );
    }

    protected
    static
    void
    assertSqlImpl3Rows( Connection conn,
                        String testId,
                        int len )
        throws Exception
    {
        inputs.notNull( conn, "conn" );
        inputs.notNull( testId, "testId" );
        inputs.nonnegativeI( len, "len" );

        List< Map< String, Object > > l =
            Sql.selectListOfMaps( conn,
                "select * from sql_impl3 where test_id = ?", testId );
 
        state.equalInt( len, l.size() );

        int i = 0;
        for ( Map< String, Object > m : l ) assertSqlImpl3Row( m, testId, i++ );
    }

    protected
    static
    void
    insertSqlImpl3Rows( Connection conn,
                        String testId,
                        int len )
        throws Exception
    {
        inputs.notNull( conn, "conn" );
        inputs.notNull( testId, "testId" );
        inputs.positiveI( len, "len" );

        PreparedStatement st = 
            conn.prepareStatement( 
                "insert into " +
                "sql_impl3( test_id, long1, str1, double1, bool1, blob1, " +
                "           date1, datetime1, timestamp1, time1 ) " +
                "values ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )"
            );
 
        try
        {
            for ( int i = 0; i < len; ++i ) addSqlImpl3Row( testId, i, st );
            st.executeBatch();
        }
        finally { st.close(); }
    }
    
    protected
    abstract
    class AbstractTest
    implements TestCall
    {
        // to simplify impl of LabeledTestObject
        public Object getInvocationTarget() { return this; }

        protected
        abstract
        void
        startSql()
            throws Exception;

        // This used to have more init logic in it before calling startSql();
        // we'll leave that in place in case we find ourselves needing to add
        // logic again at some point
        public
        final
        void
        call()
            throws Exception
        {
            startSql();
        }
    }

    protected
    abstract
    class AbstractSqlTest
    extends AbstractTest
    {
        protected
        abstract
        void
        useConnection( Connection conn )
            throws Exception;
        
        protected
        final
        void
        startSql()
            throws Exception
        {
            Sql.useConnection(
                new ConnectionUser< Void >() {
                    public 
                    Void 
                    useConnection( Connection conn )
                        throws Exception 
                    {
                        AbstractSqlTest.this.useConnection( conn );
                        return null;
                    }
                },
                sqlCtx.getDataSource()
            );
        }
    }

    protected
    class SetValueTest
    extends AbstractSqlTest
    implements LabeledTestObject
    {
        private final Object toSet;
        private final SqlType sqlTyp;
        private final String colNm;
        private final Object expctVal;
        private final Object tag;

        protected
        SetValueTest( Object toSet,
                      SqlType sqlTyp,
                      String colNm,
                      Object expctVal,
                      String tag )
        {
            this.toSet = toSet;
            this.sqlTyp = sqlTyp;
            this.colNm = colNm;
            this.expctVal = expctVal;
            this.tag = tag;
        }
        
        protected
        SetValueTest( Object toSet,
                      SqlType sqlTyp,
                      String colNm,
                      Object expctVal )
        {
            this( toSet, sqlTyp, colNm, expctVal, null );
        }

        // overridable
        protected
        void
        setValue( Object toSet,
                  SqlType sqlTyp,
                  PreparedStatement ps,
                  int indx )
            throws Exception
        {
            Sql.setValue( toSet, sqlTyp, ps, indx );
        }

        public
        CharSequence
        getLabel()
        {
            StringBuilder res = new StringBuilder();
            
            res.append(
                Strings.crossJoin( "=", ",",
                    "toSet", 
                        toSet == null ? null : toSet.getClass().getSimpleName(),
                    "sqlTyp", sqlTyp,
                    "colNm", colNm
                )
            );
            
            if ( tag != null ) res.append( ",tag=" ).append( tag );

            return res;
        }

        private
        void
        assertVal( Object val )
        {
            if ( state.sameNullity( expctVal, val ) )
            {
                Class< ? > c = expctVal.getClass();

                if ( c.equals( Integer.class ) )
                {
                    state.equalInt( 
                        ( (Number) expctVal ).intValue(),
                        ( (Number) val ).intValue() 
                    );
                }
                else if ( expctVal instanceof ByteBuffer )
                {
                    state.equal( expctVal, ByteBuffer.wrap( (byte[]) val ) );
                }
                else if ( expctVal instanceof Timestamp )
                {
                    state.equal(
                        ( (Timestamp) expctVal ).getTime(),
                        ( (Timestamp) val ).getTime()
                    );
                }
                else if ( expctVal instanceof Time )
                {
                    state.equalString(
                        expctVal.toString(),
                        ( (Time) val ).toString()
                    );
                }
                else if ( expctVal instanceof Date )
                {
                    Date d;

                    if ( val instanceof Timestamp )
                    {
                        d = new Date( ( (Timestamp) val ).getTime() );
                    }
                    else d = (Date) val;

                    state.equal( ( (Date) expctVal ).toString(), d.toString() );
                }
                else state.equal( expctVal, val );
            }
        }

        protected
        final
        void
        useConnection( Connection conn )
            throws Exception
        {
            String id = Lang.randomUuid();

            PreparedStatement ps = 
                conn.prepareStatement( 
                    "insert into sql_impl3( test_id, `" + colNm + "` ) " +
                    "values ( ?, ? )"
                );
 
            setValue( toSet, sqlTyp, ps, 2 );
            ps.setString( 1, id );
            state.equalInt( 1, ps.executeUpdate() );

            assertVal(
                Sql.selectOne( conn,
                    "select `" + colNm + "` from sql_impl3 where test_id = ?",
                    id
                )
            );
        }
    }

    @Retention( RetentionPolicy.RUNTIME )
    @Target( ElementType.METHOD )
    public
    static
    @interface TestWithConn
    {}

    private
    final
    class TestWithConnCall
    extends AbstractSqlTest
    implements LabeledTestObject
    {
        private final Method m;

        private TestWithConnCall( Method m ) { this.m = m; }

        public CharSequence getLabel() { return m.getName(); }

        protected
        void
        useConnection( Connection conn )
            throws Exception
        {
            ReflectUtils.invoke( m, AbstractSqlTests.this, conn );
        }
    }

    @InvocationFactory
    private
    List< LabeledTestObject >
    connTests()
        throws Exception
    {
        Collection< Method > l = 
            ReflectUtils.getDeclaredMethods( getClass(), TestWithConn.class );

        List< LabeledTestObject > res = Lang.newList( l.size() );

        for ( Method m : l )
        {
            m.setAccessible( true );

            Class< ? >[] params = m.getParameterTypes();

            state.isTrue(
                params.length == 1 && params[ 0 ].equals( Connection.class ),
                "Method", m, "should have one parameter of type Connection"
            );

            res.add( new TestWithConnCall( m ) );
        }

        return res;
    }
}
