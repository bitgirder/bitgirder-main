require 'bitgirder/core'
require 'bitgirder/io'
require 'mingle/io'

module Mingle
module Io
module Stream

MESSAGE_VERSION1 = 1

TYPE_CODE_HEADERS = 1
TYPE_CODE_MESSAGE_BODY = 2

class Message < BitGirder::Core::BitGirderClass

    bg_attr :headers, 
            :processor => lambda { |val|
                Mingle::Io::Headers.as_headers( val )
            },
            :default => lambda { Mingle::Io::Headers.as_headers( {} ) }

    bg_attr :body, :default => "".freeze
end

class Connection < BitGirder::Core::BitGirderClass

    include Mingle

    bg_attr :reader
    bg_attr :writer

    private
    def impl_initialize

        @enc = Io::Encoder.new( @writer )
        @dec = Io::Decoder.new( @reader )
    end

    private
    def write_headers( hdrs )

        @enc.write_int32( TYPE_CODE_HEADERS )
        @enc.write_headers( hdrs )
    end

    private
    def write_body( body )
        
        @enc.write_int32( TYPE_CODE_MESSAGE_BODY )
        @enc.write_int64( sz = body.bytesize )
        @writer.write( body ) unless sz == 0
    end

    public
    def write_message( msg )

        msg = Message.as_instance( msg )

        @enc.write_int32( MESSAGE_VERSION1 )
        write_headers( msg.headers )
        write_body( msg.body )
        @writer.flush
    end

    private
    def read_body
        sz = @dec.read_int64
        @dec.read_full( sz )
    end

    public
    def read_message
        
        @dec.expect_version( MESSAGE_VERSION1, "message" )

        @dec.expect_type_code( TYPE_CODE_HEADERS )
        hdrs = @dec.read_headers

        @dec.expect_type_code( TYPE_CODE_MESSAGE_BODY )
        body = read_body
        
        Message.new( :headers => hdrs, :body => body )
    end

end

class MinglePeer < BitGirder::Core::BitGirderClass

    include Mingle
    
    private_class_method :new

    bg_attr :proc_builder

    private
    def start
        @peer = @proc_builder.popen( "r+" )
        @conn = Connection.new( :reader => @peer, :writer => @peer )
    end

    public
    def exchange_message( msg )
        
        @conn.write_message( msg )
        @conn.read_message
    end

    public
    def await_exit( opts = { :expect_success => true } )

        not_nil( opts, :opts )

        msg = nil

        begin
            pid, stat = 
                BitGirder::Io.debug_wait2( :pid => @peer.pid, :name => :peer )
    
            unless stat.success?
                msg = "Peer #{pid} exited with non-success #{stat.exitstatus}"
            end
        rescue Errno::ECHILD
            msg = "Peer #{pid} appears to have exited already"
        end
        
        if msg
            if opts[ :expect_success ] then raise msg else warn msg end
        end
    end

    def self.open( *argv )
        
        res = self.send( :new, *argv )
        res.send( :start )

        res
    end
end

end
end
end
