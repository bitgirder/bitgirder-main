package com.bitgirder.jetty7;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.Stoppable;

import org.eclipse.jetty.util.component.LifeCycle;

class Jetty7LifeCycleManager< L extends LifeCycle >
extends AbstractVoidProcess
implements Stoppable
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final L lifeCycle;

    Jetty7LifeCycleManager( Builder< L, ? > b )
    {
        super( b );

        this.lifeCycle = inputs.notNull( b.lifeCycle, "lifeCycle" );
    }

    Jetty7LifeCycleManager( L lifeCycle )
    {
        this(
            new Builder< L, Builder< L, ? > >() {}.setLifeCycle( lifeCycle ) );
    }

    final L lifeCycle() { return lifeCycle; }

    private
    final
    class LifeCycleListener
    extends AbstractLifeCycleListener
    {
        @Override
        public
        void
        lifeCycleFailure( LifeCycle event,
                          Throwable cause )
        {
            fail( 
                new Exception( 
                    "LifeCycle failed (see chained cause)", cause ) );
        }

        @Override public void lifeCycleStopped( LifeCycle event ) { exit(); }
    }

    protected 
    final
    void 
    startImpl()
        throws Exception
    {
        lifeCycle.addLifeCycleListener( new LifeCycleListener() );
        lifeCycle.start();
    }

    private
    void
    initStop()
        throws Exception
    {
        if ( ! ( lifeCycle.isStopped() || lifeCycle.isStopping() ) ) 
        {
            lifeCycle.stop();
        }
    }

    public
    final
    void
    stop()
    {
        submit(
            new AbstractTask() { 
                protected void runImpl() throws Exception { initStop(); } } );
    }

    static
    class Builder< L extends LifeCycle, B extends Builder< L, ? > >
    extends AbstractVoidProcess.Builder< B >
    {
        private L lifeCycle;

        public
        final
        B
        setLifeCycle( L lifeCycle )
        {
            this.lifeCycle = inputs.notNull( lifeCycle, "lifeCycle" );
            return castThis();
        }
    }
}
