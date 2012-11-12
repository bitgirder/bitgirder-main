require 'bitgirder/core'
include BitGirder::Core

require 'mingle'
include Mingle

require 'net/http'

module BitGirder
module Http

class HttpHeaders < BitGirderClass
    
    private_class_method :new

    bg_attr :pairs, 
            :default => [],
            :processor => lambda { |pairs| pairs.sort!.freeze }

    def self.as_header_name( nm )
        nm.to_s.downcase
    end

    def self.as_header_val( val )
        val.to_s.strip
    end

    def self.as_header_val_list( val )
        
        case val

            when Array, MingleList 
                val.empty? ? [ "" ] : val.map { |v| self.as_header_val( v ) }

            else [ self.as_header_val( val ) ]
        end
    end

    map_instance_of( Hash ) do |h|
        
        pairs = {}

        h.each_pair do |k, v|
            arr = ( pairs[ self.as_header_name( k ) ] ||= [] )
            arr.push( *( self.as_header_val_list( v ) ) )
        end

        self.send( :new, :pairs => pairs.to_a )
    end

    map_instance_of( Array ) do |arr|
        
        pairs = arr.map do |pair|
 
            if pair.size == 2
                [ self.as_header_name( pair[ 0 ] ), 
                  self.as_header_val_list( pair[ 1 ] ) ]
            else
                raise "Not a http header pair: #{pair.inspect}"
            end
        end

        self.send( :new, :pairs => pairs )
    end

    map_instance_of( Net::HTTPHeader ) do |hdrs|

        pairs = []
        hdrs.each_name { |nm| pairs << [ nm.downcase, hdrs.get_fields( nm ) ] }

        self.send( :new, :pairs => pairs )
    end

    public
    def to_mingle_struct

        MingleStruct.new(
            :type => :"bitgirder:http@v1/HttpHeaders",
            :fields => { :pairs => @pairs.map { |p| MingleList.new( p ) } }
        )
    end

    def self.from_mingle_struct( ms )
        self.as_instance( ( ms[ :pairs ] || [] ).to_a )
    end
end

class HttpEndpoint < BitGirderClass
    
    bg_attr :host, :required => false
    bg_attr :port, :required => false
    bg_attr :is_ssl, :required => false

    map_instance_of( Net::HTTP ) do |http|

        new( 
            :host => http.address, 
            :port => http.port, 
            :is_ssl => http.use_ssl?
        )
    end

    def to_mingle_struct

        MingleStruct.new(
            :type => :"bitgirder:http@v1/HttpEndpoint",
            :fields => {
                :host => @host,
                :port => @port,
                :is_ssl => @is_ssl
            }
        )
    end

    def self.from_mingle_struct( ms )
        
        self.new(
            :host => ms.fields.get_string( :host ),
            :port => ms.fields.get_int( :port ),
            :is_ssl => ms.fields.get_boolean( :is_ssl )
        )
    end
end

PROC_TO_BIN = lambda { |b|
    RubyVersions.when_19x( b ) { b.encode( Encoding::BINARY ) }
}

class HttpRequest < BitGirderClass
 
    bg_attr :headers, :processor => HttpHeaders, :required => false
    bg_attr :path, :required => false
    bg_attr :body, :processor => PROC_TO_BIN, :required => false

    def to_mingle_struct
        
        MingleStruct.new(
            :type => :"bitgirder:http@v1/HttpRequest",
            :fields => {
                :headers => @headers ? @headers.to_mingle_struct : nil,
                :path => @path,
                :body => @body ? MingleBuffer.new( @body ) : nil
            }
        )
    end

    def self.from_mingle_struct( ms )
        
        mg_hdrs, mg_body = ms[ :headers ], ms[ :body ]

        self.new(
            :headers => 
                mg_hdrs ? HttpHeaders.from_mingle_struct( mg_hdrs ) : nil,
            :path => ms.fields.get_string( :path ),
            :body => mg_body ? mg_body.buf : nil
        )
    end
end

class HttpStatus < BitGirderClass
    
    bg_attr :code
    bg_attr :message
    bg_attr :version

    map_instance_of( Net::HTTPResponse ) do |resp|
        
        ver =
            if ( ver_base = resp.http_version ).start_with?( "HTTP/" )
                ver_base
            else
                "HTTP/#{ver_base}"
            end

        self.new( 
            :message => resp.message, :code => resp.code, :version => ver )
    end

    public
    def to_mingle_struct
        MingleStruct.new(
            :type => :"bitgirder:http@v1/HttpStatus",
            :fields => {
                :code => @code,
                :message => @message,
                :version => @version
            }
        )
    end

    def self.from_mingle_struct( ms )

        self.new(
            :code => ms[ :code ].to_i,
            :message => ms[ :message ].to_s,
            :version => ms[ :version ].to_s
        )
    end
end

class HttpResponse < BitGirderClass
 
    bg_attr :status, :processor => HttpStatus, :required => false
    bg_attr :headers, :processor => HttpHeaders, :required => false
    bg_attr :body, :processor => PROC_TO_BIN, :required => false

    def self.from_net_http_response( resp, opts = {} )
 
        not_nil( resp, :resp )
        not_nil( opts, :opts )

        log_body = opts[ :log_body ]

        attrs = {
            :status => HttpStatus.as_instance( resp ),
            :headers => HttpHeaders.as_instance( resp )
        }

        if opts[ :log_body ] && resp.class.body_permitted?
            attrs[ :body ] = resp.body
        end

        new( attrs )
    end

    map_instance_of( Net::HTTPResponse ) do |resp|
        self.from_net_http_response( resp, :log_body => false )
    end

    def to_mingle_struct

        MingleStruct.new(
            :type => :"bitgirder:http@v1/HttpResponse",
            :fields => {
                :status => @status ? @status.to_mingle_struct : nil,
                :headers => @headers ? @headers.to_mingle_struct : nil,
                :body => @body ? MingleBuffer.new( body ) : nil
            }
        )
    end

    def self.from_mingle_struct( ms )
 
        stat, hdrs, body = ms[ :status ], ms[ :headers ], ms[ :body ]

        self.new(
            :status => stat ? HttpStatus.from_mingle_struct( stat ) : nil,
            :headers => hdrs ? HttpHeaders.from_mingle_struct( hdrs ) : nil,
            :body => body ? body.buf : nil
        )
    end
end

end
end
