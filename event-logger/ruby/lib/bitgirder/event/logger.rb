require 'bitgirder/core'
include BitGirder::Core

require 'thread'
require 'set'

module BitGirder
module Event

# Base library for BitGirder event logging. Generators of an event will call
# Engine.log_event, which dispatches the event to a dynamic set of listeners.
#
# Typically an application will create a single Engine instance during
# application startup, register some number of listeners, and then make the
# engine available to the rest of the application for logging events via
# Engine#log_event.
#
# Listeners registered with an Engine (see Engine#add_listener) should process
# events immediately and without blocking, since they execute in the call stack
# of the event producer.  In particular, implementations which serialize events
# and perform IO (write to disk, send over a network) should be careful not to
# perform the IO directly upon receiving the event, but should delegate this to
# some ancillary thread.
#
# Events can be any ruby object. It is up to each listener to decide if/how to
# handle a given event. 
#
# Development roadmap for this library:
#
# - Add skeletal listener implementations to simplify offloading IO or other
#   intensive processing to a helper thread, as well as managing queuing
#   policies and overload behavior (drop new events, drop old events, grow
#   unbounded, etc).
#
module Logger

# An Engine coordinates dispatching of logged events a dynamic set of listeners.
# All methods are safe to call from concurrent threads.
class Engine < BitGirderClass
    
    bg_attr :error_handler,
            :required => false,
            :description => <<-END_DESC
                If set, this object will be called for various error methods to
                which it responds.
            END_DESC
 
    # Creates a new instance with an empty listener set.
    def impl_initialize # :nodoc:
        
        super

        @listeners = Set.new
        @mut = Mutex.new
    end

    # In case we later decide to allow @mut to be optional or of varying
    # implementations (not all apps which log will necessarily be threaded)
    # we'll only need to modify this block to run code which should be
    # threadsafe
    private
    def guarded( &blk )
        @mut.synchronize( &blk )
    end

    # Gets the number of listeners managed by this engine
    public
    def listener_count
        guarded { @listeners.size }
    end

    private
    def listener_failed( l, ev, ex )
        
        if @error_handler.respond_to?( :listener_failed )
            @error_handler.listener_failed( l, ev, ex )
        else
            warn( ex, "Listener #{l} failed for event #{ev}" )
        end
    end

    # Logs an event. After this method returns all listeners will have received
    # the event.
    public
    def log_event( ev )

        # Accumulate failures here so we can process them after releasing the
        # lock
        failures = []

        guarded do 
            @listeners.each do |l| 
                begin 
                    l.event_logged( ev )
                rescue Exception => ex
                    failures << [ l, ev, ex ]
                end
            end
        end

        failures.each { |arr| listener_failed( *arr ) }
    end

    # Adds the given listener to this engine if it is not already present. All
    # subsequent calls to log_event will dispatch to +l+ until a call is made to
    # remove_listener
    #
    # This method returns +l+ to allow callers to chain listener
    # creation/addition:
    #
    #   list = eng.add_listener( SomeListener.new( ... ) )
    #
    public
    def add_listener( l )

        guarded do
            if l.respond_to?( :event_logged ) 
                @listeners << l unless @listeners.include?( l )
            else
                raise "Not an event listener: #{l} (#{l.class})" 
            end
        end

        l
    end

    # Removes +l+ from the dispatch set. It is not an error to call this
    # method for an unregistered listener.
    public
    def remove_listener( l )
        guarded { @listeners.delete( l ) }
    end
end

end
end
end
