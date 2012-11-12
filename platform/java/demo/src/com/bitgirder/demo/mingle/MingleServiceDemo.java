package com.bitgirder.demo.mingle;

import com.bitgirder.process.ProcessRpcClient;
import com.bitgirder.process.ProcessExit;
import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.Processes;

import com.bitgirder.mingle.model.MingleServiceRequest;
import com.bitgirder.mingle.model.MingleServiceResponse;
import com.bitgirder.mingle.model.MingleModels;

import com.bitgirder.mingle.service.MingleServices;
import com.bitgirder.mingle.service.MingleServiceEndpoint;
import com.bitgirder.mingle.service.MingleRpcClient;

import com.bitgirder.demo.Demo;

import java.util.List;

// Demo to create a NativeDemoService, place it behind a MingleServiceEndpoint,
// and show how to make calls through that endpoint.
@Demo
final
class MingleServiceDemo
extends AbstractVoidProcess
{
    // The service object
    private NativeDemoService ns;

    // Will be accessed by instances of TestClient in their own process thread,
    // but that's okay since we assign to ep before creating the clients and
    // make no changes thereafter.
    private MingleServiceEndpoint ep;

    // The demo itself needs to include ProcessRpcClient only for use in our
    // calls to Processes.sendStop()
    private MingleServiceDemo() { super( ProcessRpcClient.create() ); }

    // A process that will send some mingle service calls to the demo service
    // behind the endpoint addressed at the given ( namespace, service id )
    // combo.
    private
    final
    class TestClient
    extends AbstractVoidProcess
    {
        // address of the service we'll call
        private final CharSequence ns;
        private final CharSequence svc;

        // number of outstanding calls before we exit
        private int waitCount;

        // the MingleRpcClient configured to send requests to the service
        // endpoint ep
        private MingleRpcClient mgCli;

        private
        TestClient( CharSequence ns,
                    CharSequence svc )
        {
            super( ProcessRpcClient.create() );

            this.ns = ns;
            this.svc = svc;
        }

        // conditionally exit if all calls are done
        private 
        void 
        callDone() 
        { 
            if ( --waitCount == 0 ) 
            {
                code( "Calls have completed; exiting" );
                exit(); 
            }
        }

        // Simple rpc handler for op1 -- just print the mingle response, but
        // fail on an rpc-level failure (that is, the rpc framework itself
        // failed to deliver the request or response -- we have no idea as to
        // the status of the actual request)
        private
        final
        class Op1Handler
        implements MingleRpcClient.Handler
        {
            public
            void
            rpcFailed( Throwable th,
                       MingleServiceRequest req,
                       MingleRpcClient cli )
            {
                fail( th );
            }

            public
            void
            rpcSucceeded( MingleServiceResponse resp,
                          MingleServiceRequest req,
                          MingleRpcClient cli )
            {
                code( 
                    "Got mingle response", MingleModels.inspect( resp ),
                    "for request", MingleModels.inspect( req, true ) );
 
                callDone();
            }
        }

        // start a do-op1 request that should succeed. We make do-op1 requests
        // using the native mingle objects and via mgCli
        private
        void
        startDoOp1Request()
        {
            MingleServiceRequest.Builder b =
                new MingleServiceRequest.Builder().
                    setNamespace( ns ).
                    setService( svc ).
                    setOperation( "do-op1" );
            
            b.params().setString( "string", "floppy" );
            b.params().setIntegral( "copies", 3 );
            b.params().setBoolean( "reverse", true );

            mgCli.beginRpc( b.build(), new Op1Handler() );
            ++waitCount;
        }

        // start a do-op1 request that should fail with a message about an
        // invalid parameter (copies < 0)
        private
        void
        startDoOp1RequestInvalidCopiesValue()
        {
            MingleServiceRequest.Builder b =
                new MingleServiceRequest.Builder().
                    setNamespace( ns ).
                    setService( svc ).
                    setOperation( "do-op1" );
            
            b.params().setString( "string", "fail this" );
            b.params().setIntegral( "copies", -10 );

            mgCli.beginRpc( b.build(), new Op1Handler() );

            ++waitCount;
        }

        private
        void
        startRequests()
        {
            startDoOp1Request();
            startDoOp1RequestInvalidCopiesValue();
        }

        // configure the clients and fire off the requests
        protected 
        void 
        startImpl()
        {
            mgCli =
                MingleServices.
                    createRpcClient( ep, behavior( ProcessRpcClient.class ) );

            startRequests();
        }
    }

    // wait for the TestClients to exit and then send stops to the services and
    // the endpoint using Processes.sendStop()
    @Override
    protected
    void
    childExited( AbstractProcess< ? > child,
                 ProcessExit< ? > exit )
    {
        if ( ! exit.isOk() ) fail( exit.getThrowable() );

        if ( getChildCount() == 3 ) // just the endpoint and 2 services
        {
            ProcessRpcClient cli = behavior( ProcessRpcClient.class );

            Processes.sendStop( ep, cli );
            Processes.sendStop( ns, cli );
        }

        if ( ! hasChildren() ) exit();
    }

    protected
    void
    startImpl()
        throws Exception
    {
        spawn( ns = new NativeDemoService() );

        ep =
            new MingleServiceEndpoint.Builder().
                addRoute( "mingle:demo", "native-service", ns ).
                build();
        
        spawn( ep );

        spawn( new TestClient( "mingle:demo", "native-service" ) );
    }
}
