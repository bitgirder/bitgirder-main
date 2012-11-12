require 'bitgirder/testing'
require 'bitgirder/core'

require 'set'

module BitGirder
module Testing

    class TestRunnerTests

        TEST_DURATION = 1

        AFTERS_EXPCT = 
            Set.new( [ :async_method, :immediate_method, :async_class ] )

        AFTERS_ACTUAL = Set.new
 
        include TestClassMixin

        attr_reader :befores_actual

        def initialize

            super

            @befores_expct = 
                Set.new( [ :async_method, :immediate_method, :async_class ] )

            @befores_actual = Set.new

            @instance_var1 = :hello
        end

        def self.assert_after_ran
            raise "Not all afters ran" unless AFTERS_EXPCT == AFTERS_ACTUAL
        end

        define_before :before_stuff
        def before_stuff
            @befores_actual << :immediate_method
        end

        define_before :before_stuff_method_async
        def before_stuff_method_async( ctx )

            EM.add_timer( TEST_DURATION ) do 
                @befores_actual << :async_method
                ctx.succeed
            end
        end

        class AsyncBefore
            
            include BeforePhase
            
            public
            def start_invocation( ctx )

                EM.add_timer( TEST_DURATION ) do
                    ctx.inst.befores_actual << :async_class
                    ctx.complete_invocation
                end
            end
        end

        define_after :after_stuff
        def after_stuff
            AFTERS_ACTUAL << :immediate_method
        end

        define_after :after_async
        def after_async( ctx )

            AFTERS_ACTUAL << :async_method
            ctx.invocation_complete
        end

        class AfterAsync
            
            include AfterPhase

            public
            def start_invocation( ctx )
                
                EM.add_timer( TEST_DURATION ) do
                    AFTERS_ACTUAL << :async_class
                    ctx.invocation_complete
                end
            end
        end

        def test_setups_ran
            assert_equal( @befores_expct, @befores_actual )
        end

        define_test :is_a_test
        def is_a_test; end

        def not_a_test
            raise "Should not be run as a test"
        end

        ignore_test :test_ignore_of_tests
        def test_ignore_of_tests
            raise "Should not be run as a test"
        end

        # We also use this to get some coverage of various passing assertions
        def test_immediate_success

            assert( true )
            assert_equal( "yes", "yes" )
            assert_nil( nil )

            assert_raised( Exception ) { raise Exception.new( "raised" ) }
        end

        def test_immediate_fail
            assert( false, "test-immediate-fail" )
        end 

        def test_async_success( ctx )
            EM.add_timer( TEST_DURATION ) { ctx.succeed }
        end

        # This also tests that the async dispatch is being honored, since if it
        # is not then this test will complete with success and we'll catch that
        # in the assert reporter
        def test_async_failure( ctx )
            EM.add_timer( TEST_DURATION ) do
                ctx.fail_test( Exception.new( "async-failure" ) )
            end
        end

        def test_async_unchecked_failure( ctx )
            fail_test( "async-unchecked-failure" )
        end

        def test_context_complete_with_block_success( ctx )
            ctx.complete_invocation {}
        end

        def test_context_complete_with_block_fails( ctx )
            ctx.complete_invocation { fail_test( "complete-with-block-fails" ) }
        end

        def assert_not_em_thread
            assert_false( EM.reactor_thread? )
        end

        def test_async_blocking_test_explicit( ctx )
            ctx.blocking { assert_not_em_thread }
        end

        def test_async_blocking_test_explicit_fails( ctx )

            ctx.blocking do

                assert_not_em_thread
                raise Exception.new( "async-blocking-explicit-fails" )
            end
        end

        blocking :test_annotated_blocking
        def test_annotated_blocking
            assert_not_em_thread
        end

        blocking :test_annotated_blocking_fails
        def test_annotated_blocking_fails
            assert_not_em_thread
            raise Exception.new( "annotated-blocking-fails" )
        end

        private
        def check_inv_fact_follows_befores
            
            if @befores_actual.size < @befores_expct.size
                raise "inv fact running ahead of before tasks"
            end
        end

        invocation_factory :get_inv_fact_tests
        def get_inv_fact_tests
            
            check_inv_fact_follows_befores

            {
                # In this call we check that the block is correctly invoked in
                # the scope of the test instance without loss of generality, we
                # don't currently check this in the rest of the blocks.
                :"test_inv_fact/succ_immediate" => lambda {
                    assert( @instance_var1 == :hello )
                },

                :"test_inv_fact/succ_async" => lambda { |ctx|
                    EM.add_timer( TEST_DURATION ) { ctx.complete }
                },

                :"test_inv_fact/block_uncaught_exception" => lambda {
                    raise Exception.new( "test-exception" )
                },

                :"test_inv_fact/async_fail" => lambda { |ctx|
                    ctx.fail_test( Exception.new( "test-exception" ) )
                }
            }
        end

        class TestClassSuccessImmediateTest
 
            include TestPhase

            public
            def start_invocation( ctx )
                ctx.succeed
            end
        end

        class TestClassFailImmediateTest
 
            include TestPhase

            public
            def start_invocation( ctx )
                raise Exception.new( "test-class-fail-immediate" )
            end
        end

        class TestClassFailAsyncTest < TestRunClass
            
            private
            def start_impl
                EM.add_timer( TEST_DURATION ) do
                    @ctx.fail_invocation( 
                        Exception.new( "test-class-fail-async" ) )
                end
            end
        end

        class TestClassSucceedAsyncTest < TestRunClass
            
            # Also test som basic methods in AbstractPhaseClass
            private
            def start_impl

                inst_set( :test1, "stuff" )
                assert_equal( "stuff", inst_get( :test1 ) )

                EM.add_timer( TEST_DURATION ) { @ctx.complete_invocation }
            end
        end
    end

    class TestClassWithBeforeFailure
        
        include TestClassMixin

        define_before :before_that_fails
        def before_that_fails
            raise "failure-from-before"
        end

        def test_which_should_not_run
            raise "test should not have run"
        end

        define_after :after_that_should_still_run
        def after_that_should_still_run
            @@after_ran = true
        end

        def self.assert_after_ran
            raise "After did not run" unless @@after_ran
        end
    end

    class TestWithOnlyClassInvocations
        
        include TestClassMixin

        class ThisRunsTest

            include TestPhase
            
            public
            def start_invocation( ctx )
                ctx.complete_invocation
            end
        end
    end

    class TestClassBase < BitGirder::Core::BitGirderClass
        
        include TestClassMixin
        is_test_base

        def test_meth1; end

        bg_abstract :impl_get_fact_tests

        invocation_factory :get_fact_tests
        def get_fact_tests
            { :fact_test1 => lambda {} }.merge( impl_get_fact_tests )
        end
    end

    class TestSubClassA < TestClassBase
        
        include TestClassMixin

        def test_sub_meth1; end

        def impl_get_fact_tests
            { :sub_fact_test1 => lambda {} }
        end
    end

    class TestSubClassB < TestClassBase

        include TestClassMixin

        def test_sub_meth2; end

        def impl_get_fact_tests
            { :sub_fact_test2 => lambda {} }
        end
    end

    # A reporter which asserts the test run results we expect from the above
    class TestRunnerAssertReporter < BitGirderClass
        
        private
        def make_name( cls, meth )
            "#{cls}.#{meth}"
        end

        private
        def make_class_test_name( cls )
            
            # Duplicating logic used in unit-test-runner on purpose so that we
            # can catch unintentional changes there by failing on a mismatch
            # here
            if md = /(.+)::([^:]+)Test$/.match( cls.to_s )

                lc_underscore = md[ 2 ].gsub( /[A-Z]/ ) { |m| "_#{m.downcase}" }
                "#{md[ 1 ]}.test#{lc_underscore}"
            else
                raise "Couldn't parse name: #{cls}"
            end
        end
 
        private
        def create_fail_expector( excpt_cls, msg )
            
            lambda do |nm, res|
                
                if res
                    if res.is_a?( excpt_cls )
                        unless ( msg_actual = res.message ) == msg
                            raise "Message #{msg_actual} != #{msg} for #{nm}"
                        end
                    else
                        raise "Got exception #{res} but expected instance of " +
                              excpt_cls.to_s + " for #{nm}"
                    end
                else
                    raise "Expected a failure for #{nm}"
                end
            end
        end

        private
        def get_verifiers
            
            cls = TestRunnerTests

            expct_success = lambda do |nm, res| 
                if res
                    raise "Expected success for #{nm} but got :failure => " \
                          "#{res}\n#{res.backtrace.join( "\n" )}" 
                end
            end

            res = {}

            [ :is_a_test, 
              :test_immediate_success, 
              :test_async_success,
              :test_context_complete_with_block_success,
              :test_async_blocking_test_explicit,
              :test_annotated_blocking,
              :test_setups_ran,
              :"test_inv_fact/succ_immediate",
              :"test_inv_fact/succ_async" ].each do |meth|
                res[ make_name( cls, meth ) ] = expct_success
            end
 
            res.merge!(

                make_name( cls, :test_immediate_fail ) =>
                    create_fail_expector( 
                        AssertionFailure, "test-immediate-fail" ),
            
                make_name( cls, :test_async_failure ) =>
                    create_fail_expector( Exception, "async-failure" ),

                make_name( cls, :test_async_unchecked_failure ) =>
                    create_fail_expector( 
                        AssertionFailure, "async-unchecked-failure" ),

                make_name( cls, :test_context_complete_with_block_fails ) =>
                    create_fail_expector( 
                        AssertionFailure, "complete-with-block-fails" ),
                
                make_name( cls, :test_async_blocking_test_explicit_fails ) =>
                    create_fail_expector(
                        Exception, "async-blocking-explicit-fails" ),

                make_name( cls, :test_annotated_blocking_fails ) =>
                    create_fail_expector(
                        Exception, "annotated-blocking-fails" ),
                
                make_name( cls, :"test_inv_fact/block_uncaught_exception" ) =>
                    create_fail_expector( Exception, "test-exception" ),

                make_name( cls, :"test_inv_fact/async_fail" ) =>
                    create_fail_expector( Exception, "test-exception" )
            )

            [ cls::TestClassSuccessImmediateTest ].each do |t_cls|
                res[ make_class_test_name( t_cls ) ] = expct_success
            end

            res.merge!(
                
                make_class_test_name( cls::TestClassFailImmediateTest ) =>
                    create_fail_expector(
                        Exception, "test-class-fail-immediate" ),
                
                make_class_test_name( cls::TestClassFailAsyncTest ) =>
                    create_fail_expector( Exception, "test-class-fail-async" )
            )

            [ cls::TestClassSucceedAsyncTest,
              TestWithOnlyClassInvocations::ThisRunsTest ].each do |c| 
                res[ make_class_test_name( c ) ] = expct_success
            end

            [ :test_meth1, 
              :fact_test1, 
              :sub_fact_test1, 
              :test_sub_meth1 ].each do |m|
                res[ make_name( cls::TestSubClassA, m ) ] = expct_success
            end

            [ :test_meth1, 
              :fact_test1, 
              :sub_fact_test2,
              :test_sub_meth2 ].each do |m|
                res[ make_name( cls::TestSubClassB, m ) ] = expct_success
            end

            res
        end

        private
        def get_test_results( cycles )
 
            cycles.values.
                inject( {} ) do |h, cyc| 
                    h.merge!( cyc.phase_results( :tests ) )
                    h
                end
        end

        public
        def report( cycles )
            
            verifiers = get_verifiers

            get_test_results( cycles ).each_pair do |inv, res|
                
                if verify = verifiers.delete( nm = inv.to_s )
                    verify.call( nm, res )
                else
                    raise "Unexpected test invocation: #{inv}"
                end
            end 

            unless verifiers.empty?
                raise "Some invocations did not run: #{verifiers.keys}"
            end

            TestRunnerTests.assert_after_ran
            TestClassWithBeforeFailure.assert_after_ran

            code( "Test runner assertion passed" )
        end
    end

end
end
