require 'bitgirder/testing'
require 'bitgirder/testing/integ'
require 'mingle/http'
require 'mingle/test-support'
require 'mingle/io'
require 'mingle/codec'
require 'mingle/json'
require 'mingle/bincodec'
require 'bitgirder/ops/java'

require 'time'

module Mingle
module Http

class HttpClientTests < BitGirder::Core::BitGirderClass

    include TestClassMixin
    is_test_base

    include BitGirder::Ops::Java
    include Mingle::Codec
    include Mingle::Json
    include Mingle

    HDR_CALL_ID = "x-client-call-id"

    SVC1_OPTS = { :namespace => "mingle:tck@v1", :service => :service1 }

    SVC3_OPTS = {
        :namespace => "mingle:tck@v1",
        :service => :service3,
        :operation => :who_am_i
    }

    TYPE_AUTH_EXCEPTION1 =
        MingleTypeReference.get( "mingle:tck@v1/AuthException1" )

    TYPE_AUTH_EXCEPTION1_SUBTYPE =
        MingleTypeReference.get( "mingle:tck@v1/AuthException1Subtype" )

    AUTH_RES_TYPES = 
        %w{ accept bad_auth bad_auth2 undeclared_fail internal_fail }.
            map { |s| s.to_sym }

    AUTH_NAME1 = "BigBird"

    AUTH_TOK1 =
        MingleStruct.new(
            :type => "mingle:tck@v1/AuthToken1",
            :fields => { :res_type => :accept, :name => AUTH_NAME1 }
        )

    TYPE_OP_STRUCT1 = MingleTypeReference.get( "mingle:tck@v1/OpStruct1" )

    TYPE_INTERNAL_SERVICE_EXCEPTION =
        MingleTypeReference.get( "service@v1/InternalServiceException" )

    TYPE_EXCEPTION1 = MingleTypeReference.get( "mingle:tck@v1/Exception1" )

    TYPE_VALIDATION_EXCEPTION = 
        MingleTypeReference.get( "mingle:core@v1/ValidationException" )

    ECHO_PARAMS_DEFAULT_VALS = 
    [
        "hello",
        :A1,
        1234567,
        0,
        3.1415,
        42.0,
        "null",
        true,
        "1991-12-11T02:30:47.012300000-08:00",
        "null",
        [],
        [],
        [],
        []
    ]

    ECHO_VALUE1 = MingleString.new( "echo_val1" )

    ECHO_PARAMS_PARAM_NAMES = 
        %w{ string1 enum1 int1 int2 dec1 dec2 buf1 bool1 time1
            op_struct1 strings1 enums1 ints1 op_structs1 }.map { |s| s.to_sym }
    
    BUFFER1 = MingleBuffer.new( "\x00\x01\x02\x03\x04\x05".encode!( "binary" ) )

    def initialize
        @peer_codec = JsonMingleCodec.new
    end

    private
    def create_peer_builder
 
        ig_ctx = BitGirder::Testing::Integ::IntegrationTesting.get_integ_context

        jv_env = 
            JavaEnvironment.new( ig_ctx[ :java ].expect_string( :java_home ) )

        classpath = ig_ctx[ :java ].expect( :classpath ).to_a
        if rsrcs = ENV[ "JBUILDER_TEST_RESOURCES" ]
            classpath << rsrcs
        end

        JavaRunner.create_mingle_app_runner(
            :java_env => jv_env,
            :classpath => classpath,
            :app_class => "com.bitgirder.mingle.http.v1.HttpTestServer",
            :sys_props => {
                "com.bitgirder.application.Applications.logStream" => "STDERR" 
            },
            :argv => %w{}
        ).
        process_builder
    end

    private
    def get_http_servers
        
        resp = 
            @peer.exchange_message( 
                :mods => { :command => :get_http_servers, :codec => :json } )

        @servers = MingleCodecs.decode( @peer_codec, resp[ :body ] )
    end

    define_before :init
    def init( ctx )

        @peer = 
            Mingle::Io::MinglePeer.
                open( :proc_builder => create_peer_builder )

        ctx.defer_block { ctx.complete { get_http_servers } }
    end

    define_after :close_peer
    def close_peer( ctx )

        ctx.defer_block do 

            ctx.complete do
                
                @peer.exchange_message( :mods => { :command => :close } )
                @peer.await_exit
            end
        end
    end

    class TestCall < BitGirder::Core::BitGirderClass
        
        bg_attr :name
        bg_attr :call_opts
        bg_attr :identifier => :with_result
        bg_attr :identifier => :with_exception

        attr_reader :call_id
    
        private
        def next_call_id
            Array.new( 16 ) { |i| sprintf( "%02x", rand( 256 ) ) }.join( "" )
        end

        def initialize( *argv )

            super( *argv )

            @with_result ||= lambda { |res| raise "Unexpected result: #{res}" }

            @with_exception ||=
                lambda { |me|
                    code( "Raising me: #{resp.exception.inspect}" )
                    MingleModels.raise_as_ruby_exception( resp.exception )
                }
            
            @call_id = next_call_id
        end
    end

    class TestCallStats < BitGirder::Core::BitGirderClass

        bg_attr :elapsed
    end

    class NetHttpReactorImpl < BitGirder::Core::BitGirderClass
        
        bg_attr :test_call

        public
        def complete_request( req )
            req[ HDR_CALL_ID ] = @test_call.call_id
        end

        public
        def response_received( resp )
            code( "Got resp headers: #{resp.to_hash}" )
        end
    end

    private
    def net_http_cli_for( opts )

        srv_info = has_key( opts, :srv_info )

        NetHttpMingleRpcClient.new(
            :location => ServerLocation.new(
                :host => srv_info.fields.expect_string( :host ),
                :port => srv_info.fields.expect_int( :port ),
                :uri => srv_info.fields.expect_string( :uri )
            ),
            :reactor => NetHttpReactorImpl.new( has_key( opts, :test_call ) ),
            :codec_ctx => has_key( opts, :codec_ctx )
        )
    end

    private
    def get_server_info( nm )
        
        @servers[ :servers ].find do |info|
            info.fields.expect_string( :name ) == nm
        end or raise "No server with name #{nm}"
    end

    private
    def deliver_resp( resp, stats, test_call )
    
        if resp.ok?
            argv = [ resp.result ]
            block = test_call.with_result
        else
            argv = [ resp.exception ]
            block = test_call.with_exception
        end
            
        argv << stats if block.arity == 2
        block.call( *argv )
    end

    private
    def exec_net_http_call( opts )
        
        test_ctx = has_key( opts, :test_ctx )
        test_call = has_key( opts, :test_call )

        test_ctx.defer_block do
 
            start_t = Time.now
            cli = net_http_cli_for( opts )
            resp = cli.call( test_call.call_opts )
            stats = TestCallStats.new( :elapsed => ( Time.now - start_t ) * 1000 )
            
            test_ctx.complete { deliver_resp( resp, stats, test_call ) }
        end
    end

    private
    def exec_test_call( opts )
        
        case t = has_key( opts, :transport )
            when :net_http then exec_net_http_call( opts ) 
            else raise "Unexpected transport: #{t}"
        end
    end

    private
    def make_op_struct_echo_str( ms )
        
        raise "Not an OpStruct1" unless ms.type == TYPE_OP_STRUCT1
        "OpStruct1:{ f1: #{ms[ :f1 ]} }"
    end

    private
    def make_echo_params_res( argv )
        
        argv.map do |arg|
            case arg
                when Array then "[#{make_echo_params_res( arg )}]"
                when MingleBuffer then Zlib.crc32( arg.buf )
                when MingleStruct then make_op_struct_echo_str( arg )
                when MingleEnum then arg.value.format( :lc_underscore ).upcase
                else arg.to_s
            end
        end.join( "|" )
    end

    private
    def op_struct1( f1 )
        MingleStruct.new( :type => TYPE_OP_STRUCT1, :fields => { :f1 => f1 } )
    end

    ignore_test :test_enum1
    private
    def test_enum1( val )
        MingleEnum.new( :type => :"mingle:tck@v1/TestEnum1", :value => val )
    end

    private
    def create_echo_params_call( name, params = nil )

        TestCall.new(
            :name => name,
            :call_opts => SVC1_OPTS.merge( 
                :operation => :echo_params,
                :parameters => params || {}
            ),
            :with_result => lambda { |res|

                param_vals = 
                    if params
                        params.values_at( *ECHO_PARAMS_PARAM_NAMES )
                    else
                        ECHO_PARAMS_DEFAULT_VALS
                    end

                assert_equal( make_echo_params_res( param_vals ), res.to_s )
            }
        )
    end

    private
    def assert_delay( stats, async, delay )
        assert( stats.elapsed >= delay ) if ( async && delay )
    end

    private
    def assert_error( ex, err_mode )

        case err_mode

            when :declared_exception
                assert_equal( TYPE_EXCEPTION1, ex.type )

            when :internal_exception, :internal_error, :undeclared_exception
                assert_equal( TYPE_INTERNAL_SERVICE_EXCEPTION, ex.type )
 
            when :general_exception
                assert_equal( TYPE_VALIDATION_EXCEPTION, ex.type )

            when nil then MingleModels.raise_as_ruby_exception( ex )

            else raise "Unhandled err mode: #{err_mode}"
        end
    end

    private
    def assert_asyncable_res( res, async, stats, delay )
        
        assert_delay( stats, async, delay )
        assert_equal( ECHO_VALUE1.to_s, res.to_s )
    end

    private
    def assert_asyncable_error( ex, err_mode, async, stats, delay )
        
        # Otherwise the server will have failed quickly
        assert_delay( stats, async, delay )
        assert_error( ex, err_mode )
    end

    private
    def create_asyncable_test_call( async, delay, err_mode )
            
        TestCall.new(
            :name => "op=do_asyncable,async=#{async},delay=#{delay}," \
                  "err_mode=#{err_mode}",
            :call_opts => SVC1_OPTS.merge(
                :operation => :do_asyncable,
                :parameters => {
                    :echo_value => ECHO_VALUE1,
                    :delay => delay,
                    :error_mode => err_mode,
                    :async => async
                }
            ),
            :with_result => lambda { |res, stats| 
                assert_asyncable_res( res, async, stats, delay )
            },
            :with_exception => lambda { |ex, stats|
                assert_asyncable_error( ex, err_mode, async, stats, delay )
            }
        )
    end

    private
    def get_do_asyncable_calls
        
        res = []

        [ true, false ].each { |async|
        [ nil, 1000 ].each { |delay|
        [ nil, :internal_exception, :declared_exception ].each { |err_mode|
            res << create_asyncable_test_call( async, delay, err_mode )
        }}}

        res
    end

    private
    def get_produce_error_calls
        
        [ :internal_error, 
          :internal_exception, 
          :general_exception,
          :declared_exception, 
          :undeclared_exception ].inject( [] ) do |res, err_mode|
            
            res << TestCall.new(
                :name => "op=produce-error1,err_mode=#{err_mode}",
                :call_opts => SVC1_OPTS.merge(
                    :operation => :produce_error1,
                    :parameters => { :mode => err_mode }
                ),
                :with_exception => lambda { |ex| assert_error( ex, err_mode ) }
            )
        end
    end

    private
    def assert_auth_exception( ex, auth_res_type )
        
        case auth_res_type
            
            when :internal_fail, :undeclared_fail
                assert_equal( TYPE_INTERNAL_SERVICE_EXCEPTION, ex.type )

            when :bad_auth 
                assert_equal( TYPE_AUTH_EXCEPTION1, ex.type )

            when :bad_auth2
                assert_equal( TYPE_AUTH_EXCEPTION1_SUBTYPE, ex.type )
            
            else raise "Unhandled auth res type: #{auth_res_type}"
        end
    end

    private
    def create_who_am_i_call( auth_res_type, async )

        opts = {
            :name => "op=who-am-i,auth_res_type=#{auth_res_type},async=#{async}",
            :call_opts => SVC3_OPTS.merge( 
                :authentication => {
                    :res_type => auth_res_type, :async => async, :name => AUTH_NAME1
                }
            )
        }

        if auth_res_type == :accept
            opts[ :with_result ] = lambda { |res|
                ModelTestInstances.assert_equal( AUTH_TOK1[ :name ], res )
            }
        else
            opts[ :with_exception ] = lambda { |ex|
                assert_auth_exception( ex, auth_res_type )
            }
        end

        TestCall.new( opts )
    end

    private
    def get_who_am_i_calls
        
        res = []

        AUTH_RES_TYPES.each { |auth_res_type|
        [ true, false ].each { |async|
            res << create_who_am_i_call( auth_res_type, async )
        }}

        res
    end

    private
    def get_test_calls

        [
            TestCall.new(
                :name => :void_op,
                :call_opts => SVC1_OPTS.merge( :operation => :void_op ),
                :with_result => lambda { |res| assert_nil( res ) }
            ),

            TestCall.new(
                :name => :get_string_list_non_empty,
                :call_opts => SVC1_OPTS.merge(
                    :operation => :get_string_list,
                    :parameters => { :str => :hello, :copies => 4 }
                ),
                :with_result => lambda { |res|
                    assert_equal(
                        Array.new( 4, :hello ).join( "|" ), res.join( "|" ) )
                }
            ),

            TestCall.new(
                :name => :get_string_list_empty,
                :call_opts => SVC1_OPTS.merge(
                    :operation => :get_string_list,
                    :parameters => { :str => "hello", :copies => 0 }
                ),
                :with_result => lambda { |res| assert( res.empty? ) }
            ),

            create_echo_params_call( :echo_params_with_defaults ),

            create_echo_params_call( 
                :echo_params_with_cli_vals, 
                {
                    :string1 => "str1",
                    :enum1 => test_enum1( :a2 ),
                    :int1 => 2 << 50,
                    :int2 => 5555,
                    :dec1 => 12.0,
                    :dec2 => -1.9,
                    :buf1 => BUFFER1,
                    :bool1 => true,
                    :time1 => MingleTimestamp.now,
                    :op_struct1 => op_struct1( "os1" ),
                    :strings1 => %w{ s1 s2 s3 },
                    :enums1 => [ test_enum1( :a1 ), test_enum1( :a2 ) ],
                    :ints1 => [ 1, 3, 4, 7, 11 ],
                    :op_structs1 => 
                        Array.new( 3 ) { |i| op_struct1( "os1.#{i + 1}" ) }
                }
            ),

            get_do_asyncable_calls,
            get_produce_error_calls,

            TestCall.new(
                :name => :echo_opaque_value,
                :call_opts => SVC1_OPTS.merge(
                    :operation => :echo_opaque_value,
                    :parameters => { :value => ECHO_VALUE1 }
                ),
                :with_result => lambda { |res| 
                    ModelTestInstances.assert_equal( ECHO_VALUE1, res )
                }
            ),

            TestCall.new(
                :name => :get_opaque_list,
                :call_opts => SVC1_OPTS.merge(
                    :operation => :get_opaque_list,
                    :parameters => { :value => ECHO_VALUE1, :copies => 5 }
                ),
                :with_result => lambda { |res|
                    ModelTestInstances.assert_equal(
                        MingleList.new( Array.new( 5 ) { |i| ECHO_VALUE1 } ),
                        res
                    )
                }
            ),

            TestCall.new(
                :name => :validation_op1,
                :call_opts => SVC1_OPTS.merge(
                    :operation => :validation_op1,
                    :parameters => { :str => "bbb" }
                ),
                :with_exception => lambda { |ex|
                    
                    assert_equal( TYPE_VALIDATION_EXCEPTION, ex.type )
                    assert_equal(
                        %q{Value does not match "^a+$": "bbb"},
                        ex[ :message ].to_s
                    )
                }
            ),

            get_who_am_i_calls,
 
        ].flatten
    end

    private
    def get_codec_contexts

        [
            HttpCodecContext.new(
                :codec => JsonMingleCodec.new, :content_type => "application/json" ),
            
            HttpCodecContext.new(
                :codec => Mingle::BinCodec::MingleBinCodec.new, 
                :content_type => "application/mingle-binary" 
            )
        ] 
    end

    private
    def make_name( srv_nm, opts )
        
        opt_str =
        [
            [ "server", srv_nm ],
            [ "call", has_key( opts, :test_call ).name ],
            [ "codec", has_key( opts, :codec_ctx ).content_type ],
            [ "transport", has_key( opts, :transport ) ]
        ].
        map { |pair| pair.join( "=" ) }.join( "," )
        
        "test_tck_call/#{opt_str}"
    end

    private
    def get_tests_for_name( nm )

        res = {}
        srv_inf = get_server_info( nm )

        # Order of loop nesting is important, since each tc has its own call id
        # and other accounting information, and needs to be created new for each
        # test call we'll execute
        get_codec_contexts.each { |codec_ctx|
        get_test_calls.each { |tc|

            call_opts = {
                :test_call => tc,
                :srv_info => srv_inf,
                :transport => :net_http,
                :codec_ctx => codec_ctx
            }
            
            res[ test_nm = make_name( nm, call_opts ) ] = lambda { |ctx| 
                
                code( "Starting test call #{test_nm}, call_id: #{tc.call_id}" )
                exec_test_call( call_opts.merge( :test_ctx => ctx ) ) 
            }
        }}

        res
    end

    bg_abstract :get_test_server_names

    invocation_factory :get_tck_call_tests
    def get_tck_call_tests
 
        res = {}

        get_test_server_names.each do |nm| 
            res.merge!( get_tests_for_name( nm ) )
        end

        res
    end
end

end
end
