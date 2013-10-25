require 'bitgirder/core'
require 'bitgirder/io'
require 'mingle'
require 'mingle/codec'

module Mingle
module BinCodec

TYPE_CODE_NULL = 0x00
TYPE_CODE_BOOLEAN = 0x0a
TYPE_CODE_STRING = 0x0b
TYPE_CODE_INT32 = 0x0c
TYPE_CODE_INT64 = 0x0d
TYPE_CODE_UINT32 = 0x0e
TYPE_CODE_UINT64 = 0x0f
TYPE_CODE_FLOAT32 = 0x10
TYPE_CODE_FLOAT64 = 0x11
TYPE_CODE_TIMESTAMP = 0x12
TYPE_CODE_BUFFER = 0x13
TYPE_CODE_ENUM = 0x14
TYPE_CODE_SYMBOL_MAP = 0x15
TYPE_CODE_FIELD = 0x16
TYPE_CODE_STRUCT = 0x17
TYPE_CODE_LIST = 0x19
TYPE_CODE_END = 0x1a

Io = BitGirder::Io

class MingleBinCodec < BitGirder::Core::BitGirderClass
    
    include BitGirder::Core
    include Mingle
    include Mingle::Codec::MingleCodecImpl

    BYTE_ORDER = Io::ORDER_LITTLE_ENDIAN

    def initialize
        @conv = Io::BinaryConverter.new( :order => BYTE_ORDER )
    end

    private
    def append_type_code( wr, code )
        wr.write_int8( code )
    end

    private
    def append_sized_buffer( wr, buf )
        wr.write_buffer32( buf )
    end

    private
    def append_type_reference( wr, typ )
        BinWriter.as_bin_writer( wr ).write_type_reference( typ )
    end

    private
    def append_identifier( wr, id )
        BinWriter.as_bin_writer( wr ).write_identifier( id )
    end

    private
    def append_boolean( wr, val )

        append_type_code( wr, TYPE_CODE_BOOLEAN )
        wr.write_bool( val.to_bool )
    end

    private
    def append_string( wr, str )
        
        append_type_code( wr, TYPE_CODE_STRING )
        wr.write_utf8( str.to_s )
    end

    private
    def append_num( wr, mg_num, type_code, enc_meth )
 
        append_type_code( wr, type_code )
        wr.send( enc_meth, mg_num.num )
    end

    private
    def append_int64( wr, val )
        append_num( wr, val, TYPE_CODE_INT64, :write_int64 )
    end

    private
    def append_int32( wr, val )
        append_num( wr, val, TYPE_CODE_INT32, :write_int32 )
    end

    private
    def append_uint32( wr, val )
        append_num( wr, val, TYPE_CODE_UINT32, :write_uint32 )
    end

    private
    def append_uint64( wr, val )
        append_num( wr, val, TYPE_CODE_UINT64, :write_uint64 )
    end

    private
    def append_float64( wr, val )
        append_num( wr, val, TYPE_CODE_FLOAT64, :write_float64 )
    end

    private
    def append_float32( wr, val )
        append_num( wr, val, TYPE_CODE_FLOAT32, :write_float32 )
    end

    private
    def append_buffer( wr, val )
        
        append_type_code( wr, TYPE_CODE_BUFFER )
        append_sized_buffer( wr, val.buf )
    end

    private
    def append_enum( wr, val )

        append_type_code( wr, TYPE_CODE_ENUM )
        append_type_reference( wr, val.type )
        append_identifier( wr, val.value )
    end

    private
    def append_timestamp( wr, val )
        
        append_type_code( wr, TYPE_CODE_TIMESTAMP )
        wr.write_int64( val.time.to_i )
        wr.write_int32( val.time.nsec )
    end

    private
    def append_symbol_map( wr, val )
        
        append_type_code( wr, TYPE_CODE_SYMBOL_MAP )
        append_fields( wr, val )
    end

    private
    def append_list( wr, val )
        
        append_type_code( wr, TYPE_CODE_LIST )
        wr.write_int32( -1 )
        val.each { |elt| append_value( wr, elt ) }
        append_type_code( wr, TYPE_CODE_END )
    end

    private
    def append_value( wr, val )
        
        case val

            when MingleBoolean then append_boolean( wr, val )
            when MingleString then append_string( wr, val )
            when MingleInt32 then append_int32( wr, val )
            when MingleInt64 then append_int64( wr, val )
            when MingleUint32 then append_uint32( wr, val )
            when MingleUint64 then append_uint64( wr, val )
            when MingleFloat32 then append_float32( wr, val )
            when MingleFloat64 then append_float64( wr, val )
            when MingleBuffer then append_buffer( wr, val )
            when MingleEnum then append_enum( wr, val )
            when MingleTimestamp then append_timestamp( wr, val )
            when MingleStruct then append_struct( wr, val )
            when MingleSymbolMap then append_symbol_map( wr, val )
            when MingleList then append_list( wr, val )
            when MingleNull then append_type_code( wr, TYPE_CODE_NULL )

            else raise "Unhandled value type: #{val.class}"
        end
    end

    private
    def append_fields( wr, flds )
        
        flds.each_pair do |fld, val|
            
            unless val.is_a?( MingleNull )
                append_type_code( wr, TYPE_CODE_FIELD )
                append_identifier( wr, fld )
                append_value( wr, val )
            end
        end

        append_type_code( wr, TYPE_CODE_END )
    end

    private
    def append_struct( wr, ms )
        
        append_type_code( wr, TYPE_CODE_STRUCT )
        wr.write_int32( -1 )
        append_type_reference( wr, ms.type )
        append_fields( wr, ms.fields )
    end

    public
    def as_buffer( obj )
        
        not_nil( obj, :obj )
        obj.is_a?( MingleStruct ) or codec_raise( "Not a struct: #{obj}" )

        buf = RubyVersions.when_19x( StringIO.new ) do |io|
            io.set_encoding( "binary" )
        end

        wr = Io::BinaryWriter.new( :order => BYTE_ORDER, :io => buf )

        append_struct( wr, obj )

        buf.string
    end

    private
    def to_hex_byte( i )
        sprintf( "0x%02x", i % 256 )
    end

    # Useful when reporting errors; returns the index just before pos, which was
    # presumably related to the error
    private
    def last_pos( scanner )
        scanner.pos - 1
    end

    private
    def cur_pos( scanner )
        scanner.pos
    end

    private
    def decode_raise( pos_obj, msg )
        
        off = case pos_obj
            when Fixnum then pos_obj
            when Io::BinaryReader then last_pos( pos_obj )
            else raise "Unexpected pos_obj of type #{pos_obj.class}"
        end

        codec_raise( "[offset #{off}]: #{msg}" )
    end

    private
    def raise_unrecognized_value_code( tc, pos )
        decode_raise( pos, sprintf( "Unrecognized value code: 0x%02x", tc ) )
    end

    private
    def type_code_expect_raise( code_sym, code_act, pos )
        raise_unrecognized_value_code( code_act, pos )
    end

    private
    def expect_type_code( scanner, code_sym )

        code_val = Mingle::BinCodec.const_get( code_sym )
        
        if ( b = scanner.read_int8 ) == code_val
            code_val
        else
            type_code_expect_raise( code_sym, b, last_pos( scanner ) )
        end
    end

    private
    def expect_type_code_end( scanner )
        expect_type_code( scanner, :TYPE_CODE_END )
    end

    private
    def bin_reader_result( scanner )
        
        off = cur_pos( scanner )

        begin
            yield( BinReader.as_bin_reader( scanner ) )
        rescue => err
            decode_raise( off, err )
        end
    end

    private
    def read_type_reference( scanner )

        bin_reader_result( scanner ) do |br| 
            qn = br.read_qualified_type_name 
            AtomicTypeReference.create( name: qn )
        end 
    end

    private
    def read_identifier( scanner )
        bin_reader_result( scanner ) { |br| br.read_identifier }
    end

    private
    def read_field( scanner, flds )
 
        id = read_identifier( scanner )
        fld = read_value( scanner )

        flds[ id ] = fld
    end

    private
    def read_fields( scanner )

        flds = {}

        while ( tc = scanner.read_int8 ) != TYPE_CODE_END
            if tc == TYPE_CODE_FIELD
                read_field( scanner, flds )
            else
                type_code_expect_raise( 
                    :TYPE_CODE_FIELD, tc, cur_pos( scanner ) )
            end
        end

        MingleSymbolMap.create( flds )
    end

    private
    def read_mg_boolean( scanner )
        
        case b = scanner.read_int8
            when 0x00 then MingleBoolean::FALSE
            when 0x01 then MingleBoolean::TRUE
            else raise "Unexpected bool val: #{to_hex_byte( b )}"
        end
    end

    private
    def read_mg_string( scanner )
        MingleString.new( scanner.read_utf8 )
    end

    private
    def read_mg_struct( scanner )
 
        code( "reading struct, cur pos: #{cur_pos( scanner )}" )
        sz = scanner.read_int32
        code( "sz is #{sz}, cur_pos: #{cur_pos( scanner )}" )
        typ = read_type_reference( scanner )
        flds = read_fields( scanner )

        MingleStruct.new( :type => typ, :fields => flds )
    end

    [
        [ :int64, MingleInt64, :read_int64 ],
        [ :int32, MingleInt32, :read_int32 ],
        [ :uint32, MingleUint32, :read_uint32 ],
        [ :uint64, MingleUint64, :read_uint64 ],
        [ :float64, MingleFloat64, :read_float64 ],
        [ :float32, MingleFloat32, :read_float32 ],
    ].
    each do |arr|
        
        meth, num_cls, scan_meth = *arr

        define_method( :"read_mg_#{meth}" ) do |scanner|
            num_cls.new( scanner.send( scan_meth ) )
        end
    end

    private
    def read_mg_buffer( scanner )
        MingleBuffer.new( scanner.read_buffer32 )
    end

    private
    def read_mg_enum( scanner )
        
        typ = read_type_reference( scanner )
        value = read_identifier( scanner )

        MingleEnum.new( :type => typ, :value => value )
    end

    private
    def read_mg_timestamp( scanner )
        
        secs, nsec = scanner.read_int64, scanner.read_int32
        t = Time.at( secs, nsec.to_f / 1000.0 )

        MingleTimestamp.new( t, false )
    end

    private
    def read_mg_list( scanner )
        
        res = []
        len = scanner.read_int32 # ignored for now

        while ( tc = scanner.read_int8 ) != TYPE_CODE_END
            res << read_value( scanner, tc )
        end

        MingleList.new( res )
    end

    private
    def read_value( scanner, typ = scanner.read_int8 )
        
        case typ 

            when TYPE_CODE_BOOLEAN then read_mg_boolean( scanner )
            when TYPE_CODE_INT64 then read_mg_int64( scanner )
            when TYPE_CODE_INT32 then read_mg_int32( scanner )
            when TYPE_CODE_UINT32 then read_mg_uint32( scanner )
            when TYPE_CODE_UINT64 then read_mg_uint64( scanner )
            when TYPE_CODE_FLOAT64 then read_mg_float64( scanner )
            when TYPE_CODE_FLOAT32 then read_mg_float32( scanner )
            when TYPE_CODE_STRING then read_mg_string( scanner )
            when TYPE_CODE_BUFFER then read_mg_buffer( scanner )
            when TYPE_CODE_ENUM then read_mg_enum( scanner )
            when TYPE_CODE_TIMESTAMP then read_mg_timestamp( scanner )
            when TYPE_CODE_STRUCT then read_mg_struct( scanner )
            when TYPE_CODE_SYMBOL_MAP then read_fields( scanner )
            when TYPE_CODE_LIST then read_mg_list( scanner )
            when TYPE_CODE_NULL then MingleNull::INSTANCE
            else raise_unrecognized_value_code( typ, scanner )
        end
    end

    private
    def validate_from_buffer_args( buf )

        not_nil( buf, :buf )

        RubyVersions.when_19x do 
            buf.encoding == Encoding::BINARY or 
                codec_raise( "Buffer encoding is not binary" )
        end

        unless ( tc = @conv.read_int8( buf[ 0, 1 ] ) ) == TYPE_CODE_STRUCT 
            raise_unrecognized_value_code( tc, 0 )
        end
    end

    public
    def from_buffer( buf )
 
        validate_from_buffer_args( buf )

        scanner = Io::BinaryReader.new( 
            :order => BYTE_ORDER, :io => StringIO.new( buf ) )

        if ( res = read_value( scanner ) ).is_a?( MingleStruct )
            res
        else
            raise "Decode res wasn't a struct; got #{res.class}" 
        end
    end
end

end
end
