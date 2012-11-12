require 'mingle/codec'
require 'mingle/codec-tests'
require 'mingle/bincodec'
require 'bitgirder/testing'

module Mingle
module BinCodec

class BinCodecTests < BitGirder::Core::BitGirderClass
    
    include TestClassMixin
    include Mingle::Codec

    def test_trailing_input
        
        codec = MingleBinCodec.new

        ms = MingleStruct.new( :type => :"ns1@v1/S1" )
        enc_buf = MingleCodecs.encode( codec, ms )
        enc_buf << "test-tail"

        ms2 = MingleCodecs.decode( codec, enc_buf )
        ModelTestInstances.assert_equal( ms, ms2 )
    end
end

class StandardTests < Mingle::Codec::StandardCodecTests

    include BitGirder::Testing::TestClassMixin

    def get_codec_id; :binary; end
    def get_codec; MingleBinCodec.new; end

    def expected_error_message_for( spec )
        
        case spec.id.to_s

            when "test-rfc3339-str-fail"
                %q{[offset 50]: Invalid rfc3339 timestamp: invalid date: "2009-23-22222"}
 
            else super
        end
    end
end

end
end
