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

    private
    def assert_retry_runtime( retr, completion_time = Time.now )
        
        # Check that the actual elapsed is within 20% of the idealized
        # delay
        elapsed = completion_time - retr.start_time
        ideal = retr.seed_secs * ( ( 2 ** ( retr.attempt - 1 ) ) - 1 )
        code( "elapsed: #{elapsed}, ideal: #{ideal}" )
        assert( elapsed >= ideal ) # A basic check -- we should not be faster
        assert( ( elapsed - ideal ).abs / ideal < 0.2 )
    end

    public
    def test_retry_success( ctx )
        
        Retry.run do |b|

            b.retries = 3
            b.retry_on MarkerError
            b.seed_secs = 4 

            b.action do |r| 
                if r.attempt < 3 then raise MarkerError else "result-ok" end
            end

            b.complete do |retr, res|

                assert_equal( "result-ok", res )
                assert_equal( 3, retr.attempt )
                assert_retry_runtime( retr )

                ctx.complete_invocation
            end

            b.failed { |err| ctx.fail_invocation( err ) }
        end
    end

    public
    def test_retry_fails_on_retries_exceeded( ctx )

        Retry.run do |b|
            
            remain = 3

            b.retries = 3
            b.seed_secs = 4

            b.action { remain -= 1; raise MarkerError }

            # Also get coverage of retry_on with a list of exception classes
            b.retry_on MarkerError, AnotherMarkerError 

            b.complete { fail_test( "complete block was called" ) }

            b.failed do |retr, err| 
                
                ctx.succeed do
                    assert_equal( 0, remain )
                    assert_equal( 3, retr.attempt )
                    assert_retry_runtime( retr )
                    assert( err.is_a?( MarkerError ) )
                end
            end
        end
    end
    
    # Test coverage for complete block with single arg (the result)
    def test_retry_complete_result_only( ctx )
        
        Retry.run do |b|
            
            b.action { "ok" }
            b.complete { |res| ctx.succeed { assert_equal( "ok", res ) } }
            b.failed { |err| ctx.fail_invocation( err ) }
        end
    end

    def test_retry_complete_no_args( ctx )
        
        Retry.run do |b|
            
            b.action {}
            b.complete { ctx.succeed }
            b.failed { |err| ctx.fail_invocation( err ) }
        end
    end

    def test_retry_using_custom_block( ctx )
        
        Retry.run do |b|
            
            b.action do |r| 
                if r.attempt <= 2 then raise MarkerError; else "ok" end
            end

            b.complete do |retr, res| 
                ctx.succeed do
                    assert_equal( "ok", res )
                    assert_equal( 3, retr.attempt )
                end
            end


            b.failed { |err| ctx.fail_invocation( err ) }
            b.retry_on { |err| err.is_a?( MarkerError ) }
        end
    end

    def test_retry_async_action_success( ctx )
        
        Retry.run do |b|

            b.retries = 3
            b.seed_secs = 0.5
            b.retry_on MarkerError
            
            b.async_action do |r|
                r.complete do
                    if r.attempt <= 2 then raise MarkerError; else "ok" end
                end
            end

            b.complete do |retr, res|
                ctx.succeed do
                    assert_equal( 3, retr.attempt )
                    assert_equal( "ok", res )
                end
            end

            b.failed { |err| ctx.fail_invocation( err ) }
        end
    end

    def test_retry_async_action_fail_exceed_retries( ctx )
        
        Retry.run do |b|
            
            remain = 3
            b.retries = 3
            b.retry_on MarkerError
            
            b.async_action do |r| 
                remain -= 1
                r.fail_attempt( MarkerError.new )
            end

            b.complete { ctx.complete { fail_test( "got normal completion" ) } }

            b.failed do |retr, err|
                ctx.complete do
                    assert_equal( 0, remain )
                    assert( err.is_a?( MarkerError ) )
                end
            end
        end
    end

    def test_retry_async_action_fail_non_retriable( ctx )
        
        Retry.run do |b|
            
            # Set retry_on to something that will not be retried at all
            b.retry_on MarkerError 

            b.async_action { |r| r.complete { raise Exception } }
            b.complete { ctx.complete { fail_test( "got normal completion" ) } }

            b.failed do |retr, err| 
                ctx.complete do
                    assert_equal( 1, retr.attempt )
                    assert( err.is_a?( Exception ) )
                end
            end
        end
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
            run.complete { |r| ctx.succeed { assert_equal( 0, count ) } }
        end
    end

    def assert_em_defer( ret_val, do_fail, &blk )
        
        EmUtils.defer( blk ) do
        
            assert_false( EM.reactor_thread? )
    
            if do_fail
                raise MarkerError
            else
                ret_val
            end
        end
    end

    def test_em_defer_success( ctx )
        
        assert_em_defer( 3, false ) do |comp|
            ctx.succeed { assert_equal( 3, comp.get ) }
        end
    end

    def test_em_defer_failure( ctx )
 
        assert_em_defer( nil, true ) do |comp|
            ctx.succeed { assert_raised( MarkerError ) { comp.get } }
        end
    end
end

end
end
