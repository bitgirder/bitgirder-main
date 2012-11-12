require 'bitgirder/io'

module Mingle
module Codec

require 'tempfile'

class MingleCodecError < StandardError; end

module MingleCodecImpl

    private
    def codec_raise( *argv )
        raise Mingle::Codec::MingleCodecError, *argv
    end
end

module MingleCodecs

    @@bgm = BitGirder::Core::BitGirderMethods

    def decode( codec, obj )

        @@bgm.not_nil( codec, :codec )
        @@bgm.not_nil( obj, :obj )

        case obj

            when String then codec.from_buffer( obj )

            when IO, Tempfile
                data = BitGirder::Io.slurp_io( obj ) || ""
                decode( codec, data ) # recurse

            else 
                raise "Don't know how to decode obj #{obj} of type #{obj.class}"
        end
    end

    module_function :decode

    def encode( codec, mv )
 
        @@bgm.not_nil( codec, :codec )
        @@bgm.not_nil( mv, :mv )

        codec.as_buffer( mv )
    end

    module_function :encode
end

end
end
