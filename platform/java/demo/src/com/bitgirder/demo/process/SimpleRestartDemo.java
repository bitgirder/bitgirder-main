package com.bitgirder.demo.process;

import com.bitgirder.validation.State;

import com.bitgirder.process.ProcessRpcServer;
import com.bitgirder.process.ProcessRpcClient;
import com.bitgirder.process.ProcessExit;
import com.bitgirder.process.AbstractPulse;
import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.AbstractVoidProcess;

import com.bitgirder.concurrent.Duration;

import com.bitgirder.demo.Demo;

import java.util.concurrent.TimeoutException;

// A demonstration making use of ProcessBehavior and ProcessActivity. The core
// of this demo comes from SimpleRestarter, so consult that class for details
// about its own operation.
//
// This demo consists of two directly instantiated processes. One is the demo
// itself, which includes the SimpleRestarter behavior. The demo process
// (via the restarter) spawns some number of RestartableProcesses, which don't
// do much other than respond to a single GetStatus request and which exit after
// a fixed amount of time (the restarter will restart them).
//
// The second process is a ProxyCaller, spawned as a child of the demo, which
// starts a pulse that calls via a SimpleRestarter.Proxy to get the status of
// whichever RestartableProcess is currently running. After a fixed amount of
// time this caller exits, leading the demo to begin its exit.
@Demo
final
class SimpleRestartDemo
extends AbstractVoidProcess
{
    private final static State state = new State();

    // include the SimpleRestarter behavior
    private SimpleRestartDemo() { super( new SimpleRestarter() ); }

    // A simple process which responds to a single rpc call and which exits
    // after a fixed amount of time.
    private
    final
    static
    class RestartableProcess
    extends AbstractVoidProcess
    {
        // Type used to invoke the get status rpc operation
        private final static class GetStatus {}

        // start time in millis, set in startImpl()
        private long startTime;

        private
        RestartableProcess()
        {
            super( ProcessRpcServer.createStandard() );
        }

        // Returns a status string about this process to the caller
        @ProcessRpcServer.Responder
        private
        String
        getStatus( GetStatus gs )
        {
            return "[pid " + getPid() + ", uptime " +
                ( System.currentTimeMillis() - startTime ) + "ms ]";
        }

        // Sets the start time and schedules this process's exit at some time in
        // the future.
        protected
        void
        startImpl()
        {
            startTime = System.currentTimeMillis();

            submit(
                new AbstractTask() { protected void runImpl() { exit(); } },
                Duration.fromMillis( 1500 )
            );
        }

        // exposed for RestartFactoryImpl below
        private void stop() { exit(); }
    }

    // RestartFactory implementation which manages instanceof of
    // RestartableProcess
    private
    final
    static
    class RestartFactoryImpl
    implements SimpleRestarter.RestartFactory
    {
        public
        AbstractProcess< ? >
        newProcess() 
        { 
            return new RestartableProcess(); 
        }

        public
        void
        stopProcess( AbstractProcess< ? > proc )
        {
            ( (RestartableProcess) proc ).stop();
        }
    }

    // The child process to call through a SimpleRestarter.Proxy to get the
    // status of the currently running RestartableProcess
    private
    final
    static
    class ProxyCaller
    extends AbstractVoidProcess
    {
        // SimpleRestarter from the parent process. We use it in startImpl() to
        // generate a proxy
        private final SimpleRestarter restarter;

        private
        ProxyCaller( SimpleRestarter restarter )
        {
            // Make sure to include rpc client behavior. We don't use it
            // directly in the source here, but it is needed by the
            // SimpleRestarter.Proxy
            super( ProcessRpcClient.create() );

            this.restarter = restarter;
        }
    
        // Pulse to periodically get the status of the active restartable
        // process
        private
        final
        class CallPulse
        extends AbstractPulse
        {
            // proxy through which rpc calls will go
            private final SimpleRestarter.Proxy proxy;
    
            private
            CallPulse( SimpleRestarter.Proxy proxy )
            {
                super( 
                    Duration.fromMillis( 200 ),
                    ProxyCaller.this.getActivityContext()
                );
    
                this.proxy = proxy;
            }
    
            // callback for get status calls
            private
            final
            class ProxyRpcHandler
            extends ProcessRpcClient.AbstractResponseHandler
            {
                @Override
                protected
                void
                rpcSucceeded( Object resp )
                {
                    code( "Current proxy target replied:", resp );
                    pulseDone();
                }

                // we're willing to ignore timeouts since they're most likely
                // just poor timing, representing the case in which a
                // restartable process exits but this pulse begins an rpc to it
                // before the exit notification is propagated to the proxy. In
                // that case the rpc will never complete without a timeout.
                @Override
                protected
                void
                rpcFailed( Throwable th )
                {
                    if ( ! ( th instanceof TimeoutException ) ) fail( th );
                }
            }
    
            protected
            void
            beginPulse()
            {
                // start an rpc with an aggressive timeout, since if the rpc is
                // received at the target then it will complete very quickly.
                // See note above in rpcFailed()
                proxy.beginRpc( 
                    new RestartableProcess.GetStatus(), 
                    Duration.fromMillis( 40 ),
                    new ProxyRpcHandler() 
                );
            }
        }
 
        // start the call pulse and then schedule this process to exit after
        // some fixed time
        protected
        void
        startImpl()
        {
            code( "proxy caller starting" );

            SimpleRestarter.Proxy proxy = 
                restarter.createProxy( "p1", getActivityContext() );

            new CallPulse( proxy ).start();

            submit(
                new AbstractTask() { protected void runImpl() { exit(); } },
                Duration.fromSeconds( 10 )
            );
        }
    }

    // We only take action on exit of the ProxyCaller. The other child (the
    // active RestartableProcess) will be stopped once we call exit() by the
    // SimpleRestarter behavior in its shutdown sequence.
    @Override
    protected
    void
    childExited( AbstractProcess< ? > child,
                 ProcessExit< ? > exit )
    {
        if ( ! exit.isOk() ) fail( exit.getThrowable() );

        if ( child instanceof ProxyCaller ) 
        {
            exit();
            code( "Called exit()" );
        }
    }

    // Arrange for management of a process keyed as 'p1' and then spawn a
    // ProxyCaller to call into it
    protected
    void
    startImpl()
    {
        behavior( SimpleRestarter.class ).
            manage( "p1", new RestartFactoryImpl() );
 
        spawn( new ProxyCaller( behavior( SimpleRestarter.class ) ) );
    }
}
