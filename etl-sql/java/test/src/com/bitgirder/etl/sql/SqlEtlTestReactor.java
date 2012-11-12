package com.bitgirder.etl.sql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.Processes;
import com.bitgirder.process.ProcessActivity;
import com.bitgirder.process.ComputePool;

import com.bitgirder.sql.ConnectionService;
import com.bitgirder.sql.SqlTestRuntime;

import com.bitgirder.etl.EtlTestReactor;
import com.bitgirder.etl.EtlTests;

import java.util.List;

import javax.sql.DataSource;

public
final
class SqlEtlTestReactor
implements EtlTestReactor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final ConnectionService connSvc;
    private final boolean manageConnSvc;

    private final List< AbstractProcess< ? > > procs = Lang.newList();

    private AbstractProcess< ? > stateMgr = EtlTests.createMemoryStateManager();

    private
    SqlEtlTestReactor( ConnectionService connSvc,
                       boolean manageConnSvc )
    {
        this.connSvc = connSvc;
        this.manageConnSvc = manageConnSvc;
    }

    public ConnectionService connectionService() { return connSvc; }

    private
    void
    spawn( AbstractProcess< ? > proc,
           ProcessActivity.Context ctx )
    {
        ctx.spawn( proc, Processes.< Object >getNoOpExitListener() );
        procs.add( proc );
    }

    public
    void
    startTestProcesses( ProcessActivity.Context procCtx,
                        Runnable onComplete )
        throws Exception
    {
        if ( manageConnSvc ) spawn( connSvc, procCtx );
        spawn( stateMgr, procCtx );

        onComplete.run();
    }

    public
    void
    stopTestProcesses( ProcessActivity.Context ctx )
    {
        for ( AbstractProcess< ? > proc : procs ) 
        {
            Processes.sendStop( proc, ctx );
        }
    }

    public AbstractProcess< ? > getStateManager() { return stateMgr; }

    public
    static
    SqlEtlTestReactor
    forRuntime( SqlTestRuntime srt )
    {
        inputs.notNull( srt, "srt" );
        return new SqlEtlTestReactor( srt.connectionService(), false );
    }

    public
    static
    SqlEtlTestReactor
    forDataSource( DataSource ds )
    {
        inputs.notNull( ds, "ds" );

        ConnectionService connSvc = 
            new ConnectionService.Builder().
                setDataSource( ds ).
                setTaskRunner( ComputePool.createFixedPool( 1 ) ).
                build();
        
        return new SqlEtlTestReactor( connSvc, true );
    }
}
