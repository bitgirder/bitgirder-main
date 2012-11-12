package com.bitgirder.systest;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Strings;

import com.bitgirder.concurrent.Duration;

import com.bitgirder.process.ProcessRpcClient;
import com.bitgirder.process.ProcessRpcServer;
import com.bitgirder.process.ProcessExit;
import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.TestServer;

import com.bitgirder.application.ApplicationProcess;

final
class ProcessRpcBenchmark2
extends ApplicationProcess
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final int numMessages;
    private final int tickSize;

    private long startNanos;

    private
    ProcessRpcBenchmark2( Configurator c )
    {
        super( c.mixin( ProcessRpcClient.create() ) );

        this.numMessages = inputs.positiveI( c.numMessages, "numMessages" );
        this.tickSize = inputs.positiveI( c.tickSize, "tickSize" );
    }

    private
    void
    out( Object... msg )
    {
        System.out.println( Strings.join( " ", msg ) );
    }

    @Override
    protected
    void
    childExited( AbstractProcess< ? > child,
                 ProcessExit< ? > exit )
    {
        if ( ! exit.isOk() ) fail( exit.getThrowable() );
        if ( ! hasChildren() ) exit();
    }
    
    private
    final
    class Handler
    extends ProcessRpcClient.DefaultResponseHandler
    {
        private Handler() { super( ProcessRpcBenchmark2.this ); }

        private
        void
        conclude( AbstractProcess< ? > proc )
        {
            long runTimeMillis = ( System.nanoTime() - startNanos ) / 1000000L;
            out(
                "Total run time:",
                Duration.fromMillis( runTimeMillis ).toStringSeconds() );

            behavior( ProcessRpcClient.class ).
                sendAsync( proc, new ProcessRpcServer.Stop() );
        }

        @Override
        public
        void
        rpcSucceeded( Object resp,
                      ProcessRpcClient.Call call )
        {
            int next = ( (Integer) resp ).intValue() + 1;

            if ( next % tickSize == 0 ) out( "Made", next, "calls" );

            if ( next == numMessages ) conclude( call.getDestination() );
            else startRpc( call.getDestination(), next, this );
        }
    }

    private
    void
    startRpc( AbstractProcess< ? > proc,
              int i,
              ProcessRpcClient.ResponseHandler h )
    {
        TestServer.EchoImmediate< Integer > req =
            new TestServer.EchoImmediate< Integer >( i );

        behavior( ProcessRpcClient.class ).beginRpc( proc, req, h );
    }

    protected
    void
    startImpl()
    {
        TestServer srv = TestServer.create();
        spawn( srv );
        
        startNanos = System.nanoTime();
        startRpc( srv, 0, new Handler() );
    }

    private
    final
    static
    class Configurator
    extends ApplicationProcess.Configurator
    {
        private int numMessages;
        private int tickSize = 100;

        @Argument
        private
        void
        setNumMessages( int numMessages )
        {
            this.numMessages = numMessages;
        }

        @Argument
        private
        void
        setTickSize( int tickSize )
        {
            this.tickSize = tickSize;
        }
    }
}
