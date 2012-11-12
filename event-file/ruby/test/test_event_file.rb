require 'bitgirder/core'
include BitGirder::Core

require 'bitgirder/event/testing'
include BitGirder::Event::Testing

require 'bitgirder/event/file'

require 'bitgirder/testing'
require 'bitgirder/io'
require 'bitgirder/io/testing'

require 'digest/md5'
require 'tmpdir'

module BitGirder
module Event
module File

class IoWrapper < BitGirderClass
    
    bg_attr :io

    public
    def open_file
        OpenResult.new( :io => @io )
    end

    public
    def close_file( io ); end
end

class StringIOFactory < BitGirderClass
    
    include Io::Testing

    bg_attr :data, :default => []

    public
    def open_file
        OpenResult.new( :io => new_string_io )
    end

    public
    def close_file( io )

        io.close_write
        io.seek( 0, IO::SEEK_SET )
        @data << io
    end
end

class TestFileFactory < BitGirderClass
    
    bg_attr :files, :default => []
    bg_attr :dir, :default => lambda { Dir.mktmpdir }
    bg_attr :event_handler, :required => false

    private
    def impl_initialize
        
        @ev_ff = EventFileFactory.open(
            :dir => @dir,
            :path_generator => lambda { "file#{@files.size}" },
            :event_handler => @event_handler
        )
    end

    public
    def open_file

        # Don't re-add the file if it is being closed again after being reopened
        @ev_ff.open_file.tap do |res| 
            unless @files.include?( p = res.io.path )
                @files << p
            end
        end
    end

    public
    def close_file( io )
        @ev_ff.close_file( io )
    end

    public
    def remove_all
        Io.fu.remove_entry_secure( @dir )
    end

    public
    def copy( with_opts = {} )
        
        opts = { 
            :dir => @dir, 
            :files => @files, 
            :event_handler => @event_handler
        }.merge( with_opts )

        self.class.new( opts )
    end
end

class EmptyCodec < BitGirderClass

    READ_VALUE = ""
    
    public
    def decode_event( *argv ); READ_VALUE; end

    public
    def encode_event( *argv ); end
end

class EventFileTests < BitGirder::Testing::TestHolder

    include BitGirder::Io::Testing

    def new_digest
        Digest::MD5.new
    end

    def reset( io )
        io.tap { io.seek( 0, IO::SEEK_SET ) }
    end

    def as_file_reader( str )

        EventFileReader.new( 
            :io => new_string_io( str ), 
            :codec => TestCodec.new 
        )
    end

    def output_files_of( fact )
        case fact
        when StringIOFactory then fact.data
        when TestFileFactory then fact.files
        else raise "Unhandled factory type: #{fact.class}"
        end
    end

    def file_reader_for( obj, opts = {} )
        
        opts = opts.merge( :codec => TestCodec.new )

        opts[ :io ] = 
            case obj
            when StringIO then obj
            when String then ::File.open( obj, "rb" )
            else raise "Unhandled file object: #{obj.class}"
            end
        
        EventFileReader.new( opts )
    end

    def output_file_size_of( obj )
        case obj
        when StringIO then Io::DataSize.as_instance( obj.string.bytesize )
        when String then Io::DataSize.as_instance( Io.fsize( obj ) )
        else raise "Unhandled file object: #{obj.class}"
        end
    end

    def read_with_index( rd )
        
        idx = 0

        rd.each_with_loc do |ev, loc| 
            yield( ev, loc, idx )
            idx += 1
        end

        idx
    end

    def write_int_recs( wr, num_recs, opts = {} )
        
        offset = opts[ :offset ] || 0

        num_recs.times do |i|
            wr.write_event( Int32Event.new( i + offset ) )
        end

        wr.close unless opts[ :close ] == false
    end

    def write_buffer_recs( wr, num_recs, buf_sz, opts = {} )

        dig = opts[ :digest ]

        num_recs.times do |i|

            buf = rand_buf( buf_sz )
            dig.update( buf ) if dig
            wr.write_event( BufferEvent.new( buf ) )
        end

        wr.close unless opts[ :close ] == false
    end

    def assert_int_recs( rd, num_recs )

        assert_equal(
            num_recs,
            read_with_index( rd ) do |ev, loc, i|
                assert_equal( Int32Event.new( i ), ev )
            end
        )
    end

    # Create a new reader and set some of its internal state in order to assert
    # that loc is the correct first byte of the record ev
    def assert_loc( ev, loc, io, cdc )

        save_pos = io.pos

        begin
            io.seek( loc, IO::SEEK_SET )
            rd = EventFileReader.new( :io => io, :codec => cdc )
            rd.instance_variable_set( :@read_file_header, true )
            assert_equal( ev, rd.read_event[ 0 ] )
        ensure
            io.seek( save_pos, IO::SEEK_SET )
        end
    end

    def assert_read_write( num_recs )

        io, cdc = new_string_io, TestCodec.new

        wr = EventFileWriter.new( 
            :file_factory => IoWrapper.new( io ), 
            :codec => cdc 
        )

        write_int_recs( wr, num_recs )

        io.seek( 0, IO::SEEK_SET )
        rd = EventFileReader.new( :io => io, :codec => cdc )

        evs_read = read_with_index( rd ) do |ev, loc, idx|
            assert_equal( ev, Int32Event.from_i( idx ) )
            assert_loc( ev, loc, io, cdc )
        end

        assert_equal( num_recs, evs_read )
    end

    def test_read_write_roundtrip
        [ 0, 10 ].each { |num_recs| assert_read_write( num_recs ) }
    end

    def test_empty_records
        
        io, cdc = new_string_io, EmptyCodec.new
        
        wr = EventFileWriter.new( 
            :file_factory => IoWrapper.new( io ),
            :codec => cdc 
        )

        10.times { |i| wr.write_event( "" ) }
        wr.close

        io.seek( 0, IO::SEEK_SET )
        rd = EventFileReader.new( :io => io, :codec => cdc )
 
        evs_read = read_with_index( rd ) do |ev, loc, idx|
            assert_equal( EmptyCodec::READ_VALUE, ev )
        end

        assert_equal( 10, evs_read )
    end

    def test_bad_file_magic
        
        assert_raised( 
            'Unrecognized file magic: "abcdefgh"', FileMagicError ) do 
            as_file_reader( "abcdefghij" ).read_event
        end

        7.times do |i|
            assert_raised(
                'Missing or incomplete file magic', FileMagicError ) do
                as_file_reader( "a" * i ).read_event
            end
        end
    end

    def test_bad_file_version

        io = new_string_io
        bin = Io::BinaryWriter.new( 
            :io => io, :order => Io::ORDER_LITTLE_ENDIAN )

        bin.write_full( FILE_MAGIC )
        bin.write_utf8( "bad-version" )

        msg = 'Unrecognized file version: "bad-version"'

        assert_raised( msg, FileVersionError ) do
            as_file_reader( io.string ).read_event
        end
    end

    def test_lazy_file_open
 
        # since :file_factory is just an Object, we'll fail with
        # MethodMissingError if any attempts are made to access it and create or
        # close a file
        wr = EventFileWriter.new(
            :file_factory => Object.new,
            :codec => TestCodec.new
        ).close
    end

    def assert_rotate( opts )
        
        wr = has_key( opts, :writer )

        act_files = output_files_of( wr.file_factory )
        expct_files = has_key( opts, :expect_files )
        expct_files_pad = opts[ :expect_files_pad ] || 1
        expct_files_rng = ( expct_files .. expct_files + expct_files_pad )
        assert( expct_files_rng.include?( act_files.size ) )

        dig = new_digest

        act_files.each do |f|
            assert( output_file_size_of( f ) <= wr.rotate_size )
            file_reader_for( f ).each { |ev| dig.update( ev.buf ) }
        end

        assert_equal( has_key( opts, :digest ).digest, dig.digest )
    end

    def run_basic_rotate_test( opts )
        
        h = { :digest => new_digest }

        wr = h[ :writer ] = EventFileWriter.new(
            :rotate_size => "10k",
            :buffer_size => "5k",
            :file_factory => has_key( opts, :file_factory ),
            :codec => TestCodec.new
        )

        expct_files = h[ :expect_files ] = 4
        buf_sz = Io::DataSize.as_instance( "750b" )
        num_recs = expct_files * ( wr.rotate_size.bytes / buf_sz.bytes )
        write_buffer_recs( wr, num_recs, buf_sz, :digest => h[ :digest ] )

        assert_rotate( h )
    end

    def test_stringio_rotate_basic
        run_basic_rotate_test( :file_factory => StringIOFactory.new )
    end

    def test_file_rotate_basic

        ff = TestFileFactory.new

        begin
            run_basic_rotate_test( :file_factory => ff )
        ensure
            ff.remove_all
        end
    end

    # The general rotate test is for basic properties (file size <= rotate size)
    # and sequencing. It could happen in that test that all events fall out so
    # that the last event in each runs exactly to rotate_size (though unlikely),
    # so we test here explicitly for handling of records that overflow a file
    # and which should carry over into the rotated file.
    def test_rotate_boundary_overflow
        
        h = { :expect_files => 2, :expect_files_pad => 0 }

        wr = h[ :writer ] = EventFileWriter.new(
            :rotate_size => "3k",
            :buffer_size => "3k",
            :file_factory => StringIOFactory.new,
            :codec => TestCodec.new
        )

        dig = h[ :digest ] = new_digest

        2.times do 
            buf = rand_buf( 2000 ).tap { |buf| dig.update( buf ) }
            wr.write_event( BufferEvent.new( buf ) )
        end

        wr.close

        assert_rotate( h )
    end

    class WriteBatchCallAsserter < BitGirderClass
        
        include BitGirder::Testing::AssertMethods

        bg_attr :buf_sz

        private
        def impl_initialize
            
            super
            @calls = []
        end

        public
        def wrote_buffer( sz )
            @calls << { :buffer_size => sz }
        end

        public
        def assert_calls( expct_bufs )
            
            rng = ( expct_bufs .. expct_bufs + 1 )
            assert( rng.include?( @calls.size ) )

            rec_sz = EventFileIo::HEADER_SIZE + 8 # size of an Int32Event
            min_sz = @buf_sz.bytes - rec_sz

            # Only check size of first expct_bufs calls
            expct_bufs.times do |i|
                assert( @calls[ i ][ :buffer_size ] >= min_sz )
            end
        end
    end

    def test_write_batching
        
        buf_sz = Io::DataSize.as_instance( "10k" )

        wr = EventFileWriter.new(
            :file_factory => IoWrapper.new( new_string_io ),
            :rotate_size => "50m", # don't rotate
            :codec => TestCodec.new,
            :buffer_size => buf_sz,
            :event_handler => WriteBatchCallAsserter.new( buf_sz )
        )

        expct_bufs, rec_sz = 3, EventFileIo::HEADER_SIZE + 8
        num_recs = ( buf_sz.bytes / rec_sz * expct_bufs ) 
        write_int_recs( wr, num_recs )

        rd = file_reader_for( reset( wr.file_factory.io ) )
        assert_int_recs( rd, num_recs )
        wr.event_handler.assert_calls( expct_bufs )
    end

    # Ensure that a write which exactly fills up the buffer leads to an
    # immediate flushing to the underlying io
    def test_write_flushes_on_buffer_precise_fill
        
        wr = EventFileWriter.new(
            :file_factory => IoWrapper.new( new_string_io ),
            :codec => TestCodec.new,
            :buffer_size => "512b"
        )

        wr.write_event( Int32Event.new( 0 ) )

        buf_remain = wr.instance_variable_get( :@buf_remain )
        buf = rand_buf( buf_remain - 4 - EventFileIo::HEADER_SIZE ) 
        wr.write_event( BufferEvent.new( buf ) )
        assert_equal( wr.buffer_size.bytes, wr.file_factory.io.pos )

        wr.close
        rd = file_reader_for( reset( wr.file_factory.io ) )
        assert_equal( Int32Event.new( 0 ), rd.read_event[ 0 ] )
        assert_equal( BufferEvent.new( buf ), rd.read_event[ 0 ] )
        assert( rd.io.eof? )
    end

    def test_close_flushes_buffered_data
        
        wr = EventFileWriter.new(
            :file_factory => IoWrapper.new( new_string_io ),
            :codec => TestCodec.new,
            :buffer_size => "1k"
        )

        wr.write_event( Int32Event.new( 0 ) )
        assert_equal( 0, wr.file_factory.io.pos ) # No data should be written

        wr.close # Now data should be written

        assert_int_recs( file_reader_for( reset( wr.file_factory.io ) ), 1 )
    end

    def test_reopen_basic

        ff = TestFileFactory.new 
        wr = EventFileWriter.new( :file_factory => ff, :codec => TestCodec.new )
        write_int_recs( wr, 10 )
        assert_equal( 1, ff.files.size )

        ff = ff.copy
        wr = EventFileWriter.new( :file_factory => ff, :codec => TestCodec.new )
        write_int_recs( wr, 10, :offset => 10 )
        assert_equal( 1, ff.files.size )

        assert_int_recs( file_reader_for( ff.files.first ), 20 )
    end

    def test_reopen_truncates_partial_record
        
        ff = TestFileFactory.new
        wr = EventFileWriter.new( :file_factory => ff, :codec => TestCodec.new )
        write_int_recs( wr, 10 )
        assert_equal( 1, ff.files.size )
        ::File.open( ff.files.first, "r+b" ) do |io| 
            io.truncate( Io.fsize( io ) - 1 )
        end

        ff = ff.copy
        wr = EventFileWriter.new( :file_factory => ff, :codec => TestCodec.new )
        write_int_recs( wr, 10, :offset => 9 )
        assert_equal( 1, ff.files.size )

        assert_int_recs( file_reader_for( ff.files.first ), 19 )
    end

    def test_reopen_skips_corrupt_file
        
        ff = TestFileFactory.new
        wr = EventFileWriter.new( :file_factory => ff, :codec => TestCodec.new )
        write_int_recs( wr, 1 )
        ::File.open( ff.files.first, "r+b" ) { |io| io.truncate( 7 ) }

        class <<( eh = {} )
            def reopen_target_corrupt( f, ex )
                raise "Multiple calls" unless self.empty?
                merge!( :file => f, :exception => ex )
            end
        end

        ff = ff.copy( :event_handler => eh )
        wr = EventFileWriter.new( :file_factory => ff, :codec => TestCodec.new )
        write_int_recs( wr, 10 )
        assert_equal( 2, ff.files.size )

        assert_equal( 7, Io.fsize( ff.files.first ) )
        assert_int_recs( file_reader_for( ff.files[ 1 ] ), 10 )
        assert( eh[ :exception ].is_a?( FileFormatError ) )
        assert_equal( ff.files.first, eh[ :file ] )
    end

    # Check that if the first record after a reopen causes a file size overflow
    # then we rotate without appending and append to the new file
    def test_reopen_and_immediate_overflow
        
        ff = TestFileFactory.new
        wr = EventFileWriter.new( :file_factory => ff, :codec => TestCodec.new )
        write_int_recs( wr, 100 )
        assert_equal( 1, ff.files.size )

        ff = ff.copy
        wr = EventFileWriter.new(
            :file_factory => ff,
            :codec => TestCodec.new,
            :buffer_size => 100,
            :rotate_size => Io.fsize( ff.files.first ) + 1
        )
        write_int_recs( wr, 100 )
        assert_equal( 2, ff.files.size )

        ff.files.each { |f| assert_int_recs( file_reader_for( f ), 100 ) }
    end

    def test_file_factory_detects_existent_file
        
        Io.open_tempfile do |tmp|
            
            msg = "File already exists: #{tmp.path}"

            assert_raised( msg, EventFileExistsError ) do 
                EventFileFactory.new(
                    :dir => ::File.dirname( tmp.path ),
                    :path_generator => lambda { ::File.basename( tmp.path ) }
                ).
                open_file
            end
        end
    end

    def test_header_larger_than_rotate_size_fails
        
        wr = EventFileWriter.new(
            :rotate_size => 10,
            :buffer_size => 5,
            :file_factory => StringIOFactory.new,
            :codec => TestCodec.new
        )

        msg = 'File header length 30 >= rotate size 10'

        assert_raised( msg, WriteError ) do
            wr.write_event( Int32Event.new( 1 ) )
        end

        assert( wr.closed? )
    end

    def test_clean_scan_of_header_only_file
        
        io = new_string_io

        bin = Io::BinaryWriter.new(
            :io => io, 
            :order => Io::ORDER_LITTLE_ENDIAN 
        )

        bin.write_full( FILE_MAGIC )
        bin.write_utf8( FILE_VERSION )

        io.seek( 0, IO::SEEK_SET )

        rd = EventFileReader.new( :io => io, :codec => TestCodec.new )
        assert_equal( 0, read_with_index( rd ) {} )
    end

    def test_write_event_fails_after_close
        
        wr = EventFileWriter.new(
            :file_factory => StringIOFactory.new,
            :codec => TestCodec.new
        )

        wr.close

        assert_raised( ClosedError ) { wr.write_event( "" ) }
    end

    def test_detect_buffer_size_larger_than_rotate_size
        
        assert_raised( "Buffer size 1000 > rotate size 500", ArgumentError ) do
            EventFileWriter.new( 
                :file_factory => StringIOFactory.new,
                :buffer_size => 1000, 
                :rotate_size => 500,
                :codec => TestCodec.new
            )
        end
    end

    def test_event_file_reader_enumerable
        
        wr = EventFileWriter.new(
            :file_factory => IoWrapper.new( new_string_io ),
            :codec => TestCodec.new
        )

        write_int_recs( wr, 100 )

        rd = EventFileReader.new(
            :io => reset( wr.file_factory.io ),
            :codec => TestCodec.new
        )

        assert_equal( 
            [ 0, 33, 66, 99 ], 
            rd.select { |ev| ev.i % 33 == 0 }.map { |ev| ev.i }
        )
    end

    # Should flush any pending writes and then go straight to io with large
    # event, bypassing buffer. This is a regression, since the first version of
    # this lib went into an infinite loop in this case
    def test_event_larger_than_buf_size
        
        wr = EventFileWriter.new(
            :file_factory => IoWrapper.new( new_string_io ),
            :codec => TestCodec.new,
            :buffer_size => "1k"
        )

        wr.write_event( Int32Event.new( 0 ) )
        wr.write_event( BufferEvent.new( buf = rand_buf( 2048 ) ) )

        rd = EventFileReader.new(
            :io => reset( wr.file_factory.io ), :codec => wr.codec )
        
        assert_equal( Int32Event.new( 0 ), rd.read_event[ 0 ] )
        assert_equal( BufferEvent.new( buf ), rd.read_event[ 0 ] )
        assert( rd.io.eof? )
    end

    def test_event_larger_than_rotate_size_fails
        
        wr = EventFileWriter.new(
            :file_factory => IoWrapper.new( new_string_io ),
            :codec => TestCodec.new,
            :buffer_size => "1k",
            :rotate_size => "1k"
        )

        msg = 'Record is too large for rotate size 1024'

        assert_raised( msg, WriteError ) do
            wr.write_event( BufferEvent.new( "\x00" * 2048 ) )
        end
    end

    def test_default_buffer_size
        
        h = { :file_factory => StringIOFactory.new, :codec => TestCodec.new }

        assert_equal(
            EventFileWriter::DEFAULT_BUFFER_SIZE,
            EventFileWriter.new( h.merge( :rotate_size => "1g" ) ).buffer_size
        )

        assert_equal(
            Io::DataSize.as_instance( "1k" ),
            EventFileWriter.new( h.merge( :rotate_size => "1k" ) ).buffer_size
        )

        assert_equal(
            Io::DataSize.as_instance( "1k" ), 
            EventFileWriter.new( h.merge( :buffer_size => "1k" ) ).buffer_size
        )
    end

    def test_event_file_logger_basic

        log = EventFileLogger.start(
            :writer => EventFileWriter.new(
                :file_factory => IoWrapper.new( new_string_io ),
                :codec => TestCodec.new
            )
        )

        num_recs, delay, start = 5, 3, Time.now
        
        num_recs.times do |i|
            ev = Int32Event.new( i )
            ev = DelayEvent.new( :delay => delay, :event => ev ) if i == 1
            log.event_logged( ev )
        end
        
        assert( Time.now - start <= 0.5 ) # event_logged should itself be fast
        log.shutdown
        assert( Time.now - start >= delay ) # But we should have delayed in log

        rd = file_reader_for( reset( log.writer.file_factory.io ) )
        assert_int_recs( rd, num_recs )
    end

    def test_path_generator_bindings
        
        pg = PathGenerator.as_instance( '"abc-#{millis_hex}"' )

        t = Time.now
        millis = ( t.to_f * 1000 ).to_i 

        assert_equal(
            sprintf( "abc-%016x", millis ),
            pg.generate( t )
        )
    end
end

# To test:
#
# - Codec which produces no data does not lead to an empty record being written

end
end
end
