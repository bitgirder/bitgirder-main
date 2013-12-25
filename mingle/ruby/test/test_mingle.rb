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

        t4_sec, t4_nsec = 1, 123456789
        t4 = MingleTimestamp.from_seconds_and_nanos( t4_sec, t4_nsec )
        assert_equal( [ t4_sec, t4_nsec ], [ t4.sec, t4.nsec ] )
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
end

end
