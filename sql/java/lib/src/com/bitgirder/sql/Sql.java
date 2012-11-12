package com.bitgirder.sql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.io.IoUtils;

import java.math.BigDecimal;
import java.math.BigInteger;

import java.sql.Connection;
import java.sql.SQLException;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.ResultSetMetaData;
import java.sql.DatabaseMetaData;
import java.sql.Types;
import java.sql.Timestamp;

import javax.sql.DataSource;

import java.util.Map;
import java.util.List;
import java.util.Date;

import java.nio.ByteBuffer;

public
final
class Sql
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static MapRowProcessor MAP_ROW_PROCESSOR =
        new MapRowProcessor();
    
    private final static SqlColumnDescriptorRowProcessor SQL_COL_DESCRIP_PROC =
        new SqlColumnDescriptorRowProcessor();

    private Sql() {}

    // does null checks for public frontends
    private
    static
    PreparedStatement
    prepareAndBind( Connection conn,
                    CharSequence stmt,
                    Object... bindArgs )
        throws SQLException
    {
        inputs.notNull( conn, "conn" );
        inputs.notNull( stmt, "stmt" );

        // could contain null or be empty
        inputs.notNull( bindArgs, "bindArgs" ); 

        PreparedStatement st = conn.prepareStatement( stmt.toString() );

        try 
        { 
            bind( st, bindArgs ); 
            return st;
        }
        catch ( SQLException ex )
        {
            st.close();
            throw ex;
        }
    }

    public
    static
    ResultSet
    executeQuery( Connection conn,
                  CharSequence stmt,
                  Object... bindArgs )
        throws SQLException
    {
        PreparedStatement ps = prepareAndBind( conn, stmt, bindArgs );
        try { return ps.executeQuery(); } finally { ps.close(); }
    }

    private
    static
    enum ValType
    {
        OBJECT,
        INTEGER,
        LONG,
        STRING;
    }

    private
    static
    void
    assertNotMultiRow( ResultSet rs )
        throws SQLException
    {
        state.isFalse( rs.next(), "ResultSet has more than one row" );
    }

    private
    static
    ResultSet
    expectSingleResult( ResultSet rs,
                        Object... msg )
        throws SQLException
    {
        if ( rs.next() )
        {
            if ( rs.next() )
            {
                state.fail( 
                    Strings.join( " ", msg ) + ":", 
                    "result set has more than one row" );
            }
            else state.isTrue( rs.first() ); // now reset it for action
        }
        else
        {
            state.fail(
                Strings.join( " ", msg ) + ":",
                "result set has no rows" );
        }

        return rs;
    }
 
    private
    static
    < V >
    V
    expectRow( V res )
    {
        state.isFalse( res == null, "ResultSet has no rows" );
        return res;
    }

    // Some methods in this class have to declare themselves as throwing
    // Exception because that is what is thrown by interface methods used by
    // those methods, such as buildList() which calls RowProcessor.init() and
    // RowProcessor.processRow(). Sometimes though we call these methods
    // internally with processors which we know to throw only SQLException and
    // want to keep the public APIs of the calling methods limited to throwing
    // only SQLException. Since java doesn't support generic typing of thrown
    // exceptions, the method below is here to allow us to concisely express the
    // assertion in such situations that if an exception is encountered it is
    // either a SQLException or a RuntimeException. But, because java also
    // doesn't support union types, we can't type this method as returning
    // (SQLException|RuntimeException). Thus, this method returns a SQLException
    // or throws a RuntimeException. Deal with it.
    private
    static
    SQLException
    createSqlRethrow( Exception ex )
    {
        if ( ex instanceof SQLException ) return (SQLException) ex;
        else throw (RuntimeException) ex;
    }

    private
    static
    void
    bind( PreparedStatement st,
          Object[] bindArgs )
        throws SQLException
    {
        for ( int i = 0, e = bindArgs.length; i < e; ++i )
        {
            Object o = bindArgs[ i ];

            if ( o == null ) st.setNull( i + 1, Types.NULL );
            else st.setObject( i + 1, o );
        }
    }

    private
    static
    String[]
    getColumnLabels( ResultSet rs )
        throws SQLException
    {
        ResultSetMetaData md = rs.getMetaData();

        String[] res = new String[ md.getColumnCount() ];

        for ( int i = 0, e = res.length; i < e; ++i )
        {
            res[ i ] = md.getColumnLabel( i + 1 );
        }

        return res;
    }

    private
    static
    Map< String, Object >
    rowToMap( ResultSet rs,
              String[] colLabels )
        throws SQLException
    {
        Map< String, Object > res = Lang.newMap( colLabels.length );

        for ( int i = 0, e = colLabels.length; i < e; ++i )
        {
            Object val = rs.getObject( i + 1 );
            if ( val != null ) res.put( colLabels[ i ], val );
        }

        return res;
    }

    public
    static
    interface RowProcessor< V, I >
    {
        public
        I
        init( ResultSet rs )
            throws Exception;
        
        public
        V
        processRow( ResultSet rs,
                    I initObj )
            throws Exception;
    }

    public
    abstract
    static
    class DefaultRowProcessor< V >
    implements RowProcessor< V, Void >
    {
        public final Void init( ResultSet rs ) { return null; }

        protected
        abstract
        V
        implProcessRow( ResultSet rs )
            throws Exception;
        
        public
        final
        V
        processRow( ResultSet rs,
                    Void ign )
            throws Exception
        {
            return implProcessRow( rs );
        }
    }

    private
    final
    static
    class MapRowProcessor
    implements RowProcessor< Map< String, Object >, String[] >
    {
        public
        String[]
        init( ResultSet rs )
            throws SQLException
        {
            return getColumnLabels( rs );
        }

        public
        Map< String, Object >
        processRow( ResultSet rs,
                    String[] cols )
            throws SQLException
        {
            return rowToMap( rs, cols );
        }
    }

    public
    static
    < V, I >
    V
    selectObject( Connection conn,
                  CharSequence stmt,
                  RowProcessor< V, I > proc,
                  Object... bindArgs )
        throws Exception
    {
        inputs.notNull( proc, "proc" );
        PreparedStatement st = prepareAndBind( conn, stmt, bindArgs );

        try
        {
            ResultSet rs = st.executeQuery();
            try 
            {
                if ( rs.next() )
                {
                    V res = proc.processRow( rs, proc.init( rs ) );
                    assertNotMultiRow( rs );
        
                    return res;
                }
                else return null;
            }
            finally { rs.close(); }
        }
        finally { st.close(); }
    }

    public
    static
    < V, I >
    V
    expectObject( Connection conn,
                  CharSequence stmt,
                  RowProcessor< V, I > proc,
                  Object... bindArgs )
        throws Exception
    {
        return expectRow( selectObject( conn, stmt, proc, bindArgs ) );
    }

    public
    static
    Map< String, Object >
    selectOneMap( Connection conn,
                  CharSequence stmt,
                  Object... bindArgs )
        throws SQLException
    {
        try { return selectObject( conn, stmt, MAP_ROW_PROCESSOR, bindArgs ); }
        catch ( Exception ex ) { throw createSqlRethrow( ex ); }
    }

    public
    static
    Map< String, Object >
    expectOneMap( Connection conn,
                  CharSequence stmt,
                  Object... bindArgs )
        throws SQLException
    {
        return expectRow( selectOneMap( conn, stmt, bindArgs ) );
    }

    // test coverage provided now implicitly via selectListOfMaps
    public
    static
    < V, I >
    List< V >
    buildList( ResultSet rs,
               RowProcessor< ? extends V, I > proc,
               boolean closeResultSet )
        throws Exception
    {
        inputs.notNull( rs, "rs" );
        inputs.notNull( proc, "proc" );
        
        try
        {
            I initObj = proc.init( rs );
            List< V > res = Lang.newList();

            while ( rs.next() ) res.add( proc.processRow( rs, initObj ) );

            return res;
        }
        finally { if ( closeResultSet ) rs.close(); }
    }

    // will ultimately be made public
    public
    static
    < V, I >
    List< V >
    selectList( Connection conn,
                CharSequence stmt,
                RowProcessor< ? extends V, I > proc,
                Object... bindArgs )
        throws Exception
    {
        inputs.notNull( proc, "proc" );

        PreparedStatement ps = prepareAndBind( conn, stmt, bindArgs );
        try { return buildList( ps.executeQuery(), proc, true ); }
        finally { ps.close(); }
    }

    public
    static
    List< Map< String, Object > >
    selectListOfMaps( Connection conn,
                      CharSequence stmt,
                      Object... bindArgs )
        throws SQLException
    {
        try { return selectList( conn, stmt, MAP_ROW_PROCESSOR, bindArgs ); }
        catch ( Exception ex ) { throw createSqlRethrow( ex ); }
    }

    // We return something that is effectively of type V, either by explicitly
    // casting the result set value using cls, or by returning the type of
    // boxed object which would correspond to a safe cast later if it turns out
    // that cls refers to a primitive type (this would be an odd way to call the
    // rollUp method in general, but it is allowed)
    public
    static
    < V >
    V
    getObject( ResultSet rs,
               int col,
               Class< V > cls )
        throws SQLException
    {
        inputs.notNull( rs, "rs" );
        inputs.positiveI( col, "col" );
        inputs.notNull( cls, "cls" );

        Object obj;

        if ( cls.equals( BigInteger.class ) )
        {
            // could optimize this by skipping to getBigDecimal() step when
            // value is known to be more easily converted directly to bigint;
            // for now this is correct and we'll go with it
            obj = rs.getBigDecimal( col ).toBigInteger();
        }
        else if ( cls.equals( BigDecimal.class ) )
        {
            obj = rs.getBigDecimal( col );
        }
        else if ( cls.equals( Long.class ) || cls.equals( Long.TYPE ) )
        {
            obj = Long.valueOf( rs.getLong( col ) );
        }
        else if ( cls.equals( Integer.class ) || cls.equals( Integer.TYPE ) )
        {
            obj = Integer.valueOf( rs.getInt( col ) );
        }
        else if ( cls.equals( Short.class ) || cls.equals( Short.TYPE ) )
        {
            obj = Short.valueOf( rs.getShort( col ) );
        }
        else if ( cls.equals( Byte.class ) || cls.equals( Byte.TYPE ) )
        {
            obj = Byte.valueOf( rs.getByte( col ) );
        }
        else if ( cls.equals( Double.class ) || cls.equals( Double.TYPE ) )
        {
            obj = Double.valueOf( rs.getDouble( col ) );
        }
        else if ( cls.equals( Float.class ) || cls.equals( Float.TYPE ) )
        {
            obj = Float.valueOf( rs.getFloat( col ) );
        }
        else if ( cls.equals( Boolean.class ) || cls.equals( Boolean.TYPE ) )
        {
            obj = Boolean.valueOf( rs.getBoolean( col ) );
        }
        else if ( cls.equals( Timestamp.class ) ) obj = rs.getTimestamp( col );
        else if ( cls.equals( Date.class ) )
        {
            // Javadocs for java.sql.Timestamp say, "it is recommended that code
            // not view Timestamp values generically as an instance of
            // java.util.Date. The inheritance relationship between Timestamp
            // and java.util.Date really denotes implementation inheritance, and
            // not type inheritance." For that reason, we return a
            // java.util.Date rather than the Timestamp instance itself
            obj = new Date( rs.getTimestamp( col ).getTime() );
        }
        else if ( cls.equals( String.class ) ) obj = rs.getString( col );
        else if ( cls.equals( ByteBuffer.class ) )
        {
            obj = ByteBuffer.wrap( rs.getBytes( col ) );
        }
        else obj = cls.cast( rs.getObject( col ) );

        @SuppressWarnings( "unchecked" )
        V res = (V) obj;
        return res;
    }

    public
    static
    < V >
    V
    getObject( ResultSet rs,
               String colNm,
               Class< V > cls )
        throws SQLException
    {
        inputs.notNull( rs, "rs" );
        inputs.notNull( colNm, "colNm" );

        return getObject( rs, rs.findColumn( colNm ), cls );
    }

    private
    static
    < K, V >
    Map< K, V >
    buildRollUpMap( ResultSet rs,
                    String keyCol,
                    Class< K > keyColCls,
                    String valCol,
                    Class< V > valColCls )
        throws SQLException
    {
        Map< K, V > res = Lang.newMap( rs.getMetaData().getColumnCount() );

        while ( rs.next() )
        {
            res.put( 
                getObject( rs, keyCol, keyColCls ),
                getObject( rs, valCol, valColCls )
            );
        }

        return res;
    }

    public
    static
    < K, V >
    Map< K, V >
    rollUpMap( Connection conn,
               String keyCol,
               Class< K > keyColCls,
               String valCol,
               Class< V > valColCls,
               CharSequence stmt,
               Object... bindArgs )
        throws SQLException
    {
        inputs.notNull( keyCol, "keyCol" );
        inputs.notNull( keyColCls, "keyColCls" );
        inputs.notNull( valCol, "valCol" );
        inputs.notNull( valColCls, "valColCls" );

        PreparedStatement st = prepareAndBind( conn, stmt, bindArgs );
        try
        {
            ResultSet rs = st.executeQuery();
            try 
            { 
                return 
                    buildRollUpMap( rs, keyCol, keyColCls, valCol, valColCls );
            }
            finally { rs.close(); }
        }
        finally { st.close(); }
    }

    private
    static
    Object
    getValue( ResultSet rs,
              ValType vt )
        throws SQLException
    {
        switch ( vt )
        {
            case OBJECT: return rs.getObject( 1 );
            case INTEGER: return rs.getInt( 1 );
            case LONG: return rs.getLong( 1 );
            case STRING: return rs.getString( 1 );

            default: throw state.createFail( "Unexpected vt:", vt );
        }
    }

    private
    static
    Object
    getOneValue( ResultSet rs,
                 ValType vt )
        throws SQLException
    {
        if ( rs.next() )
        {
            Object res = getValue( rs, vt );
            assertNotMultiRow( rs );

            return res;
        }
        else return null;
    }

    // does null checks for public method params
    private
    static
    Object
    doSelectOne( Connection conn,
                 CharSequence stmt,
                 Object[] bindArgs,
                 ValType vt )
        throws SQLException
    {
        PreparedStatement st = prepareAndBind( conn, stmt, bindArgs );
        try
        {
            ResultSet rs = st.executeQuery();
            try { return getOneValue( rs, vt ); } finally { rs.close(); }
        }
        finally { st.close(); }
    }

    private
    static
    Object
    doExpectOne( Connection conn,
                 CharSequence stmt,
                 Object[] bindArgs,
                 ValType vt )
        throws SQLException
    {
        return expectRow( doSelectOne( conn, stmt, bindArgs, vt ) );
    }

    public
    static
    Object
    selectOne( Connection conn,
               CharSequence stmt,
               Object... bindArgs )
        throws SQLException
    {
        return doSelectOne( conn, stmt, bindArgs, ValType.OBJECT );
    }

    public
    static
    Object
    expectOne( Connection conn,
               CharSequence stmt,
               Object... bindArgs )
        throws SQLException
    {
        return doExpectOne( conn, stmt, bindArgs, ValType.OBJECT );
    }

    public
    static
    String
    selectString( Connection conn,
                  CharSequence stmt,
                  Object... bindArgs )
        throws SQLException
    {
        return (String) doSelectOne( conn, stmt, bindArgs, ValType.STRING );
    }

    public
    static
    String
    expectString( Connection conn,
                  CharSequence stmt,
                  Object... bindArgs )
        throws SQLException
    {
        return (String) doExpectOne( conn, stmt, bindArgs, ValType.STRING );
    }

    public
    static
    Integer
    selectInteger( Connection conn,
                   CharSequence stmt,
                   Object... bindArgs )
        throws SQLException
    {
        return (Integer) doSelectOne( conn, stmt, bindArgs, ValType.INTEGER );
    }

    public
    static
    int
    selectInt( Connection conn,
               CharSequence stmt,
               Object... bindArgs )
        throws SQLException
    {
        Integer res = selectInteger( conn, stmt, bindArgs );
        return res == null ? 0 : res.intValue();
    }

    public
    static
    Integer
    expectInteger( Connection conn,
                   CharSequence stmt,
                   Object... bindArgs )
        throws SQLException
    {
        return (Integer) doExpectOne( conn, stmt, bindArgs, ValType.INTEGER );
    }

    public
    static
    int
    expectInt( Connection conn,
               CharSequence stmt,
               Object... bindArgs )
        throws SQLException
    {
        return expectInteger( conn, stmt, bindArgs ).intValue();
    }

    public
    static
    Long
    selectLongObj( Connection conn,
                   CharSequence stmt,
                   Object... bindArgs )
        throws SQLException
    {
        return (Long) doSelectOne( conn, stmt, bindArgs, ValType.LONG );
    }

    public
    static
    long
    selectLong( Connection conn,
                CharSequence stmt,
                Object... bindArgs )
        throws SQLException
    {
        Long res = selectLongObj( conn, stmt, bindArgs );
        return res == null ? 0L : res.longValue();
    }

    public
    static
    Long
    expectLongObj( Connection conn,
                   CharSequence stmt,
                   Object... bindArgs )
        throws SQLException
    {
        return (Long) doExpectOne( conn, stmt, bindArgs, ValType.LONG );
    }

    public
    static
    long
    expectLong( Connection conn,
                CharSequence stmt,
                Object... bindArgs )
        throws SQLException
    {
        return expectLongObj( conn, stmt, bindArgs ).longValue();
    }

    public
    static
    int
    executeUpdate( Connection conn,
                   CharSequence stmt,
                   Object... bindArgs )
        throws SQLException
    {
        PreparedStatement st = prepareAndBind( conn, stmt, bindArgs );
        try { return st.executeUpdate(); } finally { st.close(); }
    }

    private
    static
    void
    setByteBuffer( ByteBuffer bb,
                   PreparedStatement ps,
                   int indx )
        throws SQLException
    {
        ps.setBytes( indx, IoUtils.toByteArray( bb ) );
    }

    private
    static
    void
    setNonNullValue( Object val,
                     SqlType sqlTyp,
                     PreparedStatement ps,
                     int indx )
        throws SQLException
    {
        if ( val instanceof ByteBuffer )
        {
            setByteBuffer( (ByteBuffer) val, ps, indx );
        }
        else ps.setObject( indx, val, sqlTyp.jdbcType() );
    }

    public
    static
    void
    setValue( Object val,
              SqlType sqlTyp,
              PreparedStatement ps,
              int indx )
        throws SQLException
    {
        inputs.notNull( sqlTyp, "sqlTyp" );
        inputs.notNull( ps, "ps" );
        inputs.positiveI( indx, "indx" );

        if ( val == null ) 
        {
            ps.setNull( indx, SqlType.NULL.jdbcType() );
        }
        else 
        {
            if ( sqlTyp == SqlType.NULL )
            {
                inputs.fail( "sqlTyp is NULL but val is:", val );
            }
            else setNonNullValue( val, sqlTyp, ps, indx );
        }
    }

    private
    static
    void
    debugResultSet( String msg,
                    ResultSet rs )
    {
        try { code( msg, buildList( rs, MAP_ROW_PROCESSOR, true ) ); }
        catch ( Exception ex ) { throw new RuntimeException( ex ); }
    }

    private
    static
    void
    setTable( SqlTableDescriptor.Builder b,
              String tblName,
              DatabaseMetaData md )
        throws SQLException
    {
        ResultSet rs = 
            expectSingleResult( 
                md.getTables( null, null, tblName, null ),
                "Looking for table", tblName );

        state.equalString( rs.getString( "TABLE_NAME" ), tblName );
        b.setName( tblName );

        b.setCatalog( rs.getString( "TABLE_CAT" ) );
    }

    private
    final
    static
    class SqlColumnDescriptorRowProcessor
    implements RowProcessor< SqlColumnDescriptor, Void >
    {
        public Void init( ResultSet rs ) { return null; }

        public
        SqlColumnDescriptor
        processRow( ResultSet rs,
                    Void initObj )
            throws SQLException
        {
            SqlColumnDescriptor.Builder b = new SqlColumnDescriptor.Builder();

            b.setName( state.notNull( rs.getString( "COLUMN_NAME" ) ) );
            b.setSqlType( SqlType.fromJdbcType( rs.getInt( "DATA_TYPE" ) ) );

            Object sz = rs.getObject( "COLUMN_SIZE" );
            if ( sz != null ) b.setSize( (Integer) sz );
 
            return b.build();
        }
    }

    private
    static
    List< SqlColumnDescriptor >
    getColumnDescriptors( DatabaseMetaData md,
                          String tblName )
        throws SQLException
    {
        try
        {
            return 
                buildList(
                    md.getColumns( null, null, tblName, null ),
                    SQL_COL_DESCRIP_PROC,
                    true
                );
        }
        catch ( Exception ex ) { throw createSqlRethrow( ex ); }
    }

    public
    static
    SqlTableDescriptor
    getTableDescriptor( String tblName,
                        Connection conn )
        throws SQLException
    {
        inputs.notNull( tblName, "tblName" );
        inputs.notNull( conn, "conn" );

        SqlTableDescriptor.Builder b = new SqlTableDescriptor.Builder();

        DatabaseMetaData md = conn.getMetaData();

        setTable( b, tblName, md );
        b.setColumns( getColumnDescriptors( md, tblName ) );

        return b.build();
    }

    public
    static
    SqlParameterGroupDescriptor
    getInsertParametersFor( SqlTableDescriptor td )
    {
        inputs.notNull( td, "td" );

        List< SqlColumnDescriptor > cols = td.getColumns();
        List< SqlParameterDescriptor > params = Lang.newList( cols.size() );

        int i = 1;

        for ( SqlColumnDescriptor col : cols )
        {
            params.add( 
                new SqlParameterDescriptor( 
                    col.getName(), 
                    i++,
                    col.getSqlType() 
                ) 
            );
        }

        return SqlParameterGroupDescriptor.createUnsafe( params );
    }

    public
    static
    SqlStatementWriter
    createStatementWriter( Connection conn )
        throws SQLException
    {
        inputs.notNull( conn, "conn" );

        DatabaseMetaData md = conn.getMetaData();
       
        return SqlStatementWriter.create( md );
    }

    public
    static
    < V >
    V
    useConnection( ConnectionUser< V > user,
                   Connection conn )
        throws Exception
    {
        inputs.notNull( user, "user" );
        inputs.notNull( conn, "conn" );

        try { return user.useConnection( conn ); } finally { conn.close(); }
    }

    public
    static
    < V >
    V
    useConnection( ConnectionUser< V > user,
                   DataSource ds )
        throws Exception
    {
        inputs.notNull( user, "user" );
        inputs.notNull( ds, "ds" );

        return useConnection( user, ds.getConnection() );
    }
}
