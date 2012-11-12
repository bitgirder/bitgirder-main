require 'bitgirder/core'
require 'bitgirder/io'
require 'bitgirder/io/testing'
require 'mingle'
require 'mingle/codec'
require 'mingle/codec/test-spec'
require 'mingle/io/stream'
require 'mingle/test-support'
require 'bitgirder/testing'

module Mingle
module Codec

class StandardCodecTests < BitGirder::Core::BitGirderClass

    include BitGirder::Core

    include TestClassMixin
    is_test_base

    include BitGirder::Io::Testing

    def initialize
        @peer_ops = []
    end

    def create_peer_call

        cmd = BitGirder::Io.which( "codec-tester", true )
        BitGirder::Io::UnixProcessBuilder.new( :cmd => cmd )
    end

    def run_next_peer_block
        
        EM.defer( 
            proc {
                begin
                    @peer_ops[ 0 ].call
                    nil
                rescue Exception => err
                    err
                end
            },
            proc { |err|
                if err
                    warn( err, "peer op failed" )
                else
                    @peer_ops.shift
                    run_next_peer_block unless @peer_ops.empty?
                end
            }
        )
    end

    def with_peer( ctx, &blk )

        not_nil( blk, :blk )

        @peer_ops << lambda { 
            begin blk.call; rescue Exception => e; ctx.fail_invocation( e ); end
        }

        run_next_peer_block if @peer_ops.size == 1
    end
    
    def exec_request( req )
        
        resp = @peer.exchange_message( req )

        if err = resp.headers.fields.get_string( :exception )
            raise "Remote error: #{err}"
        end

        resp
    end

    bg_abstract :get_codec_id
    bg_abstract :get_codec

    def get_spec_codec( spec )
        get_codec
    end

    def get_spec_keys
        
        req = { 
            :headers => { 
                :command => :get_spec_keys, :codec_id => get_codec_id 
            }
        }
        opt_encode( exec_request( req ).body, "utf-8" ).split( /,/ )
    end

    define_before :init_test
    def init_test( ctx )
        
        @peer = 
            Mingle::Io::Stream::MinglePeer.
                open( :proc_builder => create_peer_call )

        with_peer( ctx ) { ctx.complete { @spec_keys = get_spec_keys } }
    end

    define_after :close_peer
    def close_peer( ctx )

        if @peer
            with_peer( ctx ) do 
                ctx.complete do
                    exec_request( :headers => { :command => :close } )
                    @peer.await_exit( :expect_success => false )
                end
            end
        else
            ctx.complete
        end
    end
    
    def read_spec( spec_key, codec )

        resp = exec_request( 
            :headers => {
                :command => :get_spec,
                :spec_key => spec_key,
                :codec_id => get_codec_id,
            }
        )
        
        TestSpec.decode( resp.body, codec )
    end

    def call_check_encode( val, spec, codec )

        exec_request(
            :headers => { 
                :command => :check_encode,
                :codec_id => get_codec_id,
                :spec_key => spec.key,
            },
            :body => MingleCodecs.encode( codec, val )
        )
    end

    def do_round_trip( spec, codec )
        
        expct = ModelTestInstances.expect_std_test_struct( spec.id )
        ModelTestInstances.assert_equal( expct, spec.action.struct )

        # Do our own check in addition to the remote one
        enc_buf = MingleCodecs.encode( codec, spec.action.struct )
        dec_val = MingleCodecs.decode( codec, enc_buf )
        ModelTestInstances.assert_equal( spec.action.struct, dec_val )

        call_check_encode( spec.action.struct, spec, codec )
    end 

    # Overrideable
    def expected_error_message_for( spec )
        spec.action.error_message
    end

    def do_fail_decode( spec, codec )
        
        msg_expct = expected_error_message_for( spec )

        begin
            MingleCodecs.decode( codec, spec.action.input )
            fail_test( "Decode succeeded (expected #{msg_expct.inspect})" )
        rescue MingleCodecError => e
            raise e unless e.message == msg_expct
        end
    end

    def do_decode_input( spec, codec )
        
        val = MingleCodecs.decode( codec, spec.action.input )
        ModelTestInstances.assert_equal( spec.action.expect, val )
    end

    def do_encode_value( spec, codec )
        call_check_encode( spec.action.value, spec, codec )
    end

    def do_spec_action( spec, codec )
        
        case spec.action
            when RoundTrip then do_round_trip( spec, codec )
            when FailDecode then do_fail_decode( spec, codec )
            when DecodeInput then do_decode_input( spec, codec )
            when EncodeValue then do_encode_value( spec, codec )
        end
    end

    def execute_spec( spec_key )

        codec = get_codec
        spec = read_spec( spec_key, codec )
        spec_codec = get_spec_codec( spec )
        do_spec_action( spec, spec_codec )
    end

    invocation_factory :get_spec_tests
    def get_spec_tests

        @spec_keys.inject( {} ) do |h, spec_key|
 
            nm = "test_codec/spec=#{spec_key}"
            h[ nm ] = lambda { |ctx| 
                with_peer( ctx ) { ctx.complete { execute_spec( spec_key ) } }
            }

            h
        end
    end
end

end
end
