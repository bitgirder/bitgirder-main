require 'bitgirder/core'
require 'mingle'
require 'mingle/service'
require 'mingle/codec'

require 'uri'

module Mingle
module Http

class ServerLocation < BitGirder::Core::BitGirderClass

    bg_attr :host
    bg_attr :port

    bg_attr :uri,
            :processor => lambda { |s| s.start_with?( "/" ) ? s : "/#{s}" }
    
    public
    def to_url
        "http://#{host}:#{port}#{uri}"
    end
end

class HttpCodecContext < BitGirder::Core::BitGirderClass
    
    bg_attr :codec
    bg_attr :content_type
end

# Simple blocking client using ruby builtin http lib
class NetHttpMingleRpcClient < BitGirder::Core::BitGirderClass
    
    require 'net/http'

    include Mingle::Codec
    include Mingle::Service

    bg_attr :location
    bg_attr :codec_ctx
    bg_attr :reactor, :required => false

    private
    def react( meth, *argv )
 
        case @reactor

            when Hash
                if func = @reactor[ meth ] then func.call( *argv ) end

            else
                if @reactor.respond_to?( meth )
                    @reactor.send( meth, *argv )
                end
        end
    end

    # Input checker/converter for call()
    private
    def get_mg_req( argv )
        
        # Fail if nothing at all was given
        not_nil( argv[ 0 ], :mg_req )

        case argv.size
            when 1 
                case obj = argv[ 0 ]
                    when MingleServiceRequest then obj
                    when Hash then MingleServiceRequest.new( obj )
                    else raise "Invalid or incomplete call param: #{obj}"
                end
            else
                MingleServiceRequest.new(
                    :namespace => argv[ 0 ],
                    :service => argv[ 1 ],
                    :operation => argv[ 2 ],
                    :parameters => argv[ 3 ],
                    :authentication => argv[ 4 ] )
        end
    end

    private
    def create_request( mv )
         
        req = Net::HTTP::Post.new( @location.uri )

        req.body = MingleCodecs.encode( @codec_ctx.codec, mv )
        req.content_type = @codec_ctx.content_type
        req[ "connection" ] = "close"

        react( :complete_request, req )

        req
    end

    private
    def as_service_response( body )
        
        ms = MingleCodecs.decode( @codec_ctx.codec, body )
        MingleServices.as_service_response( ms )
    end

    public
    def call( *argv )
        
        mv = MingleServices.as_mingle_struct( get_mg_req( argv ) )
        req = create_request( mv )

        resp = 
            Net::HTTP.new( @location.host, @location.port ).start do |http|
                http.request( req )
            end
 
        react( :response_received, resp )

        case resp
            when Net::HTTPOK then as_service_response( resp.body )
            else raise "Got non-OK response: #{resp} (#{resp.body})"
        end
    end

#    class ServiceClient < BitGirder::Core::BitGirderClass
#        
#        bg_attr :cli
#        bg_attr :namespace
#        bg_attr :service
#        bg_attr( :identifier => :authentication )
#
#        def initialize( opts )
#            super
#        end
#
#        public
#        def call( op, args = {}, auth = nil )
#
#            @cli.call(
#                @namespace,
#                @service,
#                not_nil( op, :op ),
#                not_nil( args, :args ),
#                auth || @authentication
#            )
#        end
#    end
#
#    # returns a ServiceClient object bound with the given preset context. opts
#    # must contain the keys :namespace ans :service and may optionally contain
#    # an auth object :authentication
#    public
#    def create_service_client( opts )
#        ServiceClient.new( opts.merge( :cli => self ) )
#    end
end

end
end
