package com.bitgirder.demo.process;

import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.ProcessRpcClient;
import com.bitgirder.process.ProcessExit;

import com.bitgirder.concurrent.Duration;

import com.bitgirder.demo.Demo;

import java.util.Random;

// Demo of making rpc calls. See EchoServer.java for more about the server
// itself.
@Demo
final
class ProcessRpcDemo
extends AbstractVoidProcess
{
    // srv is instantiated immediately; started in startImpl();
    private EchoServer srv = new EchoServer( Duration.fromMillis( 500 ) );

    // Our client process. Note that we make this an innner class so it has
    // direct access to the srv field in the enclosing instance. Since that
    // value is final and, from the perspective of this class, immutable, it is
    // okay to reference it from inside this child's process thread for the
    // purposes of starting rpc calls.
    private
    final
    class ExampleClient
    extends AbstractVoidProcess
    {
        // track how many calls remain before we exit. All increments complete
        // before any decrements, so we can safely use 0 as the indicator that
        // all calls are done and that it is time to exit
        private int callsRemaining;

        // make sure to include the rpc client behavior for this type
        private ExampleClient() { super( ProcessRpcClient.create() ); }

        // conditionally exit if the previous call was the last
        private void callCompleted() { if ( --callsRemaining == 0 ) exit(); }

        // implementation of rpc handler which prints a small message about the
        // call result and decrements the call count
        private
        final
        class RpcHandler
        extends ProcessRpcClient.AbstractResponseHandler
        {
            private
            void
            rpcComplete( ProcessRpcClient.Call call,
                         String msgTail )
            {
                code( "Rpc", call.getRequest(), msgTail );
                callCompleted();
            }

            @Override
            public
            void
            rpcFailed( Throwable th,
                       ProcessRpcClient.Call call )
            {
                rpcComplete( 
                    call, "failed with: " + th.getClass().getSimpleName() );
            }

            @Override
            public
            void
            rpcSucceeded( Object resp,
                          ProcessRpcClient.Call call )
            {
                rpcComplete( call, "succeeded with: " + resp );
            }
        }

        // util method to begin an rpc call with some request object. This
        // method uses the simple call invocation of ProcessRpcClient in which a
        // destination, request, and callback are provided. This is a compact
        // and simple way to begin calls which will not have timeouts and which
        // do not require any additional request context or disambiguation.
        private
        void
        beginRpc( Object req )
        {
            behavior( ProcessRpcClient.class ).
                beginRpc( srv, req, new RpcHandler() );

            ++callsRemaining;
        }

        // start some basic calls with no timeouts; one success and failure each
        private
        void
        startSimpleCalls()
        {
            beginRpc( new EchoServer.ImmediateEcho( "immediate" ) );
            beginRpc( new EchoServer.ImmediateEcho( "failMe" ) );

            beginRpc( new EchoServer.ChildEcho( "child" ) );
            beginRpc( new EchoServer.ChildEcho( "failMe" ) );

            beginRpc( new EchoServer.AsyncEcho( "async" ) );
            beginRpc( new EchoServer.AsyncEcho( "failMe" ) );
        }

        // start a call that will timeout
        private
        void
        startTimeoutCall()
        {
            behavior( ProcessRpcClient.class ).
                beginRpc(
                    new ProcessRpcClient.Call.Builder().
                        setDestination( srv ).
                        setRequest( new EchoServer.AsyncEcho( "timeout" ) ).
                        setTimeout( Duration.fromMillis( 100 ) ).
                        setResponseHandler( new RpcHandler() ).
                        build()
                );
 
            ++callsRemaining;
        }

        // upon start we fire off all of our various rpc calls
        protected
        void
        startImpl()
        {
            code( "Starting rpc calls" );

            startSimpleCalls();
            startTimeoutCall();
        }
    }

    // In the demo process we wait until all of the children exit. Once they do
    // there will be only one process left -- the server, which we then stop.
    //
    // Note that if any child fails for some reason we immediately fail,
    // possibly leaving behind child processes. In production code where the
    // parent may exit long before the application the parent should ensure that
    // all children are stopped before exiting or failing. For this demo though
    // we omit that extra logic.
    @Override
    protected
    void
    childExited( AbstractProcess< ? > child,
                 ProcessExit< ? > exit )
    {
        // fail if the child failed
        if ( ! exit.isOk() ) fail( exit.getThrowable() );

        // otherwise if this was the last child stop the server
        if ( child instanceof ExampleClient )
        {
            code( "ExampleClient", child.getPid(), "exited" );
            if ( getChildCount() == 1 ) srv.stop();
        }
        else code( "EchoServer", child.getPid(), "exited" );

        // exit when all children are gone
        if ( ! hasChildren() ) exit();
    }

    // start the demo. We spawn the server first and then some number of
    // children. To keep things slightly interesting we introduce a little bit
    // of random jitter around the start time of the various children.
    protected
    void
    startImpl()
    {
        code( "Spawned server as", spawn( srv ) );

        Random rand = new Random();

        for ( int i = 0; i < 4; ++i ) 
        {
            // do the child spawn at some future time according to a slight
            // randomized delay
            submit(
                new AbstractTask() {
                    protected void runImpl() {
                        code( "Spawned client", spawn( new ExampleClient() ) );
                    }
                },
                Duration.fromMillis( rand.nextInt( 50 ) )
            );
        }
    }
}
