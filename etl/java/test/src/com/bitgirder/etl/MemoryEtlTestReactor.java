package com.bitgirder.etl;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.ProcessActivity;
import com.bitgirder.process.Processes;

public
final
class MemoryEtlTestReactor
implements EtlTestReactor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();
 
    private final AbstractProcess< ? > stateMgr = 
        EtlTests.createMemoryStateManager();

    public AbstractProcess< ? > getStateManager() { return stateMgr; }

    public
    void
    startTestProcesses( ProcessActivity.Context ctx,
                        Runnable onComplete )
    {
        ctx.spawn( stateMgr, Processes.< Object >getNoOpExitListener() );
        onComplete.run();
    }

    public
    void
    stopTestProcesses( ProcessActivity.Context ctx )
    {
        Processes.sendStop( stateMgr, ctx );
    }
}
