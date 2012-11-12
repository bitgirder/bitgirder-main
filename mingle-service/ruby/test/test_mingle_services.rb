require 'bitgirder/core'
require 'mingle'
require 'mingle/test-support'
require 'mingle/service'
require 'bitgirder/testing'

module Mingle
module Service

class MingleServiceTests < BitGirder::Core::BitGirderClass

    include TestClassMixin

    include Mingle

    def assert_req_roundtrip( flds )

        req1 = MingleServiceRequest.new( flds )
        str1 = MingleServices.as_mingle_struct( req1 )

        ModelTestInstances.assert_equal(
            MingleStruct.new( 
                :type => :"service@v1/ServiceRequest", 
                :fields => flds
            ),
            str1
        )

        req2 = MingleServices.as_service_request( str1 )
        ModelTestInstances.assert_equal( req1, req2 )
    end

    def test_svc_req_to_struct_all_fields_set
 
        assert_req_roundtrip(
            :namespace => :"ns1@v1",
            :service => :svc1,
            :operation => :op1,
            :parameters => { :param1 => :val1 },
            :authentication => :auth1
        )
    end

    def test_svc_req_to_struct_empty_fields
        
        assert_req_roundtrip(
            :namespace => :"ns1@v1",
            :service => :svc1,
            :operation => :op1
        )
    end

    def assert_resp_roundtrip( flds )
        
        resp = 
            if flds[ :exception ]
                MingleServiceResponse.create_failure( flds[ :exception ] )
            else
                flds = {} if flds[ :result ] == nil
                MingleServiceResponse.create_success( flds[ :result ] )
            end

        str1 = MingleServices.as_mingle_struct( resp )

        ModelTestInstances.assert_equal( 
            MingleStruct.new(
                :type => :"service@v1/ServiceResponse",
                :fields => flds
            ),
            str1
        )

        resp2 = MingleServices.as_service_response( str1 )
        ModelTestInstances.assert_equal( resp, resp2 )
    end

    def test_svc_resp_success
        assert_resp_roundtrip( :result => MingleString.new( "great" ) )
    end

    def test_svc_resp_nil_success
        assert_resp_roundtrip( {} )
    end

    def test_svc_resp_exception
        assert_resp_roundtrip( :exception => MingleString.new( "bad" ) )
    end

    def test_as_mingle_struct_fail_unrecognized_obj
        
        msg = %q{Can't convert to mingle struct: Fixnum}
            
        assert_raised( msg, Exception ) do
            MingleServices.as_mingle_struct( Fixnum )
        end
    end

    def test_as_service_request_fail_bad_mg_type
 
        assert_raised( 
            "Invalid service request type: foo@v1/Blah", Exception ) do

            MingleServices.as_service_request( 
                MingleStruct.new( :type => :"foo@v1/Blah" ) )
        end
    end

    def test_as_service_request_fail_invalid_req
        
        assert_raised( "Map has no value for key: operation", Exception ) do
            MingleServices.as_service_request(
                MingleStruct.new(
                    :type => MingleServices::TYPE_SERVICE_REQUEST,
                    :fields => { :namespace => :"ns1@v1", :service => :svc1 }
                )
            )
        end
    end

    def test_service_resp_fail_invalid_type
        
        assert_raised( 
            "Invalid service response type: ns1@v1/Blah", 
            Exception ) do

            MingleServices.as_service_response(
                MingleStruct.new( :type => :"ns1@v1/Blah" )
            )
        end
    end

    def test_service_resp_non_nil_excpt_and_res
        
        assert_raised( 
            "Response has non-nil result and exception", Exception ) do

            MingleServices.as_service_response(
                MingleStruct.new(
                    :type => MingleServices::TYPE_SERVICE_RESPONSE,
                    :fields => { :result => 1, :exception => 2 }
                )
            )
        end
    end
end

end
end
