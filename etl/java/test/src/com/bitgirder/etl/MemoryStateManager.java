package com.bitgirder.etl;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.ProcessRpcServer;

import com.bitgirder.mingle.model.MingleIdentifiedName;

import java.util.Map;

final
class MemoryStateManager
extends AbstractVoidProcess
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Map< MingleIdentifiedName, Object > states = Lang.newMap();

    private final Map< MingleIdentifiedName, Object > feedPositions =
        Lang.newMap();

    MemoryStateManager() { super( ProcessRpcServer.createStandard() ); }

    protected void startImpl() {}

    @ProcessRpcServer.Responder
    private
    Object
    handle( EtlProcessors.GetProcessorState gps )
    {
        return states.get( gps.getId() );
    }

    @ProcessRpcServer.Responder
    private
    void
    handle( EtlProcessors.SetProcessorState sps )
    {
        states.put( sps.getId(), sps.getObject() );
    }

    @ProcessRpcServer.Responder
    private
    Object
    handle( EtlProcessors.GetProcessorFeedPosition req )
    {
        return feedPositions.get( req.getId() );
    }

    @ProcessRpcServer.Responder
    private
    void
    handle( EtlProcessors.SetProcessorFeedPosition req )
    {
        feedPositions.put( req.getId(), req.getObject() );
    }
}
