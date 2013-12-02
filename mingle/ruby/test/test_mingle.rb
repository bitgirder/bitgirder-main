require 'bitgirder/core'

require 'bitgirder/testing'
Testing = BitGirder::Testing

require 'bitgirder/io'

require 'mingle'
require 'mingle/test-support'

module Mingle

class MingleTests

    include TestClassMixin

    @@mti = ModelTestInstances # shorthand

    def test_identifier_format_of
       
        { :lc_underscore => :lc_underscore,
          :lc_camel_capped => "lc-camel-capped",
          :lc_hyphenated => MingleIdentifier.get( "lc_hyphenated" )
        }.each_pair do |expct, val|
            assert_equal( expct, MingleIdentifier.as_format_name( val ) )
        end
    end

    def test_identifier_format
        
        id = MingleIdentifier.parse( "p1-p2" )

        assert_equal( "p1-p2", id.format( :lc_hyphenated ) )
        assert_equal( "p1_p2", id.format( :lc_underscore ) )
        assert_equal( "p1P2", id.format( :lc_camel_capped ) )

        assert_raised( Exception ) { id.format( :blah ) }
    end

    def test_identifier_to_sym
        
        assert_equal( :id1, MingleIdentifier.parse( "id1" ).to_sym )

        # Also important to make sure that we use lc_underscore in to_sym
        assert_equal( :id1_stuff, MingleIdentifier.parse( "id1_stuff" ).to_sym )
    end

    def test_identifier_ext_form

        id = MingleIdentifier.get( "shouldHaveHyphen" )

        assert_equal( "should-have-hyphen", id.external_form )
        assert_equal( id.external_form, id.to_s )
    end

    def test_identifier_empty_part_fails
        
        msg = "Empty id part at index 1"

        assert_raised( msg, StandardError ) do
            MingleIdentifier.send( :new, :parts => [ "ok", "", "bad" ] )
        end
    end

    def assert_symbol_map_access( msm, key, expct )
        
        [ :[], :expect ].each do |method|
            @@mti.assert_equal( expct, msm.send( method, key ) )
        end
    end

    def test_symbol_map_drops_mingle_null
        
        m = MingleSymbolMap.create( :a => 1, :b => MingleNull::INSTANCE )
        assert_equal( 1, m.get_int( :a ) )
        assert_equal( 1, m.size )
    end

    def test_symbol_map_size
        
        m = MingleSymbolMap.create( :val1 => "val1" )
        assert_equal( 1, m.size )
        assert_false( m.empty? )

        m = MingleSymbolMap.create()
        assert_equal( 0, m.size )
        assert( m.empty? )
    end

    def test_symbol_map_get_and_expect
        
        expct = MingleString.new( "hello" )
        m = MingleSymbolMap.create( :key1 => expct )

        keys = [ MingleIdentifier.get( "key1" ), :key1, "key1" ]

        10.times do |i|
            keys.each { |k| assert_symbol_map_access( m, k, expct ) }
        end

        # Assert that the vals_by_sym cache didn't create extraneous entries
        assert_equal( 1, m.instance_variable_get( :@vals_by_sym ).size )
        
        [ MingleIdentifier.get( "key2" ), :key2, "key2" ].each do |k|
            assert_raised( MingleSymbolMap::NoSuchKeyError ) { m.expect( k ) }
        end
    end

    def test_symbol_map_iters
        
        m = MingleSymbolMap.create( :a => 1, :b => 2 )

        expct = [ 
            [ MingleIdentifier.get( "a" ), MingleInt32.new( 1 ) ], 
            [ MingleIdentifier.get( "b" ), MingleInt32.new( 2 ) ]
        ]
        chk = lambda { |arr| assert( arr == expct || arr = expct.reverse ) }

        arr = []
        m.each_pair { |k, v| arr << [ k, v ] }
        chk.call( arr )

        arr = []
        m.each { |pair| arr << pair }
        chk.call( arr )
    end        

    def test_symbol_map_to_hash
        
        m = MingleSymbolMap.create( :f1 => "val1" )
        h = m.to_hash
        assert_equal( 1, h.size )
        assert_equal( 
            h[ MingleIdentifier.get( :f1 ) ], MingleString.new( "val1" ) )
    end

    RubyVersions.when_19x do
    
        def test_mingle_buffer_encodings
            
            b1 = MingleBuffer.new( "a".encode( "binary" ) )
            assert_equal( 
                b1, MingleBuffer.new( "a".encode( "utf-8" ), :copy ) )
    
            s = "a".encode( "utf-8" )
            b2 = MingleBuffer.new( s, :in_place )
            assert_equal( b1, b2 )
            assert( s.equal?( b2.buf ) )
    
            assert_equal( 
                b1, MingleBuffer.new( "a".encode( "binary" ), :none ) )
            
            assert_raised( MingleBuffer::EncodingError ) do |e|
                MingleBuffer.new( "a".encode( "utf-8" ) )
            end
        end
    end

    # Test comparison function and coverage that Comparable is included
    def assert_comparisons( small, large, small_dup = nil )
        
        assert( ( small <=> large ) < 0 )
        assert( ( large <=> small ) > 0 )

        assert( ( small <=> small_dup ) == 0 ) if small_dup

        assert( small < large )
    end

    def test_mingle_string_compare
 
        s1 = MingleString.new( "stuff" )
        s2 = MingleString.new( "stuff" )

        assert( s1 == s2 )
        assert( [ s1 ].include?( s2 ) )

        [ s1, s2 ].each { |k| assert( { s1 => 1 }.key?( k ) ) }

        assert_comparisons( s1, MingleString.new( "xxx" ), s2 )
    end

    def test_mingle_string_ruby_compat_ops
        
        s = MingleString.new( "hello" )

        # Test a handful of [] ops. Really one would suffice given the current
        # impl since we just forward to the underlying ruby string, but if we
        # later change that impl these tests should be sure to cover those
        # changes
        assert_equal( "hel", s[ 0 .. 2 ] )
        assert_equal( RUBY_VERSION >= "1.9" ? "e" : ?e, s[ 1 ] )
        assert_equal( "ll", s[ /l+/ ] )

        assert_equal( 2, s =~ /l/ )

        assert_equal( :"hello", s.to_sym )

        assert_equal( %w{ h llo }, s.split( /e/ ) )
    end

    def test_symbol_map_accessors
        
        ts = MingleTimestamp.now

        bin = RubyVersions.when_19x( "this is binary data" ) do |s|
            s.force_encoding( "binary" )
        end

        m =
            MingleSymbolMap.create(
                :a_string => "hello",
                :an_int => "1234",
                :a_decimal => "-12.34",
                :a_true => "true",
                :a_false => false,
                :a_buffer => BitGirder::Io.strict_encode64( bin ),
                :a_timestamp => ts.rfc3339,
                :a_list => [ 1, 2 ],
                :a_map => { :a => 1 }
            )
        
        assert_equal( "hello", m.expect_string( :a_string ) )
        assert_nil( m.get_string( :a_nonexistent_string ) )
        assert_equal( 1234, m.expect_mingle_int64( :an_int ).num )
        assert_equal( 1234, m.get_mingle_int64( :an_int ).num )
        assert_equal( 1234, m.expect_int( :an_int ) )
        assert_equal( 1234, m.get_int( :an_int ) )
        assert_equal( -12.34, m.expect_mingle_float64( :a_decimal ).num )
        assert_equal( true, m.expect_mingle_boolean( :a_true ).as_boolean )
        assert_equal( true, m.get_boolean( :a_true ) )
        assert_equal( false, m.get_boolean( :a_false ) )
        assert_nil( m.get_boolean( :not_present ) )

        assert_equal( 
            MingleBuffer.new( bin ), m.expect_mingle_buffer( :a_buffer ) )
        
        assert_equal( ts, m.expect_mingle_timestamp( :a_timestamp ) )
        assert_equal( ts, m.expect_timestamp( :a_timestamp ) )

        assert_equal( 
            MingleList.new( [ 1, 2 ] ), m.expect_mingle_list( :a_list ) )

        assert_equal(
            MingleSymbolMap.create( :a => 1 ), 
            m.expect_mingle_symbol_map( :a_map ) )
    end

    def test_as_mingle_value_type_failures
        [ :as_mingle_list, :as_mingle_symbol_map ].each do |m|
            assert_raised( TypeError ) do
                MingleModels.send( m, MingleString.new( "s" ) )
            end
        end
    end

    def test_mingle_symbol_map_has_compatibility
        
        msm = MingleSymbolMap.create( 
            :key1 => "val1", :key2 => "val2", :key3 => "val3" )

        assert_equal( "val2|val1", msm.values_at( :key2, :key1 ).join( "|" ) )
    end

    def test_mingle_symbol_map_iterators
        
        msm = MingleSymbolMap.create( :key1 => "key1", :key2 => 2 )

        h = {}
        msm.each_pair { |k, v| h[ k ] = v }
        @@mti.assert_equal( msm, MingleSymbolMap.create( h ) )

        h = {}
        msm.each { |pair| h[ pair[ 0 ] ] = pair[ 1 ] }
        @@mti.assert_equal( msm, MingleSymbolMap.create( h ) )

        h = msm.inject( {} ) { |acc, pair| acc.merge( pair[ 0 ] => pair[ 1 ] ) }
        @@mti.assert_equal( msm, MingleSymbolMap.create( h ) )
    end

    def test_mingle_service_request_from_syms
        
        expct = 
            MingleServiceRequest.new(
                :namespace => "a:namespace@v1",
                :service => "a-service",
                :operation => "an-operation" )
        
        actual =
            MingleServiceRequest.new(
                :namespace => :"a:namespace@v1",
                :service => :a_service,
                :operation => :an_operation )
        
        @@mti.assert_equal( expct, actual )
    end

    # Will add more as needed
    def test_as_mingle_value
        
        t1 = Time.now
        l1 = MingleList.new( [ 1, 2 ] )
        s1 = MingleSymbolMap.create( :a => 1 )

        [
            { :expect => MingleTimestamp.new( t1 ), :value => t1 },
            { :expect => MingleString.new( "hello" ), :value => "hello" },
            { :expect => MingleString.new( "hello" ), :value => :"hello" },

            # Check as_value when presented with an object which needs no
            # conversion
            { :expect => MingleString.new( "hello" ),
              :value => MingleString.new( "hello" ) },
            
            # Check default num handling and boundaries
            { :expect => MingleInt64.new( 2 ** 40 ), :value => 2 ** 40 },
            { :expect => MingleInt32.new( 2 ** 20 ), :value => 2 ** 20 },

            { :expect => MingleInt32.new( ( 2 ** 31 ) - 1 ), 
              :value => ( 2 ** 31 ) - 1 },
            
            { :expect => MingleInt32.new( -( 2 ** 31 ) ),
              :value => -( 2 ** 31 ) },
            
            { :expect => MingleInt64.new( ( 2 ** 63 ) - 1 ),
              :value => ( 2 ** 63 ) - 1 },
            
            { :expect => MingleInt64.new( -( 2 ** 63 ) ),
              :value => -( 2 ** 63 ) },
            
            { :value => 2 ** 100,
              :error => {
                :message =>
                    "Number is out of range for mingle integer types: " \
                    "#{2 ** 100}"
              }
            },

            # Various checks of container conversions and handling of nested
            # values; also checks of specific coercion methods (:as_mingle_list,
            # :as_mingle_symbol_map)
            { :expect => l1, :value => [ 1, 2 ] },
            { :expect => l1, :value => l1 },
            { :expect => l1, :value => [ 1, MingleInt32.new( 2 ) ] },
            { :expect => l1, :value => l1, :method => :as_mingle_list },
            { :expect => s1, :value => { :a => 1 } },
            { :expect => s1, :value => { "a" => 1 } },
            { :expect => s1, :value => { :a => MingleInt32.new( 1 ) } },
            { :expect => s1, :value => s1 },
            { :expect => s1, :value => s1, :method => :as_mingle_symbol_map }

        ].each do |t|
            
            meth = ( t[ :method ] || :as_mingle_value )
            inv = lambda { MingleModels.send( meth, t[ :value ] ) }
            
            if err = t[ :error ]
                err_cls = err[ :error_class ] || Exception
                assert_raised( err[ :message ], err_cls ) { inv.call }
            else
                # Check exact equality -- not with conversions like
                # ModelTestInstances.assert_equal
                assert_equal( t[ :expect ], inv.call )
            end
        end
    end

    def test_mingle_list_methods
        
        l = MingleList.new( [ 1, 2, "hello", 3, false, 4 ] )
        
        assert_equal( 6, l.size )
        assert( ! l.empty? )

        # Test enumerable included
        assert_equal( 
            10, 
            l.select { |v| v.is_a?( MingleInt32 ) }.
              inject( 0 ) { |s, i| s + i.to_i } 
        )

        assert_equal( "1|2|hello|3|false|4", l.join( "|" ) )

        assert( MingleList.new( [] ).empty? )
    end

    def test_mingle_list_add
        
        l1 = MingleList.new( [ 1, 2 ] )

        assert_equal( 
            "1|2|3|4", ( l1 + MingleList.new( [ 3, 4 ] ) ).join( "|" ) )

        assert_equal( "1|2|3|4", ( l1 + [ 3, 4 ] ).join( "|" ) )
    end

    def test_mingle_timestamp_integral_factories
        
        t1 = MingleTimestamp.rfc3339( "2011-03-07T21:45:04.123456Z" )

        t2 = MingleTimestamp.from_seconds( t1.time.to_i )
        assert_equal( t1.rfc3339[ 0 .. -11 ] + ( "0" * 9 ) + "Z", t2.rfc3339 )

        time_ms = ( t1.time.to_i * 1000 ) + ( t1.time.usec / 1000 )
        t3 = MingleTimestamp.from_millis( time_ms )
        assert_equal( t1.rfc3339[ 0 .. -8 ] + ( "0" * 6 ) + "Z", t3.rfc3339 )
    end

    # Assumes proper behavior of MingleTimestamp.from_seconds(), as tested
    # elsewhere in this file
    def test_mingle_timestamp_comparison
        
        t1 = MingleTimestamp.from_seconds( 1000000 )
        t2 = MingleTimestamp.from_seconds( 1000001 )
        t3 = MingleTimestamp.from_seconds( 1000001 )

        assert( ( t1 <=> t2 ) < 0 )
        assert( ( t2 <=> t1 ) > 0 )
        assert( ( t2 <=> t3 ) == 0 )
        assert( t1 < t2 )
        assert( t2 > t1 )
    end

    def test_mingle_timestamp_rfc3339_parse_fail
        
        str = "2001-01-01T01-01-99.000Z"

        assert_raised( 
            "invalid date: \"#{str}\"", MingleTimestamp::Rfc3339FormatError ) do
        
            MingleTimestamp.rfc3339( str )
        end
    end

    def test_mingle_int64_basic
        
        i1 = MingleInt64.new( 1 )
        i2 = MingleInt64.new( 2 )

        assert_equal( 1, i1.to_i )
        assert_equal( "1", i1.to_s )

        assert_comparisons( i1, i2, MingleInt64.new( 1 ) )

        # Test that floats are truncated by MingleInt64 constructor
        assert_equal( 2, MingleInt64.new( 2.2 ).num )
    end

    # Also tests mixed integral/decimal comparison
    def test_mingle_float64_basic
        
        d1 = MingleFloat64.new( 1.1 )
        d2 = MingleFloat64.new( 2.3 )

        assert_equal( 1.1, d1.to_f )
        assert_equal( "1.1", d1.to_s )

        assert_comparisons( d1, d2, MingleFloat64.new( 1.1 ) )

        i1 = MingleInt64.new( 2 )
        assert_comparisons( d1, i1 )
        assert_comparisons( i1, d2 )
    end

    def test_hash_and_eql_impls
        
        t = Time.now
        
        arr = 
            Array.new( 2 ) do |i|
                {
                    MingleString.new( "str1" ) => "val1",
                    MingleInt64.new( 1 ) => "val2",
                    MingleInt32.new( 1 ) => "val2",
                    MingleUint32.new( 1 ) => "val2",
                    MingleUint64.new( 1 ) => "val2",
                    MingleFloat64.new( 1.1 ) => "val3",
                    MingleFloat32.new( 1.1 ) => "val3",
                    MingleTimestamp.new( t ) => "val4",
                    MingleBoolean::TRUE => "val5",
                    MingleBoolean::FALSE => "val6",
                    MingleBuffer.new( "abc", :in_place ) => "val7"
                }
            end
        
        arr[ 0 ].each do |pair| 
            assert_equal( pair[ 1 ], arr[ 1 ][ pair[ 0 ] ] )
        end
    end

    # Simple regression for an early bug which over-zealously disallowed
    # creation of struct without fields
    def test_empty_struct_allowed
 
        ms = MingleStruct.new( :type => :"test@v1/Type" )
        assert( ms.fields.empty? )
    end

    def test_mingle_models_raise
        
        mg_ex = MingleStruct.new(
            :type => :"test@v1/Err1",
            :fields => { :message => "test-message" }
        )

        begin
            MingleModels.raise_as_ruby_error( mg_ex )
        rescue GenericRaisedMingleError => e
            assert_equal( mg_ex.type, e.type )
            assert_equal( "#{mg_ex.type}: #{mg_ex[ :message ]}", e.message )
        end
    end

    def test_void_service_response
        
        resp = MingleServiceResponse.create_success( MingleNull::INSTANCE )
        assert_nil( resp.result )
    end
end

#class ParseErrorExpectation < BitGirderClass
#    bg_attr :col, :processor => :integer, :validation => :nonnegative
#    bg_attr :message
#end
#
#class RestrictionErrorExpectation < BitGirderClass
#    bg_attr :message
#end
#
#PARSE_OVERRIDES = {
#
#    :string => {
#        "\"a\\udd1e\\ud834\"" => 'Trailing surrogate with no lead: \uDD1E'
#    },
#
#    :identifier => {
#        'giving-mixedMessages' => 
#            'Illegal start of identifier part: "M" (0x4D)'
#    },
#
#    :declared_type_name => { 
#        'Bad$Char' => 'Unrecognized token: "$" (0x24)',
#        'Bad_Char' => 'Unrecognized token: "_" (0x5F)'
#    },
#
#    :qualified_type_name => {
#        "ns1@v1/T1#T2" => 'Unrecognized token: "#" (0x23)'
#    },
#
#    :type_reference => {
#            
#        "mingle:core@v1/Float32~[1,2)" => {
#            :external_form => 'mingle:core@v1/Float32~[1.0,2.0)'
#        },
#        "Float32~[1,2)" => {
#            :external_form => 'mingle:core@v1/Float32~[1.0,2.0)'
#        },
#        "mingle:core@v1/Timestamp~[\"2012-01-01T12:00:00Z\",\"2012-01-02T12:00:00Z\"]" => {
#            :external_form => 'mingle:core@v1/Timestamp~["2012-01-01T12:00:00.000000000Z","2012-01-02T12:00:00.000000000Z"]'
#        },
#        "Timestamp~[\"2012-01-01T12:00:00Z\",\"2012-01-02T12:00:00Z\"]" => {
#            :external_form => 'mingle:core@v1/Timestamp~["2012-01-01T12:00:00.000000000Z","2012-01-02T12:00:00.000000000Z"]'
#        },
#        "mingle:core@v1/String ~= \"sdf\"" => 'Unrecognized token: "=" (0x3D)',
#        "mingle:core@v1/String~" => "Expected type restriction but found: END",
#        "Int32~[1,3}" => {
#            :err_message => 'Unexpected char in integer part: "}" (0x7D)',
#            :err_col => 11
#        },
#        "Int32~[-\"abc\",2)" => {
#            :err_message => 'Unexpected char in integer part: "\"" (0x22)',
#            :err_col => 9
#        },
#        "Int32~[--3,4)" => {
#            :err_message => 'Number has empty or invalid integer part',
#            :err_col => 8
#        },
#        "String~\"ab[a-z\"" => 
#            'Invalid regex: premature end of ' +
#            ( RUBY_VERSION >= "1.9" || RubyVersions.jruby? ? 
#                "char-class" : "regular expression" ) +
#            ': /ab[a-z/',
#        
#        "mingle:core@v1/String~=\"sdf\"" => 'Unrecognized token: "=" (0x3D)',
#        "Timestamp~[\"2001-0x-22\",)" => 'invalid date: "2001-0x-22"'
#    }
#}
#
#class ParseTest < BitGirderClass
#
#    include Testing::AssertMethods
#
#    bg_attr :test_type, :processor => :symbol
#    bg_attr :input
#    bg_attr :external_form, :required => false
#    bg_attr :expect, :required => false
#    bg_attr :error, :required => false
#
#    private
#    def get_override
#
#        res = ( PARSE_OVERRIDES[ @test_type ] || {} )[ @input ] || {}
#        res = { :err_message => res } if res.is_a?( String )
#
#        res
#    end
#
#    private
#    def expect_token( cls )
#        
#        lx = MingleLexer.as_instance( @input )
#        tok, loc = lx.expect_token
#        raise "Expected #{cls} but got #{tok.class}" unless tok.is_a?( cls )
#        raise "Trailing input" unless lx.eof?
#
#        tok
#    end
#
#    private
#    def call_parse
#        
#        case @test_type
#        when :string then expect_token( StringToken )
#        when :number then ParsedNumber.parse( @input )
#        when :identifier then MingleIdentifier.parse( @input )
#        when :namespace then MingleNamespace.parse( @input )
#        when :declared_type_name then DeclaredTypeName.parse( @input )
#        when :qualified_type_name then QualifiedTypeName.parse( @input )
#        when :identified_name then MingleIdentifiedName.parse( @input )
#        when :type_reference then MingleTypeReference.parse( @input )
#        else raise "Unhandled test type: #@test_type"
#        end
#    end
#
#    private
#    def assert_num_roundtrip( n )
#        
#        n_str = n.external_form
#        lx = MingleLexer.as_instance( n_str )
#        
#        n2, _ = ParsedNumber.parse( n_str )
#        assert_equal( n, n2 )
#    end
#
#    private
#    def assert_result( res )
#
#        assert_equal( @expect, res )
#        assert_equal( @expect.hash, res.hash )
#        assert( @expect.eql?( res ) )
#
#        unless @external_form.empty?
#            expct = get_override[ :external_form ] || @external_form
#            assert_equal( expct, res.external_form ) 
#        end
#
#        assert_num_roundtrip( res ) if res.is_a?( ParsedNumber )
#    end
#
#    private
#    def get_message_expect
# 
#        msg_expct = get_override[ :err_message ]
#        msg_expct ||= @error.message.gsub( /U\+00([[:xdigit:]]{2})/, '0x\1' )
#    end
#
#    private
#    def assert_parse_error( pe )
#        
#        assert_equal( get_message_expect, pe.err )
#
#        assert_equal( 1, pe.loc.line )
#        assert_equal( get_override[ :err_col ] || @error.col, pe.loc.col )
#    end
#
#    private
#    def assert_restriction_error( re )
#        assert_equal( get_message_expect, re.message )
#    end
#
#    private
#    def assert_error( e )
#        
#        case 
#        when e.is_a?( MingleParseError ) && 
#             @error.is_a?( ParseErrorExpectation )
#            assert_parse_error( e )
#        when e.is_a?( RestrictionTypeError ) &&
#             @error.is_a?( RestrictionErrorExpectation )
#            assert_restriction_error( e )
#        else raise e
#        end
#    end
#
#    public
#    def call
# 
#        begin
#
##            code( "Parsing: #{self.inspect}" )
#            res = call_parse
#
#            raise "Got #{res}, but expected error #{@error.inspect}" if @error
#            assert_result( res )
#
#        rescue => e
#            assert_error( e )
#        end
#    end
#end
#
#class ParseTests < Testing::TestHolder
#
#    FILE_VERSION1 = 0x01
#    
#    # int8
#    FLD_CODES = {
#        0x01 => :test_type,
#        0x02 => :input,
#        0x03 => :expect,
#        0x04 => :error,
#        0x05 => :end,
#        0x06 => :external_form
#    }
#
#    # int8
#    VAL_TYPES = {
#        0x01 => :identifier,
#        0x02 => :namespace,
#        0x03 => :declared_type_name,
#        0x05 => :qualified_type_name,
#        0x06 => :identified_name,
#        0x07 => :regex_restriction,
#        0x08 => :range_restriction,
#        0x09 => :atomic_type_reference,
#        0x0a => :list_type_reference,
#        0x0b => :nullable_type_reference,
#        0x0c => :nil,
#        0x0d => :int32,
#        0x0e => :int64,
#        0x0f => :float32,
#        0x10 => :float64,
#        0x11 => :string,
#        0x12 => :timestamp,
#        0x13 => :boolean,
#        0x14 => :parse_error,
#        0x15 => :restriction_error,
#        0x16 => :string_token,
#        0x17 => :numeric_token,
#        0x18 => :uint32,
#        0x19 => :uint64
#    }
#    
#    # int8
#    ELT_TYPES = { 0x00 => :file_end, 0x01 => :parse_test }
#
#    def unread8( i )
#
#        if @unread8
#            raisef "Attempt to unread %0x when %0x is already set", @unread8, i
#        else
#            @unread8 = i
#        end
#    end
#
#    def read_int8( rd )
#        ( @unread8 || rd.read_int8 ).tap { @unread8 = nil }
#    end
#
#    def read_named8( rd, map, err_nm, opts = {} )
#        
#        i = read_int8( rd )
#        unread8( i ) if opts[ :unread ]
#
#        map[ i ] or raisef "Unknown #{err_nm} value: 0x%02x", i
#    end
# 
#    def peek_named8( *argv )
#        read_named8( *( argv + [ :unread => true ] ) ) 
#    end
#
#    def read_val_type( rd )
#        read_named8( rd, VAL_TYPES, "value type" )
#    end
#
#    def peek_val_type( rd )
#        peek_named8( rd, VAL_TYPES, "val type" )
#    end
#
#    def read_header( rd )
#        
#        unless ( ver = rd.read_int32 ) == FILE_VERSION1
#            raisef "Unexpected file version: %04x", ver
#        end
#    end
#
#    def expect_value_type( rd, nm )
#        
#        unless ( typ = read_val_type( rd ) ) == nm
#            raise "Expected #{nm}, saw: #{typ}"
#        end
#    end
#
#    def read_prim_value( rd )
#        
#        case vt = read_val_type( rd )
#        when :nil then nil
#        when :boolean then MingleBoolean.for_boolean( rd.read_bool )
#        when :int32 then MingleInt32.new( rd.read_int32 )
#        when :int64 then MingleInt64.new( rd.read_int64 )
#        when :uint32 then MingleUint32.new( rd.read_uint32 )
#        when :uint64 then MingleUint64.new( rd.read_uint64 )
#        when :float32 then MingleFloat32.new( rd.read_float32 )
#        when :float64 then MingleFloat64.new( rd.read_float64 )
#        when :timestamp then MingleTimestamp.rfc3339( rd.read_utf8 )
#        when :string then MingleString.new( rd.read_utf8 )
#        else raise "Unhandled primitive type: #{vt}"
#        end
#    end
#
#    def expect_bool( rd )
#        
#        expect_value_type( rd, :boolean )
#        rd.read_bool
#    end
#
#    def read_string_token( rd )
#        
#        expect_value_type( rd, :string_token )
#        StringToken.new( :val => rd.read_utf8 )
#    end
#
#    def read_numeric_token( rd )
#        
#        expect_value_type( rd, :numeric_token )
#
#        ParsedNumber.new(
#            :negative => expect_bool( rd ),
#            :num => NumericToken.new(
#                :int => rd.read_utf8,
#                :frac => rd.read_utf8,
#                :exp => rd.read_utf8,
#                :exp_char => rd.read_utf8
#            )
#        )
#    end 
#
#    def read_identifier( rd )
#
#        expect_value_type( rd, :identifier )
#
#        parts = Array.new( rd.read_int32 ) { rd.read_utf8 }
#        MingleIdentifier.send( :new, :parts => parts )
#    end
#
#    def read_identifiers( rd )
#        Array.new( rd.read_int32 ) { read_identifier( rd ) }
#    end
#
#    def read_namespace( rd )
# 
#        expect_value_type( rd, :namespace )
#
#        MingleNamespace.send( :new,
#            :parts => read_identifiers( rd ),
#            :version => read_identifier( rd )
#        )
#    end
#
#    def read_declared_type_name( rd )
#        
#        expect_value_type( rd, :declared_type_name )
#        DeclaredTypeName.send( :new, rd.read_utf8 )
#    end
#
#    def read_qualified_type_name( rd )
#        
#        expect_value_type( rd, :qualified_type_name )
#
#        QualifiedTypeName.new(
#            :namespace => read_namespace( rd ),
#            :name => read_declared_type_name( rd )
#        )
#    end
#
#    def read_identified_name( rd )
#        
#        expect_value_type( rd, :identified_name )
#        
#        MingleIdentifiedName.new(
#            :namespace => read_namespace( rd ),
#            :names => read_identifiers( rd )
#        )
#    end
#
#    def read_type_name( rd )
#        
#        case vt = peek_val_type( rd )
#        when :declared_type_name then read_declared_type_name( rd )
#        when :qualified_type_name then read_qualified_type_name( rd )
#        else raise "Unexpected type name: #{vt}"
#        end
#    end
#
#    def read_regex_restriction( rd )
#        RegexRestriction.new( :ext_pattern => rd.read_utf8 )
#    end
#
#    def read_range_restriction( rd )
#        
#        RangeRestriction.new(
#            :min_closed => expect_bool( rd ),
#            :min => read_prim_value( rd ),
#            :max => read_prim_value( rd ),
#            :max_closed => expect_bool( rd )
#        )
#    end
#
#    def read_type_restriction( rd )
#        
#        case vt = read_val_type( rd )
#        when :nil then nil
#        when :regex_restriction then read_regex_restriction( rd )
#        when :range_restriction then read_range_restriction( rd )
#        else raise "Unrecognized restriction type: #{vt}"
#        end
#    end
#
#    def read_atomic_type_reference( rd )
#        
#        expect_value_type( rd, :atomic_type_reference )
#
#        AtomicTypeReference.send( :new,
#            :name => read_type_name( rd ),
#            :restriction => read_type_restriction( rd )
#        )
#    end
#
#    def read_list_type_reference( rd )
#        
#        expect_value_type( rd, :list_type_reference )
#
#        ListTypeReference.send( :new,
#            :element_type => read_type_reference( rd ),
#            :allows_empty => read_prim_value( rd ).to_bool
#        )
#    end
#
#    def read_nullable_type_reference( rd )
#        
#        expect_value_type( rd, :nullable_type_reference )
#
#        NullableTypeReference.send( :new, :type => read_type_reference( rd ) )
#    end
#
#    def read_type_reference( rd )
#
#        case vt = peek_val_type( rd )        
#        when :atomic_type_reference then read_atomic_type_reference( rd )
#        when :list_type_reference then read_list_type_reference( rd )
#        when :nullable_type_reference then read_nullable_type_reference( rd )
#        else raise "Unrecognized type reference type: #{vt}"
#        end
#    end
#        
#    def read_expect_val( rd )
#
#        case vt = peek_val_type( rd )
#        when :string_token then read_string_token( rd )
#        when :numeric_token then read_numeric_token( rd )
#        when :identifier then read_identifier( rd )
#        when :namespace then read_namespace( rd )
#        when :declared_type_name then read_declared_type_name( rd )
#        when :qualified_type_name then read_qualified_type_name( rd )
#        when :identified_name then read_identified_name( rd )
#
#        when :atomic_type_reference, 
#             :list_type_reference,
#             :nullable_type_reference 
#            read_type_reference( rd )
#
#        else raise "Unhandled value type: #{vt}"
#        end
#    end
#
#    def read_parse_error( rd )
#        
#        ParseErrorExpectation.new(
#            :col => rd.read_int32,
#            :message => rd.read_utf8
#        )
#    end
#
#    def read_restriction_error( rd )
#        RestrictionErrorExpectation.new( :message => rd.read_utf8 )
#    end
#
#    def read_error( rd )
#        
#        case err_typ = read_named8( rd, VAL_TYPES, "val type" )
#        when :parse_error then read_parse_error( rd )
#        when :restriction_error then read_restriction_error( rd )
#        else raise "Unhandled error type: #{err_typ}"
#        end
#    end
#
#    def read_next( rd )
# 
#        attrs = {}
#
#        until ( fc = read_named8( rd, FLD_CODES, "field code" ) ) == :end
#            attrs[ fc ] = case fc
#                when :test_type then rd.read_utf8.gsub( '-', '_' )
#                when :input, :external_form then rd.read_utf8
#                when :expect then read_expect_val( rd )
#                when :error then read_error( rd )
#                else raise "Unhandled attribute: #{fc}"
#            end
#        end
#
#        ParseTest.new( attrs )
#    end
#
#    def read_parse_test( rd )
#        
#        case fc = read_named8( rd, ELT_TYPES, "element type" )
#        when :parse_test then read_next( rd )
#        when :file_end then nil
#        else raise "Unhandled: #{fc}"
#        end
#    end
#
#    # Impl note: we first pull the entire test file into a buffer and then wrap
#    # it with StringIO, since using an IO instance directly gives pretty
#    # terrible performance -- the read loop below, as of this writing, took 10s
#    # using io directly, but only .04s when wrapped in a StringIO. Perhaps we'll
#    # later find/build some transparent buffered reader, but for now this will
#    # do.
#    def read_parse_tests
# 
#        res = []
#
#        File.open( Testing.find_test_data( "parser-tests.bin" ) ) do |io|
#
#            sz = BitGirder::Io.fsize( io )
#            buf = BitGirder::Io.read_full( io, sz )
#            str_io = StringIO.new( buf )
#            rd = BitGirder::Io::BinaryReader.new_le( :io => str_io )
#
#            read_header( rd )
#
#            start = Time.now
#            until ( test = read_parse_test( rd ) ) == nil
#                res << test
#            end
#        end
#
#        res
#    end
#
#    # In addition to testing parser coverage, these tests also give us coverage
#    # of external form and equality
#    def test_parse
#        read_parse_tests.each { |t| t.call }
#    end
#end

end
