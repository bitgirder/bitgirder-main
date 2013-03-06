require 'bitgirder/core'

require 'bigdecimal'
require 'tempfile'

# A default value for TEST_MIXIN, set below
module EmptyTestModule; end

# We set the tests up in this class to be able to run either as instances of
# Test::Unit::TestCase or as BitGirder::Testing tests, depending on whether this
# file is loaded under the guise of a BG test run. We do this here in core since
# we want a fallback way to test core if it turns out that there are problems in
# here which prevent successfully loading or executing other modules, namely
# BitGirder::Testing
if $is_bitgirder_test_runner_run
    require 'bitgirder/testing'
    TEST_CLS = Object
    TEST_MIXIN = BitGirder::Testing::TestClassMixin
    ASSERT_MOD = BitGirder::Testing::AssertMethods
else
    require 'test/unit'
    TEST_CLS = Test::Unit::TestCase
    TEST_MIXIN = EmptyTestModule
    ASSERT_MOD = Test::Unit::Assertions
end

class TestClassBase < TEST_CLS

    include TEST_MIXIN
    include ASSERT_MOD
    
    def assert_raised( msg, err_cls )
        begin
            yield
            raise "Expected #{err_cls} with message #{msg}"
        rescue err_cls => e
            assert_equal( e.message, msg )
        end
    end
end

module BitGirder
module Core

class RubyVersionsTests < TestClassBase
    
    def test_geq_result
        
        assert_equal( 1, RubyVersions.when_geq( "100.0", 1 ) { 2 } )
        assert_equal( nil, RubyVersions.when_geq( "100.0" ) { 2 } )
        assert_equal( 2, RubyVersions.when_geq( "1", 1 ) { 2 } )
        assert_equal( 2, RubyVersions.when_geq( "1", 1 ) { |i| i + 1 } )
    end
end

class ObjectHolder
    
    attr_reader :obj

    def initialize( obj )
        @obj = obj
    end
end

class TestClass1 < BitGirderClass
 
    bg_attr :attr1

    bg_attr :attr2, :processor => :boolean, :required => false

    bg_attr :attr3, :default => lambda { "HELLO" }

    install_hash

    INST1 = 
        TestClass1.new( 
            :attr1 => "hello",
            :attr2 => true,
            :attr3 => "goodbye"
        )
    
    # Used to test that subclasses of BitGirderClass themselves have direct call
    # paths to methods in BitGirderMethods
    def self.foo
        not_nil( 1, :ignore )
    end
end

class TestClass2 < BitGirderClass
    
    bg_attr :attr1

    attr_reader :did_impl_init

    private
    def impl_initialize

        super # Make sure super safe to call

        raise "@attr1 not set" unless @attr1
        @did_impl_init = true
    end
end

class TestClass3 < BitGirderClass
    
    bg_attr :attr1, :default => 12, :mutable => true
end

class TestClass4 < BitGirderClass

    bg_attr :attr1
    bg_attr :attr2, :processor => TestClass3, :required => false

    map_instances(
        :processor => lambda { |val| val.is_a?( ObjectHolder ) && val.obj }
    )

    public
    def foo1
        "stuff"
    end
end

class TestClass5 < BitGirderClass
    
    extend ASSERT_MOD
    
    bg_attr :attr1, :default => 1
    bg_attr :attr2, :default => lambda { [] }
    bg_attr :attr3, :default => Array
    bg_attr :attr4, :default => []
    bg_attr :attr5, :default => [ 1, 2 ]
    bg_attr :attr6, :default => {}
    bg_attr :attr7, :default => { :a => 1 }

    def self.assert_base_instance
        
        self.new.tap do |res|
    
            { :attr1 => 1, :attr2 => [], :attr3 => [], :attr4 => [],
              :attr5 => [ 1, 2 ], :attr6 => {}, :attr7 => { :a => 1 }
            }.
            each_pair do |attr, expct|

                act = res.send( attr )
                msg = "For :#{attr}, wanted #{expct} but got #{act}"
                assert_equal( expct, act, msg )
            end
        end
    end
end

class TestClass6 < BitGirderClass
    
    bg_attr :attr1,
            :is_list => true,
            :required => true,
            :default => lambda { [] },
            :processor => lambda { |elt| elt.to_s }
end

class TestError1 < BitGirderError
    bg_attr :attr1
end

# Used to test that attr :message overrides StandardError#message
class TestError2 < BitGirderError

    bg_attr :message
end

# Used to test that explicit to_s() is used by StandardError#message
class TestError3 < BitGirderError

    bg_attr :val

    public
    def to_s
        "test-message: #@val"
    end
end

class AbstractBase < BitGirderClass
    bg_abstract :foo
end

class Concrete1 < AbstractBase
    def foo( val )
        val
    end
end

class AsInstanceOrderTester < BitGirderClass
    
    # Ensure that the instance mappers are called in reverse order of
    # registration
    map_instances( :processor => lambda { |val| raise "Called unexpectedly" } )
    map_instances( :processor => lambda { |val| self.new } )
end

class AsInstanceShorthandTester < BitGirderClass
    
    bg_attr :val

    # Basic single-type mapper
    map_instance_of( Array ) { |val| new( val.size ) }

    # List of sibling types
    map_instance_of( String, Symbol ) { |val| new( val.to_s.upcase ) }

    # List of concrete classes (dec, float) and a base class (int)
    map_instance_of( BigDecimal, Float, Integer ) { |val| new( val * 2 ) }

    # Should override the builtin Hash selector
    map_instance_of( Hash ) { |val| new( val[ :key1 ] ) }
end

class NonBitGirderClass; end

class MarkerError < StandardError; end

class CoreTests < TestClassBase
    
    include BitGirderMethods

    def test_structure_equalities

        s1A = TestClass1.new( :attr1 => "a1" )
        s1B = TestClass1.new( :attr1 => "a1" )

        assert_equal( s1A, s1B )
        assert( s1A.eql?( s1B ) )

        assert_false( s1A == TestClass1::INST1 )
        assert_false( s1A.eql?( TestClass1::INST1 ) )
    end

    def test_install_hash
        
        cd = BitGirderClassDefinition.for_class( TestClass1 )

        assert_equal(
            cd.hash_instance( TestClass1::INST1 ),
            TestClass1::INST1.hash
        )
    end

    def test_initialize_with_string_attrs
        assert( TestClass1::INST1 != TestClass1.new( "attr1" => "bye" ) )
    end

    def test_has_env_success
        assert_equal( ENV[ "HOME" ], has_env( "HOME" ) )
    end

    def test_has_env_failure
        
        var = "test-env-#{rand( 1 << 100 )}"

        begin
            has_env( var )
            raise "Got env val for #{var}"
        rescue Exception => e

            assert_equal( 
                e.message, "Environment Variable #{var.inspect} cannot be nil" )
        end
    end

    @@class_def_errors = {}

    def assert_class_def_error( sym )
        assert( @@class_def_errors[ sym ] )
    end

    begin
        class TestClass < BitGirderClass
            bg_attr :an_ident, :validatin => :note_validation_misspelled
        end
    rescue BitGirderAttribute::InvalidModifier => e
        @@class_def_errors[ :mistyped_attr ] = true
    end
    
    def test_class_def_with_mistyped_attr_arg
        assert_class_def_error( :mistyped_attr )
    end

    begin
        class Redefine1 < BitGirderClass
            bg_attr :attr1
            bg_attr :attr2
            bg_attr :attr1
        end
    rescue BitGirderAttribute::RedefinitionError => e
        if e.message == "Attribute :attr1 already defined"
            @@class_def_errors[ :attr_redefined_same_class ] = true
        else
            raise
        end
    end

    def test_bg_attr_redefined_fails_same_class
        assert_class_def_error( :attr_redefined_same_class )
    end

    begin
        class Redefine2 < BitGirderClass
            bg_attr :attr1
        end

        class Redefine3 < Redefine2
            bg_attr :attr2
            bg_attr :attr1
        end

    rescue BitGirderAttribute::RedefinitionError => e
        if e.message == "Attribute :attr1 already defined"
            @@class_def_errors[ :attr_redefined_subclass ] = true
        else
            raise
        end
    end

    def test_bg_attr_redefined_fails_subclass
        assert_class_def_error( :attr_redefined_subclass )
    end

    def test_bg_class_extends_bg_methods
        TestClass1.foo
    end

    def test_impl_init_called
        tc2 = TestClass2.new( :attr1 => "hello" )
        assert_equal( "hello", tc2.attr1 ) # Make sure normal attrs set too
        assert( tc2.did_impl_init )
    end

    def test_mutable_attrs
        tc3 = TestClass3.new
        assert_equal( 12, tc3.attr1 )
        tc3.attr1 = 13
        assert_equal( 13, tc3.attr1 )
        tc3.attr1 += 1
        assert_equal( 14, tc3.attr1 )
    end

    def test_default_attr_instantiations
        
        # Check base defaults and also ensure that we're creating new instances
        # for things like list/hash literals, and not sharing the actual default
        # literal across instances
        o1 = TestClass5.assert_base_instance
        o1.attr4 << "o1val"
        o1.attr5 << "o1val"
        o1.attr6[ :o1val ] = "o1val"
        o1.attr7[ :o1val ] = "o1val"

        TestClass5.assert_base_instance
    end

    def test_attr_missing_message

        assert_raised( "attr1: Missing value", RuntimeError ) do 
            TestClass1.new
        end
    end

    def test_has_key

        h = { :a => 1, :b => 2 }

        assert_equal( 1, has_key( h, :a ) )

        assert_raised( "Value for key :c cannot be nil", RuntimeError ) do
            has_key( h, :c )
        end
        
        assert_equal( [ 2, 1 ], has_keys( h, :b, :a ) )
        assert_raised( "Value for key :c cannot be nil", RuntimeError ) do
            has_keys( h, :a, :c )
        end
    end

    def test_as_instance

        chk = lambda { |val| 
            TestClass4.as_instance( val ).tap do |obj|
                assert_equal( 1, obj.attr1 )
                assert( obj.attr2.is_a?( TestClass3 ) )
                assert_equal( 2, obj.attr2.attr1 )
            end
        }
 
        obj = chk.call( { :attr1 => 1, :attr2 => { :attr1 => 2 } } )
        assert( obj.equal?( chk.call( obj ) ) )

        assert( obj.equal?( chk.call( ObjectHolder.new( obj ) ) ) )

        msg =
            %q{Don't know how to convert String to BitGirder::Core::TestClass4}
        assert_raised( msg, TypeError ) { chk.call( "cannot-handle-string" ) }
    end

    def test_as_instance_select_order
        AsInstanceOrderTester.as_instance( "" )
    end

    def test_as_instance_shorthands
        
        chk = lambda { |expct, val|
            obj = AsInstanceShorthandTester.as_instance( val )
            assert_equal( expct, obj.val )
        }

        chk.call( 1, [ "" ] )

        [ :a, "a" ].each { |v| chk.call( v.to_s.upcase, v ) }

        [ 1.0, BigDecimal.new( "2" ), 2 ** 100, 2 ].each do |v| 
            chk.call( v * 2, v )
        end
    end

    def test_custom_hash_instance_mapper
        assert_equal( 
            1, AsInstanceShorthandTester.as_instance( :key1 => 1 ).val )
    end

    def test_attr_processor

        [
            [ :boolean, [ "true", true ], true ],
            [ :symbol, [ "s", :s ], :s ],
            [ :integer, [ 1, "1", 1.1, "1.1" ], 1 ],
            [ :float, [ 1.0, "1.0", 1 ], 1.0 ],
            [ :expand_path, [ "/a/b", "/a/b/c/../../b" ], "/a/b" ],
            [ TestClass1, [ TestClass1::INST1 ], TestClass1::INST1 ]

        ].each do |test|
 
            bg_attr = BitGirderAttribute.new(
                :identifier => :id,
                :processor => test[ 0 ] 
            )
            
            test[ 1 ].each do |val|
                assert_equal( test[ 2 ], bg_attr.processor.call( val ) )
            end
        end
    end

    def test_list_proc_applied_to_default_val
        
        assert_equal( [], TestClass6.new.attr1 )
        assert_equal( [ "1", "2" ], TestClass6.new( :attr1 => [ 1, 2 ] ).attr1 )
    end

    def test_error_basic

        e = TestError1.new( :attr1 => 1 )
        assert_equal( 1, e.attr1 )
        assert( e.is_a?( StandardError ) )
    end

    def test_error_with_message_attr

        assert_equal(
            "test-message",
            TestError2.new( :message => "test-message" ).message
        )
    end

    def test_derived_error_message
        
        e = TestError3.new( :val => 3 )

        [ :message, :to_s ].each do |m|
            assert_equal( "test-message: 3", e.send( m ) )
        end
    end

    def test_abstract_methods
        
        begin
            AbstractBase.new.foo
        rescue RuntimeError => e
            msg = "Abstract method BitGirder::Core::AbstractBase#foo not " \
                  "implemented"
            assert_equal( msg, e.message )
        end

        assert_equal( 1, Concrete1.new.foo( 1 ) )
    end

    def test_attr_processor_bad_class
        
        msg = "Not a BitGirder::Core::BitGirderClass: " \
              "BitGirder::Core::NonBitGirderClass"

        assert_raised( msg, ArgumentError ) do
            BitGirderAttribute.new( 
                :identifier => :id, 
                :processor =>  NonBitGirderClass
            )
        end
    end

    def assert_raisef( cls, argv, err_expct )
            
        begin
            BitGirderMethods.raisef( *argv )
            raise "Nothing raised"
        rescue cls => err
            assert_equal( err_expct.message, err.message )
        end
    end

    def test_raisef
        
        fmt, val, msg = "A val: 0x%04x", 0x11, "A val: 0x0011"

        [ MarkerError, RuntimeError ].each do |cls|

            is_rt = cls == RuntimeError
            argvs = [ [ fmt, val ], [ msg ] ].map do |arr|
                is_rt ? arr : [ cls ] + arr
            end

            argvs << [ cls.new( msg ) ]

            argvs.each { |argv| assert_raisef( cls, argv, cls.new( msg ) ) }

            assert_raisef( 
                cls, 
                is_rt ? [] : [ cls ], 
                is_rt ? cls.new( "" ) : cls.new 
            )
        end
    end
end

class ReflectTests < TestClassBase

    def test_instance_methods
        Reflect.instance_methods_of( String ).include?( :to_s )
    end
end

class ObjectPathTests < TestClassBase

    def test_default_formatting
        
        [
            [ ObjectPath.get_root( "a" ), "a" ],

            [ ObjectPath.get_root( "a" ).
                descend( "b" ).
                descend( "c" ).
                start_list.
                descend( "d" ).
                start_list.
                next.
                next.
                next,
              "a.b.c[ 0 ].d[ 3 ]"
            ],

            [ ObjectPath.get_root_list, "[ 0 ]" ],

            [ ObjectPath.get_root_list.next.descend( "a" ), "[ 1 ].a" ],

        ].each do |pair|
            
            path, expct = *pair
            assert_equal( expct, path.format )
        end
    end
end

class WaitConditionTests < TestClassBase

    class WaitValue < BitGirderClass
        
        bg_attr :complete_at
        bg_attr :do_fail, :required => false
        
        attr_reader :value, :calls

        private
        def impl_initialize

            @value = rand( 1 << 64 )
            @calls = 0
        end

        public
        def get
            
            begin
                if @calls == @complete_at
                    raise MarkerError, @value.to_s if @do_fail
                    @value
                end
            ensure @calls += 1 end
        end
    end

    def assert_wait( v, opts )
        
        raise "Need an expect val (even if nil)" unless opts.key?( :expect )
        expct = opts[ :expect ]

        start = Time.now
        res = yield
        elaps = Time.now - start

        assert_equal( expct, res )
        assert_equal( has_key( opts, :calls ), v.calls )
        assert( elaps >= has_key( opts, :min_wait ) )
    end

    def test_wait_nontrivial_poll
        
        v = WaitValue.new( :complete_at => 3 )

        assert_wait( v, :expect => v.value, :calls => 4, :min_wait => 1.5 ) do
            WaitCondition.wait_poll( :poll => 0.5, :max_tries => 4 ) { v.get }
        end
    end

    def test_wait_immediate_success
        
        v = WaitValue.new( :complete_at => 0 )

        assert_wait( v, :calls => 1, :expect => v.value, :min_wait => 0 ) do
            WaitCondition.wait_poll( :poll => 0.5, :max_tries => 1 ) { v.get }
        end
    end

    def test_wait_exhausted
        
        v = WaitValue.new( :complete_at => 4 )

        assert_wait( v, :calls => 3, :expect => nil, :min_wait => 1 ) do
            WaitCondition.wait_poll( :poll => 0.5, :max_tries => 3 ) { v.get }
        end
    end

    def test_wait_raise
        
        v = WaitValue.new( :complete_at => 3, :do_fail => true )

        start = Time.now
        assert_raised( v.value.to_s, MarkerError ) do
            WaitCondition.wait_poll( :poll => 0.5, :max_tries => 4 ) { v.get }
        end
        elaps = Time.now - start

        assert( elaps >= 1.5 )
        assert_equal( 4, v.calls )
    end

    def test_wait_backoff
 
        v = WaitValue.new( :complete_at => 4 )

        assert_wait( v, :calls => 5, :expect => v.value, :min_wait => 1.5 ) do
            WaitCondition.
                wait_backoff( :seed => 0.1, :max_tries => 5 ) { v.get }
        end
    end
end

end
end
