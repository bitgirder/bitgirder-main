package com.bitgirder.etl.mingle.sql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;
import com.bitgirder.lang.TestSums;

import com.bitgirder.sql.Sql;
import com.bitgirder.sql.SqlTests;
import com.bitgirder.sql.SqlTestContext;
import com.bitgirder.sql.SqlStatementWriter;
import com.bitgirder.sql.ConnectionService;
import com.bitgirder.sql.ConnectionOperation;
import com.bitgirder.sql.SqlParameterGroupDescriptor;
import com.bitgirder.sql.SqlTableDescriptor;
import com.bitgirder.sql.SqlParameterMapper;
import com.bitgirder.sql.SqlUpdateOperation;
import com.bitgirder.sql.SqlUpdateExecutor;

import com.bitgirder.etl.AbstractEtlProcessorTest;
import com.bitgirder.etl.EtlTestRecordGenerator;

import com.bitgirder.etl.sql.SqlEtlTestReactor;

import com.bitgirder.etl.mingle.MingleRecordProcessor;
import com.bitgirder.etl.mingle.MingleEtlTests;

import com.bitgirder.mingle.model.MingleStruct;
import com.bitgirder.mingle.model.MingleSymbolMapAccessor;

import com.bitgirder.mingle.sql.MingleSqlParameterMapper;

import com.bitgirder.test.Test;
import com.bitgirder.test.Before;
import com.bitgirder.test.TestFactory;
import com.bitgirder.test.AbstractLabeledTestObject;

import java.util.List;

import java.sql.Connection;
import java.sql.PreparedStatement;

final
class MingleSqlEtlTests
extends AbstractLabeledTestObject
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final SqlTestContext ctx;

    private
    MingleSqlEtlTests( SqlTestContext ctx,
                       CharSequence lbl )
    {
        super( lbl );
        this.ctx = ctx;
    }

    @Before
    private
    final
    class DbInit
    extends SqlTests.AbstractDbInit
    {
        private DbInit() { super( ctx ); }

        private
        void
        createTableEtlInteg1( Connection conn )
            throws Exception
        {
            Sql.executeUpdate( conn, "drop table if exists etl_integ1" );

            Sql.executeUpdate( conn,
                "create table etl_integ1 ( " +
                "   load_id char( 36 ) charset ascii not null, " +
                "   uid char( 36 ) charset ascii not null, " +
                "   type varchar( 255 ) not null, " +
                "   val int( 31 ) unsigned not null " +
                ") charset=utf8, engine=myisam"
            );
        }

        protected
        void
        doInit( Connection conn )
            throws Exception
        {
            createTableEtlInteg1( conn );
        }
    }

    private
    final
    static
    class Loader
    extends MingleRecordProcessor
    {
        private final ConnectionService connSvc;

        private final String loadId = Lang.randomUuid().toString();

        private CharSequence insertStmt;
        private SqlParameterGroupDescriptor paramGrp;
        private MingleSqlParameterMapper mapper;

        private Loader( ConnectionService connSvc ) { this.connSvc = connSvc; }

        private
        final
        class TableDefInit
        extends SqlUpdateOperation
        {
            private
            TableDefInit()
            {
                super( connSvc, Loader.this.getActivityContext() );
            }

            public
            void 
            executeUpdate( Connection conn )
                throws Exception
            {
                SqlTableDescriptor td = 
                    Sql.getTableDescriptor( "etl_integ1", conn );

                paramGrp = Sql.getInsertParametersFor( td, conn );

                SqlStatementWriter w = Sql.createStatementWriter( conn );
                insertStmt = w.getInsertIgnore( td );
            }

            public void useResult() { resumeRpcRequests(); }
        }

        @Override
        protected
        void
        completeInit()
        {
            holdRpcRequests();
            new TableDefInit().start();

            mapper = 
                new MingleSqlParameterMapper.Builder().
                    mapConstant( "load_id", loadId ).
                    mapStraight( "uid" ).
                    mapStraight( "val" ).
                    mapType( "type" ).
                    setParameters( paramGrp ).
                    build();
        }

        private
        final
        class Processor
        extends MingleBatchProcessor
        {
            protected 
            void 
            startBatch() 
            {
                new SqlUpdateExecutor.Builder< MingleStruct >().
                    setActivityContext( getActivityContext() ).
                    setConnectionService( connSvc ).
                    setBatch( structs() ).
                    setSql( insertStmt ).
                    setMapper( mapper ).
                    setOnComplete(
                        new AbstractTask() {
                            protected void runImpl() { batchDone( null ); }
                        }
                    ).
                    start();
            }
        }
    }

    @Test
    private
    final
    class LoadTest
    extends AbstractEtlProcessorTest< Loader, SqlEtlTestReactor >
    {
        protected
        EtlTestRecordGenerator
        createGenerator()
        {
            return MingleEtlTests.createStructGenerator();
        }

        protected
        SqlEtlTestReactor
        createReactor()
            throws Exception
        {
            return SqlEtlTestReactor.forContext( ctx );
        }

        protected
        Loader
        createTestProcessor()
        {
            return new Loader( reactor().connectionService() );
        }

        private
        final
        class Assert
        extends ConnectionOperation< Void >
        {
            private final Loader l;
            private final Runnable onComp;

            private
            Assert( Loader l,
                    Runnable onComp )
            {
                super( 
                    reactor().connectionService(), 
                    LoadTest.this.getActivityContext()
                );

                this.l = l;
                this.onComp = onComp;
            }

            public
            Void
            useConnection( Connection conn )
                throws Exception
            {
                int sumExpct = TestSums.ofSequence( 0, getFeedLength() / 2 );

                state.equalString( 
                    Strings.crossJoin( "=", ";",
                        MingleEtlTests.TYPE0.getExternalForm(), sumExpct,
                        MingleEtlTests.TYPE1.getExternalForm(), sumExpct ),
                    Sql.selectString( conn,
                        "select group_concat( sum_str separator ';' ) " +
                        "from ( " +
                        "   select concat( type, '=', sum( val ) ) " +
                        "       as sum_str " +
                        "   from etl_integ1 where load_id = ? " +
                        "   group by type order by type asc " +
                        ") as t1",
                        l.loadId
                    )
                );
 
                return null;
            }
 
            protected void useResult() { onComp.run(); }
        }

        protected
        void
        beginAssert( Loader l,
                     Object lastState,
                     Runnable onComp )
        {
            new Assert( l, onComp ).start();
        }
    }

    @TestFactory
    private
    static
    List< MingleSqlEtlTests >
    createSuite()
        throws Exception
    {
        return SqlTests.createSqlSuite( MingleSqlEtlTests.class );
    }
}
