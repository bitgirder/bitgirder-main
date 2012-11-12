package com.bitgirder.demo.process;

import com.bitgirder.validation.Inputs;

import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.ProcessRpcServer;

import com.bitgirder.concurrent.Duration;

// Simple process that performs various rpc calls in which a caller's string is
// echoed back, possibly after some delay and optionally failing if the test
// value 'failMe' is received.
final
class EchoServer
extends AbstractVoidProcess
{
    private final static Inputs inputs = new Inputs();

    // Marker exception to indicate that the failure-triggering input was
    // received
    final static class EchoException extends Exception {}
 
    // how long to simulate long-running echoes
    private final Duration asyncDelay;

    EchoServer( Duration asyncDelay ) 
    { 
        // To be able to receive and respond to rpc requests a process must
        // include a ProcessRpcServer instance
        super( ProcessRpcServer.createStandard() ); 
        
        this.asyncDelay = inputs.notNull( asyncDelay, "asyncDelay" );
    }

    // Base request class
    private
    static
    abstract
    class Echo
    {
        private final String echoObj;

        private 
        Echo( String echoObj ) 
        { 
            this.echoObj = inputs.notNull( echoObj, "echoObj" ); 
        }

        @Override
        public
        final
        String
        toString()
        {
            return 
                new StringBuilder().
                    append( getClass().getSimpleName() ).
                    append( "[ echoObj: " ).
                    append( echoObj ).
                    append( " ]" ).
                    toString();
        }
    }

    // This process does nothing of its own volition so we just have an empty
    // start
    protected void startImpl() {}

    // Simple direct method by which this process can be stopped by code with a
    // reference to this instance. exit() is a protected method and can't be
    // called directly, so we just put this proxy in front of it.
    void stop() { exit(); }

    // util method used by the various responders below to return the echo
    // result or fail as appropriate
    private
    String
    getEcho( Echo e )
        throws EchoException
    {
        if ( "failMe".equals( e.echoObj ) ) throw new EchoException();
        else return e.echoObj;
    }

    // echo request used to trigger the immediate responder below
    final
    static
    class ImmediateEcho
    extends Echo
    {
        ImmediateEcho( String echoObj ) { super( echoObj ); }
    }

    // Responder methods and classes must be annotated with the @Responder
    // annotation.
    //
    // This method is immediate and so either returns the rpc result (or fails)
    // immediately
    @ProcessRpcServer.Responder
    private 
    String 
    echo( ImmediateEcho e ) 
        throws EchoException
    { 
        return getEcho( e ); 
    }

    // request for the async responder below
    final
    static
    class AsyncEcho
    extends Echo
    {
        AsyncEcho( String echoObj ) { super( echoObj ); }
    }

    // An async responder. Like an immediate responder, this runs directly in
    // the server process's process thread. This responder simulates a
    // long-running operation by taking the input and the response context and
    // finally completing after the asyncDelay.
    @ProcessRpcServer.Responder
    private
    void
    echo( final AsyncEcho e,
          final ProcessRpcServer.ResponderContext< String > ctx )
    {
        submit(
            new AbstractTask() {
                protected void runImpl() 
                { 
                    try { ctx.respond( getEcho( e ) ); }
                    catch ( Throwable th ) { ctx.fail( th ); }
                }
            },
            asyncDelay
        );
    }

    // request for child responder below
    final
    static
    class ChildEcho
    extends Echo
    {
        ChildEcho( String echoObj ) { super( echoObj ); }
    }

    // A child responder handler. When a ChildEcho request is received an
    // instnace of this class is created, passing the request to the
    // constructor, and the instance is spawned. Whenever the instance exits,
    // the rpc will complete in the parent with the the exit result (or
    // failure).
    @ProcessRpcServer.Responder
    private
    final
    class ChildResponder
    extends AbstractProcess< String >
    {
        private final ChildEcho e;

        private ChildResponder( ChildEcho e ) { this.e = e; }

        // A fairly simplistic implementation. Like the async responder, we'd
        // likely be doing something long running or generally worth the
        // overhead of performing in a child process. 
        protected
        void
        startImpl()
            throws EchoException
        {
            exit( getEcho( e ) ); 
        }
    }
}
