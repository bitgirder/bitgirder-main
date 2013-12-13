require 'mingle'

require 'bitgirder/core'
require 'bitgirder/io/testing'

module Mingle

module ModelTestInstances

    require 'set'

    include BitGirder::Core
    extend BitGirderMethods

    extend BitGirder::Io::Testing

    TEST_NS = MingleNamespace.get( "mingle:test@v1" )

    TEST_LIST1 = 
        MingleList.new( 
            Array.new( 5 ) { |i| MingleString.new( "string#{i}" ) } )
    
    TEST_SYM_MAP1 = {
        :string_sym1 => "something to do here",
        :int_sym1 => 1234,
        :decimal_sym1 => 3.14,
        :bool_sym1 => false,
        :list_sym1 => TEST_LIST1
    }

    TEST_TIMESTAMP1 = 
        MingleTimestamp.rfc3339( "2007-08-24T13:15:43.12345-08:00" );

    TEST_TIMESTAMP2 = MingleTimestamp.rfc3339( "2007-08-24T13:15:43-08:00" );

    TEST_ENUM1_CONSTANT1 = 
        MingleEnum.new( 
            :type => "mingle:test@v1/TestEnum1", 
            :value => :constant1 )

    module_function

    def make_test_buffer( sz )
        
        res = opt_encode( "", "binary" )
        sz.times { |i| res << i }

        res
    end

    TEST_BYTE_BUFFER1 = MingleBuffer.new( make_test_buffer( 150 ) )

    TEST_STRUCT1_INST1 = 
        MingleStruct.new(
            :type => "mingle:test@v1/TestStruct1",
            :fields => {
                :string1 => "hello",
                :bool1 => true,
                :int1 => 32234,
                :int2 => MingleInt64.new( ( 2 ** 63 ) - 1 ),
                :int3 => MingleInt32.new( ( 2 ** 31 ) - 1 ),
                :double1 => MingleFloat64.new( 1.1 ),
                :float1 => MingleFloat32.new( 1.1 ),
                :buffer1 => TEST_BYTE_BUFFER1,
                :list1 => TEST_LIST1,
                :timestamp1 => TEST_TIMESTAMP1,
                :timestamp2 => TEST_TIMESTAMP2,
                :enum1 => TEST_ENUM1_CONSTANT1,
                :symbol_map1 => TEST_SYM_MAP1,
                :struct1 =>
                    MingleStruct.new(
                        :type => "mingle:test@v1/TestStruct2",
                        :fields => { :i1 => 111 }
                    )
            }
        )
    
    TYPE_COV_STRUCT1 = 
        MingleStruct.new(
            :type => "mingle:test@v1/TypeCov",
            :fields => {
                :f1 => MingleString.new( "hello" ),
                :f2 => MingleBoolean::TRUE,
                :f3 => MingleInt32.new( 1 ),
                :f4 => MingleInt64.new( 1 ),
                :f5 => MingleUint32.new( 1 ),
                :f6 => MingleUint64.new( 1 ),
                :f7 => MingleFloat32.new( 1.0 ),
                :f8 => MingleFloat64.new( 1.0 ),
                :f9 => TEST_BYTE_BUFFER1,
                :f10 => TEST_ENUM1_CONSTANT1,
                :f11 => TEST_TIMESTAMP1,
                :f12 => MingleList.new( [] ),
                :f13 => MingleList.new( [
                    MingleInt32.new( 1 ),
                    MingleString.new( "hello" ),
                    MingleList.new( [ MingleBoolean::TRUE ] )
                ] ),
                :f14 => MingleSymbolMap.create( {} ),
                :f15 => MingleSymbolMap.create(
                    :k1 => MingleString.new( "hello" ),
                    :k2 => MingleList.new(
                        %w{ a b c }.map { |s| MingleString.new( s ) }
                    ),
                    :k3 => MingleSymbolMap.create( :kk1 => MingleBoolean::TRUE )
                ),
                :f16 => TEST_STRUCT1_INST1
            }
        )

    # Regardless of ruby version, GCLEF will be a utf8 encoded g-clef
    GCLEF = RubyVersions.when_19x( "\360\235\204\236" ) do |s| 
        s.force_encoding( "utf-8" )
    end

    STD_TEST_STRUCTS =
        MingleSymbolMap.create(

            :test_struct1_inst1 => TEST_STRUCT1_INST1,

            :type_cov_struct1 => TYPE_COV_STRUCT1,

            :empty_struct => MingleStruct.new( :type => :"ns1@v1/S1" ),

            :empty_val_struct =>
                MingleStruct.new(
                    :type => :"ns1@v1/S1",
                    :fields => { 
                        :buf1 => MingleBuffer.new( "", :in_place ),
                        :str1 => "", 
                        :list1 => [], 
                        :map1 => {} 
                    }
                ),
            
            :nulls_in_list =>
                MingleStruct.new(
                    :type => :"ns1@v1/S1",
                    :fields => { :list1 => [ "s1", nil, nil, "s4" ] }
                ),
            
            :unicode_handler =>
                MingleStruct.new(
                    :type => :"ns1@v1/S1",
                    :fields => {
                        :s0 => "hello",
                        # s1 is Utf-8 of U+01fe (Ǿ)
                        :s1 => opt_force_encoding( "\307\276", "utf-8" ), 
                        :s2 => GCLEF
                    }
                )
        )

    def expect_std_test_struct( obj_id )
        
        STD_TEST_STRUCTS.expect_mingle_struct( obj_id ) or 
            raise "No such test struct: #{obj_id}"
    end

    def fail_assert( msg, path )
        raise "#{path.format}: #{msg}"
    end

    def assert( res, msg, path )
        fail_assert( msg, path ) unless res
    end

    def assert_equal_objects( expct, actual, path )

        assert( 
            expct == actual,
            "Objects are not equal: expected #{expct.inspect} " \
            "(#{expct.class}) but got #{actual.inspect} (#{actual.class})", 
            path )
    end

    def assert_equal_with_coercion( expct, actual, path )
        
        actual = MingleModels.as_mingle_instance( actual, expct.class )
        assert_equal_objects( expct, actual, path )
    end

    def assert_equal_type_ref( expct, actual, path )

        assert_equal_objects( expct.type, actual.type, path.descend( "type" ) )
    end

    def key_set_to_s( set )
        set.map { |sym| sym.external_form }
    end

    def assert_equal_symbol_map( m1, m2, path )
        
        k1 = m1.keys
        k2 = m2.keys

        assert( 
            Set.new( k1 ) == Set.new( k2 ),
            "Symbol map key sets are not :equal => " \
            "#{key_set_to_s( k1 )} != #{key_set_to_s( k2 )}", path )
        
        k1.each { |k| assert_equal( m1[ k ], m2[ k ], path.descend( k ) ) }
    end

    def assert_equal_struct( expct, actual, path )
        
        assert_equal_objects( expct.class, actual.class, path )
        assert_equal_type_ref( expct, actual, path )
        
        assert_equal_symbol_map( expct.fields, actual.fields, path )
    end

    def assert_equal_buffers( expct, actual, path )
        
        assert_equal_objects( 
            expct.buf, MingleModels.as_mingle_buffer( actual ).buf, path )
    end

    def assert_equal_timestamps( expct, actual, path )
        
        assert_equal_objects(
            expct, MingleModels.as_mingle_timestamp( actual ), path )
    end

    def assert_equal_enum( expct, actual, path )
        
        case actual

            when MingleEnum
                assert_equal_objects( expct.type, actual.type, path )
                assert_equal_objects( expct.value, actual.value, path )

            when MingleString
                assert_equal_objects(
                    expct.value, MingleIdentifier.get( actual ), path )
            
            else fail_assert( "Unhandled enum value: #{val} (#{val.class})" )
        end
    end

    def assert_equal_lists( expct, actual, path )

        path = path.start_list

        expct.zip( actual ) { |pair|

            assert_equal( pair.shift, pair.shift, path )
            path = path.next
        }
    end

    def assert_equal_service_requests( expct, actual, path )
        
        [ :namespace, :service, :operation, :authentication ].each do |sym|

            assert_equal_objects( 
                expct.send( sym ), 
                actual.send( sym ), 
                path.descend( sym.to_s ) )
        end

        assert_equal( 
            expct.parameters, actual.parameters, path.descend( :parameters ) )
    end

    def assert_equal_service_responses( expct, actual, path )
        
        key = expct.ok? ? :result : :error
 
        fld_key = "get_#{key}"

        assert_equal( 
            expct.send( fld_key ), actual.send( fld_key ), path.descend( key ) )
    end

    def assert_equal( expct, 
                      actual, 
                      path = BitGirder::Core::ObjectPath.get_root( "expct" ) )
        case expct

            when MingleNull, NilClass 
                assert_equal_objects( expct, actual, path )

            when MingleString, MingleBoolean, MingleInt64, MingleInt32,
                 MingleUint32, MingleUint64, MingleFloat64, MingleFloat32
                assert_equal_with_coercion( expct, actual, path )

            when MingleStruct then assert_equal_struct( expct, actual, path )

            when MingleBuffer then assert_equal_buffers( expct, actual, path )

            when MingleTimestamp 
                assert_equal_timestamps( expct, actual, path )

            when MingleEnum
                assert_equal_enum( expct, actual, path )

            when MingleList then assert_equal_lists( expct, actual, path )

            when MingleSymbolMap 
                assert_equal_symbol_map( expct, actual, path )

            when MingleServiceRequest
                assert_equal_service_requests( expct, actual, path )

            when MingleServiceResponse
                assert_equal_service_responses( expct, actual, path )

            else 
                fail_assert( 
                    "Don't know how to assert equality for #{expct.class}", 
                    path )
        end
    end
end

#module TestServiceConstants
#
#    require 'set'
#
#    TEST_NS = MingleNamespace.get( "mingle:service:test@v1" );
# 
#    TEST_SVC = MingleIdentifier.get( "test-service" );
#
#    OP_GET_TEST_STRUCT1_INST1 = 
#        MingleIdentifier.get( "get-test-struct1-inst1" )
#
#    OP_REVERSE_STRING = MingleIdentifier.get( "reverse-string" );
#
#    OP_DO_DELAYED_ECHO = MingleIdentifier.get( "do-delayed-echo" );
#
#    OP_DO_AUTHENTICATED_ACTION = 
#        MingleIdentifier.get( "do-authenticated-action" )
#    
#    OP_DO_AUTHORIZED_ACTION = MingleIdentifier.get( "do-authorized-action" );
#
#    OP_TEST_FAILURES = MingleIdentifier.get( "test-failures" );
# 
#    OP_TEST_ASYNC_FAILURES = MingleIdentifier.get( "test-async-failures" );
#
#    THE_UNACCEPTABLE = "a string that should not be passed";
#
#    UNACCEPTABLE_STRING_VALUE_MESSAGE = "Unacceptable string value";
#
#    ID_STR = MingleIdentifier.get( "str" );
#
#    ID_AUTH_TOKEN = MingleIdentifier.get( "auth-token" );
#    
#    ID_AUTH_EXPIRED = MingleIdentifier.get( "auth-expired" );
#
#    ID_FAILURE_TYPE = MingleIdentifier.get( "failure-type" );
#    
#    ID_ECHO_VALUE = MingleIdentifier.get( "echo-value" );
#    
#    ID_DELAY_MILLIS = MingleIdentifier.get( "delay-millis" );
#
#    FAIL_TYPE_EXCEPTION = "exception";
#    FAIL_TYPE_ERROR = "error";
#
#    VALID_AUTH_TOKEN1 = "golden ticket";
#    VALID_AUTH_TOKEN2 = "snickety snicket";
#
#    AUTHENTICATED_TOKENS = 
#        Set.new( [ VALID_AUTH_TOKEN1, VALID_AUTH_TOKEN2 ] ).freeze
#
#    AUTHORIZED_TOKENS = Set.new( [ VALID_AUTH_TOKEN1 ] ).freeze
#
#end
#
#class AbstractMingleServiceTest < BitGirder::Core::BitGirderClass
#
#    require 'test/unit/assertions'
#
#    @@anns = {}
#
#    include Test::Unit::Assertions
#
#    bg_attr :cli
#
#    def self.annotation_context( cls = self )
#        @@anns[ cls ] ||= { :add_method_pending => false, :tests => [] }
#    end
#
#    def self.method_added( name )
#        
#        ctx = annotation_context
#
#        if ctx[ :add_method_pending ]
#
#            ctx[ :tests ] << name
#            ctx[ :add_method_pending ] = false
#        end
#    end
#
#    def self.test_method
#        
#        ctx = annotation_context
#
#        if ctx[ :add_method_pending ]
#            BitGirder::Core::BitGirderLogger.get_logger.warn(
#                "Repeat calls to #{__method__} at #{caller( 1 )[ 0 ]} " \
#                "(missing actual test method?)" )
#        else
#            ctx[ :add_method_pending ] = true
#        end
#    end
#
#    # Returns a modifiable copy of the test method names. TODO: handle inherited
#    # tests.
#    def self.test_methods( cls )
#        Array.new( annotation_context( not_nil( cls, "cls" ) )[ :tests ] )
#    end
#
#    def expect_response( req, ctx )
# 
#        begin
#            @cli.begin( req ) do |resp, ex|
# 
#                begin
#                    if ex
#                        raise ex
#                    else
#                        yield( req, resp )
#                    end
#        
#                    ctx.exit
#     
#                rescue Exception => e 
#                    ctx.fail( e )
#                end
#            end
#    
#        rescue => e
#            ctx.fail( e )
#        end
#    end
# 
#    # blk takes |req, val| where req is the original request and val is the
#    # result value if the response was a success response
#    def expect_success( req, ctx )
#        
#        expect_response( req, ctx ) do |req, resp|
# 
#            if resp.ok?
#                yield req, resp.result 
#            else
#                raise "Request succeeded but service :failed => " +
#                      resp.exception.inspect
#            end
#        end
#    end
#    
#    def expect_failure( req, ctx )
#        
#        expect_response( req, ctx ) do |req, resp|
#            
#            if resp.ok?
#                raise "Request succeeded (expected exception from service"
#            else
#                yield req, resp.exception
#            end
#        end
#    end
#end

end
