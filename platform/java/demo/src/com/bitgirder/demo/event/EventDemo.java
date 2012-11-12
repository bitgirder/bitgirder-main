package com.bitgirder.demo.event;

import com.bitgirder.validation.State;

import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.ProcessExit;

import com.bitgirder.event.EventManager;
import com.bitgirder.event.EventBehavior;
import com.bitgirder.event.EventTopic;
import com.bitgirder.event.ImmediateEventReceiver;

import com.bitgirder.concurrent.Duration;

import com.bitgirder.demo.Demo;

import java.util.UUID;
import java.util.Random;

// Demonstration of publishing and receiving process events
@Demo
final
class EventDemo
extends AbstractVoidProcess
{
    private final static State state = new State();

    // The event manager used by all processes in this application. Just set
    // here as a static field (since it is referenced in this class's
    // constructor).
    private final static EventManager evMgr = EventManager.create();

    // We create two topics just to illustrate their use: "data" simulates
    // actual application data; "control" simulates data about the processes
    // here themselves.
    
    private final static EventTopic< String > dataTopic = 
        EventTopic.create( "data" );

    private final static EventTopic< String > controlTopic =
        EventTopic.create( "control" );

    // Number of publishers that we'll create
    private final int publisherCount = 10;

    // The current number of starts observed via the control topic
    private int starts;

    // The current number of active publishers as inferred via activity on the
    // control topic
    private int active;

    private EventDemo() { super( EventBehavior.create( evMgr ) ); }

    // Base class for our test events. Each event is associated with its
    // publisher by publisherId and contains a unique id to differentiate it
    // from others of the same type from the same publisher.
    private
    static
    abstract
    class TestEvent
    {
        private final UUID uid = UUID.randomUUID();

        private final int publisherId;

        private TestEvent( int publisherId ) { this.publisherId = publisherId; }

        final int getPublisherId() { return publisherId; }

        @Override
        public
        final
        String
        toString()
        {
            return 
                getClass().getSimpleName() + 
                "[ uid: " + uid + ", publisherId: " + publisherId + " ]";
        }
    }

    // Simulated application data
    private
    final
    static
    class DataEvent
    extends TestEvent
    {
        private DataEvent( int publisherId ) { super( publisherId ); }
    }

    // Sent when a publisher starts
    private
    final
    static
    class PublisherStart
    extends TestEvent
    {
        private PublisherStart( int publisherId ) { super( publisherId ); }
    }

    // Sent as a publisher prepares to exit
    private
    final
    static
    class PublisherStop
    extends TestEvent
    {
        private PublisherStop( int publisherId ) { super( publisherId ); }
    }
    
    // Exit if all publishers have started, all stop events have been received,
    // and all publishers have exited
    private
    void
    exitConditional()
    {
        if ( starts == publisherCount && active == 0 && ! hasChildren() ) 
        {
            exit();
        }
    }

    // Reap a publisher exit
    @Override
    protected
    void
    childExited( AbstractProcess< ? > child,
                 ProcessExit< ? > exit )
    {
        if ( ! exit.isOk() ) fail( exit.getThrowable() );
        exitConditional();
    }

    // Simplistic event receiver that just prints its event object to the log
    private
    final
    class DataReceiver
    extends ImmediateEventReceiver
    {
        private DataReceiver() { super( EventDemo.this.getActivityContext() ); }

        protected void receiveEventImpl( Object ev ) { code( "Received", ev ); }
    }

    // Receiver for control events
    private
    final
    class ControlReceiver
    extends ImmediateEventReceiver
    {
        private 
        ControlReceiver()
        {
            super( EventDemo.this.getActivityContext() );
        }

        // Update active/starts counts as appropriate
        protected
        void
        receiveEventImpl( Object ev )
        {
            if ( ev instanceof PublisherStart ) 
            {
                code( "Got start event from publisher",
                    ( (PublisherStart) ev ).getPublisherId() );

                ++active;
                ++starts;
            }
            else if ( ev instanceof PublisherStop ) 
            {
                code( "Got stop event from publisher",
                    ( (PublisherStop) ev ).getPublisherId() );
                
                --active;
                exitConditional();
            }
        }
    }

    // Publisher process which publishes various data and control events for
    // awhile and then exits
    private
    final
    static
    class EventPublisher
    extends AbstractVoidProcess
    {
        // publisher id
        private final int id;

        // used to introduce some random jitter into the publish rate
        private final Random rand = new Random();

        private
        EventPublisher( int id )
        {
            super( EventBehavior.create( evMgr ) );

            this.id = id;
        }

        private
        final
        class StopTask
        extends AbstractTask
        {
            protected 
            void 
            runImpl()
            {
                // publish stop before calling exit().
                behavior( EventBehavior.class ).
                    publish( controlTopic, new PublisherStop( id ) );

                exit();
            }
        }

        // Publish an event, pause for a bit, and repeat for as long as this
        // publisher is alive
        private
        final
        class PublishTask
        extends AbstractTask
        {
            protected
            void
            runImpl()
            {
                behavior( EventBehavior.class ).
                    publish( dataTopic, new DataEvent( id ) );

                // create a random pause before publishing the next event
                Duration delay = Duration.fromMillis( rand.nextInt( 3000 ) );
                submit( this, delay );
            }
        }

        // begin the actual publishing
        private
        void
        startPublishing()
        {
            // broadcast start event
            behavior( EventBehavior.class ).
                publish( controlTopic, new PublisherStart( id ) );

            // set some random exit delay between [3s,4s)
            Duration exitDelay = 
                Duration.fromMillis( 3000 + rand.nextInt( 1000 ) );

            // exit eventually
            submit( new StopTask(), exitDelay );

            // begin publishing
            submit( new PublishTask() );
        }

        protected
        void
        startImpl()
        {
            // wait a little random amount to begin publishing, just to prevent
            // all publishers started in the main loop from starting totally in
            // sync with each other
            Duration startDelay = Duration.fromMillis( rand.nextInt( 1500 ) );

            submit(
                new AbstractTask() {
                    protected void runImpl() { startPublishing(); }
                },
                startDelay
            );
        }
    }

    // set the event receivers and start the publishers
    protected
    void
    startImpl()
    {
        behavior( EventBehavior.class ).
            subscribe( dataTopic, new DataReceiver() );
        
        behavior( EventBehavior.class ).
            subscribe( controlTopic, new ControlReceiver() );

        for ( int i = 0; i < publisherCount; ++i ) 
        {
            spawn( new EventPublisher( i ) );
        }
    }
}
