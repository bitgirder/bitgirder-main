package com.bitgirder.demo.process;

import com.bitgirder.process.AbstractPulse;
import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.ProcessActivity;

import com.bitgirder.concurrent.Duration;

import com.bitgirder.demo.Demo;

// Demo of how to manage periodic activities with an AbstractPulse. This process
// fires off two separate pulses -- one demonstrating the behavior of a fast
// activity which completes immediately on the process thread, and another that
// simulates the behavior of a pulse for which the activity is some asynchronous
// operation which completes in a separate location than the method
// (beginPulse()) in which it begins.
@Demo
final
class AbstractPulseDemo
extends AbstractVoidProcess
{
    // Pulse for an activity which completes immediately in the method in which
    // it is started.
    private
    final
    static
    class SimpleActivityPulse
    extends AbstractPulse
    {
        private
        SimpleActivityPulse( ProcessActivity.Context ctx,
                             Duration pulse )
        {
            super( pulse, ctx );
        }

        // beginPulse() is called every time the pulse is triggered. It is up to
        // this class to call pulseDone() at some point. In this case it is
        // called right here.
        protected
        void
        beginPulse()
        {
            // this is the simple activity: print something to the log
            code( "Simple activity executing" );

            // signal that this specific run is done
            pulseDone();
        }
    }

    // pulse demonstrating how to manage a pulsed activity that involves some
    // async operation which begins in beginPulse() but completes in some
    // callback at some later point.
    private
    final
    static
    class AsyncOpPulse
    extends AbstractPulse
    {
        // how long should our simulated async op take
        private final Duration opDelay;;

        private
        AsyncOpPulse( Duration opDelay,
                      ProcessActivity.Context ctx,
                      Duration pulse )
        {
            super( pulse, ctx );

            this.opDelay = opDelay;
        }

        protected
        void
        beginPulse()
        {
            code( "Starting async op" );

            // simulate some async activity by just submitting a delayed task
            // that completes this pulse. In more realistic scenarios what would
            // happen here is that we would call some method to start some
            // operation and end up passing in a callback that ultimately calls
            // pulseDone() at the end of its execution
            submit(
                new AbstractTask() { 
                    protected void runImpl() 
                    { 
                        code( "Completing async op" );
                        pulseDone(); 
                    }
                },
                opDelay
            );
        }
    }

    protected 
    void 
    startImpl() 
    { 
        // start a simple pulse
        Duration simplePulse = Duration.fromMillis( 500 );
        new SimpleActivityPulse( getActivityContext(), simplePulse ).start(); 

        // start a pulse that performs some async op
        Duration opDelay = Duration.fromMillis( 100 );
        Duration asyncPulse = Duration.fromSeconds( 1 );
        new AsyncOpPulse( opDelay, getActivityContext(), asyncPulse ).start();

        // set a timer to exit the program after a few seconds
        submit(
            new AbstractTask() { protected void runImpl() { exit(); } },
            Duration.fromSeconds( 5 )
        );
    }
}
