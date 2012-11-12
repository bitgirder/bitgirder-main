package com.bitgirder.etl;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.process.ProcessRpcServer;
    
public
abstract
class AbstractRecordSetProcessor
extends ProcessRpcServer.AbstractAsyncResponder< Object >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    @ProcessRpcServer.Request
    private EtlRecordSet rs;

    protected AbstractRecordSetProcessor() {}

    // Called from within package by AbstractEtlProcessor
    final
    void
    abortProcess()
    {
        fail( new EtlRecordSetAbortedException() );
    }

    protected final EtlRecordSet recordSet() { return rs; }

    protected 
    abstract
    void
    startProcess()
        throws Exception;

    private
    AbstractEtlProcessor
    etlProc()
    {
        return (AbstractEtlProcessor) getEnclosingProcess();
    }

    protected 
    final 
    void 
    startImpl() 
        throws Exception 
    { 
        etlProc().setActiveProcessor( this );

        startProcess(); 
    }

    @Override
    protected
    final
    void
    afterRespond()
    {
        etlProc().clearActiveProcessor( this );
    }
}
