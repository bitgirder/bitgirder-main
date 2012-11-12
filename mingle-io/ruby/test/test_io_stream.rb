require 'bitgirder/testing'
require 'bitgirder/core'
require 'bitgirder/io/testing'
require 'mingle/test-support'
require 'mingle/io/stream'

module Mingle
module Io
module Stream

include BitGirder::Core
include BitGirder::Testing
include Mingle

class StreamTests < BitGirderClass
    
    include TestClassMixin
    include BitGirder::Io::Testing

    # we can use the same io for reader/writer so long as we do all of the write
    # first and then rewind io before the read.
    private
    def assert_msg_rt( msg, copies = 1 )
        
        io = new_string_io
        conn = Connection.new( :reader => io, :writer => io )

        copies.times { conn.write_message( msg ) }
        io.rewind

        copies.times do

            msg2 = conn.read_message
            assert_equal( msg.body, msg2.body )

            ModelTestInstances.
                assert_equal( msg.headers.fields, msg2.headers.fields )
        end
    end

    public
    def test_empty_msg_rt
        
        empties = []
        empties << Message.new
        empties << Message.new( :headers => {}, :body => "" )

        [ 1, 10 ].each do |copies|
            empties.each { |msg| assert_msg_rt( msg, copies ) }
        end
    end

    public
    def test_nonempty_msg_rt

        msg = 
            Message.new(
                :headers => { :f1 => 1, :f2 => "f2", :f3 => true },
                :body => "\x01\x02\x03"
            )

        [ 1, 10 ].each { |copies| assert_msg_rt( msg, copies ) }
    end

    private
    def assert_read_failure( input, ex_cls, msg )
        
        ex = assert_raised( ex_cls ) do

            io = StringIO.new( opt_encode( input, "binary" ) )
            conn = Connection.new( :writer => io, :reader => io )

            conn.read_message
        end

        assert_equal( msg, ex.message )
    end

    public
    def test_bad_message_version
        assert_read_failure(
            "\x03\x03\x03\x03",
            Io::InvalidVersionError,
            'Invalid message :version => 0x03030303 (expected 0x00000001)'
        )
    end

    public
    def test_bad_headers_type_code
        assert_read_failure(
            "\x01\x00\x00\x00\x03\x03\x03\x03",
            Io::InvalidTypeCodeError,
            'Invalid type :code => 0x03030303 (expected 0x00000001)'
        )
    end

    public
    def test_bad_body_type_code
        
        io = new_string_io
        enc = Io::Encoder.new( io )
        enc.write_int32( MESSAGE_VERSION1 )
        enc.write_int32( TYPE_CODE_HEADERS )
        enc.write_headers( Io::Headers.new() )
        enc.write_int32( 7654321 ) # bad body type code
        
        assert_read_failure(
            io.string,
            Io::InvalidTypeCodeError,
            'Invalid type :code => 0x0074cbb1 (expected 0x00000002)'
        )
    end
end

end
end
end
