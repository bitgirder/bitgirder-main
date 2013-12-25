require 'mingle'

require 'mingle/test-support'

require 'bitgirder/io'
include BitGirder::Io

require 'bitgirder/testing'

require 'bitgirder/core/testing'

require 'thread'
require 'stringio'

module Mingle

class AbstractCoreIoTest < BitGirderClass

    EXPCT_VALS = {}

    include BitGirder::Testing::AssertMethods

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
        when ObjectPath then reader.read_identifier_path
        when MingleIdentifier then reader.read_identifier
        when MingleNamespace then reader.read_namespace
        when QualifiedTypeName then reader.read_qualified_type_name
        when DeclaredTypeName then reader.read_declared_type_name
        when MingleTypeReference then reader.read_type_reference
        else raise "Unhandled expect value: #{expct.class}"
        end
    end

    private
    def assert_read_matches( expct, act )
        
        case expct
        when MingleValue then ModelTestInstances.assert_equal( expct, act )
        when ObjectPath 
            ObjectPathTestMethods.assert_equal_with_format( expct, act )
        when MingleIdentifier, MingleNamespace, QualifiedTypeName, 
             DeclaredTypeName, MingleTypeReference
            assert_equal( expct, act )
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

        @writer
    end

    private
    def write_value( val )

        case val
        when MingleValue then writer.write_value( val )
        when ObjectPath then writer.write_identifier_path( val )
        when MingleIdentifier then writer.write_identifier( val )
        when MingleNamespace then writer.write_namespace( val )
        when QualifiedTypeName then writer.write_qualified_type_name( val )
        when DeclaredTypeName then writer.write_declared_type_name( val )
        when MingleTypeReference then writer.write_type_reference( val )
        else raise "unhandled write val: #{val.class}"
        end
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

    def run_test
        
        err = assert_raised( BinIoError ) { reader.read_value }
        assert_equal( @error, err.to_s )
        @ctx.complete
    end
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

    def run_test
        
        get_expect_value.each do |expct|
            
            act = assert_matching_val( expct )
            write_value( act )
        end

        assert( reader.rd.eof? )

        check_write_buf
    end
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
    def add_expect_val_with_prefix( pref, name, val )
        AbstractCoreIoTest::EXPCT_VALS[ "#{pref}/#{name}" ] = val
    end

    private
    def add_expect_vals_with_prefix( pref, h )
        h.each_pair { |k, v| add_expect_val_with_prefix( pref, k, v ) }
    end

    private
    def add_value_roundtrip_expect_vals

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

            "float32-val1" => MingleFloat32.new( 1.0 ),

            "float32-max" => 
                MingleFloat32.new( "\xFF\xFF\x7F\x7F".unpack( 'e' ).shift ),

            "float32-smallest-nonzero" =>
                MingleFloat32.new( "\x01\x00\x00\x00".unpack( 'e' ).shift ),
            
            "float64-val1" => MingleFloat64.new( 1.0 ),
            
            "float64-max" =>
                MingleFloat64.new( 
                    "\xff\xff\xff\xff\xff\xff\xef\x7f".unpack( 'E' ).shift ),
            
            "float64-smallest-nonzero" =>
                MingleFloat64.new(
                    "\x01\x00\x00\x00\x00\x00\x00\x00".unpack( 'E' ).shift ),

            "time-val1" => 
                MingleTimestamp.rfc3339( "2013-10-19T02:47:00-08:00" ),

            "enum-val1" => 
                MingleEnum.new( :type => :"ns1@v1/E1", :value => :val1 ),

            "symmap-empty" => MingleSymbolMap::EMPTY,

            "symmap-flat" => MingleSymbolMap.create(
                "k1" => MingleInt32.new( 1 ),
                "k2" => MingleInt32.new( 2 )
            ),

            "symmap-nested" => MingleSymbolMap.create(
                "k1" => MingleSymbolMap.create( "kk1" => MingleInt32.new( 1 ) )
            ),

            "struct-empty" => MingleStruct.new( :type => :"ns1@v1/T1" ),

            "struct-flat" =>
                MingleStruct.new( 
                    :type => :"ns1@v1/T1",
                    :fields => { :k1 => MingleInt32.new( 1 ) }
                ),

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
    def add_id_path_roundtrip_expect_vals

        id = lambda { |idx| MingleIdentifier.get( "id#{idx}" ) }

        add = lambda do |nm, path|
            add_expect_val_with_prefix( "roundtrip", nm, path )
            path
        end

        p = add.call( "p1", ObjectPath.get_root( id.call( 1 ) ) )
        p = add.call( "p2", p.descend( id.call( 2 ) ) )
        p = add.call( "p3", p.start_list.tap { |lp| lp.index = 2 } )
        p = add.call( "p4", p.descend( id.call( 3 ) ) )
        add.call( "p5", ObjectPath.get_root_list.descend( id.call( 1 ) ) )
    end

    private
    def add_definition_roundtrip_expect_vals
        
        add_expect_vals_with_prefix( "roundtrip", {

            "Identifier/id1" => MingleIdentifier.get( "id1" ),
            "Identifier/id1-id2" => MingleIdentifier.get( "id1-id2" ),

            "Namespace/ns1@v1" => MingleNamespace.get( "ns1@v1" ),
            "Namespace/ns1:ns2@v1" => MingleNamespace.get( "ns1:ns2@v1" ),

            "DeclaredTypeName/T1" => DeclaredTypeName.get( "T1" ),

            "QualifiedTypeName/ns1:ns2@v1/T1" => 
                QualifiedTypeName.get( "ns1:ns2@v1/T1" ),

            "AtomicTypeReference/T1" => MingleTypeReference.get( "T1" ),

            "AtomicTypeReference/mingle:core@v1/String~\"a\"" =>
                MingleTypeReference.get( "String~\"a\"" ),

            "AtomicTypeReference/mingle:core@v1/String~[\"a\",\"b\"]" =>
                MingleTypeReference.get( "String~[\"a\",\"b\"]" ),

            "AtomicTypeReference/mingle:core@v1/Timestamp~[\"2012-01-01T00:00:00Z\",\"2012-02-01T00:00:00Z\"]" =>
                MingleTypeReference.get( "Timestamp~[\"2012-01-01T00:00:00Z\",\"2012-02-01T00:00:00Z\"]" ),

            "AtomicTypeReference/mingle:core@v1/Int32~(0,10)" =>
                MingleTypeReference.get( "Int32~(0,10)" ),

            "AtomicTypeReference/mingle:core@v1/Int64~[0,10]" =>
                MingleTypeReference.get( "Int64~[0,10]" ),

            "AtomicTypeReference/mingle:core@v1/Uint32~(0,10)" =>
                MingleTypeReference.get( "Uint32~(0,10)" ),

            "AtomicTypeReference/mingle:core@v1/Uint64~[0,10]" =>
                MingleTypeReference.get( "Uint64~[0,10]" ),

            "AtomicTypeReference/mingle:core@v1/Float32~(0,1]" =>
                MingleTypeReference.get( "Float32~(0.0,1.0]" ),

            "AtomicTypeReference/mingle:core@v1/Float64~[0,1)" =>
                MingleTypeReference.get( "Float64~[0.0,1.0)" ),

            "AtomicTypeReference/mingle:core@v1/Float64~(,)" =>
                MingleTypeReference.get( "Float64~(,)" ),

            "ListTypeReference/T1*" => MingleTypeReference.get( "T1*" ),
            "ListTypeReference/T1+" => MingleTypeReference.get( "T1+" ),

            "NullableTypeReference/T1*?" => MingleTypeReference.get( "T1*?" ),

            "AtomicTypeReference/ns1@v1/T1" =>
                MingleTypeReference.get( "ns1@v1/T1" ),

            "ListTypeReference/ns1@v1/T1*" =>
                MingleTypeReference.get( "ns1@v1/T1*" ),

            "NullableTypeReference/ns1@v1/T1?" =>
                MingleTypeReference.get( "ns1@v1/T1?" )
        })
    end

    private
    def add_roundtrip_expect_vals
        
        add_value_roundtrip_expect_vals
        add_id_path_roundtrip_expect_vals
        add_definition_roundtrip_expect_vals
    end

    private
    def add_sequence_expect_vals

        add_expect_val_with_prefix( 
            "sequence-roundtrip", 
            "struct-sequence",
            [ 
                MingleStruct.new( :type => :"ns1@v1/S1" ),

                MingleStruct.new( 
                    :type => :"ns1@v1/S1", 
                    :fields => { :f1 => MingleInt32.new( 1 ) }
                )
            ]
        )
    end

    private
    def add_expect_vals
        add_roundtrip_expect_vals
        add_sequence_expect_vals
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
