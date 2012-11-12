require 'bitgirder/core'
require 'bitgirder/testing'
require 'bitgirder/io/testing'
require 'mingle/io'
require 'mingle/test-support'

module Mingle
module Io

include BitGirder::Testing
include BitGirder::Core
include Mingle

class HeadersTests < BitGirderClass
    
    include TestClassMixin
    include BitGirder::Io::Testing

    public
    def test_headers_create

        [ 
            [ Headers.new, {} ],

            [ Headers.new( :fields => {} ), {} ],

            [ Headers.new( :fields => MingleSymbolMap.create( :f1 => 1 ) ),
              { :f1 => "1" }
            ],

            [ Headers.new( :fields => { :f1 => 1, :f2 => true, :f3 => "3" } ),
              { :f1 => "1", :f2 => "true", :f3 => "3" }
            ],

            [ Headers.new( :fields => { :f1 => MingleIdentifier.get( "id1" ) } ),
              { :f1 => "id1" }
            ],

        ].each do |pair|
            expct = MingleSymbolMap.create( pair[ 1 ] )
            flds = pair[ 0 ].fields
            flds.each_pair { |k, v| assert( v.is_a?( MingleString ) ) }
            ModelTestInstances.assert_equal( expct, pair[ 0 ].fields )
        end
    end

    private
    def assert_headers_rt( hdrs, copies = 1 )
 
        io = new_string_io
        enc = Encoder.new( io )

        copies.times { enc.write_headers( hdrs ) }

        io.pos = 0
        dec = Decoder.new( io )
        
        copies.times do
            hdrs2 = dec.read_headers
            ModelTestInstances.assert_equal( hdrs.fields, hdrs2.fields )
        end
    end

    public
    def test_rt_nontrivial_headers

        hdrs = Headers.as_headers( :f1 => "val1", :f2 => true, :f3 => 3 )
        [ 1, 10 ].each { |copies| assert_headers_rt( hdrs, copies ) }
    end

    public 
    def test_rt_empty_headers
        
        [ 1, 10 ].each do |copies|
            assert_headers_rt( Headers.as_headers( {} ), copies )
        end
    end

    private
    def assert_exception( input, ex_cls, msg_expct )

        ex = assert_raised( ex_cls ) do
            str = opt_force_encoding( input, "binary" )
            Decoder.new( StringIO.new( str ) ).read_headers
        end

        assert_equal( msg_expct, ex.message )
    end

    public
    def test_bad_headers_version
        
        assert_exception(
            "\x00\x02\x02\x02", 
            InvalidVersionError,
            'Invalid headers :version => 0x02020200 (expected 0x00000001)'
        )
    end

    public
    def test_bad_type_code
 
        assert_exception(
            "\x01\x00\x00\x00\x0f\x00\x00\x00",
            InvalidTypeCodeError,
            "Unknown type :code => 0x0000000f"
        )
    end
end

end
end
