require 'bitgirder/event/logger'
require 'bitgirder/event/testing'

require 'bitgirder/core'
include BitGirder::Core

require 'bitgirder/testing'

require 'thread'

module BitGirder
module Event
module Logger

# Provides helpers for testing event flow in an application. In addition to
# testing the basic operation of an event listener (does it serialize correctly,
# filter appropriately, etc), thoroughly tested applications will want to add
# coverge of specific events.
#
# For example, suppose an application expects to log an event upon every login
# attempt, including the user id, login completion time, and result (_success_,
# <i>unknown user</i>, <i>bad password</i>, etc). This data might be used in
# realtime to look for ongoing attacks on an account, or over time to track user
# engagement. In any event, the events are an important part of the application,
# but an inadvertent change in the login handling code might cause logins to not
# be logged. This module helps make it easy to include assertions about event
# delivery right alongside other assertions (that a login succeeded or failed as
# expected).
#
module Testing

class CodecRoundtripper < BitGirderClass
 
    bg_attr :codec
    bg_attr :listener, :required => false

    public
    def event_logged( ev )

        ev = BitGirder::Event::Testing.roundtrip( ev, @codec )
        @listener.event_logged( ev ) if @listener
    end
end

# A simple event listener (see BitGirder::EventLogger) which accumulates all
# events into an unbounded list. Test classses should make sure to remove this
# listener from its associated engine after assertions are complete, either
# explicitly or via ensure_removed, lest it continue to amass large amounts of
# unneeded events.
class EventAccumulator < BitGirderClass
    
    include BitGirder::Testing::AssertMethods

    bg_attr :engine

    # Creates a new instance associated with the given engine
    def impl_initialize
        
        @mut = Mutex.new
        @events = []
    end

    # Adds +ev+ to this instance's event list.
    public
    def event_logged( ev )
        @mut.synchronize { @events << ev }
    end

    # Gets a snapshot of the events accumulated so far.
    public
    def events
        @mut.synchronize { Array.new( @events ) }
    end

    public
    def assert_logged( expct = nil, &blk )
        
        if expct
            if blk
                raise "Illegal combination of expect val and block"
            else
                blk = lambda { |ev| ev == expct }
            end
        else
            raise "Block missing" unless blk
        end

        assert events.find( &blk ), "Block did not match any events"
    end

    # Executes an arbitrary block, ensuring that this instance is removed from
    # the engine with which it is associated whether or not the block completes
    # normally. This method returns the block's result or raises its exception.
    #
    # This instance is itself provided to the block, enabling (in conjunction
    # with EventAccumulator.create()) test code to easily and reliably wrap an
    # entire test:
    #
    #   def test_login_success
    #
    #       EventAccumulator.create( ev_eng ).ensure_removed do |acc|
    #
    #           # Do a login and first check that it succeeds as wanted
    #           login_res = do_login( "ezra", "somepass" )
    #           assert( login_res.ok? )
    #
    #           # Now also check that login event was generated
    #           ev_expct = { :event => :login, :user => "ezra", :result => :ok }
    #           acc.assert_logged { |ev| ev == ev_expct }
    #
    #           ... # Possibly more test code, cleanup, etc
    #       end
    #   end
    #
    public
    def ensure_removed
        begin
            yield( self )
        ensure
            @engine.remove_listener( self )
        end
    end

    # Convenience method to create, register, and return an instance which
    # accumulates events from the given EventLogger::Engine
    def self.create( *argv )
        self.new( *argv ).tap { |acc| acc.engine.add_listener( acc ) }
    end

    def self.while_accumulating( *argv )

        acc = self.create( *argv )
        acc.ensure_removed { yield( acc ) }
    end
end

end
end
end
end
