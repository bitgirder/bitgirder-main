require 'bitgirder/core'
require 'bitgirder/concurrent'
require 'bitgirder/testing'

module BitGirder
module Concurrent

class ConcurrentTests
    
    include BitGirder::Testing::TestClassMixin

    class MarkerError < StandardError; end
    class AnotherMarkerError < StandardError; end

    OKS = [ :ok, :ok?, :is_ok, :is_ok ]

    def test_completion_success_impl
        
        c = Completion.create_success( 1 )

        [ :get, :result, :get_result ].each do |m| 
            assert_equal( 1, c.send( m ) )
        end

        OKS.each { |m| assert( c.send( m ) ) }

        assert_equal(
            "Attempt to call get_exception when ok? returns true",
            assert_raised( Exception ) { c.exception }.message
        )
    end

    def test_completion_failure_impl
        
        ex = MarkerError.new( "HI" )
        c = Completion.create_failure( ex )

        [ :exception, :get_exception ].each do |m|
            assert_equal( ex, c.send( m ) )
        end

        OKS.each { |m| assert_false( c.send( m ) ) }

        assert_equal( ex, assert_raised( ex.class ) { c.get } )

        assert_equal(
            "Attempt to call get_result when ok? returns false",
            assert_raised( Exception ) { c.get_result }.message
        )
    end
    
    # This covers various properties during the lifetime of a rendezvous; other
    # tests cover more specific subcases and exception states
    def test_rendezvous_normal
        
        rendezvous_ran = false

        r = Rendezvous.new { rendezvous_ran = true }
    
        2.times { |i| r.fire }

        assert( r.open? && ! r.closed? )
        assert_equal( 2, r.remain )

        r.arrive
        assert( r.open? && ! r.closed? )
        assert_equal( 1, r.remain )

        r.close
        assert( r.closed? && ! r.open? )
        assert_equal( 1, r.remain )

        r.arrive
        assert( r.closed? && ! r.open? )
        assert_equal( 0, r.remain )
        assert( rendezvous_ran )
    end

    # The test above handles the case in which the join occurs on the last
    # arrival. Here we also cover the case in which the join should occur with
    # the call to close
    def test_rendezvous_all_arrive_before_close
        
        rendezvous_ran = false

        r = Rendezvous.new { rendezvous_ran = true }

        r.fire
        r.arrive
        r.close

        assert( rendezvous_ran )
    end

    def test_rendezvous_underflow
 
        r = Rendezvous.new

        r.fire
        r.arrive
        assert_raised( Rendezvous::UnderflowError ) { r.arrive }
    end

    def test_rendezvous_closed_exception
 
        # Test enter after close
        r = Rendezvous.new
        r.close
        assert_raised( Rendezvous::ClosedError ) { r.fire }

        # Test close after close
        r = Rendezvous.new
        r.close
        assert_raised( Rendezvous::ClosedError ) { r.close }
    end

    def test_rendezvous_run( ctx )
        
        count = 3

        Rendezvous.run do |run|
            
            count.times { |i| run.fire { |r| count -= 1; r.arrive } }
            run.complete { |r| ctx.complete { assert_equal( 0, count ) } }
        end
    end
end

end
end
