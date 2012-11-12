require 'bitgirder/core'
require 'mingle'

module Mingle
module Service

module MingleServices
 
    extend BitGirder::Core::BitGirderMethods
    include Mingle
    
    TYPE_SERVICE_REQUEST = 
        MingleTypeReference.get( :"service@v1/ServiceRequest" )

    TYPE_SERVICE_RESPONSE =
        MingleTypeReference.get( :"service@v1/ServiceResponse" )

    module_function

    def check_type( ms, typ_expct, err_type )
        
        ( typ = ms.type ) == typ_expct or
            raise "Invalid #{err_type} type: #{typ}"
    end

    def from_svc_req( req )
        
        flds = {
            :namespace => req.namespace.external_form,
            :service => req.service.external_form,
            :operation => req.operation.external_form
        }

        if ( ( val = req.parameters ) && ( ! val.empty? ) )
            flds[ :parameters ] = val 
        end

        if ( val = req.authentication ) then flds[ :authentication ] = val end

        MingleStruct.new( :type => TYPE_SERVICE_REQUEST, :fields => flds )
    end

    def from_svc_resp( resp )
        
        flds = 
            if resp.ok?
                resp.result ? { :result => resp.result } : {}
            else
                { :exception => resp.error }
            end
 
        MingleStruct.new( :type => TYPE_SERVICE_RESPONSE, :fields => flds )
    end

    def as_mingle_struct( obj )
        
        not_nil( obj, :obj )

        case obj
            when MingleServiceRequest then from_svc_req( obj )
            when MingleServiceResponse then from_svc_resp( obj )
            else raise "Can't convert to mingle struct: #{obj}"
        end
    end

    def as_service_request( ms )

        not_nil( ms, :ms )
        check_type( ms, TYPE_SERVICE_REQUEST, "service request" )

        f = ms.fields

        MingleServiceRequest.new( 
            :namespace => MingleNamespace.get( f.expect_string( :namespace ) ),
            :service => MingleIdentifier.get( f.expect_string( :service ) ),
            :operation => MingleIdentifier.get( f.expect_string( :operation ) ),
            :parameters => f[ :parameters ],
            :authentication => f[ :authentication ]
        )
    end

    def get_non_nil( val )

        case val
            when MingleNull then nil
            else val
        end
    end

    def as_service_response( ms )
        
        not_nil( ms, :ms )
        check_type( ms, TYPE_SERVICE_RESPONSE, "service response" )

        ex = get_non_nil( ms[ :exception ] )
        res = get_non_nil( ms[ :result ] )

        ( ex == nil || res == nil ) or 
            raise "Response has non-nil result and exception"
        
        if ex 
            MingleServiceResponse.create_failure( ex )
        else
            MingleServiceResponse.create_success( res )
        end
    end
end

end
end
