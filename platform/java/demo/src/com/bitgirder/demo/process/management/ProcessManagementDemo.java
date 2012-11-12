package com.bitgirder.demo.process.management;

import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.ProcessExit;
import com.bitgirder.process.ProcessRpcServer;
import com.bitgirder.process.ProcessRpcClient;
import com.bitgirder.process.AbstractPulse;

import com.bitgirder.process.management.ProcessManager;
import com.bitgirder.process.management.ProcessThrashTracker;
import com.bitgirder.process.management.ManagerProxy;
import com.bitgirder.process.management.AbstractProcessControl;

import com.bitgirder.event.EventBehavior;
import com.bitgirder.event.EventManager;

import com.bitgirder.concurrent.Duration;

import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.demo.Demo;

import java.util.Map;

// Demonstration of some uses of ProcessManager. We start 3 managed processes.
// The first we set up to exit quickly and often, eventually seeing it exceed
// its restart threshold. The second and third we have exit less often so that
// they are restarted without exceeding the threshold. We make repeated rpc
// calls to the 2nd and 3rd managed processes to have each report its own
// identity. For the 2nd we call directly to the process using the most recently
// stored reference set by an event receiver that we have listening for manager
// events. For the 3rd we call to the active process via a ManagerProxy.
@Demo
final
class ProcessManagementDemo
extends AbstractVoidProcess
{
    private final static State state = new State();

    // Request signature class for Worker rpc call
    private final static class GetIdentity {}

    // Written to by the manager event receiver; read from by the rpc caller
    // which calls directly to the managed worker
    private final Map< String, Worker > workers = Lang.newMap();

    // The manager proxy we'll put in front of the 3rd managed process id
    private ManagerProxy mp;

    private
    ProcessManagementDemo()
    {
        // Note that we have to include ProcessRpcServer for use with
        // ManagerProxy (which makes rpc calls to the manager to bootstrap
        // itself when it is spawned)
        super(
            ProcessRpcClient.create(),
            ProcessRpcServer.createStandard(),
            EventBehavior.create( EventManager.create() ),
            ProcessManager.create()
        );
    }

    // Event receiver to process selected events from this process's process
    // manager (and only that manager). In less trivial applications this would
    // be active in a process that is not the same as the process including the
    // manager itself.
    private
    final
    class ManagerEventReceiver
    extends ProcessManager.AbstractManagerEventReceiver
    {
        private
        ManagerEventReceiver()
        {
            super( 
                ProcessManagementDemo.this.behavior( ProcessManager.class ),
                ProcessManagementDemo.this.getActivityContext()
            );
        }

        @Override
        protected
        void
        processStarted( ProcessManager.ProcessStarted ps )
        {
            code( 
                "Got start event for", ps.getId() + ":",
                ps.getProcess().getPid() );

            workers.put( (String) ps.getId(), (Worker) ps.getProcess() );
        }

        @Override
        protected
        void
        processExited( ProcessManager.ProcessExited pe )
        {
            code( "Got exit for", pe.getId() + ":", pe.getExit().getPid() );
            workers.remove( (String) pe.getId() );
        }

        @Override
        protected
        void
        restartDeclined( ProcessManager.RestartDeclined rd )
        {
            code( "Further restarts will not be attempted for", rd.getId() );
        }
    }

    // Simple managed process. Responds to a basic rpc call and exits of its own
    // volition (simulating crash) after ttl.
    private
    final
    static
    class Worker
    extends AbstractVoidProcess
    {
        // Time-to-live for this worker after start
        private final Duration ttl;

        private
        Worker( Duration ttl )
        {
            super( ProcessRpcServer.createStandard() );

            this.ttl = ttl;
        }

        // Simple rpc call to get some info about this worker
        @ProcessRpcServer.Responder
        private
        String
        respond( GetIdentity req )
        {
            return "I am " + getClass().getSimpleName() + " " + getPid();
        }

        protected 
        void 
        startImpl()
        {
            // Set to exit after ttl
            submit(
                new AbstractTask() { protected void runImpl() { exit(); } },
                ttl
            );
        }

        // Expose exit() for use by WorkerControl.stopProcess()
        private void stop() { exit(); }
    }

    // ProcessControl that manages instances of type Worker
    private
    final
    static
    class WorkerControl
    extends AbstractProcessControl< Worker >
    {
        // passed to all Worker instances started by this control
        private final Duration ttl;

        private
        WorkerControl( Duration ttl )
        {
            // All instances of this control will refuse to perform more than 10
            // starts in a 5 second period
            super( 
                ProcessThrashTracker.create( 10, Duration.fromSeconds( 5 ) ) 
            );

            this.ttl = ttl;
        }

        // Called by process manager when (re)starting a managed target
        protected Worker newProcessImpl() { return new Worker( ttl ); }

        // Called by process manager when stopping a managed target
        public void stopProcess( Worker w ) { w.stop(); }
    }

    // Start managing a process that will exceed its restart threshold.
    private
    void
    startThrasher()
    {
        behavior( ProcessManager.class ).
            manage( "id1", new WorkerControl( Duration.fromMillis( 10 ) ) );
    }

    // Abstract pulse class to begin an rpc every pulse. 
    private
    abstract
    class RpcCallPulse
    extends AbstractPulse
    {
        // Id with which this pulse is associated. Used in log statements and
        // optionally to determine the current rpc destination
        private final String id;

        private
        RpcCallPulse( String id )
        {
            super( Duration.fromMillis( 750 ), self() );

            this.id = id;
        }

        // Gets the current rpc destination for the given id. May return null if
        // there is no known target at the moment (for instance, if this call
        // arrives when the previous target at id has exited but before a new
        // one has taken its place).
        abstract
        AbstractProcess< ? >
        getRpcDestination( String id );

        // start an rpc call if possible and immediately complete the pulse. We
        // don't tie the calling of pulseDone() to the rpc handler completion
        // since it could be the case that dest is non-null but has exited
        // already, in which case the rpc will never complete. Not a big deal
        // (one less print statement), but we don't want to have the pulse
        // itself hang forever because of this unlikely scenario.
        protected
        void
        beginPulse()
        {
            AbstractProcess< ? > dest = getRpcDestination( id );

            if ( dest != null )
            {
                beginRpc( dest, new GetIdentity(), new DefaultRpcHandler() {
                    @Override protected void rpcSucceeded( Object resp ) 
                    {
                        code( 
                            "Managed process with id", id, "responded:", resp );
                    }
                });
            }

            pulseDone();
        }
    }

    // RpcCallPulse that calls directly into the current Worker instance managed
    // for the given id.
    private
    final
    class DirectCallPulse
    extends RpcCallPulse
    {
        private DirectCallPulse( String id ) { super( id ); }

        // could potentially be null
        AbstractProcess< ? >
        getRpcDestination( String id )
        {
            return workers.get( id );
        }
    }

    // manage and begin sending rpc calls directly to a Worker
    private
    void
    startManuallyTrackedRpcTarget()
    {
        behavior( ProcessManager.class ).
            manage( "id2", new WorkerControl( Duration.fromSeconds( 1 ) ) );

        new DirectCallPulse( "id2" ).start();
    }

    // An RpcCallPulse that sends all calls through the ManagerProxy mp
    private
    final
    class ManagerProxyCallPulse
    extends RpcCallPulse
    {
        private ManagerProxyCallPulse( String id ) { super( id ); }

        AbstractProcess< ? > getRpcDestination( String id ) { return mp; }
    }

    // Manage a process and start sending it rpc calls via a ManagerProxy
    private
    void
    startManagerProxy()
    {
        mp = 
            behavior( ProcessManager.class ).
                manageAndProxy( 
                    "id3", new WorkerControl( Duration.fromSeconds( 1 ) ) );
        
        spawn( mp );

        new ManagerProxyCallPulse( "id3" ).start();
    }

    @Override
    protected
    void
    childExited( AbstractProcess< ? > child,
                 ProcessExit< ? > exit )
    {
        state.isTrue( child instanceof ManagerProxy );
        if ( ! exit.isOk() ) fail( exit.getThrowable() );
        
        // there may still be children (Workers, which are right now also being
        // stopped by the process manager) but the only child on which we needed
        // to wait was mp, so for our purposes as the including process shutdown
        // is now done
        shutdownDone();
    }

    protected
    void
    startImpl()
    {
        // set the event receiver to manage the workers map
        behavior( EventBehavior.class ).
            subscribe( 
                behavior( ProcessManager.class ).getLifecycleTopic(), 
                new ManagerEventReceiver() 
            );

        // start managed processes and associated rpc call pulses
        startThrasher();
        startManuallyTrackedRpcTarget();
        startManagerProxy();

        // schedule our exit for a little bit later
        submit( 
            new AbstractTask() { 
                protected void runImpl() { mp.stop(); exit(); } 
            },
            Duration.fromSeconds( 4 ) 
        );
    }
}
