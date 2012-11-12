require 'mingle/http_integ'

module Mingle
module Http
module Test

class BitGirderServerTests < Mingle::Http::HttpClientTests

    include TestClassMixin

    private
    def get_test_server_names
        [ "bitgirder:mingle:http:test@v1/bgServer/plain" ]
    end
end

end
end
end
