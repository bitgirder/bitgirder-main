require 'mingle'

require 'bitgirder/core'
require 'bitgirder/io/testing'

module Mingle

module ModelTestInstances

    require 'set'

    include BitGirder::Core
    extend BitGirderMethods

    extend BitGirder::Io::Testing

    module_function

    def fail_assert( msg, path )
        raise "#{path.format}: #{msg}"
    end

    def assert( res, msg, path )
        fail_assert( msg, path ) unless res
    end

    def assert_equal_objects( expct, actual, path )

        assert( 
            expct == actual,
            "Objects are not equal: expected #{expct.inspect} " \
            "(#{expct.class}) but got #{actual.inspect} (#{actual.class})", 
            path )
    end

    def assert_equal_with_coercion( expct, actual, path )
        
        actual = MingleModels.as_mingle_instance( actual, expct.class )
        assert_equal_objects( expct, actual, path )
    end

    def assert_equal_type_ref( expct, actual, path )

        assert_equal_objects( expct.type, actual.type, path.descend( "type" ) )
    end

    def key_set_to_s( set )
        set.map { |sym| sym.external_form }
    end

    def assert_equal_symbol_map( m1, m2, path )
        
        k1 = m1.keys
        k2 = m2.keys

        assert( 
            Set.new( k1 ) == Set.new( k2 ),
            "Symbol map key sets are not :equal => " \
            "#{key_set_to_s( k1 )} != #{key_set_to_s( k2 )}", path )
        
        k1.each { |k| assert_equal( m1[ k ], m2[ k ], path.descend( k ) ) }
    end

    def assert_equal_struct( expct, actual, path )
        
        assert_equal_objects( expct.class, actual.class, path )
        assert_equal_type_ref( expct, actual, path )
        
        assert_equal_symbol_map( expct.fields, actual.fields, path )
    end

    def assert_equal_buffers( expct, actual, path )
        
        assert_equal_objects( 
            expct.buf, MingleModels.as_mingle_buffer( actual ).buf, path )
    end

    def assert_equal_timestamps( expct, actual, path )
        
        assert_equal_objects(
            expct, MingleModels.as_mingle_timestamp( actual ), path )
    end

    def assert_equal_enum( expct, actual, path )
        
        case actual

            when MingleEnum
                assert_equal_objects( expct.type, actual.type, path )
                assert_equal_objects( expct.value, actual.value, path )

            when MingleString
                assert_equal_objects(
                    expct.value, MingleIdentifier.get( actual ), path )
            
            else fail_assert( "Unhandled enum value: #{val} (#{val.class})" )
        end
    end

    def assert_equal_lists( expct, actual, path )

        path = path.start_list

        expct.zip( actual ) { |pair|

            assert_equal( pair.shift, pair.shift, path )
            path = path.next
        }
    end

    def assert_equal( expct, 
                      actual, 
                      path = BitGirder::Core::ObjectPath.get_root( "expct" ) )
        case expct

            when MingleNull, NilClass 
                assert_equal_objects( expct, actual, path )

            when MingleString, MingleBoolean, MingleInt64, MingleInt32,
                 MingleUint32, MingleUint64, MingleFloat64, MingleFloat32
                assert_equal_with_coercion( expct, actual, path )

            when MingleStruct then assert_equal_struct( expct, actual, path )

            when MingleBuffer then assert_equal_buffers( expct, actual, path )

            when MingleTimestamp 
                assert_equal_timestamps( expct, actual, path )

            when MingleEnum
                assert_equal_enum( expct, actual, path )

            when MingleList then assert_equal_lists( expct, actual, path )

            when MingleSymbolMap 
                assert_equal_symbol_map( expct, actual, path )

            else 
                fail_assert( 
                    "Don't know how to assert equality for #{expct.class}", 
                    path )
        end
    end
end

module MingleTestStructFile

    QN_FILE_END = QualifiedTypeName.get( :"mingle:testgen@v1/TestFileEnd" )

    def self.each_struct_in( f )
        
        File.open( Testing.find_test_data( f ) ) do |io|
            
            rd = BinReader.as_bin_reader( io )
            
            while ( s = rd.read_value ) && ( s.type != QN_FILE_END )
                yield( s )
            end 
        end
    end
end

end
