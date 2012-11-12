package com.bitgirder.etl.sql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.process.ProcessRpcServer;
import com.bitgirder.process.ProcessRpcClient;
import com.bitgirder.process.AbstractVoidProcess;

import com.bitgirder.etl.EtlProcessors;

import com.bitgirder.sql.ConnectionService;
import com.bitgirder.sql.ConnectionOperation;
import com.bitgirder.sql.Sql;

import java.sql.Connection;

public
final
class SqlEtlStateManager
extends AbstractVoidProcess
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final ConnectionService connSvc;

    // hardcoding for now
    private final CharSequence tblName = "etl_processor_state";

    private
    SqlEtlStateManager( Builder b )
    {
        super( 
            ProcessRpcServer.createStandard(),
            ProcessRpcClient.create()
        );

        this.connSvc = inputs.notNull( b.connSvc, "connSvc" );
    }

    protected void startImpl() {}

    @ProcessRpcServer.Responder
    private
    final
    class GetStateProcessor
    extends ProcessRpcServer.AbstractAsyncResponder< Object >
    {
        private EtlProcessors.GetProcessorState req;

        private
        final
        class GetState
        extends ConnectionOperation< Object >
        {
            private
            GetState()
            {
                super( connSvc, GetStateProcessor.this.getActivityContext() );
            }

            public
            Object
            useConnection( Connection conn )
                throws Exception
            {
                return 
                    Sql.selectOne(
                        conn, 
                        "select proc_state from `" + tblName + "` " +
                        "where proc_id = ?",
                        req.getId().getExternalForm().toString()
                    );
            }

            public void useResult( Object res ) { respond( res ); }
        }

        protected void startImpl() { new GetState().start(); } 
    }

    @ProcessRpcServer.Responder
    private
    final
    class SetProcessorState
    extends ProcessRpcServer.AbstractAsyncResponder< Void >
    {
        private EtlProcessors.SetProcessorState req;

        private
        final
        class SetOp
        extends ConnectionOperation< Void >
        {
            private
            SetOp()
            {
                super( connSvc, SetProcessorState.this.getActivityContext() );
            }

            public
            Void
            useConnection( Connection conn )
                throws Exception
            {
                Sql.executeUpdate( conn,
                    "insert into " + tblName + " ( proc_id, proc_state ) " +
                    "values ( ?, ? ) on duplicate key update " +
                    "proc_state = values( proc_state )",
                    req.getId().getExternalForm().toString(),
                    req.getObject()
                );

                return null;
            }

            public void useResult() { respond( null ); }
        }

        protected void startImpl() { new SetOp().start(); }
    }

    public
    final
    static
    class Builder
    {
        private ConnectionService connSvc;

        public
        Builder
        setConnectionService( ConnectionService connSvc )
        {
            this.connSvc = inputs.notNull( connSvc, "connSvc" );
            return this;
        }

        public
        SqlEtlStateManager
        build()
        {
            return new SqlEtlStateManager( this ); 
        }
    }
}
