package com.bitgirder.demo.process;

import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.AbstractPulse;
import com.bitgirder.process.ProcessExit;
import com.bitgirder.process.ComputePool;
import com.bitgirder.process.ProcessRpcClient;
import com.bitgirder.process.ProcessRpcServer;

import com.bitgirder.concurrent.Duration;

import com.bitgirder.demo.Demo;

// Demo to show two ways to use a ComputePool by mimicking what a process might
// do to provide access to blocking JDBC libraries, allowing a process-based
// application to interact with a SQL database.
//
// We spawn two processes, one acting as the db pool, another acting as a client
// to it. The client makes one request and exits; the pool runs for a small
// fixed amount of time and exits (because all infinite demos are eventually
// boring)
@Demo
final
class ComputePoolDemo
extends AbstractVoidProcess
{
    // The db pool process.
    private
    final
    static
    class DbPool
    extends AbstractVoidProcess
    {
        // In addition to including the compute pool to use in performing JDBC
        // operations (simulated here), we also include rpc server to accept sql
        // requests from clients.
        private 
        DbPool() 
        { 
            super( 
                ComputePool.createFixedPool( 3 ),
                ProcessRpcServer.createStandard()
            ); 
        }

        // little shell to encapsulate a select request
        private
        final
        static
        class Select
        {
            private final String sql;

            private Select( String sql ) { this.sql = sql; }
        }

        // shell class to represent results of a select operation
        private
        final
        static
        class SelectResult
        {
            private final String data;

            private SelectResult( String data ) { this.data = data; }

            @Override public String toString() { return data; }
        }

        // A child responder subclassed from ComputePool.AbstractCall. Under the
        // hood the ComputePool.AbstractCall creates a
        // java.util.concurrent.Callable that returns our call result below and
        // handles submitting the Callable to the ComputePool, which we need to
        // pass in in the constructor, and calling exit() or fail() as
        // appopriate with the result of the call.
        @ProcessRpcServer.Responder
        private
        final
        class SelectRunner
        extends ComputePool.AbstractCall< Select, SelectResult >
        {
            private
            SelectRunner( Select select )
            {
                super( select, DbPool.this.behavior( ComputePool.class ) );
            }

            // Executed by a compute pool thread
            protected
            SelectResult
            call( Select select )
            {
                code( "Compute pool thread executing select:", select.sql );

                // here is where the actual jdbc stuff would happen
                return new SelectResult( "[sample result data]" );
            }
        }

        // A pulse to simulate a background check that the db is alive
        private
        final
        class ConnectionCheckPulse
        extends AbstractPulse
        {
            private
            ConnectionCheckPulse()
            {
                super( Duration.fromSeconds( 1 ), self() );
            }

            // This runnable is actually run by a compute pool thread. For that
            // reason we have to make sure to call pulseDone() from back in the
            // DbPool process thread.
            private
            final
            class ConnectionCheck
            implements Runnable
            {
                public
                void
                run()
                {
                    code( "Checking database connection" );

                    submit(
                        new AbstractTask() {
                            protected void runImpl() { pulseDone(); }
                        }
                    );
                }
            }

            // start the next check
            protected
            void
            beginPulse()
            {
                behavior( ComputePool.class ).
                    submitTask( new ConnectionCheck() );
            }
        }

        protected
        void
        startImpl()
        {
            // fire off a connection check pulse
            new ConnectionCheckPulse().start();

            // exit after awhile
            submit(
                new AbstractTask() { protected void runImpl() { exit(); } },
                Duration.fromSeconds( 5 )
            );
        }
    }

    // sample client of the db pool -- makes a single select request and exits
    // after getting the result
    private
    final
    static
    class DbPoolClient
    extends AbstractVoidProcess
    {
        private final DbPool pool;

        private 
        DbPoolClient( DbPool pool ) 
        { 
            super( ProcessRpcClient.create() );

            this.pool = pool; 
        }

        // make the request and exit
        protected
        void
        startImpl()
        {
            code( "Calling a db operation" );
            
            // call a select with a simple handler that prints the results and
            // exits or fails the process if the rpc fails (by virtue of it
            // being a ProcessRpcClient.DefaultResponseHandler)
            beginRpc( 
                pool,
                new DbPool.Select( "select * from blah" ),
                new ProcessRpcClient.DefaultResponseHandler( this ) {
                    @Override protected void rpcSucceeded( Object resp )
                    {
                        code( "Got select result:", resp );
                        exit();
                    }
                }
            );
        }
    }

    // exit when all children are gone or if one fails
    @Override
    protected
    void
    childExited( AbstractProcess< ? > child,
                 ProcessExit< ? > exit )
    {
        if ( ! exit.isOk() ) fail( exit.getThrowable() );
        if ( ! hasChildren() ) exit();
    }

    // spawn the pool, wait a small amount of time, and then start the child to
    // simulate a select
    protected
    void
    startImpl()
    {
        final DbPool pool = new DbPool();
        spawn( pool );
 
        submit(
            new AbstractTask() { 
                protected void runImpl() { spawn( new DbPoolClient( pool ) ); }
            },
            Duration.fromSeconds( 2 )
        );
    }
}
