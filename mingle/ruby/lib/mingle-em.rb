require 'bitgirder/core'
require 'mingle'

module Mingle

class EmHttpMingleRpcClient < BitGirder::Core::BitGirderClass
    
    require 'eventmachine'
    require 'em-http'

    HTTP_PRE_POST = :http_pre_post
    HTTP_RESP_RECEIVED = :http_resp_received
    MG_RESP_CREATED = :mg_resp_created

    bg_attr :endpoint

    bg_attr :identifier => :logger

    @@codec = JsonMingleCodec.new

    private
    def cli_log( ev, obj )
        
        case @logger
            
            when Proc then @logger.call( ev, obj )

            when BitGirder::Core::BitGirderLogger 
                @logger.code( "#{self} logged event #{ev} with object #{obj}" )
        end
    end

    private
    def handle_response( resp, blk )

        mg_resp, ex = [ nil, nil ]

        begin
            stat_str = resp.response_header.http_status

            if stat_str.to_i == 200
                mg_resp = @@codec.as_mingle_service_response( resp.response )
                cli_log( MG_RESP_CREATED, mg_resp )
            else
                raise "http got non-success status: #{stat_str}"
            end

        rescue Exception => ex; end

        begin
            blk.call( mg_resp, ex )
        rescue Exception => e
            warn( "Response handler block failed: #{e}\n" +
                  e.backtrace.join( "\n" ) )
        end
    end

    public
    def begin( mg_req, &blk )
        
        not_nil( mg_req, "mg_req" )
        raise "Need a response handler block" unless blk

        req = EM::HttpRequest.new( @endpoint )

        post_args = { :body => @@codec.as_codec_object( mg_req ) }
        cli_log( HTTP_PRE_POST, post_args )

        http = req.post( post_args )

        http.callback do |resp| 

            cli_log( HTTP_RESP_RECEIVED, resp )
            handle_response( resp, blk )
        end
    end

    def call( ns, svc, op, params = nil, auth = nil, &blk )
        
        req =
            MingleServiceRequest.new(
                :namespace => ns,
                :service => svc,
                :operation => op,
                :parameters => params,
                :authentication => auth )
    
        self.begin( req, &blk )
    end
end

end
