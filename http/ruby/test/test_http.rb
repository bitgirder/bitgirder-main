require 'bitgirder/http'

require 'bitgirder/testing'
include BitGirder::Testing

require 'net/http'

module BitGirder
module Http

class ObjectTests < BitGirderClass
 
    include TestClassMixin

    def assert_as_instance( defls, *tests )
        
        tests.each do |test|
            
            test = defls.merge( test )

            meth = test[ :method ] || :as_instance
            inv = lambda { test[ :class ].send( meth, test[ :in ] ) }

            if err = test[ :error ]
                assert_raised( *err ) { inv.call }
            else
                expct, act = test[ :expect ], inv.call
                assert_equal( expct, act )
            end
        end
    end

    # check various create modes and that headers are downcase-normalized and
    # sorted
    def test_headers
        
        expct = HttpHeaders.send( :new, 
            :pairs => [
                [ "a", %w{ v1 } ], [ "b", %w{ v1 v2 } ], [ "c", [ "" ] ]
            ]
        )

        ref_hash = { :a => %w{ v1 }, :b => %w{ v1 v2 }, :c => [ "" ] }

        defls = { :class => HttpHeaders, :expect => expct }

        assert_as_instance( defls,

            { :in => [ [ "a", %w{ v1 } ], [ "B", %w{ v1 v2 } ], [ "c", [] ] ] },
            { :in => [ [ "b", %w{ v1 v2 } ], [ "A", %w{ v1 } ], [ "C", [] ] ] },

            { 
                :in => [ [ :a, %w{ v1 } ], [ :b, %w{ v1 v2 } ], [ :c, [ "" ] ] ]
            },

            { :in => ref_hash },
            { :in => { "a" => %w{ v1 }, "B" => %w{ v1 v2 }, "c" => "" } },
            { :in => ref_hash.merge( :c => [ "" ] ) },

            { 
                :in => 
                    Net::HTTP::Get.new( "/ignore" ).tap do |req|
                        req.add_field( "a", "v1" )
                        req.add_field( "b", "v1" )
                        req.add_field( "B", "v2" )
                        req.add_field( "c", "" )
                        req.delete( "user-agent" )
                        req.delete( "accept" )
                    end
            },

            { 
                :in => [ [ :a, "1" ], [ :b, :c, :d ] ],
                :error => [ 
                    %q{Not a http header pair: [:b, :c, :d]}, Exception ]
            }
        ) 
    end

    def test_endpoint
 
        expct = HttpEndpoint.new(
            :host => "somewhere.net", :port => 12345, :is_ssl => true )

        assert_as_instance( { :class => HttpEndpoint, :expect => expct },
            
            {
                :in => 
                    # In some older (j)ruby versions, ssl can only be set after
                    # an include of net/https, which itself requires ssl to be
                    # installed. Rather than require all of that just to ensure
                    # that http.use_ssl? returns true, we stub it out here
                    Net::HTTP.new( "somewhere.net", 12345 ).tap do |http|
                        class <<http; def use_ssl?; true; end; end
                    end
            }
        )
    end

    def test_request
        
        expct = HttpRequest.new(
            :headers => HttpHeaders.as_instance( :a => "val1" ),
            :path => "/foo",
            :body => "stuff"
        )

        assert_as_instance( { :expect => expct, :class => HttpRequest },

            { 
                :in => {
                    :headers => { "a" => "val1" },
                    :path => "/foo",
                    :body => "stuff"
                }
            }
        )
    end

    def test_status
        
        expct =
            HttpStatus.new( 
                :code => 200, :message => "OK", :version => "HTTP/1.1" )

        defls = { :class => HttpStatus, :expect => expct }

        assert_as_instance( defls,

            { 
                :in => { 
                    :code => 200, :message => "OK", :version => "HTTP/1.1" }
            },

            { :in => Net::HTTPOK.new( "1.1", 200, "OK" ) }
        )
    end

    def test_response

        expct = HttpResponse.new(
            :status => HttpStatus.new(
                :code => 200,
                :message => "OK",
                :version => "HTTP/1.1"
            ),
            :headers => HttpHeaders.as_instance(
                :a => "val1", 
                :b => %w{ val1 val2 }
            )
        )
            
        resp = Net::HTTPOK.new( "HTTP/1.1", 200, "OK" ).tap do |resp|
            
            resp.add_field( "a", "val1" )
            resp.add_field( "b", "val1" )
            resp.add_field( "b", "val2" )

            class <<resp; def body; "stuff"; end; end
        end
        
        assert_as_instance( { :class => HttpResponse, :expect => expct },
            
            {
                :in => {
                    :status => { 
                        :code => 200, 
                        :version => "HTTP/1.1",
                        :message => "OK"
                    },
                    :headers => { :a => "val1", :b => %w{ val1 val2 } },
                }
            },

            { :in => resp }
        )

        assert_equal(
            HttpResponse.new(
                :status => expct.status,
                :headers => expct.headers,
                :body => "stuff"
            ),
            HttpResponse.from_net_http_response( resp, :log_body => true )
        )
    end

    def test_mingle_roundtrips
        
        [
            HttpHeaders.as_instance( {} ),

            HttpHeaders.as_instance(
                "h1" => "v1",
                "h2" => %w{ v1 v2 }
            ),

            HttpEndpoint.new(
                :host => "somewhere.net",
                :port => 1234,
                :is_ssl => true
            ),

            HttpEndpoint.new,

            HttpRequest.new(
                :headers => { "h1" => "v1", "h2" => %w{ v1 v2 } },
                :path => "/foo",
                :body => "\x00\x00"
            ),

            HttpRequest.new(),

            HttpStatus.new(
                :code => 200,
                :message => "OK",
                :version => "HTTP/1.1"
            ),

            HttpResponse.new(
                :status => HttpStatus.new(
                    :code => 200,
                    :message => "OK",
                    :version => "HTTP/1.1"
                ),
                :headers => { "h1" => "v1" },
                :body => "\x00\x00"
            ),

            HttpResponse.new

        ].each do |obj|
            
            ms = obj.to_mingle_struct
            obj2 = obj.class.from_mingle_struct( ms )
            assert_equal( obj, obj2 )
        end
    end

    def test_request_and_resp_body_encode_as_binary
        
        # start with utf-8
        str = RubyVersions.when_19x( "abcd" ) { |s| s.encode!( "utf-8" ) }

        req = HttpRequest.new( :body => str )
        resp = HttpResponse.new( :body => str )

        # Even without the checks below, the above codepaths are good to
        # exercise even in an earlier ruby
        RubyVersions.when_19x do 
            [ req.body, resp.body ].each do |b| 
                assert_equal( Encoding::BINARY, b.encoding )
            end
        end
    end
end

end
end
