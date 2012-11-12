package com.bitgirder.demo.process;

import com.bitgirder.validation.State;

import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.ProcessExit;

import com.bitgirder.lang.Lang;

import com.bitgirder.concurrent.Duration;

import com.bitgirder.demo.Demo;

// Demonstration of some basics of AbstractProcess, including task submission,
// AbstractTask, and child processes.
@Demo
public
final
class AbstractProcessBasics
extends AbstractVoidProcess
{
    // State object that we'll use to verify a few things along the way.
    private final static State state = new State();

    // In startImpl() we'll kick off a few separate activities, incrementing the
    // waitCount for each and decrementing it when each completes, making sure
    // not to exit until all children have exited and all activities included in
    // waitCount are complete. Note that we need no special synchronization here
    // since all access occurs on the process thread.
    private int waitCount;

    private AbstractProcessBasics() {}

    // Logic for deciding whether this process should exit
    private
    void
    exitConditional()
    {
        if ( waitCount == 0 && ! hasChildren() ) exit();
    }

    // Funnel method for our various activities to actually do their thing,
    // which in this case is just to echo some object. We also check here for
    // our exit condition. Note that this method is always called from within
    // the process thread, which is why our reads/writes to waitCount are atomic
    // with respect to this process.
    private
    void
    doEcho( Object echoObj )
    {
        code( "Doing echo with object:", echoObj );
        
        --waitCount;
        exitConditional();
    }

    // Simple implementation of AbstractTask which calls our echo method.
    private
    final
    class EchoTask
    extends AbstractTask
    {
        // Object to echo
        private final Object echoObj;

        private EchoTask( Object echoObj ) { this.echoObj = echoObj; }

        // In this case we don't do much, but we could do anything here that
        // doesn't block, and could even declare this method to throw any kind
        // of checked Exception
        protected void runImpl() { doEcho( echoObj ); }
    }

    // Submits a bunch of tasks, some with delays, which will ultimately call
    // doEcho. Note that all of the tasks run here in the parent process on the
    // parent process thread.
    //
    // Also note that we increment waitCount at the very end of the method, even
    // though the submit() calls come before. This is okay because of the
    // property that at most one activity is executing on the process thread at
    // one time. Since this method itself is running on the process thread, we
    // know that none of the submitted tasks will be running, meaning that
    // we are not concerned that waitCount might be decremented before this
    // method exits.
    private
    void
    submitEchoTasks()
    {
        // submit an echo task to run as soon as possible
        submit( new EchoTask( 21 ) );

        // submit an echo task to run no less than 200ms from now (could be more
        // if the queues are overloaded)
        submit( new EchoTask( "hello" ), Duration.fromMillis( 200 ) );

        // often it's easier and more compact to just use an anonymous subclass
        // of AbstractTask, as shown next

        final Object o = Lang.< String >asList( "a", "list", "of", "strings" );

        submit(
            new AbstractTask() { protected void runImpl() { doEcho( o ); } },
            Duration.fromSeconds( 2 )
        );
 
        waitCount = 3;
    }

    // Marker exception for us to distinguish from one we didn't expect in
    // childExited()
    private final static class TestException extends Exception {}

    // Standard callback called by the process libraries when any of our
    // children exit. This method is called on this instance's process thread.
    @Override
    protected
    void
    childExited( AbstractProcess< ? > child,
                 ProcessExit< ? > exit )
    {
        code( "Got exit from child with pid", child.getPid() );

        // some simple assertions just to tie things together: this is an
        // ExampleChild and if it failed then the failure is a TestException

        state.isTrue( child instanceof ExampleChild );

        if ( exit.isOk() ) state.equal( "hello", (String) exit.getResult() );
        else state.isTrue( exit.getThrowable() instanceof TestException );

        // possibly this is the last thing we're waiting on and we'll exit
        exitConditional();
    }

    // A child process that, either immediately or with some delay, exits with
    // 'result' if 'fail' is false or fails with a TestException if 'fail' is
    // true. Note that this is an AbstractProcess< String >. Although the type
    // parameter has to be erased in processExited() above, we still have the
    // benefit within this process of the compiler helping us enfore that we
    // exit with a value of the correct type.
    private
    final
    static
    class ExampleChild
    extends AbstractProcess< String >
    {
        private final String result; // maybe null
        private final Duration delay; // maybe null
        private final boolean fail;

        private
        ExampleChild( String result,
                      Duration delay,
                      boolean fail )
        {
            this.result = result;
            this.delay = delay;
            this.fail = fail;
        }

        private
        void
        completeChild()
        {
            code( "ExampleChild completing" );
            if ( fail ) fail( new TestException() ); else exit( result );
        }

        // This method is called by the process library at some point after this
        // instance is spawned. This method is called on this ExampleChild
        // instance's process thread.
        protected
        void
        startImpl()
        {
            code( "ExampleChild starting" );

            if ( delay == null ) completeChild(); // complete immediately
            else
            {
                // complete after delay
                submit(
                    new AbstractTask() { 
                        protected void runImpl() { completeChild(); }
                    },
                    delay
                );
            }
        }
    }

    // spawn some ExampleChild instances
    private
    void
    spawnChildren()
    {
        // will wait a bit and exit with the string "hello"
        spawn( new ExampleChild( "hello", Duration.fromMillis( 500 ), false ) );

        // will fail immediately with a TestException
        spawn( new ExampleChild( null, null, true ) );
    }

    // Though this implementation does not do so, startImpl() may declare itself
    // to throw any checked exception. Indeed, any throwable thrown from this
    // method, checked or otherwise, will result in the process failing with
    // that throwable.
    protected
    void
    startImpl()
    {
        code( "Parent starting" );

        submitEchoTasks();
        spawnChildren();
    }
}
