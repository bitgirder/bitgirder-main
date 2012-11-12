package com.bitgirder.etl;

import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.ProcessActivity;

// All methods are meant to be called from the process thread of the associated
// process
public
interface EtlTestReactor
{
    public
    void
    startTestProcesses( ProcessActivity.Context ctx,
                        Runnable onComplete )
        throws Exception;
    
    public
    void
    stopTestProcesses( ProcessActivity.Context ctx )
        throws Exception;
    
    // will not be called before the onComplete task passed to
    // startTestProcesses has been run by this reactor
    public
    AbstractProcess< ? >
    getStateManager();
}
