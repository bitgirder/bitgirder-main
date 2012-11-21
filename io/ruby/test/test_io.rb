require 'bitgirder/core'
require 'bitgirder/testing'
require 'bitgirder/io'
require 'bitgirder/io/testing'

module BitGirder
module Io

include BitGirder::Core
include BitGirder::Testing

module EncTestMethods

    def force_enc( str, enc )
        RubyVersions.when_19x( str ) { |s| s.force_encoding( enc ) }
    end

    def force_bin( str )
        force_enc( str, "binary" )
    end
end    

class IoTests < BitGirderClass
 
    include TestClassMixin

    def test_which
        
        ls_act = `which ls`.strip
        raise "Should have an ls!" if ls_act.empty?

        assert_equal( ls_act, Io.which( "ls" ) )
        
        non_exist = "a-command-we-expect-not-to-be-here"
        assert_equal( nil, Io.which( non_exist ) )
        
        expct_msg = 
            'Cannot find command "a-command-we-expect-not-to-be-here" ' \
            'in path'

        assert_raised( expct_msg, Exception ) { Io.which( non_exist, true ) }
    end

    RubyVersions.when_19x do
    
        def test_as_encoded
            
            utf8 = "abc".encode!( "utf-8" )
            bin = "abc".encode!( "binary" )
            ascii = "abc".encode!( "us-ascii" )
    
            [ utf8, bin, ascii ].each do |s| 
    
                s2 = Io.as_encoded( s, Encoding.find( "utf-8" ) )
                assert_equal( utf8, s2 )
    
                # s and s2 should reference same obj if and only if s was
                # already utf8 encoded
                assert_equal( s.equal?( utf8 ), s.equal?( s2 ) )
            end
        end
    end

    def test_slurp
 
        data = "x" * ( 2 << 16 )
 
        Io.open_tempfile do |tmp|

            tmp.print( data )
            tmp.flush

            # Check slurping given a path name
            assert_equal( data, Io.slurp( tmp.path ) )

            # Check slurping given an io
            tmp.seek( 0 )
            assert_equal( data, Io.slurp( tmp ) )
        end
    end

    def assert_load_and_dump( typ )
        
        obj = { "a" => "b", "c" => [ 1, true, nil ] }
        
        loader = :"load_#{typ}"
        dumper = :"dump_#{typ}"

        Io.open_tempfile do |tmp|

            ser = Io.send( dumper, obj, tmp.path )
            obj2 = Io.send( loader, tmp.path )
            
            assert_equal( obj, obj2 )
        end
    end

    # we use load_json and dump_json, but this also gives us implicit basic
    # coverage of to_json and parse_json
    def test_load_and_dump_json
        assert_load_and_dump( :json )        
    end

    def test_load_and_dump_yaml
        assert_load_and_dump( :yaml )
    end

    def test_strict64
        
        [ "", "abc", "x" * 1000 ].each do |str|
            
            enc = Io.strict_encode64( str )
            assert_nil( enc.index( "\n" ) )

            dec = Io.strict_decode64( enc )
            assert_equal( str, dec )
        end
    end

    def test_read_full
        
        s = "abcd"
        io = StringIO.new( s )

        s.bytesize.times do |i|
            
            io.pos = 0
            assert_equal( s[ 0, i ], Io.read_full( io, i ) )
        end

        # Also test use of pre-supplied buffer
        res = "\x00" * 3
        io.pos = 0
        assert_equal( res, res2 = Io.read_full( io, 3, res ) )
        assert( res.equal?( res2 ) )

        assert_raised( 'EOF after 4 bytes (wanted 5)', EOFError ) do 
            io.pos = 0
            Io.read_full( io, s.bytesize + 1 )
        end
    end

    def test_data_unit_as_instance
        
        # First check that we convert string to sym and are ignoring case
        assert_equal( DataUnit::BYTE, DataUnit.as_instance( "byte" ) )
        assert_equal( DataUnit::BYTE, DataUnit.as_instance( "ByTe" ) )
        assert_equal( DataUnit::BYTE, DataUnit.as_instance( :BYte ) )

        # Now get coverage of acceptable words
        {
            DataUnit::BYTE => %w{ b byte bytes },
            DataUnit::KILOBYTE => %w{ k kb kilobyte kilobytes },
            DataUnit::MEGABYTE => %w{ m mb megabyte megabytes },
            DataUnit::GIGABYTE => %w{ g gb gigabyte gigabytes },
            DataUnit::TERABYTE => %w{ t tb terabyte terabytes },
            DataUnit::PETABYTE => %w{ p pb petabyte petabytes }

        }.each_pair do |unit, strs|
            strs.each { |s| assert_equal( unit, DataUnit.as_instance( s ) ) }
        end
    end

    def test_unknown_unit
        assert_raised( 'Unknown data unit: blah', DataUnitError ) do
            DataUnit.as_instance( :blah )
        end
    end

    def test_data_size_bytes
        
        expct = 3 * ( 2 ** 50 ) # 3PB

        [
            { :size => expct, :unit => :byte },
            { :size => expct >> 10, :unit => :kilobyte },
            { :size => expct >> 20, :unit => :megabyte },
            { :size => expct >> 30, :unit => :gigabyte },
            { :size => expct >> 40, :unit => :terabyte },
            { :size => expct >> 50, :unit => :petabyte }

        ].each do |test|
            
            sz = DataSize.new( test )
            assert_equal( expct, sz.bytes )
        end
    end

    def test_data_size_equality
        
        bytes = 3 * ( 2 ** 50 )

        sizes = [
            DataSize.new( :size => bytes, :unit => :b ),
            DataSize.new( :size => bytes >> 10, :unit => :kb ),
            DataSize.new( :size => bytes >> 20, :unit => :mb ),
            DataSize.new( :size => bytes >> 30, :unit => :gb ),
            DataSize.new( :size => bytes >> 40, :unit => :tb ),
            DataSize.new( :size => bytes >> 50, :unit => :pb )
        ]

        sizes.each do |s1| 
            sizes.each do |s2| 
                assert_equal( s1, s2 )
                assert_false( s1 == DataSize.as_instance( 1 ) )
            end
        end
    end

    def test_data_size_comparable
        
        d1A = DataSize.as_instance( "1024b" )
        d1B = DataSize.as_instance( "1k" )
        d2 = DataSize.as_instance( "1m" )
        d3 = DataSize.as_instance( "500b" )

        assert_false( d1A < d1B ) # check that comparable is indeed included

        # Now just check our <=> implementation
        assert_equal( 0, d1A <=> d1B )
        assert_equal( 0, d1B <=> d1A )
        assert_equal( 1, d1A <=> d3 )
        assert_equal( -1, d1A <=> d2 )
        assert_equal( nil, d1A <=> Object.new )
    end

    def test_data_size_from_int

        assert_equal( 
            DataSize.new( :unit => :byte, :size => 12 ),
            DataSize.as_instance( 12 )
        )
    end

    # We don't exhaustively test all forms of units, since that is checked in
    # test_data_unit_of(); we only check aspects here particular to the parsing
    # algorithm. 
    def test_data_size_from_string
        
        expct = 3 * ( 2 ** 50 )

        [
            "#{expct}",
            "#{expct}b",
            "#{expct >> 10} kb", # ws between num and unit
            " #{expct >> 20} mb ", # ws everywhere
            "#{expct >> 30}g", "#{expct >> 40}t", "#{expct >> 50}p" # the rest
        
        ].each do |s|

            assert_equal( 
                DataSize.new( :size => expct, :unit => :byte ),
                DataSize.as_instance( s )
            )
        end
    end

    def test_data_size_parse_failures

        [ "", "b", "0x03b", "10 kb mb", "-1k" ].each do |s|
            msg = "Invalid data size: #{s.inspect}"
            assert_raised( msg, DataSizeError ) { DataSize.as_instance( s ) }
        end

        assert_raised( "Unknown data unit: blah", DataUnitError ) do
            DataSize.as_instance( "1blah" )
        end
    end

    # Make sure we handle fsize properly across ruby versions and object types
    def test_fsize_compat
 
        len = 10

        Io.open_tempfile do |tmp|

            tmp.write( "\x00" * len )
            tmp.flush

            File.open( tmp.path ) do |f|
                [ tmp, tmp.path, f ].each do |obj| 
                    assert_equal( len, Io.fsize( obj ) )
                end
            end
        end
    end
end

# Used to check that we are correctly converting everything to String before
# invoking things like Process.system
class StringType < BitGirderClass
    
    bg_attr :str

    def to_s
        @str
    end
end

if RubyVersions.is_19x? # Don't even test unless we're >= 1.9

class UnixProcessBuilderTests < BitGirderClass
    
    include TestClassMixin

    def test_system_success
        UnixProcessBuilder.new( :cmd => "true" ).system
    end

    def test_system_failure
        
        begin
            UnixProcessBuilder.new( :cmd => "false" ).system
            raise "System didn't fail"
        rescue RuntimeError => e
            raise e unless e.message == "Command exited with status 1"
        end
    end

    def test_popen_basic
        
        b = UnixProcessBuilder.new( 
            :cmd => "/bin/bash", 
            :argv => [ "-c", "cat" ] 
        )

        res = b.popen( "r+" ) do |io|
            io.puts "hello"
            io.gets.chomp
        end

        assert_equal( "hello", res )
    end

    def test_string_handling
        
        UnixProcessBuilder.new(
            :cmd => StringType.new( "ls" ),
            :argv => [ StringType.new( "/dev/null" ) ],
            :env => { StringType.new( "ENV_VAR1" ) => StringType.new( "VAL1" ) }
        ).
        system
    end
end

end # Conditional 1.9.x block

class BinaryConverterTests < BitGirderClass

    require 'bigdecimal'

    include TestClassMixin

    private
    def dump_hex( str )
        
        res = []
        str.each_byte { |b| res << sprintf( "%02x", b ) }

        res.join( " " )
    end

    private
    def assert_num_rt_res( num, num2, opts )

        if f = opts[ :check_num2 ]
            f.call( num2 )
        else
            assert_equal( num, num2 )
        end
    end

    def assert_num_rt_str( act, expct )
        
#        code( "str: #{dump_hex( str )}" )
        case expct
            when Regexp then assert( expct.match( act ) )
            else
                expct = RubyVersions.when_19x( expct ) do |s|
                    s.dup.force_encoding( "binary" )
                end
                assert_equal( expct, act )
        end
    end

    # 'rt' short for 'roundtrip'
    def assert_num_rt( opts )
        
        type = has_key( opts, :type )
        num = has_key( opts, :num )
        order = has_key( opts, :order )
        expct_str = has_key( opts, :expct_str ) 

        enc = BinaryConverter.new( :order => order )
        
        if plat_order = opts[ :plat_order ] 
            enc.instance_variable_set( :@plat_order, plat_order )
        end

        str = enc.send( :"write_#{type}", num )
        assert_num_rt_str( str, expct_str )

        num2 = enc.send( :"read_#{type}", str )
        assert_num_rt_res( num, num2, opts )
    end

    # ruby 1.9 has Float::INFINITY and Float::NAN, but we can't assume them but
    # compute them using some literal expressions
    INFINITY = 1.0 / 0.0 
    NAN = 0.0 / 0.0

    CHECK_NAN = lambda { |num| raise "Expected NAN" unless num.nan? }
 
    [
        [ :int8, :big_endian, 12, "\x0c" ],
        
        [ :int8, :little_endian, 12, "\x0c" ],

        [ :uint8, :big_endian, 12, "\x0c" ],
        
        [ :uint8, :little_endian, 12, "\x0c" ],

        [ :uint8, :big_endian, 254, "\xfe" ],
        
        [ :uint8, :little_endian, 254, "\xfe" ],

        [ :int32, :big_endian, 12, "\x00\x00\x00\x0c", :pos ],
        
        [ :int32, :little_endian, 12, "\x0c\x00\x00\x00", :pos ],
        
        [ :int32, :big_endian, 0, "\x00\x00\x00\x00", :zero ],
        
        [ :int32, :little_endian, 0, "\x00\x00\x00\x00", :zero ],
        
        [ :int32, :big_endian, -12, "\xff\xff\xff\xf4", :neg ],
 
        [ :int32, :little_endian, -12, "\xf4\xff\xff\xff", :neg ],

        [ :uint32, :big_endian, 12, "\x00\x00\x00\x0c", :pos ],
        
        [ :uint32, :little_endian, 12, "\x0c\x00\x00\x00", :pos ],
        
        [ :uint32, :big_endian, 0, "\x00\x00\x00\x00", :zero ],
        
        [ :uint32, :little_endian, 0, "\x00\x00\x00\x00", :zero ],
        
        [ :uint32, :big_endian, 4294967284, "\xff\xff\xff\xf4", :neg ],
 
        [ :uint32, :little_endian, 4294967284, "\xf4\xff\xff\xff", :neg ],
        
        [ :int64, :big_endian, 12, ( "\x00" * 7 ) + "\x0c", :pos ],
        
        [ :int64, :little_endian, 12, "\x0c" + ( "\x00" * 7 ), :pos ],
        
        [ :int64, :big_endian, 0, "\x00" * 8, :zero ],
        
        [ :int64, :little_endian, 0, "\x00" * 8, :zero ],
        
        [ :int64, :big_endian, -12, ( "\xff" * 7 ) + "\xf4", :neg ],
        
        [ :int64, :little_endian, -12, "\xf4" + ( "\xff" * 7 ), :neg ],
        
        [ :uint64, :big_endian, 12, ( "\x00" * 7 ) + "\x0c", :pos ],
        
        [ :uint64, :little_endian, 12, "\x0c" + ( "\x00" * 7 ), :pos ],
        
        [ :uint64, :big_endian, 0, "\x00" * 8, :zero ],
        
        [ :uint64, :little_endian, 0, "\x00" * 8, :zero ],
        
        [ :uint64, :big_endian, 18446744073709551604, 
            ( "\xff" * 7 ) + "\xf4", :neg ],
        
        [ :uint64, :little_endian, 18446744073709551604, 
            "\xf4" + ( "\xff" * 7 ), :neg ],
        
        [ :float32, :big_endian, INFINITY, "\x7f\x80\x00\x00", 
            :pos_inf ],
        
        [ :float32, :little_endian, INFINITY, "\x00\x00\x80\x7f", 
            :pos_inf ],
        
        [ :float32, :big_endian, 1.0, "\x3f\x80\x00\x00", :pos1 ],
        
        [ :float32, :little_endian, 1.0, "\x00\x00\x80\x3f", :pos1 ],
        
        [ :float32, :big_endian, 0.0, "\x00" * 4, :pos_zero ],
        
        [ :float32, :little_endian, 0.0, "\x00" * 4, :pos_zero ],
        
        [ :float32, :big_endian, -0.0, "\x80" + ( "\x00" * 3 ), :neg_zero ],
        
        [ :float32, :little_endian, -0.0, ( "\x00" * 3 ) + "\x80", :neg_zero ],
        
        [ :float32, :big_endian, -1.0, "\xbf\x80\x00\x00", :neg1 ],
        
        [ :float32, :little_endian, -1.0, "\x00\x00\x80\xbf", :neg1 ],

        [ :float32, :big_endian, -INFINITY, "\xff\x80\x00\x00", 
            :neg_inf ],
        
        [ :float32, :little_endian, -INFINITY, "\x00\x00\x80\xff", 
            :neg_inf ],
        
        [ :float32, :big_endian, NAN, /^(\xff|\x7f)\xc0\x00{2}$/, :nan, 
            CHECK_NAN ],
        
        [ :float32, :little_endian, NAN, /^\x00{2}\xc0(\xff|\x7f)$/, :nan, 
            CHECK_NAN ],
        
        [ :float64, :big_endian, INFINITY, "\x7f\xf0" + ( "\x00" * 6 ),
            :pos_inf ],

        [ :float64, :little_endian, INFINITY, 
            ( "\x00" * 6 ) + "\xf0\x7f", :pos_inf ],

        [ :float64, :big_endian, 1.0, "\x3f\xf0" + ( "\x00" * 6 ), :pos1 ],
        
        [ :float64, :little_endian, 1.0, ( "\x00" * 6 ) + "\xf0\x3f", :pos1 ],
        
        [ :float64, :big_endian, 0.0, "\x00" * 8, :pos_zero ],
        
        [ :float64, :little_endian, 0.0, "\x00" * 8, :pos_zero ],
        
        [ :float64, :big_endian, -0.0, "\x80" + ( "\x00" * 7 ), :neg_zero ],
        
        [ :float64, :little_endian, -0.0, ( "\x00" * 7 ) + "\x80", :neg_zero ],
        
        [ :float64, :big_endian, -1.0, "\xbf\xf0" + ( "\x00" * 6 ), :neg1 ],
        
        [ :float64, :little_endian, -1.0, ( "\x00" * 6 ) + "\xf0\xbf", :neg1 ],

        [ :float64, :big_endian, -INFINITY,
            "\xff\xf0" + ( "\x00" * 6 ), :neg_inf ],
        
        [ :float64, :little_endian, -INFINITY,
            ( "\x00" * 6 ) + "\xf0\xff", :neg_inf ],
        
        [ :float64, :big_endian, NAN, /^(\x7f|\xff)\xf8\x00{6}$/, :nan,
            CHECK_NAN ],
        
        [ :float64, :little_endian, NAN, /^\x00{6}\xf8(\xff|\x7f)$/, :nan,
            CHECK_NAN ],
        
        [ :bignum, :big_endian, 2 ** 81, 
            "\x00\x00\x00\x0b" + "\x02" + ( "\x00" * 10 ), :pos1 ],
        
        [ :bignum, :little_endian, 2 ** 81,
            "\x0b\x00\x00\x00" + ( "\x00" * 10 ) + "\x02", :pos1 ],
        
        [ :bignum, :big_endian, 0, "\x00\x00\x00\x01\x00", :zero ],

        [ :bignum, :little_endian, 0, "\x01\x00\x00\x00\x00", :zero ],

        [ :bignum, :big_endian, -1, "\x00\x00\x00\x01\xff", :neg1 ],

        [ :bignum, :little_endian, -1, "\x01\x00\x00\x00\xff", :neg1 ],

        # In the bigdec tests, 'gt_one' is an abbreviation for 'having absolute
        # value greater than one', and similarly with 'lt' meaining 'less than'

        [ :bigdec, :big_endian, BigDecimal.new( "1.2345" ),
            "\x00\x00\x00\x02\x30\x39\xff\xff\xff\xfc", :pos_gt_one ],
        
        [ :bigdec, :little_endian, BigDecimal.new( "1.2345" ),
            "\x02\x00\x00\x00\x39\x30\xfc\xff\xff\xff", :pos_gt_one ],
        
        [ :bigdec, :big_endian, BigDecimal.new( "0.0012345" ),
            "\x00\x00\x00\x02\x30\x39\xff\xff\xff\xf9", :pos_lt_one ],

        [ :bigdec, :little_endian, BigDecimal.new( "0.0012345" ),
            "\x02\x00\x00\x00\x39\x30\xf9\xff\xff\xff", :pos_lt_one ],
        
        [ :bigdec, :big_endian, BigDecimal.new( "0" ),
            "\x00\x00\x00\x01\x00\x00\x00\x00\x00", :zero ],
        
        [ :bigdec, :little_endian, BigDecimal.new( "0" ),
            "\x01\x00\x00\x00\x00\x00\x00\x00\x00", :zero ],
        
        [ :bigdec, :big_endian, BigDecimal.new( "-1.2345" ),
            "\x00\x00\x00\x02\xcf\xc7\xff\xff\xff\xfc", :neg_gt_one ],
        
        [ :bigdec, :little_endian, BigDecimal.new( "-1.2345" ),
            "\x02\x00\x00\x00\xc7\xcf\xfc\xff\xff\xff", :neg_gt_one ],
        
        [ :bigdec, :big_endian, BigDecimal.new( "-0.0012345" ),
            "\x00\x00\x00\02\xcf\xc7\xff\xff\xff\xf9", :neg_lt_one ],

        [ :bigdec, :little_endian, BigDecimal.new( "-0.0012345" ),
            "\x02\x00\x00\x00\xc7\xcf\xf9\xff\xff\xff", :neg_lt_one ],
 
    ].
    each do |argv|
 
        type, order, num, expct_str, tail, check_num2 = *argv

        meth = "test_rt_#{type}_#{order}"
        meth += "_#{tail}" if tail

        define_method( meth.to_sym ) do

            assert_num_rt(
                :type => type,
                :order => order,
                :num => num,
                :expct_str => expct_str,
                :check_num2 => check_num2
            )
        end
    end

    private
    def assert_bignum_battery( i )
        
        [ :little_endian, :big_endian ].each do |order|
            
            enc = BinaryConverter.new( :order => order )

            buf = enc.write_bignum( i )
            i2 = enc.read_bignum( buf )

            assert_equal( i, i2 )
        end
    end

    def test_bignum_battery
        
        -257.upto( 256 ) { |i| assert_bignum_battery( i ) }
        
        # Randomly test some positives and negs of all bit-lengths up to n
        n = 512
        1.upto( n ) { |bit_len|
        10.times { 
            
            max = 2 ** bit_len
            sign = rand( 2 ) == 0 ? -1 : 1
            assert_bignum_battery( rand( max ) * sign )
        }}
    end

    def test_read_bignum_with_info
        
        [ :little_endian, :big_endian ].each do |order|
            
            enc = BinaryConverter.new( :order => order )

            i = 2 ** 100 # 13 byte 2's complement

            buf = enc.write_bignum( i )
            i2, info = enc.read_bignum_with_info( buf )

            assert_equal( i, i2 )
            assert_equal( 
                { :total_len => 17, :hdr_len => 4, :num_len => 13 }, info )
        end
    end

    private
    def assert_bigdec_battery( d )
        
        [ :little_endian, :big_endian ].each do |order|
            
            enc = BinaryConverter.new( :order => order )

            buf = enc.write_bigdec( d )
            d2 = enc.read_bigdec( buf )

            assert_equal( d, d2 )
        end
    end

    def test_bigdec_regress1
        
        d = BigDecimal.new( "100" )

        6.times do

            assert_bigdec_battery( d )
            assert_bigdec_battery( -d )

            d = d / 10
        end
    end

    def test_bigdec_battery
        
        1.upto( 200 ) { |dec_len|
        10.times { 
            
            str = ""
            dec_len.times { str << rand( 10 ).to_s }

            exp = rand( dec_len + 1 )
            if exp == 0
                str = "0.#{str}"
            elsif exp < dec_len
                str = str[ 0, exp ] + "." + str[ exp .. -1 ]
            end

            str = "-#{str}" if rand( 2 ) == 0
            assert_bigdec_battery( BigDecimal.new( str ) )
        }}
    end

    def test_read_bigdec_with_info
        
        [ :little_endian, :big_endian ].each do |order|
            
            enc = BinaryConverter.new( :order => order )

            d = BigDecimal.new( "1.2345" )

            buf = enc.write_bigdec( d )
            d2, info = enc.read_bigdec_with_info( buf )

            assert_equal( d, d2 )

            assert_equal( 
                { :total_len => 10, :scale_len => 4, :unscaled_len => 6 }, 
                info 
            )
        end
    end
end

class BinIoTests < BitGirderClass
    
    include TestClassMixin
    include EncTestMethods
    include BitGirder::Io::Testing

    def test_order_factories
        
        io = StringIO.new # dummy but non-nil

        [ BinaryReader, BinaryWriter ].each do |cls|
        
            {
                ORDER_LITTLE_ENDIAN => cls.new_le( :io => io ),
                ORDER_BIG_ENDIAN => cls.new_be( :io => io )
            }.
            each_pair do |ord, bin|
                assert_equal( ord, bin.order )
                assert_equal( ord, cls.new_with_order( ord, :io => io ).order )
            end
        end
    end

    def assert_read_write( ord, io )

        calls = [
            [ :int8, 1, 1 ],
            [ :int32, 1, 5 ],
            [ :int64, 1, 13 ],
            [ :float32, 1.0, 17 ],
            [ :float64, 1.0, 25 ],
            [ :bool, true, 26 ],
            [ :bool, false, 27 ],
            [ :utf8, "hello", 36 ],
            [ :utf8, "", 40 ],
            [ :utf8, force_enc( "\xc7\x93", "utf-8" ), 46 ],
            [ :buffer32, force_bin( "" ), 50 ],
            [ :buffer32, force_bin( "\x00\x01\x02" ), 57 ],
            [ :full, force_bin( "\x00\x01\02" ), 60 ],
            [ :$base_io, force_bin( "\x03\x04\x05" ), 63 ],
            [ :uint8, 1, 64 ],
            [ :uint8, ( 2 ** 8 ) - 1, 65 ],
            [ :uint32, 1, 69 ],
            [ :uint32, ( 2 ** 32 ) - 1, 73 ],
            [ :uint64, 1, 81 ],
            [ :uint64, ( 2 ** 64 ) - 1, 89 ]
        ]

        wr = BinaryWriter.new( :order => ord, :io => io )
        calls.each do |call| 
            meth = call[ 0 ] == :$base_io ? :write : :"write_#{call[ 0 ]}"
            wr.send( meth, call[ 1 ] )
            assert_equal( call[ 2 ], wr.pos )
        end

        io.seek( 0, IO::SEEK_SET )

        rd = BinaryReader.new( :order => ord, :io => io )
        calls.each do |call| 
            args = []
            meth = call[ 0 ] == :$base_io ? :read : :"read_#{call[ 0 ]}"
            args << call[ 1 ].size if meth == :read_full || meth == :read
            rd_val = rd.send( meth, *args )
            assert_equal( call[ 1 ], rd_val )
            assert_equal( call[ 2 ], rd.pos )
        end
    end

    def test_read_write
        
        [ ORDER_LITTLE_ENDIAN, ORDER_BIG_ENDIAN ].each do |ord|
            Io.open_tempfile { |io| assert_read_write( ord, io ) }
            assert_read_write( ord, StringIO.new )
        end
    end

    # Put in as regression against a bug in which read_utf8 returned a string
    # with encodin binary under some ruby versions
    def test_read_utf8_sets_encoding
        
        io = new_string_io

        wr = BinaryWriter.new( :order => ORDER_LITTLE_ENDIAN, :io => io )
        wr.write_utf8( "hello" )

        io.seek( 0, IO::SEEK_SET )

        rd = BinaryReader.new( :order => ORDER_LITTLE_ENDIAN, :io => io )
        str = rd.read_utf8
        RubyVersions.when_19x { assert_equal( Encoding::UTF_8, str.encoding ) }
        assert_equal( "hello", str )
    end

    def test_close
        
        io = Object.new
        io.instance_variable_set( :@calls, 0 )

        f = lambda {
            opts = { :io => io, :order => ORDER_LITTLE_ENDIAN }
            BinaryReader.new( opts ).close
            BinaryWriter.new( opts ).close
        }

        # io does not respond to close(); close() in f should no-op
        f.call 
        assert_equal( 0, io.instance_variable_get( :@calls ) )

        class <<io; def close; @calls += 1 end; end

        # now io has close(); we expect 2 calls
        f.call
        assert_equal( 2, io.instance_variable_get( :@calls ) )
    end

    def test_peek
        
        io = StringIO.new

        io << "\x00\x01\x02"
        io.pos = 0
        rd = BinaryReader.new( :io => io, :order => ORDER_LITTLE_ENDIAN )

        assert_equal( 0, rd.read_int8 )
        assert_equal( 1, rd.pos )
        assert_equal( "\x01", rd.peekc )
        assert_equal( 1, rd.pos )
        assert_equal( 1, rd.peek_int8 )
        assert_equal( 1, rd.pos )
        assert_equal( "\x01\x02", rd.read_full( 2 ) )
        assert_equal( 3, rd.pos )
    end
end

end
end
