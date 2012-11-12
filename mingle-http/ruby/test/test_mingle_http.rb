require 'bitgirder/core'
require 'bitgirder/testing'
require 'mingle/http'

module Mingle
module Http

include BitGirder::Testing

class ServerLocationTests < BitGirder::Core::BitGirderClass
    
    include TestClassMixin

    def test_uri_with_lead_slash
        
        loc = ServerLocation.new( 
            :host => "host", :port => 1234, :uri => "/foo" )

        assert_equal( "http://host:1234/foo", loc.to_url.to_s )
    end

    def test_uri_with_no_lead_slash
        
        loc = ServerLocation.new( :host => "host", :port => 1234, :uri => "foo" )
        assert_equal( "http://host:1234/foo", loc.to_url.to_s )
    end
end

end
end
