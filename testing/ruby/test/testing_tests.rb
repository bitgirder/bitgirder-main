require 'bitgirder/core'
require 'bitgirder/testing'

module BitGirder
module Testing

class TestingTests
    
    include TestClassMixin

    private
    def assert_func_fails( func, msg )
        
        err = nil
        
        begin
            func.call
        rescue Exception => err; end

        case err

            when nil then fail_test( "func did not fail" )

            when BitGirder::Testing::AssertionFailure
                assert_equal( msg, err.message )

            else raise err
        end
    end

    class MarkerError < StandardError; end

    class CustomInt < BitGirder::Core::BitGirderClass
        
        bg_attr :int

        public
        def to_i
            @int
        end
    end

    class CustomString < BitGirder::Core::BitGirderClass
        
        bg_attr :str

        public
        def to_s
            @str
        end
    end

    public
    def test_assert_methods

        assert( true )
        assert_false( false )
        assert_equal( "hello", "hello" )
        assert_nil( nil )
        assert_raised( MarkerError ) { raise MarkerError.new }
        assert_equal_meth( 12, CustomInt.new( 12 ), :to_i )
        assert_equal_i( 12, CustomInt.new( 12 ) )
        assert_equal_s( "hello", CustomString.new( "hello" ) )
        assert_match( /^a*$/, "aaa" )

        [
            lambda { fail_test( "TEST" ) },
            lambda { fail_test( lambda { "TEST" } ) },
            lambda { assert( false, "TEST" ) },
            lambda { assert_equal( "hello", "goodbye", "TEST" ) },
            lambda { assert_nil( Object.new, "TEST" ) },
            lambda { assert_equal_i( 12, 13, "TEST" ) },
            lambda { assert_equal_i( 12, CustomInt.new( 13 ), "TEST" ) },
            lambda { assert_equal_s( "a", CustomString.new( "b" ), "TEST" ) },
            lambda { assert_match( /^a*$/, "b", "TEST" ) },
        ].
        each { |func| assert_func_fails( func, "TEST" ) }

        assert_func_fails(
            lambda { assert_match( /^a*$/, "b" ) },
            '"b" does not match (?-mix:^a*$)' )

        assert_func_fails(
            lambda { assert_raised( MarkerError ) {} },
            "Expected raise of one of " \
            "[BitGirder::Testing::TestingTests::MarkerError]"
        )

        assert_func_fails(
            lambda { assert_raised( "NOPE", MarkerError ) {} },
            "Expected raise of one of " \
            "[BitGirder::Testing::TestingTests::MarkerError]"
        )
    end

    def test_assert_raised_retval
 
        ex =
            assert_raised( MarkerError ) do
                raise MarkerError.new( "HI" )
            end
        
        assert_equal( "HI", ex.message )
        assert_equal( MarkerError, ex.class )
    end

    def test_assert_raised_unmatched_reraises
        
        expct = Exception.new

        begin
            assert_raised( MarkerError ) { raise expct }
        rescue Exception => ex
            assert_equal( expct, ex )
        end
    end

    def test_assert_raised_with_pattern_success
        
        [ "foo", /^foo$/ ].each do |arg|
            assert_raised( arg, MarkerError ) do
                raise MarkerError.new( "foo" )
            end
        end
    end

    def test_assert_raised_with_pattern_fails
        
        [ "foo", /^foo$/ ].each do |arg|

            ex = 
                assert_raised( Exception ) do
                    assert_raised( arg, MarkerError ) do
                        raise MarkerError.new( "bar" )
                    end
                end
            
            assert_equal(
                %q{Message "bar" does not match (?-mix:^foo$)}, ex.message )
        end
    end
end

end
end
