require 'bitgirder/core'
include BitGirder::Core

require 'bitgirder/testing'
include BitGirder::Testing

require 'bitgirder/event/testing'
include BitGirder::Event::Testing

require 'bitgirder/event/logger'
require 'bitgirder/event/logger/testing'

module BitGirder
module Event
module Logger

class MarkerError < StandardError; end

class EventAccumulatorTests < BitGirderClass
    
    include TestClassMixin
    include Logger::Testing

    def test_add_and_remove
        
        eng = Engine.new

        acc = EventAccumulator.create( eng )
        eng.log_event( 1 )
        assert_equal( [ 1 ], acc.events )
        eng.remove_listener( acc )
        eng.log_event( 2 )
        assert_equal( [ 1 ], acc.events )
    end

    def test_ensure_removed_success
        
        eng = Engine.new

        acc = EventAccumulator.create( eng )
        acc.ensure_removed {}

        eng.log_event( 1 )
        assert_equal( [], acc.events )
    end

    def test_ensure_removed_on_fail
 
        eng = Engine.new

        acc = EventAccumulator.create( eng )

        assert_raised( MarkerError ) do
            acc.ensure_removed { raise MarkerError }
        end

        eng.log_event( 1 )
        assert_equal( [], acc.events )
    end

    def test_assert_logged
        
        EventAccumulator.while_accumulating( eng = Engine.new ) do |acc|
            
            check_fail = lambda { |i|

                msg, ex = 'Block did not match any events', AssertionFailure

                assert_raised( msg, ex ) { acc.assert_logged( i ) }
                assert_raised( msg, ex ) { acc.assert_logged { |ev| ev == i } }
            }

            check_fail.call( 1 )

            eng.log_event( 1 )
            acc.assert_logged { |ev| ev == 1 }
            acc.assert_logged( 1 )

            check_fail.call( 2 )
        end
    end

    class CountingCodec < BitGirderClass
        
        bg_attr :codec, :default => TestCodec.new
        bg_attr :encodes, :default => 0
        bg_attr :decodes, :default => 0

        public
        def encode_event( *argv )
            @codec.encode_event( *argv ).tap { @encodes += 1 }
        end

        public
        def decode_event( *argv )
            @codec.decode_event( *argv ).tap { @decodes += 1 }
        end

        public
        def call_pairs
            
            if @decodes == @encodes
                @decodes
            else
                raise "@decodes != @encodes (#@decodes != #@encodes)"
            end
        end
    end

    def assert_roundtripper_counts( opts )
        
        rt = CodecRoundtripper.new( 
            :codec => CountingCodec.new,
            :listener => opts[ :listener ] # could be nil
        )

        eng = has_key( opts, :engine )
        
        eng.add_listener( rt )
        assert_equal( 0, rt.codec.call_pairs )
        eng.log_event( Int32Event.new( 0 ) )
        assert_equal( 1, rt.codec.call_pairs )
    end

    def test_codec_roundtripper_no_backing_listener
        assert_roundtripper_counts( :engine => Engine.new )
    end

    def test_codec_roundtripper_with_backing_listener
 
        acc = EventAccumulator.new( :engine => Engine.new )
        assert_roundtripper_counts( :engine => acc.engine, :listener => acc )

        assert_equal( [ Int32Event.new( 0 ) ], acc.events )
    end
end

class FailListener < BitGirderClass
 
    include AssertMethods

    bg_attr :fail_on, :required => false

    private
    def fail_message
        "test-message"
    end

    public
    def assert_error( ex )
        
        assert_equal( MarkerError, ex.class )
        assert_equal( fail_message, ex.message )
    end

    private
    def do_fail
        raise MarkerError, fail_message
    end

    public
    def event_logged( ev )
        
        do_fail if @fail_on.is_a?( Proc ) && @fail_on.call( ev )
        do_fail if @fail_on == nil || @fail_on == ev
    end
end

class LoggerTests < BitGirderClass
 
    include TestClassMixin
    include Logger::Testing

    def test_engine_basic_ops

        eng = Engine.new
        assert_equal( 0, eng.listener_count )

        EventAccumulator.while_accumulating( eng ) do |acc|
       
            assert_equal( 1, eng.listener_count )
            eng.log_event( 1 )
            assert_equal( [ 1 ], acc.events )
        end
    end

    def run_engine_error_handling_test( eng )
 
        fl = eng.add_listener( FailListener.new( 2 ) )
        acc = EventAccumulator.create( eng )

        3.times { |i| eng.log_event( i ) }

        assert_equal( [ 0, 1, 2 ], acc.events )
    end

    def test_engine_default_error_handling
        run_engine_error_handling_test( Engine.new )
    end

    def test_engine_custom_error_handling
        
        class <<( err_handler = {} )
            def listener_failed( l, ev, ex )
                self.merge!( :listener => l, :event => ev, :error => ex )
            end
        end

        eng = Engine.new( :error_handler => err_handler )
        run_engine_error_handling_test( eng )

        l = err_handler[ :listener ]
        assert( l.is_a?( FailListener ) )
        assert_equal( 2, err_handler[ :event ] )
        l.assert_error( err_handler[ :error ] )
    end

    def test_engine_reraising_error_handler
 
        class <<( err_h = Object.new )
            def listener_failed( l, ev, ex ); raise ex; end
        end

        eng = Engine.new( :error_handler => err_h )
        eng.add_listener( FailListener.new )

        assert_raised( "test-message", MarkerError ) { eng.log_event( "" ) }
    end
end

end
end
end
