package com.bitgirder.mingle.service;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.model.MingleServiceRequest;
import com.bitgirder.mingle.model.MingleServiceResponse;

import com.bitgirder.process.ProcessRpcClient;
import com.bitgirder.process.AbstractProcess;

final
class DirectMingleRpcClient
implements MingleRpcClient
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final AbstractProcess< ? > svc;
    private final ProcessRpcClient cli;

    DirectMingleRpcClient( AbstractProcess< ? > svc,
                           ProcessRpcClient cli )
    {
        this.svc = state.notNull( svc, "svc" );
        this.cli = state.notNull( cli, "cli" );
    }

    private
    final
    class HandlerWrapper
    implements ProcessRpcClient.ResponseHandler
    {
        private final Handler h;

        private HandlerWrapper( Handler h ) { this.h = h; }

        public
        void
        rpcFailed( Throwable th,
                   ProcessRpcClient.Call call )
        {
            h.rpcFailed( 
                th, 
                (MingleServiceRequest) call.getRequest(), 
                DirectMingleRpcClient.this 
            );
        }

        public
        void
        rpcSucceeded( Object resp,
                      ProcessRpcClient.Call call )
        {
            h.rpcSucceeded(
                (MingleServiceResponse) resp,
                (MingleServiceRequest) call.getRequest(),
                DirectMingleRpcClient.this 
            );
        }
    }

    public
    void
    beginRpc( MingleServiceRequest req,
                  Handler h )
    {
        inputs.notNull( req, "req" );
        inputs.notNull( h, "h" );

        cli.beginRpc( svc, req, new HandlerWrapper( h ) );
    }
}
