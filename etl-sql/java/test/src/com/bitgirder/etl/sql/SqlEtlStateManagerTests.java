package com.bitgirder.etl.sql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.etl.EtlTestReactor;
import com.bitgirder.etl.EtlStateManagerTests;

import com.bitgirder.process.ProcessActivity;
import com.bitgirder.process.AbstractVoidProcess;

import com.bitgirder.sql.Sql;
import com.bitgirder.sql.SqlTests;
import com.bitgirder.sql.SqlTestRuntime;

import com.bitgirder.test.Test;
import com.bitgirder.test.TestRuntime;
import com.bitgirder.test.Before;
import com.bitgirder.test.TestFactory;
import com.bitgirder.test.LabeledTestObject;

import java.util.List;
import java.util.Map;

import java.sql.Connection;

final
class SqlEtlStateManagerTests
extends EtlStateManagerTests
implements LabeledTestObject
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final SqlTestRuntime srt;

    private SqlEtlStateManagerTests( SqlTestRuntime srt ) { this.srt = srt; }

    public 
    final 
    CharSequence 
    getLabel() 
    { 
        return srt.sqlTestContext().getLabel(); 
    }

    public final Object getInvocationTarget() { return this; }

    // should create:
    //
    //  table: etl_processor_state
    //  cols:
    //      proc_id: varchar
    //      proc_state: integer
    private
    void
    initTableEtlProcessorState( Connection conn )
        throws Exception
    {
        Sql.executeUpdate( conn, "drop table if exists etl_processor_state" );

        Sql.executeUpdate( conn,
            "create table etl_processor_state( " +
            "   proc_id varchar( 255 ) primary key, " +
            "   proc_state integer not null " +
            ") engine=innodb, charset=utf8"
        );
    }

    @Before
    private
    final
    class InitDb
    extends SqlTests.AbstractDbInit
    {
        private InitDb() { super( srt.connectionService() ); }

        protected
        void
        doInit( Connection conn )
            throws Exception
        {
            initTableEtlProcessorState( conn );
        }
    }

    protected
    final
    EtlTestReactor
    createTestReactor()
        throws Exception
    {
        return SqlEtlTestReactor.forRuntime( srt );
    }

    protected
    final
    Object
    createStateObject( Integer i )
    {
        return i == null ? null : i;
    }

    @TestFactory
    private
    static
    List< SqlEtlStateManagerTests >
    createSuite( TestRuntime rt )
        throws Exception
    {
        return SqlTests.createSqlSuite( SqlEtlStateManagerTests.class, rt );
    }
}
