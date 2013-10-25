require 'bitgirder/core'
require 'bitgirder/io'
require 'mingle'
require 'mingle/codec'

module Mingle
module Json

include BitGirder::Core
include Mingle

class OptionsError < StandardError; end

class JsonMingleCodec < BitGirderClass

    KEY_TYPE = "$type"
    KEY_CONSTANT = "$constant"
                
    MSG_MISSING_TYPE_KEY = %Q{Missing type key ("#{KEY_TYPE}")}

    bg_attr :expand_enums, :processor => :boolean, :default => false
 
    bg_attr :id_format,
            :processor => lambda { |v| MingleIdentifier.as_format_name( v ) },
            :default => :lc_hyphenated
 
    bg_attr :omit_type_fields, :processor => :boolean, :default => false

    include Mingle::Codec::MingleCodecImpl
    include BitGirder::Core

    require 'json'
    require 'base64'

    private
    def impl_initialize

        if @omit_type_fields && @expand_enums
            raise OptionsError.new( 
                "Illegal combination of :omit_type_fields and :expand_enums" )
        end
    end

    private
    def decode_raise( path, msg )
        
        msg = "#{path.format}: #{msg}" if path
        codec_raise( msg )
    end

    private
    def from_mingle_symbol_map( map, res = {} )
        
        map.each_pair do |k, v| 
            res[ k.format( @id_format ) ] = from_mingle_value( v )
        end

        res
    end

    private
    def from_mingle_struct( val )
        
        res = from_mingle_symbol_map( val.fields, {} )
        res[ KEY_TYPE ] = val.type.external_form unless @omit_type_fields

        res
    end

    private
    def from_mingle_buffer( val )
        Base64.strict_encode64( val.buf )
    end

    private
    def from_mingle_enum( en )
 
        val = en.value.external_form

        if @expand_enums
            { KEY_TYPE => en.type.external_form, KEY_CONSTANT => val }
        else
            val
        end
    end

    private
    def from_mingle_value( val )
        
        case val
            when MingleString then val.to_s
            when MingleSymbolMap then from_mingle_symbol_map( val )
            when MingleStruct then from_mingle_struct( val )
            when MingleBoolean then val.as_boolean

            when MingleInt64, MingleInt32, MingleFloat64, MingleFloat32 
                val.num

            when MingleBuffer then from_mingle_buffer( val )
            when MingleTimestamp then val.rfc3339
            when MingleEnum then from_mingle_enum( val )
            when MingleList then val.map { |elt| from_mingle_value( elt ) }
            when MingleNull then nil
            when nil then nil
            else codec_raise "Can't convert to json an instance of #{val.class}"
        end
    end

    public
    def as_buffer( obj )
        JSON.generate( from_mingle_value( obj ) )
#        JSON.generate( from_mingle_value( obj ) ).encode( "binary" )
    end

    private
    def descend( path, key )
        path ? path.descend( key ) : ObjectPath.get_root( key )
    end

    private
    def start_list( path )
        path ? path.start_list : ObjectPath.get_root_list
    end

    private
    def parse_identifier( s, path )
        
        begin
            MingleIdentifier.parse( s )
        rescue MingleParseError => e
            decode_raise( path, e.message )
        end
    end

    private
    def parse_type_reference( s, path )
        
        begin
            MingleTypeReference.parse( s )
        rescue MingleParseError => e
            decode_raise( path, e.message )
        end
    end

    private
    def enum_val_in( h, path )
        
        if val = h[ KEY_CONSTANT ]
            if val.is_a?( String )
                parse_identifier( val, descend( path, KEY_CONSTANT ) )
            else
                decode_raise( 
                    path.descend( KEY_CONSTANT ), "Invalid constant value" )
            end
        end
    end

    # Returns nil if no KEY_TYPE val is present in h. Raises an exception if it
    # is but is not parsable, is not an atomic type ref, or the type name is not
    # a qname
    private
    def type_ref_in( h, path )
        
        if typ_str = h[ KEY_TYPE ]

            err_path = descend( path, KEY_TYPE )

            typ = parse_type_reference( typ_str, err_path )

            unless typ.is_a?( AtomicTypeReference )
                decode_raise( err_path, 
                    "Not an atomic type reference: #{typ_str}" )
            end

            return typ if typ.name.is_a?( QualifiedTypeName )
            decode_raise( err_path, "Not a qualified type name: #{typ_str}" )
        end
    end

    private
    def as_symbol_map( h, path )
        
        res = {}

        h.each_pair do |k, v|
            unless k == KEY_TYPE
                if /^\$/ =~ k
                    msg = "Unrecognized control key: #{k.inspect}"
                    decode_raise( path, msg )
                else
                    key_path = descend( path, k )
                    id = parse_identifier( k, key_path )
                    val = as_mingle_value( v, key_path )
                    res[ id ] = val
                end
            end
        end

        MingleSymbolMap.create( res )
    end

    private
    def from_json_hash( h, struct_cls, path )
        
        en_const = enum_val_in( h, path )
        type_ref = type_ref_in( h, path )

        if en_const 
            if type_ref 
                if h.size > 2
                    decode_raise path, "Enum has one or more unrecognized keys"
                else
                    MingleEnum.new( :type => type_ref, :value => en_const )
                end
            else
                decode_raise path, MSG_MISSING_TYPE_KEY
            end
        else
            flds = as_symbol_map( h, path )
            type_ref ? 
                MingleStruct.new( :type => type_ref, :fields => flds ) : flds
        end
    end

    private
    def from_json_array( arr, path )
        
        lp = start_list( path )

        vals = arr.map do |v|
            val = as_mingle_value( v, lp )
            lp = lp.next
            val
        end

        MingleList.new( vals )
    end

    private
    def as_mingle_value( val, path )
 
        case val
            when Hash then from_json_hash( val, MingleStruct, path )
            when Array then from_json_array( val, path )
            else MingleModels.as_mingle_value( val )
        end
    end

    public
    def from_buffer( buf )

        not_nil( buf, :buf )
        json = BitGirder::Io.parse_json( buf )

        if json.is_a?( Hash )

            codec_raise( MSG_MISSING_TYPE_KEY ) if json.empty?

            if ( res = as_mingle_value( json, nil ) ).is_a?( MingleStruct )
                res
            else
                codec_raise( "Expected struct" ) 
            end
        else
            codec_raise( "Unexpected top level JSON value" )
        end
    end
end

end
end
