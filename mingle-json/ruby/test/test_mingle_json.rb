require 'mingle/json'
require 'mingle/codec'
require 'mingle/codec-tests'

module Mingle
module Json

include Mingle
include BitGirder::Core
include BitGirder::Testing

class JsonMingleCodecTests < BitGirderClass

    include TestClassMixin

    def test_bad_codec_opts
        assert_raised( OptionsError, /STUB/ ) do
            JsonMingleCodec.new( 
                :omit_type_fields => true, :expand_enums => true )
        end
    end
end

class StandardTests < Codec::StandardCodecTests

    def get_codec_id; :json; end
    def get_codec; JsonMingleCodec.new; end

    def get_spec_codec( spec )
        
        hdr_flds = spec.headers.fields
        
        opts = {
            :expand_enums => hdr_flds.get_boolean( :expand_enums ),
            :omit_type_fields => hdr_flds.get_boolean( :omit_type_fields ),
        }
 
        if fmt = hdr_flds.get_string( :id_format )
            opts[ :id_format ] = fmt
        end

        JsonMingleCodec.new( opts )
    end

    def expected_error_message_for( spec )
        
        case spec.id.to_s

            when "invalid-identifier-key" 
                %q{f1.f2.2bad: [<>, line 1, col 1]: Illegal start of identifier part: "2" (0x32)}

            when "invalid-identifier-enum-val"
                %q{f1.$constant: [<>, line 1, col 1]: Illegal start of identifier part: "2" (0x32)}

            when "incomplete-type-name"
                %q{$type: [<>, line 1, col 7]: Expected type path but found: END}

            else super
        end
    end
end

end
end
