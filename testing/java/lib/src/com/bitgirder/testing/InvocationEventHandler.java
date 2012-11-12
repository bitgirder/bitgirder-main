package com.bitgirder.testing;

// Methods may be called concurrently and from any thread
public
interface InvocationEventHandler
{
    public
    void
    invocationStarted( InvocationDescriptor id,
                       long startTime );

    // th null if and only if invocation succeeded
    public
    void
    invocationCompleted( InvocationDescriptor id,
                         Throwable th,
                         long endTime );
}
