package com.bitgirder.sql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;
import com.bitgirder.lang.ObjectReceiver;
import com.bitgirder.lang.TestSums;
import com.bitgirder.lang.TypedString;
import com.bitgirder.lang.CaseInsensitiveTypedString;

import com.bitgirder.io.IoTestFactory;

import com.bitgirder.test.Test;
import com.bitgirder.test.TestFactory;
import com.bitgirder.test.TestFailureExpector;
import com.bitgirder.test.LabeledTestObject;
import com.bitgirder.test.InvocationFactory;

import java.util.Map;
import java.util.Set;
import java.util.List;
import java.util.Date;
import java.util.Collection;

import java.sql.Connection;
import java.sql.ResultSet;
import java.sql.PreparedStatement;
import java.sql.Statement;
import java.sql.Timestamp;
import java.sql.Time;
import java.sql.SQLIntegrityConstraintViolationException;

import java.nio.ByteBuffer;

import java.math.BigInteger;
import java.math.BigDecimal;

public
final
class SqlImplTests
extends AbstractSqlTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    SqlImplTests( SqlTestContext sqlCtx ) { super( sqlCtx ); }

    @TestWithConn
    private
    void
    testBasicOps( Connection conn )
        throws Exception
    {
        String uid = Lang.randomUuid();
        String testData1 = "test-data1";

        state.equalInt( 1,
            Sql.executeUpdate( conn,
                "insert into sql_impl1 ( uid, data1 ) " +
                "values ( ?, ? )",
                uid, testData1
            )
        );

        state.equalString( testData1, 
            Sql.selectString( conn,
                "select data1 from sql_impl1 where uid = ?", uid ) );
    }

    private 
    static
    enum ValType
    {
        OBJECT,
        LONG,
        LONG_OBJ,
        INT,
        INT_OBJ,
        ROW_DATA,
        MAP_DATA,
        STRING;
    }

    private
    final
    static
    class SqlImpl2RowData
    {
        private String uid;
        private int data1;
    }

    private
    final
    static
    class SqlImpl2RowProcessor
    extends Sql.DefaultRowProcessor< SqlImpl2RowData >
    {
        protected
        SqlImpl2RowData
        implProcessRow( ResultSet rs )
            throws Exception
        {
            SqlImpl2RowData res = new SqlImpl2RowData();

            res.uid = rs.getString( "uid" );
            res.data1 = rs.getInt( "data1" );

            return res;
        }
    }

    private
    abstract
    class AbstractSelectOneTest< V >
    extends AbstractSqlTest
    implements TestFailureExpector,
               LabeledTestObject
    {
        final boolean useExpct;
        final int rsSize;

        final int len = 10;
        final String uid = Lang.randomUuid();

        private
        AbstractSelectOneTest( boolean useExpct,
                               int rsSize )
        {
            this.useExpct = useExpct;
            this.rsSize = rsSize;
        }

        abstract
        void
        appendLabel( StringBuilder sb );

        public
        final
        CharSequence
        getLabel()
        {
            StringBuilder sb = new StringBuilder( getClass().getSimpleName() );
            sb.append( "," );

            sb.append( 
                Strings.crossJoin( "=", ",",
                    "useExpct", useExpct,
                    "rsSize", rsSize
                )
            );

            appendLabel( sb );

            return sb;
        }

        public
        final
        Class< ? extends Throwable >
        expectedFailureClass()
        {
            if ( ( useExpct && rsSize == 1 ) || ( rsSize < 2 && ! useExpct ) )
            {
                return null;
            }
            else return IllegalStateException.class;
        }

        public
        final
        CharSequence
        expectedFailurePattern()
        {
            switch ( rsSize )
            {
                case 0: return "ResultSet has no rows";
                case 2: return "ResultSet has more than one row";
                default: return null;
            }
        }

        abstract
        CharSequence
        getFields();

        abstract
        V
        doSelectVal( Connection conn,
                     CharSequence sql,
                     Object[] bindArgs )
            throws Exception;

        // o is the select result; i is the value of data1 in the result row
        abstract
        void
        assertResult( V o,
                      int i );

        private
        void
        insertTestData( Connection conn )
            throws Exception
        {
            PreparedStatement st =
                conn.prepareStatement( 
                    "insert into sql_impl2( uid, data1 ) values ( ?, ? )" );
            try
            {
                for ( int i = 0; i < len; ++i )
                {
                    st.setString( 1, uid );
                    st.setInt( 2, i );
                    st.addBatch();
                }

                st.executeBatch();
            }
            finally { st.close(); }
        }
 
        private
        Object[]
        getBindArgs()
        {
            switch ( rsSize )
            {
                case 0: return new Object[] { uid, len };
                case 1: return new Object[] { uid, 0 };
                case 2: return new Object[] { uid };
                default: throw state.createFail( "rsSize:", rsSize );
            }
        }

        private
        void
        doSelect( Connection conn )
            throws Exception
        {
            StringBuilder sb =
                new StringBuilder( "select " ).
                    append( getFields() ).
                    append( " from sql_impl2 where uid = ? and " ).
                    append( rsSize == 2 ? "( data1 % 2 = 0 )" : "data1 = ?" );
            
            V val = doSelectVal( conn, sb, getBindArgs() );
            assertResult( val, 0 );
        }
 
        protected
        void
        useConnection( Connection conn )
            throws Exception
        {
            insertTestData( conn );
            doSelect( conn );
        }
    }

    private
    final
    class SelectValueTest
    extends AbstractSelectOneTest< Object >
    {
        private final ValType vt;

        private
        SelectValueTest( boolean useExpct,
                         int rsSize,
                         ValType vt )
        {
            super( useExpct, rsSize );
            this.vt = vt;
        }

        void
        appendLabel( StringBuilder sb )
        {
            sb.append( ",vt=" ).append( vt );
        }

        CharSequence
        getFields()
        {
            switch ( vt )
            {
                case OBJECT:
                case LONG: 
                case LONG_OBJ:
                case INT: 
                case INT_OBJ:
                    return "data1";

                case STRING: return "cast( data1 as char ) as data1";

                case ROW_DATA:
                case MAP_DATA:
                    return "uid, data1";

                default: throw state.createFail( "Unrecognized vt:", vt );
            }
        }

        Object
        doSelectVal( Connection conn,
                     CharSequence sql,
                     Object[] bindArgs )
            throws Exception
        {
            switch ( vt )
            {
                case OBJECT:
                    return useExpct 
                        ? Sql.expectOne( conn, sql, bindArgs )
                        : Sql.selectOne( conn, sql, bindArgs );
                
                case STRING:
                    return useExpct
                        ? Sql.expectString( conn, sql, bindArgs )
                        : Sql.selectString( conn, sql, bindArgs );
                
                case INT:
                    return useExpct
                        ? Sql.expectInt( conn, sql, bindArgs )
                        : Sql.selectInt( conn, sql, bindArgs );
                
                case INT_OBJ:
                    return useExpct
                        ? Sql.expectInteger( conn, sql, bindArgs )
                        : Sql.selectInteger( conn, sql, bindArgs );
                
                case LONG:
                    return useExpct
                        ? Sql.expectLong( conn, sql, bindArgs )
                        : Sql.selectLong( conn, sql, bindArgs );
                
                case LONG_OBJ:
                    return useExpct
                        ? Sql.expectLongObj( conn, sql, bindArgs )
                        : Sql.selectLongObj( conn, sql, bindArgs );

                case ROW_DATA:
                    SqlImpl2RowProcessor p = new SqlImpl2RowProcessor();
                    return useExpct
                        ? Sql.expectObject( conn, sql, p, bindArgs )
                        : Sql.selectObject( conn, sql, p, bindArgs );
                
                case MAP_DATA:
                    return useExpct
                        ? Sql.expectOneMap( conn, sql, bindArgs )
                        : Sql.selectOneMap( conn, sql, bindArgs );

                default: throw state.createFail( "vt:", vt );
            }
        }

        // returns true if o is not null and if null is not expected; false if o
        // is null and null was expected; fails otherwise
        private
        boolean
        expectedNull( Object o )
        {
            if ( rsSize == 0 )
            {
                state.isTrue( o == null );
                return true;
            }
            else
            {
                state.isFalse( o == null );
                return false;
            }
        }

        private
        void
        assertRowDataResult( SqlImpl2RowData d,
                             int i )
        {
            if ( ! expectedNull( d ) )
            {
                state.equalInt( i, d.data1 );
                state.equal( uid, d.uid );
            }
        }

        private
        void
        assertMapDataResult( Map< String, Object > m,
                             int i )
        {
            if ( ! expectedNull( m ) )
            {
                state.equal( uid, state.get( m, "uid", "m" ) );
                state.equal( i, state.get( m, "data1", "m" ) );
            }
        }

        void
        assertResult( Object o,
                      int i )
        {
            switch ( vt )
            {
                case STRING: 
                    if ( rsSize == 0 ) state.isTrue( o == null );
                    else state.equal( Integer.toString( i ), o ); 
                    break;

                case LONG_OBJ:
                case INT_OBJ:
                case OBJECT:
                    if ( rsSize == 0 ) state.isTrue( o == null );
                    else state.equalInt( i, ( (Number) o ).intValue() );
                    break;

                case INT:
                case LONG:
                    state.equalInt( i, ( (Number) o ).intValue() );
                    break;
                
                case ROW_DATA:
                    assertRowDataResult( (SqlImpl2RowData) o, i );
                    break;

                case MAP_DATA:
                    Map< String, Object > m = Lang.castUnchecked( o );
                    assertMapDataResult( m, i );
                    break;

                default: throw state.createFail( "vt:", vt );
            }
        }
    }

    @InvocationFactory
    private
    List< SelectValueTest >
    testSelectValue()
    {
        List< SelectValueTest > res = Lang.newList();

        for ( int i = 0; i < 2; ++i )
        {
            for ( ValType vt : ValType.class.getEnumConstants() )
            {
                for ( int rsSize = 0; rsSize < 3; ++rsSize )
                {
                    res.add( new SelectValueTest( i == 0, rsSize, vt ) );
                }
            }
        }

        return res;
    }

    @TestWithConn
    private
    void
    testResultSetAsRowElidesNulls( Connection conn )
        throws Exception
    {
        String testId = Lang.randomUuid();
        insertSqlImpl3Rows( conn, testId, 1 );

        Map< String, Object > m =
            Sql.expectOneMap( conn,
                "select *, null as null_val " +
                "from sql_impl3 where test_id = ?",
                testId
            );
        
        assertSqlImpl3Row( m, testId, 0 );
    }

    // will later make protected for things like mingle-sql
    private
    final
    class GetObjectTest
    extends AbstractSqlTest
    implements LabeledTestObject
    {
        private final Object setVal;
        private final String colNm;
        private final Class< ? > cls;
        private final Object expctObj;

        private
        GetObjectTest( Object setVal,
                       String colNm,
                       Class< ? > cls,
                       Object expctObj )
        {
            this.setVal = setVal;
            this.colNm = colNm;
            this.cls = cls;
            this.expctObj = expctObj;
        }

        public
        CharSequence
        getLabel()
        {
            return Strings.crossJoin( "=", ",",
                "setVal", 
                    setVal == null ? "null" : setVal.getClass().getSimpleName(),
                "colNm", colNm,
                "cls", cls.getSimpleName(),
                "expctObj", 
                    expctObj == null 
                        ? "null" : expctObj.getClass().getSimpleName()
            );
        }

        private
        void
        insertVal( Connection conn,
                   String id )
            throws Exception
        {
            state.equalInt( 
                1,
                Sql.executeUpdate( conn,
                    "insert into sql_impl3 ( test_id, " + colNm + " ) " +
                    "values( ?, ? )", id, setVal
                )
            );
        }

        protected
        void
        assertVal( Object expct,
                   Object val,
                   Class< ? > expctCls )
        {
            if ( expctCls.equals( Timestamp.class ) ||
                 expctCls.equals( java.sql.Date.class ) ||
                 expctCls.equals( Date.class ) )
            {
                state.equal(
                    ( (Date) expct ).getTime() / 1000,
                    ( (Date) val ).getTime() / 1000
                );
            }
            else if ( expctCls.equals( byte[].class ) )
            {
                state.equal(
                    ByteBuffer.wrap( (byte[]) expct ),
                    ByteBuffer.wrap( (byte[]) val )
                );
            }
            else state.equal( expct, val );
        }

        protected
        void
        useConnection( Connection conn )
            throws Exception
        {
            String id = Lang.randomUuid();
            insertVal( conn, id );

            PreparedStatement ps = 
                conn.prepareStatement( 
                    "select " + colNm + " as `val` from sql_impl3 " +
                    "where test_id = ?" );
 
            try
            {
                ps.setString( 1, id );
                ResultSet rs = ps.executeQuery();
                try
                {
                    state.isTrue( rs.next() );
                    Object o1 = Sql.getObject( rs, "val", cls );
                    state.equal( o1, Sql.getObject( rs, 1, cls ) );
                    assertVal( expctObj, o1, cls );
                }
                finally { rs.close(); }
            }
            finally { ps.close(); }
        }
    }

    @InvocationFactory
    private
    List< GetObjectTest >
    testGetObject()
    {
        Timestamp ts = Timestamp.valueOf( "2001-01-01 12:00:00" );
        Date dt = new Date( ts.getTime() );

        List< GetObjectTest > res = Lang.newList();

        // basic string --> {string,int,decimal,bool,Date}
        res.add( new GetObjectTest( "hello", "str1", String.class, "hello" ) );
        res.add( new GetObjectTest( "12", "str1", Integer.class, 12 ) );
        res.add( new GetObjectTest( "12", "str1", Integer.TYPE, 12 ) );
        res.add( new GetObjectTest( "12.1", "str1", Double.class, 12.1d ) );
        res.add( new GetObjectTest( "12.1", "str1", Double.TYPE, 12.1d ) );
        res.add( new GetObjectTest( "true", "str1", Boolean.class, true ) );
        res.add( new GetObjectTest( "false", "str1", Boolean.TYPE, false ) );
        res.add( new GetObjectTest( "0", "str1", Boolean.class, false ) );
        res.add( new GetObjectTest( ts, "str1", Timestamp.class, ts ) );
        res.add( new GetObjectTest( ts, "str1", Date.class, dt ) );

        // int --> {byte,short,int,long,bigint,float,double,bigdec,string,bool}
        res.add( new GetObjectTest( 12, "long1", Boolean.class, true ) );
        res.add( new GetObjectTest( 12, "long1", Byte.class, (byte) 12 ) );
        res.add( new GetObjectTest( 12, "long1", Byte.TYPE, (byte) 12 ) );
        res.add( new GetObjectTest( 12, "long1", Short.class, (short) 12 ) );
        res.add( new GetObjectTest( 12, "long1", Short.TYPE, (short) 12 ) );
        res.add( new GetObjectTest( 12, "long1", Integer.class, 12 ) );
        res.add( new GetObjectTest( 12, "long1", Integer.TYPE, 12 ) );
        res.add( new GetObjectTest( 12, "long1", Long.class, 12L ) );
        res.add( new GetObjectTest( 12, "long1", Long.TYPE, 12L ) );

        res.add( 
            new GetObjectTest( 
                12, "long1", BigInteger.class, new BigInteger( "12" ) ) );

        res.add( new GetObjectTest( 12, "long1", Float.class, 12f ) );
        res.add( new GetObjectTest( 12, "long1", Float.TYPE, 12f ) );
        res.add( new GetObjectTest( 12, "long1", Double.class, 12d ) );
        res.add( new GetObjectTest( 12, "long1", Double.TYPE, 12d ) );

        res.add(
            new GetObjectTest(
                12, "long1", BigDecimal.class, new BigDecimal( "12" ) ) );
        
        // decimal --> {byte,short,int,long.bigint,float,double,bigdec}
        res.add( new GetObjectTest( 12.1d, "double1", Boolean.TYPE, true ) );
        res.add( new GetObjectTest( 12.1d, "double1", Byte.class, (byte) 12 ) );
        res.add( new GetObjectTest( 12.1d, "double1", Byte.TYPE, (byte) 12 ) );
        res.add( 
            new GetObjectTest( 12.1d, "double1", Short.class, (short) 12 ) );

        res.add( 
            new GetObjectTest( 12.1d, "double1", Short.TYPE, (short) 12 ) );

        res.add( new GetObjectTest( 12.1d, "double1", Integer.class, 12 ) );
        res.add( new GetObjectTest( 12.1d, "double1", Integer.TYPE, 12 ) );
        res.add( new GetObjectTest( 12.1d, "double1", Long.class, 12L ) );
        res.add( new GetObjectTest( 12.1d, "double1", Long.TYPE, 12L ) );
        res.add( new GetObjectTest( 12.1d, "double1", Float.class, 12.1f ) );
        res.add( new GetObjectTest( 12.1d, "double1", Float.TYPE, 12.1f ) );
        res.add( new GetObjectTest( 12.1d, "double1", Double.class, 12.1d ) );
        res.add( new GetObjectTest( 12.1d, "double1", Double.TYPE, 12.1d ) );

        res.add(
            new GetObjectTest(
                12.1d, "double1", BigInteger.class, new BigInteger( "12" ) ) );
        
        res.add(
            new GetObjectTest(
                12.1d, "double1", BigDecimal.class, new BigDecimal( "12.1" ) )
        );

        // bool --> bool
        res.add( new GetObjectTest( true, "bool1", Boolean.class, true ) );
        res.add( new GetObjectTest( true, "bool1", Boolean.TYPE, true ) );

        // {int,decimal,timestamp,bool} --> string
        res.add( new GetObjectTest( 12, "long1", String.class, "12" ) );
        res.add( new GetObjectTest( 12.1d, "double1", String.class, "12.1" ) );

        res.add( 
            new GetObjectTest( ts, "timestamp1", String.class, 
            ts.toString() ) );

        // byte[] --> byte[], ByteBuffer
        byte[] arr = IoTestFactory.nextByteArray( 100 );
        ByteBuffer buf = ByteBuffer.wrap( arr );
        res.add( new GetObjectTest( arr, "blob1", byte[].class, arr ) );
        res.add( new GetObjectTest( arr, "blob1", ByteBuffer.class, buf ) );

        return res;
    }

    @TestWithConn
    private
    void
    testRollUpMap( Connection conn )
        throws Exception
    {
        int len = 10;
        String testId = Lang.randomUuid();

        insertSqlImpl3Rows( conn, testId, len );
        
        Map< String, Long > m =
            Sql.rollUpMap( conn,
                "str1", String.class,
                "long1", Long.class,
                "select str1, long1 from sql_impl3 where test_id = ?",
                testId
            );
 
        state.equalInt( len, m.size() );

        for ( int i = 0; i < len; ++i )
        {
            state.equal( 
                Long.valueOf( i ), state.get( m, makeStringData( i ), "m" ) );
        }
    }
    
    @TestWithConn
    private
    void
    testSelectListOfMaps( Connection conn )
        throws Exception
    {
        String testId = Lang.randomUuid();
        int len = 100;

        insertSqlImpl3Rows( conn, testId, len );
        assertSqlImpl3Rows( conn, testId, len );
    }

    // Just get some basic coverage of SqlUpdateOperation
    @TestWithConn
    private
    void
    testSqlUpdateOperation( Connection conn )
        throws Exception
    {
        String uid = Lang.randomUuid();
        String data1 = "hello";

        Sql.executeUpdate( conn,
            "insert into sql_impl1( uid, data1 ) values ( ?, ? )",
            uid, data1
        );

        state.equalString( 
            data1, 
            Sql.expectString( conn, 
                "select data1 from sql_impl1 where uid = ?", uid )
        );
    }
        
    private
    void
    assertColumnDescriptor( Set< String > expctCols,
                            SqlColumnDescriptor col )
    {
        String nm = col.getName();
        state.remove( expctCols, nm, "expctCols" );

        if ( nm.equals( "test_id" ) )
        {
            state.equal( SqlType.CHAR, col.getSqlType() );
            state.equalInt( 36, col.getSize() );
        }
        else if ( nm.equals( "long1" ) )
        {
            state.equal( SqlType.BIGINT, col.getSqlType() );
        }
        else if ( nm.equals( "str1" ) )
        {
            state.equal( SqlType.VARCHAR, col.getSqlType() );
            state.equalInt( 255, col.getSize() );
        }
        else if ( nm.equals( "double1" ) )
        {
            state.equal( SqlType.DOUBLE, col.getSqlType() );
        }
        else if ( nm.equals( "bool1" ) )
        {
            state.equal( SqlType.BIT, col.getSqlType() );
        }
        else if ( nm.equals( "blob1" ) )
        {
            state.equal( SqlType.BINARY, col.getSqlType() );
            state.equalInt( 255, col.getSize() );
        }
        else if ( nm.equals( "timestamp1" ) || nm.equals( "datetime1" ) )
        {
            state.equal( SqlType.TIMESTAMP, col.getSqlType() );
        }
        else if ( nm.equals( "date1" ) )
        {
            state.equal( SqlType.DATE, col.getSqlType() );
        }
        else if ( nm.equals( "time1" ) )
        {
            state.equal( SqlType.TIME, col.getSqlType() );
        }
        else state.fail( "Unexpected col:", nm );
    }

    @TestWithConn
    private
    void
    testGetSqlTableDescriptor( Connection conn )
        throws Exception
    {
        String tbl = "sql_impl3";
        SqlTableDescriptor td = Sql.getTableDescriptor( tbl, conn );

        state.equal( tbl, td.getName() );
        state.equal( conn.getCatalog(), td.getCatalog() );

        Set< String > expctCols = 
            Lang.newSet( 
                "test_id", "long1", "str1", "double1", "bool1", "blob1",
                "timestamp1", "datetime1", "date1", "time1" );

        for ( SqlColumnDescriptor col : td.getColumns() )
        {
            assertColumnDescriptor( expctCols, col );
        }
        
        state.isTrue( expctCols.isEmpty() );
    }

    private
    void
    assertSqlImpl4Data( long id,
                        String data,
                        Connection conn )
        throws Exception
    {
        String data2 = 
            Sql.expectString( conn,
                "select data from sql_impl4 where id = ?", id );

        state.equalString( data, data2 );
    }

    @TestWithConn
    private
    void
    testInsertAndGetKeys( Connection conn )
        throws Exception
    {
        String data = Lang.randomUuid();

        Map< String, Object > keys = 
            Sql.insertAndGetKeys( conn, 
                "insert into sql_impl4 ( data ) values ( ? )", data );
        
        long id = state.cast( 
            Long.class, 
            state.get( keys, Sql.COL_GENERATED_KEY, "keys" ) 
        );

        assertSqlImpl4Data( id, data, conn );
    }

    @TestWithConn
    private
    void
    testInsertAndGetKey( Connection conn )
        throws Exception
    {
        String data = Lang.randomUuid();

        long id = Sql.insertAndGetKey( conn, Long.class,
            "insert into sql_impl4 ( data ) values ( ? )", data );

        assertSqlImpl4Data( id, data, conn );
    }
    
    private
    abstract
    class Sql1StatementsTest
    extends AbstractSqlTest
    implements LabeledTestObject,
               TestFailureExpector
    {
        public
        final
        CharSequence
        getLabel()
        {
            return getClass().getSimpleName(); 
        }

        abstract
        void
        useConnection( SqlTableDescriptor td,
                       SqlStatementWriter w,
                       String uid,
                       String data1,
                       Connection conn )
            throws Exception;

        protected
        final
        void
        useConnection( Connection conn )
            throws Exception
        {
            useConnection(
                Sql.getTableDescriptor( "sql_impl1", conn ),
                Sql.createStatementWriter( conn ),
                Lang.randomUuid(),
                Lang.randomUuid(),
                conn
            );
        }

        public
        Class< ? extends Throwable >
        expectedFailureClass()
        {
            return null;
        }

        public CharSequence expectedFailurePattern() { return null; }
    }

    private
    final
    class GetInsertNoUserIgnoreTest
    extends Sql1StatementsTest
    {
        @Override
        public
        Class< ? extends Throwable >
        expectedFailureClass()
        {
            return SQLIntegrityConstraintViolationException.class;
        }

        void
        useConnection( SqlTableDescriptor td,
                       SqlStatementWriter w,
                       String uid,
                       String data1,
                       Connection conn )
            throws Exception
        {
            String sql = SqlStatements.getInsert( td, false, w );

            Sql.executeUpdate( conn, sql, uid, data1, 0 );

            state.equalString( 
                data1, 
                Sql.expectString( conn, 
                    "select data1 from sql_impl1 where uid = ?", uid )
            );
            
            Sql.executeUpdate( conn, sql, uid, data1, 0 ); // should fail
        }
    }

    private
    final
    class GetInsertUseIgnoreTest
    extends Sql1StatementsTest
    {
        void
        useConnection( SqlTableDescriptor td,
                       SqlStatementWriter w,
                       String uid,
                       String data1,
                       Connection conn )
            throws Exception
        {
            // insert fresh data and then check that useIgnore leads to an
            // 'insert ignore' which leaves prev data intact
            for ( int i = 0; i < 2; ++i )
            {
                Sql.executeUpdate( conn,
                    SqlStatements.getInsert( td, i == 1, w ), 
                    uid, i == 0 ? data1 : Lang.randomUuid(), i
                );
    
                state.equalString(
                    data1,
                    Sql.expectString( conn, 
                        "select data1 from sql_impl1 where uid = ? " +
                        "and data2 = 0", 
                        uid ) );
            }
        }
    }

    private
    final
    class GetUpsertHandlesDupKeysTest
    extends Sql1StatementsTest
    {
        void
        useConnection( SqlTableDescriptor td,
                       SqlStatementWriter w,
                       String uid,
                       String data1,
                       Connection conn )
            throws Exception
        {
            String sql = SqlStatements.getUpsert( td, w );

            for ( int i = 0; i < 2; ++i )
            {
                Sql.executeUpdate( conn, sql, uid, data1, i );

                String selSql =
                    "select data2 from sql_impl1 where uid = ? and data1 = ?";

                state.equalInt( i, Sql.selectInt( conn, selSql, uid, data1 ) );
            }
        }
    }

    @InvocationFactory
    private
    List< Sql1StatementsTest >
    testSql1Statements()
    {
        return Lang.< Sql1StatementsTest >asList(
            new GetInsertNoUserIgnoreTest(),
            new GetInsertUseIgnoreTest(),
            new GetUpsertHandlesDupKeysTest()
        );
    }

    @TestWithConn
    private
    void
    testTableDescriptorFromView( Connection conn )
        throws Exception
    {
        SqlTableDescriptor td = 
            Sql.getTableDescriptor( "sql_impl1_view1", conn );
        
        state.equalString( "sql_impl1_view1", td.getName() );
        state.equalString( conn.getCatalog(), td.getCatalog() );

        state.equalInt( 2, td.getColumns().size() );
        for ( SqlColumnDescriptor col : td.getColumns() )
        {
            if ( col.getName().equals( "uid" ) )
            {
                state.equal( SqlType.CHAR, col.getSqlType() );
                state.equalInt( 36, col.getSize() );
            }
            else if ( col.getName().equals( "data1" ) )
            {
                state.equal( SqlType.VARCHAR, col.getSqlType() );
                state.equalInt( 255, col.getSize() );
            }
            else state.fail( "Unexpected col:", col.getName() );
        }
    }

    private
    final
    class UpdateExecutorTest
    extends AbstractSqlTest
    {
        private final String uid = Lang.randomUuid();
        private final int batchLen = 100;

        private
        List< Integer >
        createBatch()
        {
            return new BatchImpl< Integer >( batchLen ) {
                public Integer get( int i ) { return i; }
            };
        }

        private
        final
        class MapperImpl
        implements SqlParameterMapper< Integer, Void >
        {
            public
            boolean
            setParameters( Integer obj,
                           Void mapperObj,
                           PreparedStatement ps,
                           int offset )
                throws Exception
            {
                ps.setString( offset + 1, uid );
                ps.setInt( offset + 2, obj );

                return true;
            }
        }

        private
        void
        checkInserts( Connection conn )
            throws Exception
        {
            Number sum = (Number) Sql.selectOne( conn, 
                "select sum( data1 ) from sql_impl2 where uid = ?", uid );

            state.equalInt(
                TestSums.ofSequence( 0, batchLen ), sum.intValue() );

            Number num = (Number) Sql.selectOne( conn,
                "select sum( data1 ) from sql_impl2 where uid = ?", uid );
            
            state.equalInt( 
                TestSums.ofSequence( 0, batchLen ), num.intValue() );
        }

        protected
        void
        useConnection( Connection conn )
            throws Exception
        {
            new SqlUpdateExecutor.Builder< Integer, Void >().
                setBatch( createBatch() ).
                setMapper( new MapperImpl() ).
                setSql( "insert into sql_impl2( uid, data1 ) values ( ?, ? )" ).
                setDataSource( dataSource() ).
                build().
                start();
            
            checkInserts( conn );
        }
    }
    
    @Test
    private
    void
    testUpdateExecutor()
        throws Exception
    {
        new UpdateExecutorTest().call();
    }

    private
    static
    class SqlImpl3RowMapper
    extends AbstractSqlParameterMapper< Map< String, Object >, Integer >
    {
        private
        SqlImpl3RowMapper( SqlParameterGroupDescriptor grp )
        {
            super( grp );
        }

        protected
        boolean
        setParameter( Map< String, Object > obj,
                      Integer mapObj,
                      SqlParameterDescriptor param,
                      PreparedStatement ps,
                      int indx )
            throws Exception
        {
            Object val; 

            if ( param.getName().equals( "long1" ) && mapObj != null )
            {
                val = getInt( obj, "long1" ) + mapObj;
            }
            else val = state.get( obj, param.getName(), "obj" );

            ps.setObject( indx, val );

            return true;
        }
    }

    private
    abstract
    class AbstractSqlParameterMapperTest
    extends AbstractSqlTest
    {
        final String testId = Lang.randomUuid();

        final int batchLen = 100;

        private
        List< Map< String, Object > >
        createBatch()
        {
            return new BatchImpl< Map< String, Object > >( batchLen ) 
            {
                public Map< String, Object > get( int i ) 
                {
                    return
                        Lang.newMap( String.class, Object.class,
                            "test_id", testId,
                            "long1", makeLongData( i ),
                            "str1", makeStringData( i ),
                            "double1", makeDoubleData( i ),
                            "bool1", makeBooleanData( i ),
                            "blob1", makeBlobData( i ),
                            "timestamp1", makeTimestampData( i ),
                            "datetime1", makeTimestampData( i ),
                            "date1", makeDateData( i ),
                            "time1", makeTimeData( i )
                        );
                }
            };
        }

        void
        assertRows( Connection conn )
            throws Exception
        {
            assertSqlImpl3Rows( conn, testId, batchLen );
        }

        SqlParameterMapper< Map< String, Object >, Integer >
        createMapper( SqlParameterGroupDescriptor grp )
        {
            return new SqlImpl3RowMapper( grp );
        }

        void
        buildUpdateExecutor( 
            SqlUpdateExecutor.Builder< Map< String, Object >, Integer > b )
        {}

        private
        SqlUpdateExecutor.Builder< Map< String, Object >, Integer >
        initUpdateBuilder( Connection conn )
            throws Exception
        {
            SqlTableDescriptor td = 
                Sql.getTableDescriptor( "sql_impl3", conn );
            
            SqlParameterGroupDescriptor grp = Sql.getInsertParametersFor( td );

            String sql = 
                SqlStatements.getInsert(
                    td, false, Sql.createStatementWriter( conn ) );
            
            SqlUpdateExecutor.Builder< Map< String, Object >, Integer > res =
                new SqlUpdateExecutor.Builder< Map< String, Object >,
                                               Integer >().
                    setSql( sql ).
                    setMapper( createMapper( grp ) );
            
            buildUpdateExecutor( res );

            return res;
        }

        protected
        void
        useConnection( Connection conn )
            throws Exception
        {
            initUpdateBuilder( conn ).
                setDataSource( dataSource() ).
                setBatch( createBatch() ).
                build().
                start();
            
            assertRows( conn );
        }
    }

    @Test
    private
    final
    class BasicAbstractSqlParameterMapperTest
    extends AbstractSqlParameterMapperTest
    {}

    // Ensure that if any single call to
    // AbstractSqlParameterMapper.setParameter() returns false then then entire
    // record is omitted
    @Test
    private
    final
    class PartialBatchElidedSqlMapperTest
    extends AbstractSqlParameterMapperTest
    {
        private final int elideAt = 7;

        @Override
        SqlParameterMapper< Map< String, Object >, Integer >
        createMapper( SqlParameterGroupDescriptor grp )
        {
            return new SqlImpl3RowMapper( grp ) 
            {
                @Override
                protected
                boolean
                setParameter( Map< String, Object > m,
                              Integer o,
                              SqlParameterDescriptor param,
                              PreparedStatement ps,
                              int indx )
                    throws Exception
                {
                    if ( getInt( m, "long1" ) % elideAt == 0 &&
                         param.getName().equals( "str1" ) )
                    {
                        return false; // skip this col and this record
                    }
                    else return super.setParameter( m, o, param, ps, indx );
                }
            };
        }

        @Override
        void
        assertRows( Connection conn )
            throws Exception
        {
            Map< String, Object > m = 
                Sql.expectOneMap( conn,
                    "select count( long1 ) as cnt, sum( long1 ) as `sum` " +
                    "from sql_impl3 where test_id = ?", testId
                );
            
            int elideCnt = Lang.ceilI( batchLen, elideAt );

            state.equalInt( batchLen - elideCnt, getInt( m, "cnt" ) );

            int fullSum = TestSums.ofSequence( 0, batchLen );
            int elidedSum = TestSums.ofSequence( 0, elideCnt ) * elideAt;

            state.equalInt( fullSum - elidedSum, getInt( m, "sum" ) );
        }
    }

    // ends up hopefully just being another form of the test covered by
    // MapperWithNoEffectiveUpdatesTest, but worth doubly covering via an
    // instance of AbstractSqlParameterMapper
    @Test
    private
    final
    class AllRecordsElidedAbstractSqlParameterMapperTest
    extends AbstractSqlParameterMapperTest
    {
        @Override
        SqlParameterMapper< Map< String, Object >, Integer >
        createMapper( SqlParameterGroupDescriptor grp )
        {
            return new SqlImpl3RowMapper( grp )
            {
                @Override
                protected
                boolean
                setParameter( Map< String, Object > m,
                              Integer ignored,
                              SqlParameterDescriptor param,
                              PreparedStatement ps,
                              int indx )
                {
                    return false;
                }
            };
        }

        @Override
        void
        assertRows( Connection conn )
            throws Exception
        {
            List< ? > l =
                Sql.selectListOfMaps( conn,
                    "select * from sql_impl3 where test_id = ?", testId );
            
            state.isTrue( l.isEmpty() );
        }
    }

    // just get coverage that mapperObj is passed through correctly to each
    // invocation of setParameter()
    @Test
    private
    final
    class MapperObjPassedToSetParameterTest
    extends AbstractSqlParameterMapperTest
    {
        private final int addend = 1000;

        @Override
        void
        buildUpdateExecutor( 
            SqlUpdateExecutor.Builder< Map< String, Object >, Integer > b )
        {
            b.setMapperObjectGenerator(
                new SqlUpdateExecutor.MapperObjectGenerator< 
                                        Map< String, Object >, Integer >() 
                {
                    public
                    Integer
                    generateMapperObject( 
                        Collection< ? extends Map< String, Object > > coll,
                        Connection conn )
                    {
                        return addend;
                    }
                }
            );
        }

        @Override
        void
        assertRows( Connection conn )
            throws Exception
        {
            state.equalInt(
                ( addend * batchLen ) + TestSums.ofSequence( 0, batchLen ),
                Sql.expectInt( conn, 
                    "select sum( long1 ) from sql_impl3 where test_id = ?",
                     testId
                )
            );
        }
    }

    private
    abstract
    class AbstractSqlUpdateExecutorTest< M >
    extends AbstractSqlTest
    {
        final String testId = Lang.randomUuid();
        final int batchLen = 100;

        private
        List< Integer >
        createBatch()
        {
            return new BatchImpl< Integer >( batchLen ) {
                public Integer get( int i ) { return i; }
            };
        }

        abstract
        boolean
        setParameters( Integer obj,
                       M mapperObj,
                       PreparedStatement ps,
                       int offset )
            throws Exception;

        private
        final
        class MapperImpl
        implements SqlParameterMapper< Integer, M >
        {
            public
            boolean
            setParameters( Integer obj,
                           M mapperObj,
                           PreparedStatement ps,
                           int offset )
                throws Exception
            {
                return AbstractSqlUpdateExecutorTest.this.
                    setParameters( obj, mapperObj, ps, offset );
            }
        }

        abstract
        void
        assertSum( Number sum );

        private
        void
        assertRows( Connection conn )
            throws Exception
        {
            assertSum(
                (Number) Sql.selectOne( 
                    conn,
                    "select sum( data1 ) from sql_impl2 where uid = ?", 
                    testId
                )
            );
        }

        void buildUpdateExecutor( SqlUpdateExecutor.Builder< Integer, M > b ) {}

        protected
        void
        useConnection( Connection conn )
            throws Exception
        {
            SqlUpdateExecutor.Builder< Integer, M > b =
                new SqlUpdateExecutor.Builder< Integer, M >().
                    setDataSource( dataSource() ).
                    setSql( 
                        "insert into sql_impl2( uid, data1 ) values( ?, ? )" ).
                    setMapper( new MapperImpl() ).
                    setBatch( createBatch() );
            
            buildUpdateExecutor( b );
            b.build().start();

            assertRows( conn );
        }
    }

    @Test
    private
    final
    class MapperObjectGeneratorTest
    extends AbstractSqlUpdateExecutorTest< Integer >
    {
        private
        final
        class MapperObjGenerator
        implements SqlUpdateExecutor.MapperObjectGenerator< Integer, Integer >
        {
            public
            Integer
            generateMapperObject( Collection< ? extends Integer > batch,
                                  Connection conn )
            {
                int i = 0;
                for ( Integer i2 : batch ) i += i2;
                
                return i;
            }
        }
            
        boolean
        setParameters( Integer obj,
                       Integer mapperObj,
                       PreparedStatement ps,
                       int offset )
            throws Exception
        {
            ps.setString( offset + 1, testId );
            ps.setInt( offset + 2, obj.intValue() + mapperObj );

            return true;
        }

        void
        assertSum( Number sum )
        {
            int mapperObj = TestSums.ofSequence( 0, batchLen );
            state.equalInt( mapperObj * ( 1 + batchLen ), sum.intValue() );
        }

        void
        buildUpdateExecutor( SqlUpdateExecutor.Builder< Integer, Integer > b )
        {
            b.setMapperObjectGenerator( new MapperObjGenerator() );
        }
    }

    @Test
    private
    final
    class MapperWithNoEffectiveUpdatesTest
    extends AbstractSqlUpdateExecutorTest< Void >
    {
        boolean
        setParameters( Integer obj,
                       Void ignored,
                       PreparedStatement ps,
                       int offset )
        {
            return false;
        }

        void
        assertSum( Number sum )
        {
            state.isTrue( sum == null );
        }
    }

    private
    void
    addBlobTests( List< SetValueTest > l )
    {
        byte[] arr = IoTestFactory.nextByteArray( 100 );
        ByteBuffer buf = ByteBuffer.wrap( arr );

        l.add( new SetValueTest( arr, SqlType.BLOB, "blob1", buf ) );
        l.add( new SetValueTest( buf.slice(), SqlType.BLOB, "blob1", buf ) );

        // test that Sql.setValue() pays attention to pos/limit whey reaching
        // into ByteBuffer.array()
        ByteBuffer buf2 = ByteBuffer.allocate( arr.length + 10 );
        buf2.position( 5 );
        buf2.limit( buf2.position() + arr.length );
        buf2.slice().put( arr ); // set actual data without changing pos/limit
        l.add( 
            new SetValueTest( 
                buf2, SqlType.BLOB, "blob1", buf, "internalData" ) );

        ByteBuffer dirBuf = ByteBuffer.allocateDirect( arr.length );
        dirBuf.put( arr );
        dirBuf.position( 0 );
        l.add( 
            new SetValueTest( dirBuf, SqlType.BLOB, "blob1", buf, "dirBuf" ) );
    }

    private
    void
    addTimeTests( List< SetValueTest > l )
    {
        Timestamp ts = Timestamp.valueOf( "2011-01-01 12:00:00.0" );
        java.sql.Date dt = java.sql.Date.valueOf( "2011-01-01" );

        l.add( new SetValueTest( ts, SqlType.TIMESTAMP, "timestamp1", ts ) );
        l.add( new SetValueTest( ts, SqlType.DATE, "timestamp1", dt ) );
        l.add( new SetValueTest( dt, SqlType.TIMESTAMP, "timestamp1", dt ) );
        l.add( new SetValueTest( dt, SqlType.DATE, "timestamp1", dt ) );

        l.add( 
            new SetValueTest( 
                ts.toString(), SqlType.VARCHAR, "timestamp1", ts ) );
        
        l.add( new SetValueTest( ts, SqlType.TIMESTAMP, "datetime1", ts ) );
        l.add( new SetValueTest( ts, SqlType.DATE, "datetime1", dt ) );
        l.add( new SetValueTest( dt, SqlType.TIMESTAMP, "datetime1", dt ) );
        l.add( new SetValueTest( dt, SqlType.DATE, "datetime1", dt ) );

        l.add( new SetValueTest( ts, SqlType.TIMESTAMP, "date1", dt ) );
        l.add( new SetValueTest( ts, SqlType.DATE, "date1", dt ) );
        l.add( new SetValueTest( dt, SqlType.TIMESTAMP, "date1", dt ) );
        l.add( new SetValueTest( dt, SqlType.DATE, "date1", dt ) );
        
        Time tm = Time.valueOf( "12:00:00" );
        l.add( new SetValueTest( ts, SqlType.TIMESTAMP, "time1", tm ) );
        l.add( new SetValueTest( ts, SqlType.TIME, "time1", tm ) );
        l.add( new SetValueTest( tm, SqlType.TIME, "time1", tm ) );
    }

    private
    final
    static
    class CustomCharSeq
    implements CharSequence
    {
        private final CharSequence s;

        private CustomCharSeq( CharSequence s ) { this.s = s; }

        public int length() { return s.length(); }
        public String toString() { return s.toString(); }
        public char charAt( int i ) { return s.charAt( i ); }

        public
        CharSequence
        subSequence( int start,
                     int end )
        {
            return s.subSequence( start, end );
        }
    }

    // the list we build here isn't necessarily meant to be exhaustive of every
    // possible combination of calls, nor is it expected to remain fixed. for
    // now we cover most of the basics, including primitives and some of the
    // trickier types of values, including blobs, date/time, and custom
    // charseqs.
    @InvocationFactory
    private
    List< SetValueTest >
    testSetValue()
    {
        List< SetValueTest > res = Lang.newList();
        
        res.add( new SetValueTest( "hello", SqlType.CHAR, "str1", "hello" ) );

        res.add( new SetValueTest( 
            new StringBuilder( "hello" ), SqlType.CHAR, "str1", "hello" ) );

        res.add( new SetValueTest( true, SqlType.BOOLEAN, "bool1", true ) );

        res.add( 
            new SetValueTest( 
                new CustomCharSeq( "hello" ), SqlType.VARCHAR, "str1", "hello" )
        );

        res.add( new SetValueTest(
            new BigInteger( "42" ), SqlType.INTEGER, "long1", 42 ) );
            
        res.add( new SetValueTest(
            new BigInteger( "42" ), SqlType.DOUBLE, "double1", 42.0d ) );

        res.add( new SetValueTest(
            new BigInteger( "42" ), SqlType.CHAR, "str1", "42" ) );
           
        res.add( new SetValueTest(
            new BigDecimal( "42.0" ), SqlType.DOUBLE, "double1", 42.0d ) );

        res.add( new SetValueTest(
            new BigDecimal( "42.0" ), SqlType.INTEGER, "long1", 42 ) );

        res.add( new SetValueTest( null, SqlType.CHAR, "str1", null ) );

        res.add( new SetValueTest( null, SqlType.BIGINT, "long1", null ) );

        addBlobTests( res );
        addTimeTests( res );
        
        return res;
    }

    @Test( expected = IllegalArgumentException.class,
           expectedPattern = "sqlTyp is NULL but val is: hello" )
    private
    final
    class SetNonNullValueWithSqlTypeNullFailsTest
    extends AbstractSqlTest
    {
        protected
        void
        useConnection( Connection conn )
            throws Exception
        {
            PreparedStatement ps = conn.prepareStatement( "select ?" );
            try { Sql.setValue( "hello", SqlType.NULL, ps, 1 ); }
            finally { ps.close(); }
        }
    }

    private
    final
    static
    class StringType1
    extends TypedString< StringType1 >
    {
        private StringType1( String s ) { super( s ); }
    }

    private
    final
    static
    class StringType2
    extends CaseInsensitiveTypedString< StringType2 >
    {
        private StringType2( String s ) { super( s ); }
    }

    @TestWithConn
    private
    void
    testDefaultBindTypedString( Connection conn )
        throws Exception
    {
        String uid1 = Lang.randomUuid();
        String uid2 = Lang.randomUuid();

        Sql.executeUpdate( conn,
            "insert into sql_impl1 ( uid, data1 ) values ( ?, ? ), ( ?, ? )",
            uid1, new StringType1( "hello" ),
            uid2, new StringType2( "hello" )
        );
        
        for ( String uid : new String[] { uid1, uid2 } )
        {
            state.equalString( "hello",
                Sql.expectString( conn,
                    "select data1 from sql_impl1 where uid = ?", uid ) );
        }
    }

    @TestWithConn
    private
    void
    testTransactionUserBasic( Connection conn )
        throws Exception
    {
        conn.setAutoCommit( true );

        state.equalInt(
            1,
            Sql.useConnection(
                conn,
                Sql.asTransactionUser( 
                    new ConnectionUser< Integer >() {
                        public Integer useConnection( Connection conn ) 
                            throws Exception 
                        {
                            state.isFalse( conn.getAutoCommit() );
                            return Sql.expectInt( conn, "select 1" );
                        }
                    }
                )
            )
        );

        state.isTrue( conn.getAutoCommit() );
    }

    private
    final
    static
    class RollbackTestUser
    extends VoidOp
    {
        private final String uid = Lang.randomUuid();
        private boolean didFirstUpdate;

        public 
        void 
        execute( Connection conn ) 
            throws Exception 
        {
            String sql = "insert into sql_impl1 ( uid, data1 ) values ( ?, ? )";

            Sql.executeUpdate( conn, sql, uid, "test" );
            didFirstUpdate = true;
            Sql.executeUpdate( conn, sql, uid, "test" ); // should fail: dupKey
        }
    }

    @TestWithConn
    private
    void
    testTransactionUserRollback( Connection conn )
        throws Exception
    {
        RollbackTestUser u = new RollbackTestUser();

        try { Sql.asTransactionUser( u ).useConnection( conn ); }
        catch ( SQLIntegrityConstraintViolationException okay ) {}

        state.isTrue( u.didFirstUpdate );
        String sel = "select data1 from sql_impl1 where uid = ?";
        state.isTrue( Sql.selectOne( conn, sel, u.uid ) == null );
    }

    @TestFactory
    private
    static
    List< SqlImplTests >
    getTests()
        throws Exception
    {
        return SqlTests.createSqlSuite( SqlImplTests.class );
    }

    // to test:
    //
    //  - select list which returns an empty result set (and should be an empty
    //  list and never calling proc)
    //
    //  - sql update executor where mapper cols and parameter cols are not exact
    //  same set fails
    //
    //  - sql update executor where mapper ultimately doesn't set any values;
    //  make sure statement never gets executed
}
