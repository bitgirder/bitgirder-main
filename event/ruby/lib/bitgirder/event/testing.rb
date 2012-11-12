require 'bitgirder/core'
include BitGirder::Core

require 'bitgirder/io'
require 'bitgirder/io/testing'
include BitGirder::Io

module BitGirder
module Event
module Testing

class Int32Event < BitGirderClass

    TYPE_CODE = 0x0001

    bg_attr :i

    def self.from_i( i ); self.new( :i => i ); end
end

class BufferEvent < BitGirderClass
    
    TYPE_CODE = 0x0002

    bg_attr :buf
end

class DelayEvent < BitGirderClass
    
    bg_attr :delay # will be arg to sleep()
    bg_attr :event
end

class TestCodec < BitGirderClass

    include BitGirder::Io

    private
    def impl_initialize
        @conv = BinaryConverter.new( :order => Io::ORDER_LITTLE_ENDIAN )
    end

    public
    def encode_event( ev, io )

        if ev.is_a?( DelayEvent )
            sleep( ev.delay )
            ev = ev.event
        end

        io.write( @conv.write_int32( ev.class::TYPE_CODE ) )

        case ev
        when Int32Event then io.write( @conv.write_int32( ev.i ) )
        when BufferEvent then io.write( ev.buf )
        else raise TypeError.new( "Unrecognized event: #{ev.class}" )
        end
    end

    public
    def decode_event( io, len )
 
        ev_buf_len = len - 4

        case cd = @conv.read_int32( io.read( 4 ) )
        when Int32Event::TYPE_CODE
            Int32Event.from_i( @conv.read_int32( io.read( 4 ) ) )
        when BufferEvent::TYPE_CODE
            BufferEvent.new( Io.read_full( io, ev_buf_len ) )
        else raise sprintf( "Unexpected type code: %04x", cd )
        end
    end
end

def self.roundtrip( ev, codec )
    
    io = Io::Testing.new_string_io
    codec.encode_event( ev, io )
    
    len = io.pos
    io.seek( 0, IO::SEEK_SET )
    
    codec.decode_event( io, len )
end

end
end
end
