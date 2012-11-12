require 'mingle'
require 'bitgirder/core'
require 'bitgirder/io'

module Mingle
module Io

require 'forwardable'

class Headers < BitGirder::Core::BitGirderClass
    
    def self.make_fields( flds )
        
        pairs = flds.inject( {} ) do |h, pair|
            val = pair[ 1 ]
            val = val.external_form if val.is_a?( MingleIdentifier )
            val = MingleModels.as_mingle_value( val )
            val = MingleModels.as_mingle_string( val )
            h[ pair[ 0 ] ] = val
            h
        end

        MingleSymbolMap.create( pairs )
    end

    def self.as_headers( val )
        case val
            when Headers then val
            when Hash, MingleSymbolMap then Headers.new( :fields => val )
            else raise TypeError, "Invalid headers value: #{val.class}"
        end
    end

    bg_attr :fields,
            :processor => lambda { |val| self.make_fields( val ) },
            :default => lambda { {} }
end

HEADERS_VERSION1 = 0x01

TYPE_CODE_HEADERS_FIELD = 0x01
TYPE_CODE_HEADERS_END = 0x02

BYTE_ORDER = BitGirder::Io::ORDER_LITTLE_ENDIAN

class Encoder < BitGirder::Core::BitGirderClass
 
    bg_attr :writer

    private
    def impl_initialize
        @bin = BitGirder::Io::BinaryWriter.new( :order => BYTE_ORDER, :io => @writer )
    end

    public
    def write_int32( i )
        @bin.write_int32( i )
    end

    public 
    def write_int64( i )
        @bin.write_int64( i )
    end

    public
    def write_utf8( s )
        @bin.write_utf8( s )
    end

    public
    def write_headers( hdrs )

        write_int32( HEADERS_VERSION1 )
        hdrs.fields.each_pair do |k, v|
            write_int32( TYPE_CODE_HEADERS_FIELD )
            write_utf8( k.external_form )
            write_utf8( v.to_s )
        end
        write_int32( TYPE_CODE_HEADERS_END )
    end
end

class InvalidVersionError < StandardError; end
class InvalidTypeCodeError < StandardError; end

class Decoder < BitGirder::Core::BitGirderClass
    
    bg_attr :reader

    extend Forwardable
    def_delegators :@bin, 
        :read_full, :read_int32, :read_int64, :read_buffer32, :read_utf8

    private
    def impl_initialize
        @bin = BitGirder::Io::BinaryReader.new( :order => BYTE_ORDER, :io => @reader )
    end

    public
    def read_version
        read_int32
    end

    public
    def expect_version( expct, ver_typ )
        
        unless ( ver = read_version ) == expct

            tmpl = "Invalid %s :version => 0x%08x (expected 0x%08x)"
            msg = sprintf( tmpl, ver_typ, ver, expct )
            raise InvalidVersionError.new( msg )
        end
    end

    public
    def read_type_code
        read_int32
    end

    public
    def expect_type_code( expct )
        
        unless ( act = read_type_code ) == expct
            
            tmpl = "Invalid type :code => 0x%08x (expected 0x%08x)"
            msg = sprintf( tmpl, act, expct )
            raise InvalidTypeCodeError.new( msg )
        end
    end
 
    private
    def read_header_field( flds )
        flds[ read_utf8 ] = read_utf8
    end

    public
    def read_headers
        
        expect_version( HEADERS_VERSION1, "headers" )

        flds = {}
        while true do

            case tc = read_type_code

                when TYPE_CODE_HEADERS_FIELD then read_header_field( flds )

                when TYPE_CODE_HEADERS_END 
                    return Headers.new( :fields => flds )
                
                else 
                    msg = sprintf( "Unknown type :code => 0x%08x", tc )
                    raise InvalidTypeCodeError.new( msg )
            end
        end
    end
end

end
end
