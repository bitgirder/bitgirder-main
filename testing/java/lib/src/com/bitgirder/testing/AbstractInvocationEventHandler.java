package com.bitgirder.testing;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
abstract
class AbstractInvocationEventHandler
implements InvocationEventHandler
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public void invocationStarted( InvocationDescriptor id ) {}

    public
    void
    invocationStarted( InvocationDescriptor id,
                       long startTime )
    {}

    public
    void
    invocationCompleted( InvocationDescriptor id,
                         Throwable th )
    {}

    // th null if and only if invocation succeeded
    public
    void
    invocationCompleted( InvocationDescriptor id,
                         Throwable th,
                         long endTime )
    {
        invocationCompleted( id, th );
    }
}
