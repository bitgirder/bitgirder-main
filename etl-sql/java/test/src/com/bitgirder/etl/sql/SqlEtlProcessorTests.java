package com.bitgirder.etl.sql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.etl.AbstractEtlProcessorTest;
import com.bitgirder.etl.AbstractEtlProcessor;
import com.bitgirder.etl.AbstractRecordSetProcessor;
import com.bitgirder.etl.EtlRecordSet;
import com.bitgirder.etl.EtlTestReactor;
import com.bitgirder.etl.EtlTestRecordGenerator;
import com.bitgirder.etl.EtlTests;

import com.bitgirder.sql.ConnectionService;
import com.bitgirder.sql.ConnectionOperation;
import com.bitgirder.sql.Sql;
import com.bitgirder.sql.SqlTests;
import com.bitgirder.sql.SqlTestRuntime;

import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.ProcessActivity;
import com.bitgirder.process.Processes;
import com.bitgirder.process.ComputePool;

import com.bitgirder.test.Test;
import com.bitgirder.test.TestRuntime;
import com.bitgirder.test.TestFactory;
import com.bitgirder.test.Before;
import com.bitgirder.test.AbstractLabeledTestObject;

import java.util.Map;
import java.util.List;

import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.ResultSet;

final
class SqlEtlProcessorTests
extends AbstractLabeledTestObject
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final SqlTestRuntime srt;

    private
    SqlEtlProcessorTests( SqlTestRuntime srt )
    { 
        super( srt.sqlTestContext().getLabel() );
        this.srt = srt;
    }

    private
    void
    initTableEtlFact1( Connection conn )
        throws Exception
    {
        Sql.executeUpdate( conn, "drop table if exists etl_fact1" );

        Sql.executeUpdate( conn,
            "create table etl_fact1( " +
            "   proc_id char( 36 ) not null, " +
            "   data integer not null, " +
            "   unique proc_id_data_uniq( proc_id, data ) " +
            ") engine=innodb, charset=utf8 "
        );
    }
 
    @Before
    private
    final
    class InitEtlTables
    extends SqlTests.AbstractDbInit
    {
        private InitEtlTables() { super( srt.connectionService() ); }

        protected
        void
        doInit( Connection conn )
            throws Exception
        {
            initTableEtlFact1( conn );
        }
    }

    private
    final
    class TestProcessor
    extends AbstractEtlProcessor
    {
        private final ConnectionService connSvc;
        private final String procId = Lang.randomUuid();

        private
        TestProcessor( ConnectionService connSvc )
        {
            this.connSvc = connSvc;
        }

        private
        final
        class RecordProcessor
        extends AbstractRecordSetProcessor
        {
            private
            final
            class Update
            extends ConnectionOperation< Integer >
            {
                private final EtlRecordSet rs;
    
                private
                Update( EtlRecordSet rs )
                {
                    super( connSvc, TestProcessor.this.getActivityContext() );
                    this.rs = rs;
                }
    
                private
                int
                buildBatch( PreparedStatement st )
                    throws Exception
                {
                    int res = 0;
    
                    for ( Object i : rs )
                    {
                        res = (Integer) i;
                        st.setString( 1, procId );
                        st.setInt( 2, res );
    
                        st.addBatch();
                    }
    
                    return res;
                }
    
                protected
                Integer
                useConnection( Connection conn )
                    throws Exception
                {
                    PreparedStatement st = 
                        conn.prepareStatement(
                            "insert into etl_fact1( proc_id, data ) " +
                            "values ( ?, ? ) "
                        );
                    
                    try 
                    {
                        int res = buildBatch( st );
                        st.executeBatch();
    
                        return res;
                    }
                    finally { st.close(); }
                }
    
                protected 
                void 
                useResult( Integer last ) 
                { 
                    respond( -last - 1 ); 
                }
            }

            protected void startProcess() { new Update( recordSet() ).start(); }
        }
    }

    @Test
    private
    final
    class BasicTest
    extends AbstractEtlProcessorTest< TestProcessor, SqlEtlTestReactor >
    {
        private
        final
        class CountAssert
        extends ConnectionOperation< Map< String, Object > >
        {
            private final TestProcessor tp;
            private final Runnable onComp;

            private 
            CountAssert( TestProcessor tp,
                         Runnable onComp ) 
            { 
                super( 
                    reactor().connectionService(), 
                    BasicTest.this.getActivityContext() 
                );

                this.tp = tp; 
                this.onComp = onComp;
            }

            public
            Map< String, Object >
            useConnection( Connection conn )
                throws Exception
            {
                return Sql.selectOneMap(
                    conn,
                    "select count( * ) as `count`, sum( data ) as `sum` " +
                    "from etl_fact1 where proc_id = ? group by proc_id ", 
                    tp.procId
                );
            }

            public
            void
            useResult( Map< String, Object > m )
            {
                int len = getFeedLength();

                state.equal( 
                    len, ( (Number) state.get( m, "count", "m" ) ).intValue() );

                int sum = ( len - 1 ) * len / 2;
                state.equal( 
                    sum, ( (Number) state.get( m, "sum", "m" ) ).intValue() );

                onComp.run();
            }
        }

        protected
        void
        beginAssert( TestProcessor tp,
                     Object lastState,
                     Runnable onComp )
        {
            new CountAssert( tp, onComp ).start();
        }

        protected
        TestProcessor
        createTestProcessor()
        {
            return new TestProcessor( reactor().connectionService() );
        }

        protected
        SqlEtlTestReactor
        createReactor()
            throws Exception
        {
            return SqlEtlTestReactor.forRuntime( srt );
        }

        protected
        EtlTestRecordGenerator
        createGenerator()
        {
            return EtlTests.intGenerator();
        }
    }

    @TestFactory
    private
    static
    List< SqlEtlProcessorTests >
    createSuite( TestRuntime rt )
        throws Exception
    {
        return SqlTests.createSqlSuite( SqlEtlProcessorTests.class, rt );
    }
}
