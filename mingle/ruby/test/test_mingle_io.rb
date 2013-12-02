require 'mingle'

require 'mingle/test-support'

require 'bitgirder/io'
include BitGirder::Io

require 'bitgirder/testing'

require 'thread'
require 'stringio'

module Mingle

class AbstractCoreIoTest < BitGirderClass

    EXPCT_VALS = {}

    bg_attr :name
    bg_attr :data

    attr_accessor :ctx, :peer_q

    bg_abstract :run_test

    private
    def get_expect_value
        EXPCT_VALS[ @name ] or raise "no expected value"
    end

    private
    def reader
        @reader ||= BinReader.as_bin_reader( StringIO.new( @data ) )
    end

    private
    def read_matching_val( expct )
        
        case expct
        when MingleValue then reader.read_value
        end
    end

    private
    def assert_read_matches( expct, act )
        
        case expct
        when MingleValue then ModelTestInstances.assert_equal( expct, act )
        else raise "unhandled expct val: #{expct.class}"
        end
    end

    private
    def assert_matching_val( expct )
        read_matching_val( expct ).tap { |v| assert_read_matches( expct, v ) }
    end

    private
    def writer

        unless @writer
            @write_buf = StringIO.new( create_binary_string )
            @writer = BinWriter.as_bin_writer( @write_buf )
        end
    end

    private
    def write_value( val )
        writer.write_value( val )
    end

    private
    def get_write_buf
        ( @write_buf or raise "writer not initialized" ).string
    end

    private
    def read_check_res( rd )
        
        case rc = rd.read_int8
        when 0 then nil
        when 1 then raise "peer check failed: #{rd.read_utf8}"
        else raisef "unhandled response code: 0x%02x", rc
        end
    end

    # currently assumed to be the last part of a test, if called, and to only be
    # called with a non-empty @write_buf
    private
    def check_write_buf

        buf = get_write_buf # do this here, not on peer thread

        @peer_q << lambda do |io| 
            
            wr = BinaryWriter.new_le( :io => io )
            wr.write_utf8( @name )
            wr.write_buffer32( buf )
            
            rd = BinaryReader.new_le( :io => io )
            @ctx.complete { read_check_res( rd ) }
        end
    end
end

class InvalidDataTest < AbstractCoreIoTest

    bg_attr :error
end

class RoundtripTest < AbstractCoreIoTest

    def run_test
        
        expct = get_expect_value

        act = assert_matching_val( expct )

        write_value( act )
        check_write_buf
    end
end

class SequenceRoundtripTest < AbstractCoreIoTest
end

class CoreIoTests < BitGirderClass

    TC_END = 0x00
    TC_INVALID_DATA_TEST = 0x01
    TC_ROUNDTRIP_TEST = 0x02
    TC_SEQUENCE_ROUNDTRIP_TEST = 0x03

    include BitGirder::Testing

    include TestClassMixin

    private
    def read_file_header( br )
 
        unless ( i = br.read_int32 ) == 1
            raisef( "invalid file header: 0x%04x", i )
        end
    end

    private
    def read_name_buf_pair( br )
        { :name => br.read_utf8, :data => br.read_buffer32 }
    end

    private
    def read_invalid_data_test( br )
        
        InvalidDataTest.new(
            :name => br.read_utf8,
            :error => br.read_utf8,
            :data => br.read_buffer32
        )
    end

    private
    def read_roundtrip_test( br )
        RoundtripTest.new( read_name_buf_pair( br ) )
    end

    private
    def read_sequence_roundtrip_test( br )
        SequenceRoundtripTest.new( read_name_buf_pair( br ) )
    end

    private
    def add_test( t, res )
        res[ t.name ] = lambda { |ctx| 

            t.ctx = ctx
            t.peer_q = @peer_q

            t.run_test
        }
    end

    private
    def read_next_test( res, br )
        
        test = case tc = br.read_int8
        when TC_END then nil
        when TC_INVALID_DATA_TEST then read_invalid_data_test( br )
        when TC_ROUNDTRIP_TEST then read_roundtrip_test( br )
        when TC_SEQUENCE_ROUNDTRIP_TEST then read_sequence_roundtrip_test( br )
        else raisef( "unhandled type code: 0x%02x", tc )
        end

        return false unless test

        add_test( test, res )
        true
    end

    private
    def read_tests
 
        res = {}

        File.open( Testing.find_test_data( "core-io-tests.bin" ) ) do |io|
            
            br = BinaryReader.new_le( io: io )
            
            read_file_header( br )
            nil while read_next_test( res, br )
        end

        res
    end

    invocation_factory :read_tests

    private
    def add_expect_vals_with_prefix( pref, h )
        
        h.each_pair do |k, v| 
            AbstractCoreIoTest::EXPCT_VALS[ "#{pref}/#{k}" ] = v
        end
    end

    private
    def add_roundtrip_expect_vals
        
        add_expect_vals_with_prefix( "roundtrip", {

            "null-val" => MingleNull::INSTANCE,

            "string-empty" => MingleString.new( "" ),
            "string-val1" => MingleString.new( "hello" ),

            "bool-true" => MingleBoolean::TRUE,
            "bool-false" => MingleBoolean::FALSE,

            "buffer-empty" => MingleBuffer.new( "", :in_place ),
            "buffer-nonempty" => MingleBuffer.new( "\x00\x01", :in_place ),

            "int32-min" => MingleInt32.new( -( 2 ** 31 ) ),
            "int32-max" => MingleInt32.new( ( 2 ** 31 ) - 1 ),
            "int32-pos1" => MingleInt32.new( 1 ),
            "int32-zero" => MingleInt32.new( 0 ),
            "int32-neg1" => MingleInt32.new( -1 ),

            "int64-min" => MingleInt64.new( -( 2 ** 63 ) ),
            "int64-max" => MingleInt64.new( ( 2 ** 63 ) - 1 ),
            "int64-pos1" => MingleInt64.new( 1 ),
            "int64-zero" => MingleInt64.new( 0 ),
            "int64-neg1" => MingleInt64.new( -1 ),

            "uint32-max" => MingleUint32.new( ( 2 ** 32 ) - 1 ),
            "uint32-min" => MingleUint32.new( 0 ),
            "uint32-pos1" => MingleUint32.new( 1 ),

            "uint64-max" => MingleUint64.new( ( 2 ** 64 ) - 1 ),
            "uint64-min" => MingleUint64.new( 0 ),
            "uint64-pos1" => MingleUint64.new( 1 ),

#    b.setVal( "float32-val1", Float32( float32( 1 ) ) )
#    b.setVal( "float32-max", Float32( math.MaxFloat32 ) )
#    b.setVal( "float32-smallest-nonzero",
#        Float32( math.SmallestNonzeroFloat32 ) )
#    b.setVal( "float64-val1", Float64( float64( 1 ) ) )
#    b.setVal( "float64-max", Float64( math.MaxFloat64 ) )
#    b.setVal( "float64-smallest-nonzero",
#        Float64( math.SmallestNonzeroFloat64 ) )
            
#            "time-val1" =>
#                MingleTimestamp.rfc3339( "2013-10-19T02:47:00-08:00" ),

#    b.setVal( "enum-val1", MustEnum( "ns1@v1/E1", "val1" ) )

            "symmap-empty" => MingleSymbolMap::EMPTY,

            "symmap-flat" => MingleSymbolMap.create(
                "k1" => MingleInt32.new( 1 ),
                "k2" => MingleInt32.new( 2 )
            ),

            "symmap-nested" => MingleSymbolMap.create(
                "k1" => MingleSymbolMap.create( "kk1" => MingleInt32.new( 1 ) )
            ),

#    b.setVal( "struct-empty", MustStruct( "ns1@v1/T1" ) )
#    b.setVal( "struct-flat", MustStruct( "ns1@v1/T1", "k1", int32( 1 ) ) )

            "list-empty" => MingleList.as_list(),

            "list-scalars" => MingleList.as_list( 
                MingleInt32.new( 1 ), MingleString.new( "hello" ) ),
            
            "list-nested" => MingleList.as_list(
                MingleInt32.new( 1 ),
                MingleList.as_list(),
                MingleList.as_list( MingleString.new( "hello" ) ),
                MingleNull::INSTANCE
            )
        })
    end

    private
    def add_expect_vals
        add_roundtrip_expect_vals
    end

    define_before :add_expect_vals

    private
    def start_checker( cmd )
        
        pb = UnixProcessBuilder.new( :cmd => cmd )
        pb.popen( "r+" )
    end

    private
    def stop_checker
        debug_kill( "KILL", @checker.pid, { :name => "checker" } )
    end

    # run loop for peer manager
    private
    def manage_peer_q( cmd )

        while ( elt = @peer_q.pop ) != self
            case elt 
            when Proc 
                @checker ||= start_checker( cmd ) 
                elt.call( @checker )
            else raise "unhandled queue element: #{elt.class}"
            end
        end

        stop_checker if @checker
    end

    private
    def start_peer_manager
        
        @peer_q = Queue.new

        cmd = Testing.find_test_command( "check-core-io" )
        @peer_mgr = Thread.new { manage_peer_q( cmd ) }
    end

    define_before :start_peer_manager

    private
    def stop_peer_manager
        
        return unless @peer_mgr

        @peer_q << self

        # If we are able to join retrieve value in case it was an exception
        @peer_mgr.value if @peer_mgr.join( 3 )
    end

    define_after :stop_peer_manager
end

end
