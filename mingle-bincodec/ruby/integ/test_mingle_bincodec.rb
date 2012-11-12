require 'bitgirder/testing'
require 'mingle/bincodec'
require 'mingle/codec-integ'

module Mingle
module BinCodec

class BinCodecIntegTests < Mingle::Codec::CodecIntegTests

    include TestClassMixin

    def get_codec_id; :binary; end
    def get_codec; MingleBinCodec.new; end

    def debug_rt_bin_obj( bin_obj )

        msg = "bin_obj:"
        cols = 20

        bytes = bin_obj.bytes.to_a

        begin
            arr = bytes.shift( cols )
            fmt = "\n    " + Array.new( arr.size, "%02x" ).join( " " )
            msg << sprintf( fmt, *arr )
        end until bytes.empty?

        code( msg )
    end
end

end
end
