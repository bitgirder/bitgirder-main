require 'bitgirder/core'
require 'bitgirder/testing'

module BitGirder
module Testing

class TestCheckMarkerError < StandardError
end

class RunCheck < BitGirder::Core::BitGirderClass

    def impl_initialize

        @calls = []
        @unexpected_calls = []
    end

    def self.create
        self.new.tap { |res| yield( res ) if block_given? }
    end

    public
    def <<( opts )
        @calls << opts.dup
    end

    private
    def call_for_name( sym )
        @calls.find { |call| call[ :name ] == sym }
    end

    public
    def called( sym )
        
        if c = call_for_name( sym )
            c[ :called ] = true
        else
            @unexpected_calls << sym
        end
    end

    private
    def validate_expected_calls

        not_called = @calls.select do |c| 
            ! ( c[ :called ] || c[ :expect_cancel ] )
        end

        unless not_called.empty?
            raise "not called: #{not_called.map { |c| c[ :name ] } }"
        end
    end

    private
    def validate_unexpected_calls
        
        unless @unexpected_calls.empty?
            raise "unexpected calls: #@unexpected_calls" 
        end
    end

    private
    def assert_results_equal( results, calls )
        
        rk, ck = results.keys, calls.keys

        raise "unexpected result(s): #{rk - ck}" unless ck.include?( rk )
        raise "missing result(s): #{ck - rk}" unless rk.include?( ck )
        
        results.each_pair do |nm, res|
            c = calls[ nm ]
            unless c[ :called ]
                raise "got a result for #{nm}, but no corresponding call"
            end
        end
    end

    private
    def assert_invocation_error( inv, expct )
        
        act = inv.error

        return if expct == act # includes case of both nil

        raise "Expected error in #{inv}: #{expct}" if expct && ( ! act )
        raise "Unexpected error in #{inv}: #{act}" if act && ( ! expct )

        # both now known to be not nil
        case expct
        when Class
            unless act.is_a?( expct )
                raise "Expected error in #{inv} of type #{expct}, got: #{act}"
            end
        else
            unless expct == act
                raise "In #{inv}, expected error (#{expct}) != actual (#{act})"
            end
        end
    end

    private
    def assert_engine_result( inv, c )
        
        assert_invocation_error( inv, c[ :expect_error ] )

        if blk = c[ :check ] 
            inv.object.instance_eval( &blk )
        end

        if blk = c[ :check_inv ]
            blk.call( inv )
        end

        if min_time = c[ :check_min_runtime ]
            if inv.end_time - inv.start_time < min_time
                raise "#{inv} didn't run with delay >= #{min_time}"
            end
        end
    end

    private
    def assert_engine_results( calls_h, results_h )
 
        results_h.each_pair do |k, inv|

            if c = calls_h[ k ]
                if ( c[ :called ] || c[ :expect_cancel ] )
                    assert_engine_result( inv, c )
                    calls_h.delete( k )
                end
            else
                raise "unexpected result: #{k}"
            end
        end

        return if calls_h.empty?
        raise "no result for calls: #{calls_h.keys.sort}"
    end

    private
    def validate_engine_results( eng, cls )

        calls_h = @calls.inject( {} ) do |h, c|
            h.tap { h[ [ c[ :phase ], "#{cls}/#{c[ :name ]}" ] ] = c }
        end

        results = eng.results.select { |res| res.context.test_class == cls }

        results_h = {}

        results.each do |inv_set|
            inv_set.invocations.each_pair do |phase, invs|
                invs.each { |inv| results_h[ [ phase, inv.name ] ] = inv }
            end
        end

        assert_engine_results( calls_h, results_h )
    end

    public
    def validate_run( eng, cls )
        
        validate_expected_calls
        validate_unexpected_calls
        validate_engine_results( eng, cls )
    end
end

class CallTestObject < BitGirderClass

    bg_attr :run_check
    
    bg_attr :name

    bg_attr :do_fail,
            :required => false

    private
    def set_called
        @run_check.called( @name )
    end

    private
    def impl_fail
        raise TestCheckMarkerError
    end
end

class DirectCallTestObject < CallTestObject
    
    public
    def start_test
        
        set_called
        impl_fail if @do_fail
    end
end

class ContextCallTestObject < CallTestObject

    bg_attr :delay,
            :required => false
    
    attr_accessor :test_context

    public
    def start_test
        
        set_called

        Thread.start do
            sleep @delay if @delay
            @test_context.complete { impl_fail if @do_fail }
        end
    end
end

class TestClass1 < BitGirder::Core::BitGirderClass
    
    DELAY = 2

    RUN_CHECK = RunCheck.create do |rc|

        rc << { :name => :before_meth1, :phase => :before }
        
        rc << { :name => :test_meth_direct_ok, :phase => :test }

        rc << { 
            :name => :test_meth_direct_fail, 
            :phase => :test,
            :expect_error => TestCheckMarkerError
        }

        rc << { 
            :name => :test_meth_ctx_complete_direct_ok, 
            :phase => :test
        }

        rc << {
            :name => :test_meth_ctx_complete_direct_block_ok,
            :phase => :test,
            :check => proc { raise "no @val1" unless @val1 == 1 }
        }

        rc << {
            :name => :test_meth_ctx_direct_fail,
            :phase => :test,
            :expect_error => TestCheckMarkerError
        }

        rc << {
            :name => :test_meth_ctx_complete_direct_fail,
            :phase => :test,
            :expect_error => TestCheckMarkerError
        }

        rc << {
            :name => :test_meth_ctx_fail_test_direct,
            :phase => :test,
            :expect_error => TestCheckMarkerError
        }

        rc << {
            :name => :test_meth_delay,
            :phase => :test,
            :check_min_runtime => DELAY
        }

        # :check is to ensure that the block runs in the scope of inv.object
        rc << { 
            :name => :inv_fact_block_direct_ok, 
            :phase => :test,
            :check => proc { raise "no @val2" unless @val2 == 1 }
        }

        rc << { 
            :name => :inv_fact_block_direct_fail, 
            :phase => :test,
            :expect_error => TestCheckMarkerError
        }

        # Not re-checking all ctx behaviors, just that the call is correctly
        # invoked with a valid ctx object and that async completions work as
        # expected
        rc << { 
            :name => :inv_fact_block_ctx_ok, 
            :phase => :test,
            :check_min_runtime => DELAY
        }

        rc << { :name => :inv_fact_test_obj_direct_ok, :phase => :test }

        rc << { 
            :name => :inv_fact_test_obj_direct_fail, 
            :phase => :test,
            :expect_error => TestCheckMarkerError
        }

        rc << { :name => :inv_fact_test_obj_ctx_ok, :phase => :test }

        rc << { :name => :after_meth1, :phase => :after }
    end
 
    include TestClassMixin

    def before_meth1
        RUN_CHECK.called( __method__ )
    end

    def test_meth_direct_ok
        RUN_CHECK.called( __method__ )
    end

    def test_meth_ctx_complete_direct_ok( ctx )
        RUN_CHECK.called( __method__ )
        ctx.complete
    end

    def test_meth_ctx_complete_direct_block_ok( ctx )
        RUN_CHECK.called( __method__ )
        ctx.complete { @val1 = 1 }
    end

    def test_meth_ctx_direct_fail( ctx )
        RUN_CHECK.called( __method__ )
        raise TestCheckMarkerError
    end

    def test_meth_ctx_complete_direct_fail( ctx )
        RUN_CHECK.called( __method__ )
        ctx.complete { raise TestCheckMarkerError }
    end

    def test_meth_ctx_fail_test_direct( ctx )
        RUN_CHECK.called( __method__ )
        ctx.fail_test( TestCheckMarkerError.new )
    end

    def test_meth_delay( ctx )

        RUN_CHECK.called( __method__ )

        Thread.start do
            sleep DELAY
            ctx.complete
        end
    end

    def test_meth_direct_fail
        RUN_CHECK.called( __method__ )
        raise TestCheckMarkerError
    end

    def test_match_filter_1
        RUN_CHECK.called( __method__ )
    end

    def invocation_factory_1
        
        {
            :inv_fact_match_filter_1 => proc {
                RUN_CHECK.called( :inv_fact_match_filter_1 )
            },

            :inv_fact_block_direct_ok => proc { 
                RUN_CHECK.called( :inv_fact_block_direct_ok )
                @val2 = 1 
            },

            :inv_fact_block_direct_fail => proc {
                RUN_CHECK.called( :inv_fact_block_direct_fail )
                raise TestCheckMarkerError
            },

            :inv_fact_block_ctx_ok => proc { |ctx|

                RUN_CHECK.called( :inv_fact_block_ctx_ok )

                Thread.start do
                    sleep DELAY
                    ctx.complete
                end
            },

            :inv_fact_test_obj_direct_ok => DirectCallTestObject.new( 
                :run_check => RUN_CHECK,
                :name => :inv_fact_test_obj_direct_ok
            ),

            :inv_fact_test_obj_direct_fail => DirectCallTestObject.new(
                :run_check => RUN_CHECK,
                :name => :inv_fact_test_obj_direct_fail,
                :do_fail => true
            ),

            :inv_fact_test_obj_ctx_ok => ContextCallTestObject.new(
                :run_check => RUN_CHECK,
                :name => :inv_fact_test_obj_ctx_ok,
                :delay => DELAY
            )
        }
    end

    def after_meth1
        RUN_CHECK.called( __method__ )
    end
end

# Checks that befores/afters don't run if no nest matches filter
class TestClass2 < BitGirderClass
    
    include TestClassMixin

    RUN_CHECK = RunCheck.create

    def before_1; 
        RUN_CHECK.called( __method__ )
    end

    def test_match_filter_1; 
        RUN_CHECK.called( __method__ )
    end

    def invocation_factory_1

        { 
            :inv_fact_match_filter_1 => lambda { 
                RUN_CHECK.called( :inv_fact_match_filter_1 ) 
            }
        }
    end

    def after_1; 
        RUN_CHECK.called( __method__ )
    end
end

# Checks that tests and afters don't run if a before fails
class TestClass3 < BitGirderClass

    include TestClassMixin

    RUN_CHECK = RunCheck.create do |rc|
        
        rc << {
            :name => :before_1,
            :expect_error => TestCheckMarkerError,
            :phase => :before
        }

        rc << {
            :name => :test_1,
            :phase => :test,
            :expect_cancel => true,
            :expect_error => TestInvocationCancellationError
        }

        rc << {
            :name => :after_1,
            :phase => :after,
            :expect_cancel => true,
            :expect_error => TestInvocationCancellationError
        }
    end

    def before_1
        RUN_CHECK.called( __method__ )
        raise TestCheckMarkerError
    end

    def test_1
        RUN_CHECK.called( __method__ )
    end

    def after_1
        RUN_CHECK.called( __method__ )
    end
end

end
end
