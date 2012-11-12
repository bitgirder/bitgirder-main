package com.bitgirder.demo.process;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.process.ProcessBehavior;
import com.bitgirder.process.ProcessActivity;
import com.bitgirder.process.ProcessRpcClient;
import com.bitgirder.process.ProcessExit;
import com.bitgirder.process.ProcessId;
import com.bitgirder.process.AbstractProcess;

import com.bitgirder.concurrent.Duration;

import java.util.Map;
import java.util.List;

// Demonstration of a nontrivial process behavior and activity. Instances of
// this class could be used to manage a process which should be long-lived but
// may periodically exit for some reason. If it does, this class will restart it
// until shutdown begins, at which point it will stop all managed processes.
// This is by no means a complete implementation of a manger/restart framework,
// so many details and functions are left out (for a more fully-featured
// version, see the classes in com.bitgirder.process.management).
//
// This behavior has two methods that will be of interest to calling code. One
// is to manage a new process, and it takes a key and a restart factory. The
// other is to create a new proxy, which takes a key and an activity context and
// returns a new proxy that will forward rpc requests to whatever process it
// believes to be the currently active one for the key with which it was
// created.
//
// Much of the complexity in this class arises from these two areas as well:
// managing restarts and communicating restarts to the proxy in a way that takes
// into account the fact that the SimpleRestarter instance runs in one process
// thread and that the Proxy runs in another (most likely).
final
class SimpleRestarter
extends ProcessBehavior
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    // RestartFactory instances by key
    private final Map< String, RestartFactory > factories = Lang.newMap();

    // Active processes by key
    private final Map< String, AbstractProcess< ? > > activeProcs = 
        Lang.newMap();

    // Active keys by pid
    private final Map< ProcessId, String > activeKeys = Lang.newMap();

    // Any proxies associated with a given key
    private final Map< String, List< Proxy > > proxies = Lang.newMap();

    // Tracks whether shutdown has begun
    private boolean shutdownBegun;

    SimpleRestarter() {}

    // nothing to do on start
    protected void startImpl() {}

    // called by childExited() and beginShutdown() to see if shutdown is
    // complete
    private
    void
    checkShutdownState()
    {
        if ( shutdownBegun && ! hasChildren() ) 
        {
            code( "restarter shutdown is complete" );
            shutdownComplete();
        }
    }

    // shutdown any active processes (there may be none at all) and check for
    // whether or not shutdown can be completed.
    @Override
    protected
    void
    beginShutdown()
    {
        code( "restarter beginning shutdown" );

        shutdownBegun = true;

        for ( Map.Entry< String, AbstractProcess< ? > > e : 
                activeProcs.entrySet() )
        {
            RestartFactory f = state.get( factories, e.getKey(), "factories" );
            f.stopProcess( e.getValue() );
        }

        checkShutdownState();
    }

    // Receive the exit of some process, asserting that it was one spawned by
    // this restarter in the first place, and restart it if we are not in
    // shutdown mode, or check for shutdown completion otherwise.
    @Override
    protected
    void
    childExited( AbstractProcess< ? > child,
                 ProcessExit< ? > exit )
    {
        String key = state.get( activeKeys, child.getPid(), "activeKeys" );
        state.remove( activeProcs, key, "activeProcs" );

        code( "Process", child.getPid(), "keyed at", key, "exited" );

        // inform any proxies of the status change
        List< Proxy > pxList = proxies.get( key );
        if ( pxList != null ) for ( Proxy p : pxList ) p.targetExited();

        if ( shutdownBegun ) checkShutdownState(); else startNext( key );
    }

    // start the next process for the given management key
    private
    void
    startNext( String key )
    {
        AbstractProcess< ? > next =
            state.get( factories, key, "factories" ).newProcess();
        
        ProcessId pid = spawn( next );
        code( "Started next process keyed at", key, "with pid", pid );

        activeKeys.put( pid, key );
        activeProcs.put( key, next );

        // inform any proxies of the start
        List< Proxy > pxList = proxies.get( key );
        if ( pxList != null ) for ( Proxy p : pxList ) p.restarted( next );
    }

    // submit a new factory to be managed at the given key. This may only be
    // called from the process thread of the including class
    void
    manage( String key,
            RestartFactory fact )
    {
        inputs.notNull( key, "key" );
        inputs.notNull( fact, "fact" );

        Lang.putUnique( factories, key, fact );
        startNext( key );
    }

    // util method to register a proxy for the given key. This is run in the
    // including process's process thread (eg, this behavior's process thread)
    // at some point after the call to createProxy() which created the Proxy we
    // are attempting to install. Because of this we just warn if the key is
    // unmanaged. 
    //
    // This is largely a side-effect of our trying to keep the code here
    // approachable, and so we sacrifice a level of failure detection that we
    // otherwise would not in a production library (to be more thorough,
    // createProxy() wouldn't be a method call but would instead be an rpc call
    // made by the caller desiring the proxy, allowing us to communicate the
    // failure back to the caller).
    private
    void
    installProxy( String key,
                  Proxy proxy )
    {
        if ( factories.containsKey( key ) )
        {
            Lang.putAppend( proxies, key, proxy );

            AbstractProcess< ? > proc = activeProcs.get( key );
            if ( proc != null ) proxy.restarted( proc );
        }
        else warn( "Proxy created for unmanaged key:", key );
    }

    // Creates a Proxy that will eventually be wired up to this instance. The
    // "eventual" is because this method is designed to be called from the
    // process thread of the process associated with ctx, meaning that we create
    // the Proxy here but do not install it with this SimpleRestater instance
    // until the install task runs.
    //
    // The calling process represented by ctx must include the ProcessRpcClient
    // behavior, which we check.
    Proxy
    createProxy( final String key,
                 ProcessActivity.Context ctx )
    {
        inputs.notNull( key, "key" );
        inputs.notNull( ctx, "ctx" );

        inputs.isTrue( 
            ctx.hasBehavior( ProcessRpcClient.class ),
            "process doesn't include ProcessRpcClient" );

        final Proxy res = new Proxy( ctx );

        submit(
            new Runnable() {
                public void run() { installProxy( key, res ); }
            }
        );

        return res;
    }

    // Interface to create/stop managed processes
    static
    interface RestartFactory
    {
        public
        AbstractProcess< ? >
        newProcess();

        public
        void
        stopProcess( AbstractProcess< ? > proc );
    }

    // ProcessActivity that encapsulates the logic of tracking which actual
    // process is the current one for a given managed key, and transparently
    // routing rpc requests to that process.
    final
    static
    class Proxy
    extends ProcessActivity
    {
        // The current active process for the key for which this Proxy instance
        // was created. May be null if no process has been restarted or between
        // restarts.
        private AbstractProcess< ? > activeProc;

        // A list of held requests received while activeProc is null
        private final List< HeldRequest > heldReqs = Lang.newList();

        private Proxy( ProcessActivity.Context ctx ) { super( ctx ); }

        // Encapsulates the details needed to start a request once activeProc
        // becomes non-null
        private
        final
        static
        class HeldRequest
        {
            private final Object req;
            private final Duration timeout;
            private final ProcessRpcClient.ResponseHandler h;

            private
            HeldRequest( Object req,
                         Duration timeout,
                         ProcessRpcClient.ResponseHandler h )
            {
                this.req = req;
                this.timeout = timeout;
                this.h = h;
            }
        }

        // util method to actually begin an rpc to activeProc
        private
        void
        fireRequest( Object req,
                     Duration timeout,
                     ProcessRpcClient.ResponseHandler h )
        {
            behavior( ProcessRpcClient.class ).
                beginRpc(
                    new ProcessRpcClient.Call.Builder().
                        setRequest( req ).
                        setDestination( activeProc ).
                        setTimeout( timeout ).
                        setResponseHandler( h ).
                        build()
                );
        }

        // public method to start rpc. If activeProc is null we hold the request
        // until it is not null, otherwise we begin it immediately.
        //
        // Note that there is actually a bug in the setup here which is worth a
        // remark even though we ignore it for the purposes of this demo. The
        // timeout is not put into effect until it is used in the fireRequest()
        // method. But, that method may never be called if activeProc remains
        // null for a long time. To work around this we would ideally wrap the
        // response handler in one which listens to both actual rpc responses
        // but also can send a TimeoutException directly from this activity if a
        // request is Held longer than the timeout. In such a case we'd also
        // want to do things like have fireRequest notice that the request timed
        // out in the hold queue and skip it, as well as adjust the effective
        // timeout set in the rpc to deduct whatever time was spent waiting in
        // the hold queue. Again -- not an issue in our demos but worth thinking
        // about.
        public
        void
        beginRpc( Object req,
                  Duration timeout,
                  ProcessRpcClient.ResponseHandler h )
        {
            inputs.notNull( req, "req" );
            inputs.notNull( timeout, "timeout" );
            inputs.notNull( h, "h" );
        
            if ( activeProc == null ) 
            {
                heldReqs.add( new HeldRequest( req, timeout, h ) );
            }
            else fireRequest( req, timeout, h );
        }

        // callback method called by the SimpleRestarter with which this Proxy
        // is associated. This method is called from the SimpleRestarter's
        // process thread, so we complete our handling of it in our own.
        private
        void
        restarted( final AbstractProcess< ? > p )
        {
            submit(
                new AbstractTask() {
                    protected void runImpl() 
                    { 
                        activeProc = p; 

                        for ( HeldRequest hr : heldReqs )
                        {
                            fireRequest( hr.req, hr.timeout, hr.h );
                        }

                        heldReqs.clear();
                    }
                }
            );
        }

        // see note in restarted() about process thread issues
        private
        void
        targetExited()
        {
            submit(
                new AbstractTask() {
                    protected void runImpl() { activeProc = null; }
                }
            );
        }
    }
}
