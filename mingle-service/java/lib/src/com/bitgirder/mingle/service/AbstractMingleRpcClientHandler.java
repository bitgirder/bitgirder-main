package com.bitgirder.mingle.service;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.model.MingleServiceRequest;
import com.bitgirder.mingle.model.MingleServiceResponse;

// This class has no abstract methods, but subclasses are responsible for
// overriding one of the failure methods if failures are to be handled
public
abstract
class AbstractMingleRpcClientHandler
implements MingleRpcClient.Handler
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    protected void rpcSucceeded() {}

    protected
    void
    rpcSucceeded( MingleServiceResponse mgResp )
    {
        rpcSucceeded(); 
    }

    protected
    void
    rpcSucceeded( MingleServiceResponse mgResp,
                  MingleServiceRequest mgReq )
    {
        rpcSucceeded( mgResp );
    }

    public
    void
    rpcSucceeded( MingleServiceResponse mgResp,
                  MingleServiceRequest mgReq,
                  MingleRpcClient mgCli )
    {
        rpcSucceeded( mgResp, mgReq );
    }

    protected void rpcFailed() {}

    protected void rpcFailed( Throwable th ) { rpcFailed(); }

    protected
    void
    rpcFailed( Throwable th,
               MingleServiceRequest mgReq )
    {
        rpcFailed( th );
    }

    public
    void
    rpcFailed( Throwable th,
               MingleServiceRequest mgReq,
               MingleRpcClient mgCli )
    {
        rpcFailed( th, mgReq );
    }
}
