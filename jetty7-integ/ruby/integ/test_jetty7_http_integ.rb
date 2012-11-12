require 'bitgirder/mingle/http_integ'

module BitGirder
module Jetty7

class Jetty7ServerTests < Mingle::Http::HttpClientTests

    include TestClassMixin

    private
    def get_test_server_names
        [ "bitgirder:jetty7@v1/server/plain" ]
    end
end

end
end
