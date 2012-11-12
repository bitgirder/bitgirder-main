require 'bitgirder/core'
require 'mingle'
require 'mingle/io'

module Mingle
module Codec

include BitGirder::Core

ACTION_ID_ROUND_TRIP = 0x01
ACTION_ID_FAIL_DECODE = 0x02
ACTION_ID_DECODE_INPUT = 0x03
ACTION_ID_ENCODE_VALUE = 0x04

class TestSpecDecoder < Mingle::Io::Decoder
    
    include Mingle

    bg_attr :codec

    private
    def read_id
        
        if ( str = read_utf8 ).empty?
            nil
        else
            MingleIdentifier.get( str )
        end
    end

    private
    def read_struct
        MingleCodecs.decode( @codec, read_buffer32 )
    end

    private
    def read_round_trip
        RoundTrip.new( :struct => read_struct )
    end

    private
    def read_fail_decode
        FailDecode.new(
            :error_message => read_utf8,
            :input => read_buffer32
        )
    end

    private
    def read_decode_input
        DecodeInput.new(
            :input => read_buffer32,
            :expect => read_struct
        )
    end

    private
    def read_encode_value
        EncodeValue.new( :value => read_struct )
    end

    private
    def read_action
        
        act_id = read_int32

        case act_id 
            when ACTION_ID_ROUND_TRIP then read_round_trip
            when ACTION_ID_FAIL_DECODE then read_fail_decode
            when ACTION_ID_DECODE_INPUT then read_decode_input
            when ACTION_ID_ENCODE_VALUE then read_encode_value
            else raise sprintf( "Unexpected action :id => 0x%08x", act_id )
        end
    end

    public
    def read_test_spec
            
        opts = {}
        opts[ :id ] = read_id
        opts[ :codec_id ] = read_id
        opts[ :headers ] = read_headers
        opts[ :action ] = read_action

        TestSpec.new( opts )
    end
end

class RoundTrip < BitGirderClass
    bg_attr :struct
end

class FailDecode < BitGirderClass
    bg_attr :error_message
    bg_attr :input
end

class DecodeInput < BitGirderClass
    bg_attr :input
    bg_attr :expect
end

class EncodeValue < BitGirderClass
    bg_attr :value
end

class TestSpec < BitGirderClass
    
    include Mingle

    bg_attr :id
    bg_attr :codec_id, :required => false
    bg_attr :headers
    bg_attr :action

    public
    def key
        
        res = @id.external_form
        @codec_id ? res << "/" << @codec_id.external_form : res
    end

    @@bgm = BitGirder::Core::BitGirderMethods

    def self.decode( buf, codec )
        
        @@bgm.not_nil( buf, :buf )

        TestSpecDecoder.new( 
            :reader => StringIO.new( buf ),
            :codec => codec
        ).
        read_test_spec
    end
end

end
end
