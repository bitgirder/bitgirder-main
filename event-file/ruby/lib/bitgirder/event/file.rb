require 'bitgirder/core'
include BitGirder::Core

require 'bitgirder/io'

require 'stringio'

module BitGirder
module Event
module File

include BitGirder::Io

FILE_MAGIC = "3vEntF!L"
FILE_VERSION = "event-file20120703"

class EventFileIo < BitGirderClass
    
    bg_attr :codec

    HEADER_SIZE = 4
    
    DEFAULT_BUFFER_SIZE = Io::DataSize.as_instance( "4m" )
end

class OpenResult < BitGirderClass
    
    bg_attr :io
    bg_attr :is_reopen, :default => false
    bg_attr :pos, :default => 0
end

class WriteError < StandardError; end

class ClosedError < StandardError; end

class EventFileWriter < EventFileIo

    DEFAULT_ROTATE_SIZE = Io::DataSize.as_instance( "64m" )

    bg_attr :file_factory

    bg_attr :rotate_size,
            :processor => Io::DataSize,
            :default => DEFAULT_ROTATE_SIZE

    bg_attr :buffer_size, :processor => Io::DataSize, :required => false

    bg_attr :event_handler, :required => false

    private
    def impl_initialize

        @buffer_size ||= [ @rotate_size, DEFAULT_BUFFER_SIZE ].min

        if @buffer_size > @rotate_size
            raise ArgumentError, 
                  "Buffer size #@buffer_size > rotate size #@rotate_size"
        end

        @conv = Io::BinaryConverter.new(    :order => Io::ORDER_LITTLE_ENDIAN )
    end

    public
    def closed?
        @closed
    end

    private
    def send_event( ev, *argv )
        
        if @event_handler && @event_handler.respond_to?( ev )
            @event_handler.send( ev, *argv )
        end
    end

    private
    def new_string_io
        RubyVersions.when_19x( StringIO.new ) do |io|
            io.set_encoding( Encoding::BINARY )
        end
    end

    private
    def reset_bin_writer
            
        @bin =
            Io::BinaryWriter.new(
                :io => StringIO.new( @buf ),
                :order => Io::ORDER_LITTLE_ENDIAN
            )
    end 

    private
    def write_file_header
        
        @bin.write_full( FILE_MAGIC )
        @bin.write_utf8( FILE_VERSION )
        
        if ( hdr_len = @bin.pos ) >= ( rs = @rotate_size.bytes )
            
            @closed = true

            raise WriteError, 
                  "File header length #{hdr_len} >= rotate size #{rs}"
        end

        @buf_remain -= @bin.pos
    end

    private
    def update_remain_counts( file_remain )
        
        @file_remain = file_remain
        @buf_remain = [ @file_remain, @buffer_size.bytes ].min
    end

    private
    def ensure_io
 
        @buf ||= ( "\x00" * @buffer_size.bytes )

        unless @io

            open_res = @file_factory.open_file
            @io = open_res.io

            update_remain_counts( @rotate_size.bytes - open_res.pos )
            reset_bin_writer
            write_file_header unless open_res.is_reopen
        end
    end

    private
    def close_file
 
        if @io

            @file_factory.close_file( @io )
            @buf = @bin.io.string
            @io, @bin, @file_remain, @buf_remain = nil, nil, -1, -1
        end
    end

    private
    def impl_flush
 
        completes_file = @file_remain <= @buffer_size.bytes

        @buf.slice!( @bin.io.pos, @buf.size )

        @io.write( @buf )
        send_event( :wrote_buffer, @buf.size )

        # Order matters: need to use @buf before reset_bin_writer
        update_remain_counts( @file_remain - @buf.size )
        reset_bin_writer

        close_file if completes_file || @closed
    end

    private
    def write_ev_header( str_io, dest )
        dest.write( @conv.write_int32( str_io.size ) )
    end

    private
    def write_serialized_event( str_io, dest )
            
        write_ev_header( str_io, dest )
        dest.write( str_io.string[ 0, str_io.size ] )
    end

    private
    def write_overflow( str_io, io_sz )

        impl_flush

        if io_sz > @buffer_size.bytes
            
            if io_sz > @rotate_size.bytes
                raise WriteError, 
                    "Record is too large for rotate size #@rotate_size" 
            else
                write_serialized_event( str_io, @io )
            end
        else
            impl_write_event( str_io, io_sz )
        end
    end

    private
    def impl_write_event( str_io, io_sz )

        ensure_io

        if io_sz > @buf_remain
            write_overflow( str_io, io_sz )
        else
            write_serialized_event( str_io, @bin )
            if ( @buf_remain -= io_sz ) == 0 then impl_flush end
        end
    end

    public
    def write_event( ev )
        
        raise ClosedError if @closed

        str_io = new_string_io
        @codec.encode_event( ev, str_io )

        io_sz = str_io.pos + HEADER_SIZE

        impl_write_event( str_io, io_sz )
    end

    public
    def close

        @closed = true
        impl_flush if @io
    end
end

class EventFileLogger < BitGirderClass

    require 'thread'
 
    bg_attr :writer

    @@shutdown_sentinel = Object.new

    private
    def impl_initialize
        @queue = Queue.new
    end

    private
    def process_queue
 
        until ( ev = @queue.pop ) == @@shutdown_sentinel
            @writer.write_event( ev )
        end

        @writer.close
    end

    public
    def start
        @worker = Thread.start { process_queue }
    end

    public
    def event_logged( ev )
        @queue << ev
    end

    public
    def shutdown

        @queue << @@shutdown_sentinel
        @worker.join
       
        nil
    end

    def self.start( *argv )
        self.new( *argv ).tap { |l| l.start }
    end
end

class EventFileExistsError < StandardError; end

class NoOpCodec < BitGirderClass
    
    public
    def encode_event( io ); end

    public
    def decode_event( io, len )
        Io.read_full( io, len )
    end
end

class PathGenerator < BitGirderClass
    
    bg_attr :format

    public
    def generate( t = Time.now )

        millis = ( t.to_f * 1000 ).to_i
        millis_hex = sprintf( "%016x", millis )

        eval( @format, binding )
    end

    map_instance_of( String ) { |s| self.new( :format => s ) }
end

class EventFileFactory < BitGirderClass
    
    bg_attr :dir

    bg_attr :path_generator

    bg_attr :event_handler, :required => false

    # We can find a way to let callers customize this later with a block or
    # regex as needed
    private
    def find_reopen_target
        Dir.glob( "#@dir/**/*" ).max
    end

    private
    def reopen_target_corrupt( f, e )
        
        if ( eh = @event_handler ) && eh.respond_to?( :reopen_target_corrupt )
            eh.send( :reopen_target_corrupt, f, e )
        else
            warn( e, "Reopen target #{f} was invalid; skipping to next file" )
        end
    end

    private
    def init_reopen_target( f )
        
        ::File.open( f, "r+b" ) do |io|

            rd = EventFileReader.new( :codec => NoOpCodec.new, :io => io )

            trunc_at = 0

            until io.eof? || f == nil do
                begin
                    rd.read_event
                    trunc_at = io.pos
                rescue FileFormatError => e
                    reopen_target_corrupt( f, e )
                    f = nil
                rescue EOFError
                    io.truncate( trunc_at )
                end
            end
        end

        f
    end

    private
    def init_reopen
        
        if @reopen_targ = find_reopen_target
            @reopen_targ = init_reopen_target( @reopen_targ )
        end
    end

    private
    def gen_path

        base = case pg = @path_generator
            when Proc then pg.call
            when PathGenerator then pg.generate
            else raise TypeError, "Unhandled path generator: #{pg.class}"
        end
        
        "#@dir/#{base}"
    end

    public
    def open_file
        
        if @reopen_targ

            io = ::File.open( @reopen_targ, "a+b" )
            @reopen_targ = nil

            OpenResult.new( 
                :io => io, 
                :is_reopen => true, 
                :pos => Io.fsize( io )
            )
        else
            if ::File.exist?( path = gen_path )
                raise EventFileExistsError, "File already exists: #{path}"
            else
                OpenResult.new( :io => ::File.open( path, "wb" ) )
            end
        end
    end

    public
    def close_file( io )
        io.close
    end

    def self.open( opts )
        self.new( opts ).tap { |ff| ff.send( :init_reopen ) }
    end
end

class FileFormatError < StandardError; end

class FileVersionError < FileFormatError; end

class FileMagicError < FileFormatError; end

class EventFileReader < EventFileIo

    include Enumerable

    bg_attr :io

    private
    def impl_initialize
        
        super

        @bin = Io::BinaryReader.new( 
            :io => @io, 
            :order => Io::ORDER_LITTLE_ENDIAN 
        )

        @read_file_header = false
    end

    private
    def read_file_header
        
        begin
            unless ( magic = @bin.read_full( 8 ) ) == FILE_MAGIC
                raise FileMagicError, 
                      "Unrecognized file magic: #{magic.inspect}" 
            end
        rescue EOFError
            raise FileMagicError, "Missing or incomplete file magic"
        end

        unless ( ver_str = @bin.read_utf8 ) == FILE_VERSION
            raise FileVersionError, 
                  "Unrecognized file version: #{ver_str.inspect}"
        end

        @read_file_header = true
    end

    # Returns the next [ ev, loc ] pair if there is one, EOFException if
    # incomplete, and nil if this is the first read of a file which contains
    # only a valid header. 
    public
    def read_event
 
        read_file_header unless @read_file_header
        return nil if @io.eof?

        loc = @bin.pos

        len = @bin.read_int32
        buf = @bin.read_full( len )
        
        ev = @codec.decode_event( StringIO.new( buf, "r" ), len )

        [ ev, loc ]
    end

    public
    def each_with_loc
 
        until @io.eof?
            ev, loc = *( read_event )
            yield( ev, loc ) if loc
        end
    end

    public
    def each
        each_with_loc { |ev, loc| yield( ev ) }
    end
end

end
end
end
