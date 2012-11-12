package com.bitgirder.systest;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.TestServer;
import com.bitgirder.process.ProcessRpcClient;
import com.bitgirder.process.ProcessRpcServer;
import com.bitgirder.process.ProcessExit;
import com.bitgirder.process.AbstractPulse;

import com.bitgirder.concurrent.Duration;

import com.bitgirder.application.ApplicationProcess;

import java.util.Random;

import java.util.concurrent.atomic.AtomicLong;

final
class ProcessRpcBenchmark1
extends ApplicationProcess
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final int numClients;
    private final Duration reportHeartbeat;
    private final Duration thinkTime;
    private final int reportStep;
    private final Duration testDuration;

    private TestServer srv;
    private int activeClients;

    private long start;
    private long totalRpc;
    private final AtomicLong runningTotal = new AtomicLong();

    // volatile since written to by app but read from by child procs
    private volatile boolean shutdown;

    private
    ProcessRpcBenchmark1( Configurator c )
    {
        super( c.mixin( ProcessRpcClient.create() ) );

        this.numClients = inputs.positiveI( c.numClients, "numClients" );
        this.thinkTime = c.thinkTime;
        this.testDuration = c.testDuration;
        this.reportHeartbeat = c.reportHeartbeat;
        this.reportStep = c.reportStep;
    }

    private
    void
    stopServer()
    {
        behavior( ProcessRpcClient.class ).
            sendAsync( srv, new ProcessRpcServer.Stop() );
    }

    private
    CharSequence
    makeRpcRateString( long rpcCount,
                       Duration elapsed )
    {
        double rate =
            ( (double) rpcCount / ( (double) elapsed.asMillis() / 1000f ) );
 
        return String.format( "%1$f rpc/s", rate );
    }

    private
    void
    reportAndExit()
    {
        Duration run = 
            Duration.fromMillis( System.currentTimeMillis() - start );

        code( 
            "Did total of", totalRpc, "in", run.toStringSeconds(), 
            "for overall throuput of", makeRpcRateString( totalRpc, run ) );

        exit();
    }

    @Override
    protected
    void
    childExited( AbstractProcess< ? > child,
                 ProcessExit< ? > exit )
    {
        if ( ! exit.isOk() ) fail( exit.getThrowable() );

        if ( child instanceof Client )
        {
            totalRpc += (Long) exit.getResult();
            if ( --activeClients == 0 ) stopServer();
        }

        if ( ! hasChildren() ) reportAndExit();
    }

    private
    final
    class Client
    extends AbstractProcess< Long >
    implements ProcessRpcClient.ResponseHandler
    {
        private final Random r = new Random();

        private long start;
        private long rpcComplete = 0;
        private long snapshotRpc = 0;

        private Client() { super( ProcessRpcClient.create() ); }

        public
        void
        rpcFailed( Throwable th,
                   ProcessRpcClient.Call call )
        {
            fail( th );
        }

        private
        void
        logSuccess()
        {
            ++snapshotRpc;

            if ( ++rpcComplete % reportStep == 0 )
            {
                runningTotal.addAndGet( snapshotRpc );
                snapshotRpc = 0;
            }
        }

        public
        void
        rpcSucceeded( Object resp,
                      ProcessRpcClient.Call call )
        {
            logSuccess();

            if ( shutdown ) exit( rpcComplete ); 
            else 
            {
                if ( thinkTime == null ) startImpl();
                else
                {
                    long del = r.nextInt( (int) thinkTime.asMillis() );

                    submit(
                        new AbstractTask() {
                            protected void runImpl() { sendNextRequest(); }
                        },
                        Duration.fromMillis( del )
                    );
                }
            }
        }

        private
        void
        sendNextRequest()
        {
            Object req = new TestServer.EchoImmediate< Client >( this );

            behavior( ProcessRpcClient.class ).beginRpc( srv, req, this );
        }

        protected
        void
        startImpl()
        {
            start = System.currentTimeMillis();
            sendNextRequest();
        }
    }

    private
    final
    class ReportPulse
    extends AbstractPulse
    {
        private ReportPulse() { super( reportHeartbeat, self() ); }

        protected
        void
        beginPulse()
        {
            Duration elapsed = 
                Duration.fromMillis( System.currentTimeMillis() - start );

            long snap = runningTotal.get();

            code( 
                activeClients, "clients are active; throughput snapshot is ",
                makeRpcRateString( snap, elapsed ), "after", snap, "rpcs" );

            pulseDone();
        }
    }

    private
    void
    doShutdown() 
    { 
        code( "Setting shutdown flag" );
        shutdown = true; 
    }

    private
    void
    scheduleTestShutdown()
    {
        if ( testDuration != null )
        {
            submit(
                new AbstractTask() {
                    protected void runImpl() { doShutdown(); } 
                },
                testDuration
            );
        }
    }

    protected
    void
    startImpl()
    {
        start = System.currentTimeMillis();

        srv = TestServer.create();
        spawn( srv );

        for ( ; activeClients < numClients; ++activeClients ) 
        {
            spawn( new Client() );
        }

        code( "Spawned", activeClients, "clients" );

        new ReportPulse().start();
        
        scheduleTestShutdown();
    }

    @Override
    protected
    Runnable
    getShutdownTask()
    {
        return 
            new AbstractTask() { protected void runImpl() { doShutdown(); } };
    }

    private
    final
    static
    class Configurator
    extends ApplicationProcess.Configurator
    {
        private int numClients;
        private Duration thinkTime;
        private Duration testDuration;
        private Duration reportHeartbeat = Duration.fromSeconds( 30 );
        private int reportStep = 100;

        @Argument
        private
        void
        setNumClients( int numClients )
        {
            this.numClients = inputs.positiveI( numClients, "numClients" );
        }

        @Argument
        private
        void
        setThinkTime( String thinkTime )
        {
            this.thinkTime = Duration.fromString( thinkTime );
        }

        @Argument
        private
        void
        setTestDuration( String testDuration )
        {
            this.testDuration = Duration.fromString( testDuration );
        }

        @Argument
        private
        void
        setReportHeartbeat( String reportHeartbeat )
        {
            this.reportHeartbeat = Duration.fromString( reportHeartbeat );
        }

        @Argument
        private
        void
        setReportStep( int reportStep )
        {
            this.reportStep = inputs.positiveI( reportStep, "reportStep" );
        }
    }
}
