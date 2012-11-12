require 'bitgirder/core'
include BitGirder::Core

require 'bitgirder/event/testing'
include BitGirder::Event::Testing

require 'bitgirder/event/file'
require 'bitgirder/event/logger'
require 'bitgirder/io'
require 'bitgirder/testing'

module BitGirder
module Event
module File

# This class is for testing composition of an EventFileLogger with an
# Event::Engine. Tests of things that are specific only to
# EventFile(Writer|Logger) go in the event-file package.
class EventLoggerIntegTest < BitGirder::Testing::TestHolder

    EvLog = BitGirder::Event::Logger
    EvFile = BitGirder::Event::File

    include BitGirder::Io

    def create_ev_file_logger( dir )
        
        # Return files that sort lexicographically in creation order
        class <<( path_gen = [ 0 ] )
            def next
                raise "Too many files" if first > 99
                i = self.shift.tap { |i| self << i + 1 }
                sprintf( "event-file-%02i", i )
            end
        end

        EvFile::EventFileLogger.start(
            EvFile::EventFileWriter.new(
                :file_factory => EvFile::EventFileFactory.new(
                    :dir => dir,
                    :path_generator => lambda { path_gen.next }
                ),
                :rotate_size => "100k",
                :buffer_size => "20k",
                :codec => TestCodec.new
            )
        )
    end

    def event_at_index( i, opts = {} )
 
        buf_sz = opts[ :buf_size ] || Io::DataSize.as_instance( "1k" )
        delay = opts[ :delay ] || 0.1
        delay_every = opts[ :delay_every ] || 75

        ev = case i % 2
             when 0 then Int32Event.new( i )
             when 1 then BufferEvent.new( "\x00" * buf_sz.bytes )
             end
        
        unless opts[ :no_delay ] || i % delay_every != 0
            ev = DelayEvent.new( :event => ev, :delay => delay ) 
        end

        ev
    end

    def write_events( num_evs, eng )
        num_evs.times { |i| eng.log_event( event_at_index( i ) ) }
    end

    def assert_events( num_evs, dir )
            
        i, codec = 0, TestCodec.new

        Dir.glob( "#{dir}/**/*" ).sort.each do |f|
            ::File.open( f ) do |io|
                EventFileReader.new( :io => io, :codec => codec ).each do |ev|
                    assert_equal( event_at_index( i, :no_delay => true ), ev )
                    i += 1
                end
            end
        end

        assert_equal( num_evs, i )
    end

    def test_integ
 
        mktmpdir do |dir|
 
            eng = EvLog::Engine.new
            efl = eng.add_listener( create_ev_file_logger( dir ) )

            write_events( num_evs = 1500, eng )
            efl.shutdown

            assert_events( num_evs, dir )
        end
    end
end

end
end
end
