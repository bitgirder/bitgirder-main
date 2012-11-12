package com.bitgirder.mingle.service;

import com.bitgirder.mingle.model.MingleServiceRequest;
import com.bitgirder.mingle.model.MingleServiceResponse;

public
interface MingleRpcClient
{
    public
    void
    beginRpc( MingleServiceRequest req,
              Handler h );
    
    public
    interface Handler
    {
        public
        void
        rpcFailed( Throwable th,
                   MingleServiceRequest req,
                   MingleRpcClient cli );
        
        public
        void
        rpcSucceeded( MingleServiceResponse resp,
                      MingleServiceRequest req,
                      MingleRpcClient cli );
    }
}
