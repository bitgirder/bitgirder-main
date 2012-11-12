package com.bitgirder.mingle.sql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.sql.AbstractSqlTests;
import com.bitgirder.sql.SqlTests;
import com.bitgirder.sql.SqlTestRuntime;
import com.bitgirder.sql.Sql;
import com.bitgirder.sql.SqlTableDescriptor;
import com.bitgirder.sql.SqlParameterGroupDescriptor;
import com.bitgirder.sql.SqlParameterDescriptor;
import com.bitgirder.sql.SqlUpdateExecutor;
import com.bitgirder.sql.SqlStatementWriter;
import com.bitgirder.sql.SqlStatements;
import com.bitgirder.sql.SqlType;

import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleStructBuilder;
import com.bitgirder.mingle.model.MingleStruct;
import com.bitgirder.mingle.model.MingleStructure;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleInt64;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleEnum;
import com.bitgirder.mingle.model.MingleNull;
import com.bitgirder.mingle.model.MingleTimestamp;

import com.bitgirder.io.IoTestFactory;

import com.bitgirder.test.Test;
import com.bitgirder.test.TestRuntime;
import com.bitgirder.test.Before;
import com.bitgirder.test.LabeledTestObject;
import com.bitgirder.test.InvocationFactory;
import com.bitgirder.test.TestFactory;

import java.util.List;
import java.util.Map;
import java.util.Collection;

import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.Timestamp;
import java.sql.Time;
import java.sql.Date;

import java.nio.ByteBuffer;

final
class MingleSqlTests
extends AbstractSqlTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();
 
    private MingleSqlTests( SqlTestRuntime str ) { super( str ); }

    @Before
    private
    final
    class DbInit
    extends AbstractDbInit
    {
        protected
        void
        doInit( Connection conn )
            throws Exception
        {
            Sql.executeUpdate( conn, "drop table if exists mingle_sql1" );

            Sql.executeUpdate( conn,
                "create table mingle_sql1( " +
                "   test_id char( 36 ) charset ascii not null, " +
                "   some_string varchar( 255 ) default null, " +
                "   some_int integer( 31 ) unsigned default null " +
                ") engine=innodb, charset=utf8"
            );
        }
    }

    private
    final
    class MingleSqlSetValueTest
    extends SetValueTest
    {
        private
        MingleSqlSetValueTest( MingleValue toSet,
                               SqlType sqlTyp,
                               String colNm,
                               Object expctVal )
        {
            super( toSet, sqlTyp, colNm, expctVal );
        }
        
        private
        MingleSqlSetValueTest( MingleValue toSet,
                               SqlType sqlTyp,
                               String colNm,
                               Object expctVal,
                               String tag )
        {
            super( toSet, sqlTyp, colNm, expctVal, tag );
        }

        @Override
        protected
        void
        setValue( Object toSet,
                  SqlType sqlTyp,
                  PreparedStatement ps,
                  int indx )
            throws Exception
        {
            MingleSql.setValue( (MingleValue) toSet, sqlTyp, ps, indx );
        }
    }

    private
    void
    addIntegralSetValTests( List< MingleSqlSetValueTest > l )
    {
        for ( SqlType t : new SqlType[] {
                SqlType.TINYINT, SqlType.BIGINT, SqlType.INTEGER,
                SqlType.NUMERIC, SqlType.SMALLINT, SqlType.DECIMAL,
                SqlType.DOUBLE, SqlType.FLOAT, SqlType.REAL } )
        {
            l.add(
                new MingleSqlSetValueTest(
                    MingleModels.asMingleInt64( 12 ), t, "long1", 12 ) );
 
            l.add(
                new MingleSqlSetValueTest(
                    MingleModels.asMingleString( "12" ), t, "long1", 12 ) );
        }
    }

    private
    void
    addDecimalSetValTests( List< MingleSqlSetValueTest > l )
    {
        for ( SqlType t : new SqlType[] {
                SqlType.DECIMAL, SqlType.DOUBLE, SqlType.FLOAT, SqlType.REAL } )
        {
            l.add(
                new MingleSqlSetValueTest(
                    MingleModels.asMingleDouble( 12.1 ), t, "double1", 12.1d )
            );

            l.add(
                new MingleSqlSetValueTest(
                    MingleModels.asMingleString( "12.1" ), t, "double1", 
                    12.1d ) 
            );
        }
    } 

    private
    void
    addStringSetValTests( List< MingleSqlSetValueTest > l )
    {
        for ( SqlType t : new SqlType[] {
                SqlType.CHAR, SqlType.VARCHAR, SqlType.LONGVARCHAR } )
        {
            l.add(
                new MingleSqlSetValueTest(
                    MingleModels.asMingleString( "hello" ), t, "str1", "hello" )
            );
            
            l.add(
                new MingleSqlSetValueTest(
                    MingleEnum.create( "test@v1/Enum1.const1" ), t, "str1", 
                    "const1" )
            );
        }
    }

    private
    void
    addBoolSetValTests( List< MingleSqlSetValueTest > l )
    {
        for ( SqlType t : new SqlType[] { SqlType.BIT, SqlType.BOOLEAN } )
        {
            l.add(
                new MingleSqlSetValueTest(
                    MingleModels.asMingleBoolean( false ),
                    t,
                    "bool1",
                    false
                )
            );
        }
    }
    
    private
    void
    addNullSetValTests( List< MingleSqlSetValueTest > l )
    {
        for ( int i = 0; i < 2; ++i )
        {
            for ( SqlType t : new SqlType[] { SqlType.NULL, SqlType.CHAR } )
            {
                l.add( 
                    new MingleSqlSetValueTest(
                        i % 2 == 0 ? null : MingleNull.getInstance(),
                        t,
                        "str1",
                        null
                    )
                );
            }
        }
    }

    private
    void
    addTimeSetValTests( List< MingleSqlSetValueTest > l )
    {
        MingleTimestamp mgTm1 = 
            MingleTimestamp.create( "2001-01-01T12:00:00-08:00" );

        MingleTimestamp mgTm2 =
            MingleTimestamp.create( "2001-01-02T04:00:00+08:00" );
 
        Timestamp ts = new Timestamp( mgTm1.getTimeInMillis() );
        Date dt = new Date( mgTm1.getTimeInMillis() );
        Time tm = new Time( mgTm1.getTimeInMillis() );

        for ( int i = 0; i < 2; ++i )
        for ( int j = 0; j < 2; ++j )
        for ( int k = 0; k < 3; ++k )
        {{{
            l.add(
                new MingleSqlSetValueTest(
                    i % 2 == 0 ? mgTm1 : mgTm2,
                    j % 2 == 0 ? SqlType.TIMESTAMP : SqlType.DATE,
                    k == 0 ? "timestamp1" : k == 1 ? "datetime1" : "date1",
                    k == 2 ? dt : ts,
                    i % 2 == 0 ? "gmt-8" : "gmt+8"
                )
            );
        }}}

        l.add( new MingleSqlSetValueTest( mgTm1, SqlType.TIME, "time1", tm ) );
        l.add( new MingleSqlSetValueTest( mgTm2, SqlType.TIME, "time1", tm ) );
    }

    private
    void
    addBufferSetValTests( List< MingleSqlSetValueTest > l )
    {
        ByteBuffer buf = IoTestFactory.nextByteBuffer( 100 );

        for ( SqlType t : new SqlType[] {
                SqlType.BINARY, SqlType.BLOB, SqlType.LONGVARBINARY,
                SqlType.VARBINARY } )
        {
            l.add(
                new MingleSqlSetValueTest(
                    MingleModels.asMingleBuffer( buf ),
                    t,
                    "blob1",
                    buf
                )
            );
        }
    }

    @InvocationFactory
    private
    List< MingleSqlSetValueTest >
    testSetValue()
    {
        List< MingleSqlSetValueTest > res = Lang.newList();

        addIntegralSetValTests( res );
        addDecimalSetValTests( res );
        addStringSetValTests( res );
        addBoolSetValTests( res );
        addNullSetValTests( res );
        addTimeSetValTests( res );
        addBufferSetValTests( res );

        return res;
    }

    // cover all cols that are in sql_impl3
    @Test
    private
    final
    class MingleSqlMapperTest
    extends AbstractTest
    {
        private final String testId = Lang.randomUuid();
        private final int batchLen = 100;

        // just to get coverage of a custom handler, even though this could
        // just as easily be handled by MingleStructureMapper's built-in
        // handlers
        private
        final
        class CustomInt1Handler
        implements MingleStructureMapper.ParameterHandler< Void >
        {
            public
            boolean
            setParameter( MingleStructure ms,
                          Void ignore,
                          SqlParameterDescriptor param,
                          PreparedStatement ps,
                          int indx )
                throws Exception
            {
                MingleIdentifier fld = MingleIdentifier.create( "long1" );

                ps.setInt(
                    indx,
                    ( (MingleInt64) ms.getFields().get( fld ) ).intValue()
                );

                return true;
            }
        }

        private
        void
        assertRows()
        {
            new AbstractOp< Void >() 
            {
                public Void useConnection( Connection conn ) throws Exception 
                {
                    assertSqlImpl3Rows( conn, testId, batchLen );
                    return null;
                }

                public void useResult() { exit(); }
            }.
            start();
        }

        private
        MingleStructureMapper< Void >
        createMapper( SqlParameterGroupDescriptor grp )
        {
            return
                new MingleStructureMapper.Builder< Void >().
                    setParameters( grp ).
                    mapConstant( "test_id", testId ).
                    mapField( "str1", "some-string" ). // test name mapping too
                    map( "long1", new CustomInt1Handler() ).
                    mapField( "double1" ).
                    mapField( "bool1" ).
                    mapField( "blob1" ).
                    mapField( "timestamp1" ).
                    mapField( "datetime1" ).
                    mapField( "date1" ).
                    mapField( "time1" ).
                    build();
        }

        private
        List< MingleStruct >
        createBatch()
        {
            return new BatchImpl< MingleStruct >( batchLen ) {
                public MingleStruct get( int i )
                {
                    MingleTimestamp tm = 
                        MingleTimestamp.
                            fromMillis( makeTimestampData( i ).getTime() );

                    return
                        MingleModels.structBuilder(). 
                        setType( "bitgirder:mingle:sql@v1/Struct1" ).f().
                        setString( "some-string", makeStringData( i ) ).f().
                        setInt64( "long1", makeLongData( i ) ).f().
                        setBoolean( "bool1", makeBooleanData( i ) ).f().
                        setDouble( "double1", makeDoubleData( i ) ).f().
                        setBuffer( "blob1", makeBlobData( i ) ).f().
                        set( "timestamp1", tm ).f().
                        set( "datetime1", tm ).f().
                        set( "date1", tm ).f().
                        set( "time1", tm ).
                        build();
                }
            };
        }

        private
        final
        class InitOp
        extends AbstractOp< SqlUpdateExecutor.Builder< MingleStruct, Void > >
        {
            public
            SqlUpdateExecutor.Builder< MingleStruct, Void >
            useConnection( Connection conn )
                throws Exception
            {
                SqlTableDescriptor td = 
                    Sql.getTableDescriptor( "sql_impl3", conn );
                
                SqlParameterGroupDescriptor grp =
                    Sql.getInsertParametersFor( td );
                
                SqlStatementWriter w = Sql.createStatementWriter( conn );

                String sql = SqlStatements.getInsert( td, false, w );

                return 
                    new SqlUpdateExecutor.Builder< MingleStruct, Void >().
                        setMapper( createMapper( grp ) ).
                        setSql( sql );
            }

            public
            void
            useResult( SqlUpdateExecutor.Builder< MingleStruct, Void > b )
            {
                b.setActivityContext( getActivityContext() ).
                  setConnectionService( connectionService() ).
                  setBatch( createBatch() ).
                  setOnComplete(
                    new AbstractTask() {
                        protected void runImpl() { assertRows(); }
                    }
                  ).
                  build().
                  start();
            }
        }

        protected void startSql() { new InitOp().start(); }
    }

    private
    abstract
    class AbstractMapperTest< B >
    extends AbstractTest
    {
        final String testId = Lang.randomUuid();
        final int batchLen = 100;

        abstract
        void
        buildMapper( MingleStructureMapper.Builder< B > b );

        final String makeString( int i ) { return "string-" + i; }

        final
        void
        assertRow( Map< String, Object > m,
                   int i )
        {
            state.equalInt( 3, m.size() );

            state.equal( testId, m.get( "test_id" ) );

            state.equalString( 
                (CharSequence) makeString( i ),
                (CharSequence) m.get( "some_string" ) );

            state.equalInt( i, ( (Number) m.get( "some_int" ) ).intValue() );
        }

        private
        List< MingleStruct >
        createBatch()
        {
            return new BatchImpl< MingleStruct >( batchLen ) {
                public MingleStruct get( int i ) 
                {
                    return MingleModels.structBuilder().
                           setType( "test:ns@v1/Type1" ).f().
                           setString( "some-string", makeString( i ) ).f().
                           setInt64( "some-int", i ).
                           build();
                }
            };
        }

        abstract
        void
        assertRows( List< Map< String, Object > > l );

        private
        final
        class AssertOp
        extends AbstractOp< List< Map< String, Object > > >
        {
            public
            List< Map< String, Object > >
            useConnection( Connection conn )
                throws Exception
            {
                return Sql.selectListOfMaps( conn,
                    "select * from mingle_sql1 where test_id = ? " +
                    "order by some_int asc", testId
                );
            }

            public
            void
            useResult( List< Map< String, Object > > rows )
                throws Exception
            {
                state.equalInt( batchLen, rows.size() );
                assertRows( rows );
                exit();
            }
        }

        void
        buildUpdateExecutor( SqlUpdateExecutor.Builder< MingleStruct, B > b )
        {}

        private
        final
        class InitOp
        extends AbstractOp< SqlUpdateExecutor.Builder< MingleStruct, B > >
        {
            private
            MingleStructureMapper< B >
            createMapper( SqlTableDescriptor td )
            {
                SqlParameterGroupDescriptor grp = 
                    Sql.getInsertParametersFor( td );

                MingleStructureMapper.Builder< B > b =
                    new MingleStructureMapper.Builder< B >().
                        setParameters( grp ).
                        mapConstant( "test_id", testId );
                
                buildMapper( b );

                return b.build();
            }

            public
            SqlUpdateExecutor.Builder< MingleStruct, B >
            useConnection( Connection conn )
                throws Exception
            {
                SqlTableDescriptor td = 
                    Sql.getTableDescriptor( "mingle_sql1", conn );
                
                String sql =
                    SqlStatements.getInsert(
                        td, false, Sql.createStatementWriter( conn ) );

                return
                    new SqlUpdateExecutor.Builder< MingleStruct, B >().
                        setMapper( createMapper( td ) ).
                        setSql( sql );
            }

            public
            void
            useResult( SqlUpdateExecutor.Builder< MingleStruct, B > b )
            {
                b.setActivityContext( getActivityContext() ).
                  setBatch( createBatch() ).
                  setConnectionService( connectionService() ).
                  setOnComplete(
                    new AbstractTask() {
                        protected void runImpl() { new AssertOp().start(); }
                    }
                  );
                
                buildUpdateExecutor( b );
                b.build().start();
            }
        }

        protected final void startSql() { new InitOp().start(); } 
    }

    // gets coverage of setting constants both when const val is a MingleValue
    // and when it is just some other java obj
    @Test
    private
    final
    class ConstantFieldMappingTest
    extends AbstractMapperTest< Void >
    {
        private final CharSequence constStr =
            MingleModels.asMingleString( "hello" );
        
        private final int constInt = 12;

        void
        buildMapper( MingleStructureMapper.Builder< Void > b )
        {
            b.mapConstant( "some_string", constStr ).
              mapConstant( "some_int", constInt );
        }

        void
        assertRows( List< Map< String, Object > > rows )
        {
            for ( Map< String, Object > m : rows )
            {
                state.equalInt( 3, m.size() );
                state.equal( testId, m.get( "test_id" ) );

                state.equalString( 
                    (CharSequence) constStr, 
                    (CharSequence) m.get( "some_string" ) );

                state.equalInt( 
                    constInt, ( (Number) m.get( "some_int" ) ).intValue() );
            }
        }
    }

    @Test
    private
    final
    class MingleIdentifierFormsAsFieldNamesTest
    extends AbstractMapperTest< Void >
    {
        void
        buildMapper( MingleStructureMapper.Builder< Void > b )
        {
            b.mapField( "some-string" ).
              mapField( "some_int" );
        }

        void
        assertRows( List< Map< String, Object > > rows )
        {
            int i = 0;
            for ( Map< String, Object > m : rows ) assertRow( m, i++ );
        }
    }

    @Test
    private
    final
    class CustomMapperObjectGeneratorTest
    extends AbstractMapperTest< Integer >
    {
        private final int addend = 1000;

        private
        final
        class MapperObjGenerator
        implements SqlUpdateExecutor.MapperObjectGenerator< MingleStruct, 
                                                            Integer >
        {
            public
            Integer
            generateMapperObject( Collection< ? extends MingleStruct > batch,
                                  Connection conn )
            {
                return addend;
            }
        }

        @Override
        void
        buildUpdateExecutor( 
            SqlUpdateExecutor.Builder< MingleStruct, Integer > b )
        {
            b.setMapperObjectGenerator( new MapperObjGenerator() );
        }

        private
        final
        class SomeIntMapper
        implements MingleStructureMapper.ParameterHandler< Integer > 
        {
            public
            boolean
            setParameter( MingleStructure ms,
                          Integer batchObj,
                          SqlParameterDescriptor param,
                          PreparedStatement ps,
                          int indx )
                throws Exception
            {
                MingleIdentifier fld = MingleIdentifier.create( "some-int" );
                MingleInt64 mi = (MingleInt64) ms.getFields().get( fld );

                ps.setInt( indx, mi.intValue() + batchObj );

                return true;
            }
        }

        void
        buildMapper( MingleStructureMapper.Builder< Integer > b )
        {
            b.mapField( "some-string" );
            b.map( "some_int", new SomeIntMapper() );
        }

        void
        assertRows( List< Map< String, Object > > rows )
        {
            // just check some_int value; other tests covered correctness for
            // other fields

            int i = 0;

            for ( Map< String, Object > m : rows )
            {
                state.equalInt( i++ + addend, getInt( m, "some_int" ) );
            }
        }
    }

    @TestFactory
    private
    static
    List< MingleSqlTests >
    getSuite( TestRuntime rt )
        throws Exception
    {
        return SqlTests.createSqlSuite( MingleSqlTests.class, rt );
    }
}
