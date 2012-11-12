package com.bitgirder.jetty7;

import org.eclipse.jetty.util.component.LifeCycle;

public
abstract
class AbstractLifeCycleListener
implements LifeCycle.Listener
{
    public
    void
    lifeCycleFailure( LifeCycle event,
                      Throwable cause )
    {}

    public void lifeCycleStarted( LifeCycle event ) {}
    public void lifeCycleStarting( LifeCycle event ) {}
    public void lifeCycleStopped( LifeCycle event ) {}
    public void lifeCycleStopping( LifeCycle event ) {}
}
