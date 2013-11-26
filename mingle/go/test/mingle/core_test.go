package mingle

import (
    "testing"
    "fmt"
    "bitgirder/assert"
    "bitgirder/objpath"
    "reflect"
    "bytes"
//    "log"
)

type notAMingleValue struct {}

func assertAsIntValues( t *testing.T ) {
    i := 0
    vals32 := []interface{} { 
        int8( i ), int16( i ), int32( i ), Int32( int32( i ) ) }
    for _, val := range vals32 { assert.Equal( Int32( i ), MustValue( val ) ) }
    vals64 := []interface{} { i, int64( i ), Int64( int64( i ) ) }
    for _, val := range vals64 { assert.Equal( Int64( i ), MustValue( val ) ) }
    uvals32 := []interface{} { uint32( i ), Uint32( uint32( i ) ) }
    for _, val := range uvals32 {
        assert.Equal( Uint32( i ), MustValue( val ) )
    }
    uvals64 := []interface{} { uint64( i ), Uint64( i ) }
    for _, val := range uvals64 {
        assert.Equal( Uint64( i ), MustValue( val ) )
    }
}

func assertAsDecValues( t *testing.T ) {
    f32 := float32( 1.2 )
    assert.Equal( Float32( f32 ), MustValue( Float32( f32 ) ) )
    assert.Equal( Float32( f32 ), MustValue( f32 ) )
    f64 := float64( 1.2 )
    assert.Equal( Float64( f64 ), MustValue( Float64( f64 ) ) )
    assert.Equal( Float64( f64 ), MustValue( f64 ) )
}

func assertAsBufferValues( t *testing.T ) {
    buf := []byte( "abc" )
    assert.Equal( Buffer( buf ), MustValue( buf ) )
    assert.Equal( Buffer( buf ), MustValue( Buffer( buf ) ) )
}

func assertCompositeTypesAsValue( t *testing.T ) {
    m := MustSymbolMap( "key1", "val1" )
    assert.Equal( m, MustValue( m ) )
    typ := qname( "ns1@v1/T1" )
    s := &Struct{ Type: typ, Fields: m }
    assert.Equal( s, MustValue( s ) )
    l := MustList( 1, 2 )
    assert.Equal( l, MustValue( l ) )
}

func assertAsTimeValues( t *testing.T ) {
    tm := Now()
    assert.Equal( Timestamp( tm ), MustValue( tm ) )
    assert.Equal( Timestamp( tm ), MustValue( Timestamp( tm ) ) )
}

func assertAsNullValues( t *testing.T ) {
    assert.Equal( NullVal, MustValue( nil ) )
    assert.Equal( NullVal, MustValue( NullVal ) )
    assert.Equal( NullVal, MustValue( &Null{} ) )
}

func assertAsEnumValues( t *testing.T ) {
    en := MustEnum( "ns1@v1/E1", "e1" )
    assert.Equal( en, MustValue( en ) )
}

func assertAsListValues( t *testing.T ) {
    assert.Equal( MustList(), MustValue( []interface{}{} ) )
    assert.Equal(
        MustList( String( "s1" ), String( "s2" ), Int32( 3 ) ),
        MustValue( []interface{} { "s1", String( "s2" ), int32( 3 ) } ),
    )
}

func TestValueTypeErrorFormatting( t *testing.T ) {
    loc := objpath.RootedAt( "f1" )
    assert.Equal( "f1: Blah", (&ValueTypeError{ loc, "Blah" }).Error() )
    assert.Equal( "Blah", (&ValueTypeError{ nil, "Blah" }).Error() )
}

func TestAsValue( t *testing.T ) {
    assert.Equal( String( "hello" ), MustValue( "hello" ) )
    assert.Equal( String( "hello" ), MustValue( String( "hello" ) ) )
    assertAsIntValues( t )
    assertAsDecValues( t )
    assert.Equal( Boolean( true ), MustValue( true ) )
    assert.Equal( Boolean( true ), MustValue( Boolean( true ) ) )
    assertAsBufferValues( t )
    assertCompositeTypesAsValue( t )
    assertAsTimeValues( t )
    assertAsNullValues( t )
    assertAsEnumValues( t )
    assertAsListValues( t )
}

func TestAsValueBadValue( t *testing.T ) {
    assert.AssertError(
        func() ( interface{}, error ) { 
            return AsValue( notAMingleValue{} ) 
        },
        func( err error ) { 
            assert.Equal(
                "inVal: Unhandled mingle value {} (mingle.notAMingleValue)",
                err.( *ValueTypeError ).Error(), 
            )
        },
    )
}

func TestExpectValuePanics( t *testing.T ) {
    assert.AssertPanic(
        func() { MustValue( notAMingleValue{} ) },
        func( err interface{} ) {
            assert.Equal(
                "inVal: Unhandled mingle value {} (mingle.notAMingleValue)",
                err.( *ValueTypeError ).Error(),
            )
        },
    )
}

func TestAsValueNestedListErrorLocation( t *testing.T ) {
    assert.AssertError(
        func() ( interface{}, error ) {
            return AsValue( []interface{}{ "s1", &notAMingleValue{} } )
        },
        func( err error ) {
            assert.Equal(
                "inVal[ 1 ]: Unhandled mingle value &{} " +
                    "(*mingle.notAMingleValue)",
                err.( *ValueTypeError ).Error(),
            )
        },
    )
}

func assertMapLiteralError( 
    t *testing.T, expctStr string, f func() ( interface{}, error ) ) {
    assert.AssertError(
        f,
        func( err error ) {
            if mle, ok := err.( *MapLiteralError ); ok {
                assert.Equal( expctStr, mle.Error() )
            } else { assert.Fatal( err ) }
        },
    )
}

func TestCreateSymbolMapErrorBadKey( t *testing.T ) {
    assertMapLiteralError(
        t,
        "Error in map literal pairs at index 2: Unhandled id initializer: 1",
        func() ( interface{}, error ) {
            return CreateSymbolMap( "goodKey", "goodVal", 1, "valForBadKey" )
        },
    )
}

func TestCreateSymbolMapErrorBadval( t *testing.T ) {
    assertMapLiteralError(
        t,
        "Error in map literal pairs at index 1: " +
            "inVal: Unhandled mingle value &{} (*mingle.notAMingleValue)",
        func() ( interface{}, error ) {
            return CreateSymbolMap( "goodKey", &notAMingleValue{} )
        },
    )
}

func TestCreateSymbolMapOddPairLen( t *testing.T ) {
    assertMapLiteralError(
        t,
        "Invalid pairs len: 3",
        func() ( interface{}, error ) {
            return CreateSymbolMap( "f1", "v1", "f2" )
        },
    )
}

func TestCreateSymbolMapDuplicateKeyError( t *testing.T ) {
    assertMapLiteralError(
        t,
        "Multiple entries for key: f1",
        func() ( interface{}, error ) {
            return CreateSymbolMap( "f1", "v1", "f2", 1, "f1", "v2" )
        },
    )
}

func TestExpectSymbolMapPanic( t *testing.T ) {
    assert.AssertPanic(
        func() { MustSymbolMap( 1, "bad" ) },
        func( err interface{} ) {
            msg := "Error in map literal pairs at index 0: Unhandled id " +
                   "initializer: 1"
            assert.Equal( msg, err.( *MapLiteralError ).Error() )
        },
    )
}

func TestAsTypeReference( t *testing.T ) {
    expct := MustTypeReference( "ns1@v1/T1" )
    f := func( typ TypeReference ) {
        assert.Truef( expct.Equals( typ ), "%s != %s", expct, typ )
    }
    f( asTypeReference( expct ) )
    f( asTypeReference( "ns1@v1/T1" ) )
    assert.AssertPanic( 
        func() { f( asTypeReference( 12 ) ) }, 
        func( err interface{} ) { 
            msg := "mingle: Unhandled type ref initializer: int"
            assert.Equal( msg, err.( error ).Error() )
        },
    )
}

func TestNonEmptyList( t *testing.T ) {
    l := MustList( "1", 2, true, Float32( float32( 1.2 ) ) )
    assert.Equal( 4, l.Len() )
    expct := []Value { 
        String( "1" ), Int64( 2 ), Boolean( true ), Float32( float32( 1.2 ) ),
    }
    assert.Equal( expct, l.vals )
}

func TestEmptyList( t *testing.T ) {
    l := MustList()
    assert.Equal( 0, l.Len() )
    assert.Equal( []Value{}, l.vals )
}

func TestAsListError( t *testing.T ) {
    if _, err := CreateList( 1, notAMingleValue{}, "3" ); err != nil {
        assert.Equal(
            "inVal: Unhandled mingle value {} (mingle.notAMingleValue)",
            err.( *ValueTypeError ).Error(),
        )
    } else { t.Fatalf( "No error returned" ) }
}

func TestExpectListPanic( t *testing.T ) {
    assert.AssertPanic(
        func() { MustList( 1, notAMingleValue{}, "3" ) },
        func( err interface{} ) {
            assert.Equal( 
                "inVal: Unhandled mingle value {} (mingle.notAMingleValue)",
                err.( *ValueTypeError ).Error(),
            )
        },
    )
}

func TestEmptySymbolMap( t *testing.T ) {
    m := MustSymbolMap()
    assert.Equal( 0, m.Len() )
    assert.Equal( nil, m.GetById( MustIdentifier( "f1" ) ) )
}

// Test base coverage of GetByString, GetById handling of a value that is
// present and one that is not; more type-specific coverage is in
// TestSymbolMapTypedAccessors
func TestSymbolMapGettersBase( t *testing.T ) {
    m := MustSymbolMap( "f1", "val1" )
    assert.Equal( 
        String( "val1" ), m.GetById( MustIdentifier( "f1" ) ).( String ) )
    assert.Equal( nil, m.GetById( MustIdentifier( "f2" ) ) )
}

func TestEmptySymbolMapEachPair( t *testing.T ) {
    MustSymbolMap().EachPair( func( k *Identifier, v Value ) {
        t.Fatalf( "Visitor called on empty map" )
    })
}

func TestNonEmptySymbolMapEachPair( t *testing.T ) {
    m := MustSymbolMap( "k1", Int32( 1 ), "k2", Int64( 2 ) )
    vals := []Value{ Int32( 1 ), Int64( 2 ) }
    set := func( k *Identifier, v Value, kStr string, i int ) {
        if k.Equals( MustIdentifier( kStr ) ) {
            if vals[ i ] == nil { 
                t.Fatalf( "Already saw %s", kStr )
            } else {
                assert.Equal( vals[ i ], v )
                vals[ i ] = nil
            }
        }
    }
    m.EachPair( func( k *Identifier, v Value ) {
        set( k, v, "k1", 0 )
        set( k, v, "k2", 1 )
    })
    for i, val := range vals {
        if val != nil { t.Fatalf( "vals[ %d ] not visited: %v", i, val ) }
    }
}

func TestSymbolMapEachPairError( t *testing.T ) {
    m := MustSymbolMap( "k1", "v1", "k2", "v2", "k3", "v3" )
    visits := 0
    errExpct := fmt.Errorf( "test-error" )
    vis := func( fld *Identifier, val Value ) error {
        visits++
        switch visits {
        case 1: 
        case 2: return errExpct
        default: t.Fatalf( "Unexpected visit count: %d", visits )
        }
        return nil
    }
    if err := m.EachPairError( vis ); err == nil {
        t.Fatalf( "Expected error" )
    } else { assert.Equal( errExpct, err ) }
}

func TestCreateStructError( t *testing.T ) {
    assertMapLiteralError(
        t,
        "Invalid pairs len: 1",
        func() ( interface{}, error ) {
            return CreateStruct( "ns1@v1/T1", "missingVal" )
        },
    )
}

func TestExpectStructError( t *testing.T ) {
    assert.AssertPanic(
        func() { MustStruct( "ns1@v1/T1", "missingVal" ) },
        func( err interface{} ) {
            assert.Equal(
                "Invalid pairs len: 1", 
                err.( *MapLiteralError ).Error(),
            )
        },
    )
}

// Not extensively re-testing functionality of objpath.Format(); only concerned
// here with coverage of our formatter impl
func TestObjectPathFormatting( t *testing.T ) {
    str := 
        FormatIdPath(
            objpath.RootedAt( MustIdentifier( "f1" ) ).
                Descend( MustIdentifier( "someFld" ) ).
                StartList().
                Next().
                StartList().
                Descend( MustIdentifier( "some-fld2" ) ).
                StartList().
                Next().
                Next().
                Descend( MustIdentifier( "some_fld3" ) ),
        )
    assert.Equal( "f1.some-fld[ 1 ][ 0 ].some-fld2[ 2 ].some-fld3", str )
}

func TestTypeCastFormatting( t *testing.T ) {
    path := objpath.RootedAt( id( "f1" ) ).Descend( id( "f2" ) )
    t1 := typeRef( "ns1@v1/T1" )
    t2 := typeRef( "ns1@v1/T2" )
    suff := "Expected value of type ns1@v1/T1 but found ns1@v1/T2"
    err := NewTypeCastError( t1, t2, nil )
    assert.Equal( suff, err.Error() )
    err = NewTypeCastError( t1, t2, path )
    assert.Equal( FormatIdPath( path ) + ": " + suff, err.Error() )
}

func TestTypeOf( t *testing.T ) {
    assert.Equal( TypeBoolean, TypeOf( Boolean( true ) ) )
    assert.Equal( TypeBuffer, TypeOf( Buffer( []byte{} ) ) )
    assert.Equal( TypeString, TypeOf( String( "" ) ) )
    assert.Equal( TypeInt32, TypeOf( Int32( 1 ) ) )
    assert.Equal( TypeInt64, TypeOf( Int64( 1 ) ) )
    assert.Equal( TypeUint32, TypeOf( Uint32( 1 ) ) )
    assert.Equal( TypeUint64, TypeOf( Uint64( 1 ) ) )
    assert.Equal( TypeFloat32, TypeOf( Float32( 1.0 ) ) )
    assert.Equal( TypeFloat64, TypeOf( Float64( 1.0 ) ) )
    assert.Equal( TypeTimestamp, TypeOf( Now() ) )
    assert.Equal( TypeSymbolMap, TypeOf( MustSymbolMap() ) )
    assert.Equal( typeRef( "mingle:core@v1/Value*" ), TypeOf( MustList() ) )
    qn := qname( "ns1@v1/T1" )
    typ := &AtomicTypeReference{ Name: qn }
    assert.Equal( typ, TypeOf( &Enum{ Type: qn } ) )
    assert.Equal( typ, TypeOf( &Struct{ Type: qn } ) )
}

func TestAtomicTypeIn( t *testing.T ) {
    str := "ns1@v1/T1"
    typ := MustTypeReference( str )
    for _, ext := range []string{ "", "?", "*", "+", "**+", "*+?++", "??" } {
        typ2 := MustTypeReference( str + ext )
        assert.True( typ.Equals( AtomicTypeIn( typ2 ) ) )
    }
}

func TestResolveInCore( t *testing.T ) {
    f := func( nm string, expct *QualifiedTypeName ) {
        decl := &DeclaredTypeName{ nm }
        if qn, ok := ResolveInCore( decl ); ok {
            assert.True( qn.Equals( expct ) )
        } else { t.Fatalf( "Couldn't resolve: %s", nm ) }
    }
    f( "Value", QnameValue )
    f( "Boolean", QnameBoolean )
    f( "Buffer", QnameBuffer )
    f( "String", QnameString )
    f( "Int32", QnameInt32 )
    f( "Int64", QnameInt64 )
    f( "Uint32", QnameUint32 )
    f( "Uint64", QnameUint64 )
    f( "Float32", QnameFloat32 )
    f( "Float64", QnameFloat64 )
    f( "Timestamp", QnameTimestamp )
    f( "SymbolMap", QnameSymbolMap )
    f( "Null", QnameNull )
}

func TestComparer( t *testing.T ) {
    // v1 should be <= v2
    f := func( v1, v2 Comparer, eq bool ) {
        i := -1
        if eq { i = 0 }
        assert.Equal( i, v1.Compare( v2 ) )
        assert.Equal( -i, v2.Compare( v1 ) )
    }
    f( String( "a" ), String( "a" ), true )
    f( String( "a" ), String( "b" ), false )
    for _, typ := range NumericTypes {
        mkNum := func( s string ) Comparer {
            val, err := CastValue( String( s ), typ, idPathRootVal )
            if err != nil { t.Fatal( err ) }
            return val.( Comparer )
        }
        zero, one := mkNum( "0" ), mkNum( "1" )
        f( zero, zero, true )
        f( zero, one, false )
    }
    tm1A := MustTimestamp( "2012-01-01T12:00:00.0Z" )
    tm1B := MustTimestamp( "2012-01-01T11:00:00.0-01:00" )
    tm1C := MustTimestamp( "2012-01-01T10:59:00.0-01:01" )
    tm2 := MustTimestamp( "2012-01-01T13:00:00.0+00:00" )
    for _, tm := range []Timestamp { tm1A, tm1B, tm1C } { 
        f( tm, tm, true )
        f( tm, tm2, false ) 
    }
}

func TestRestrictionAccept( t *testing.T ) {
    f := func( v Value, vr ValueRestriction, expct bool ) {
        assert.Equal( expct, vr.AcceptsValue( v ) )
    }
    vr1 := &RangeRestriction{ true, Int32( 0 ), Int32( 10 ), true }
    f( Int32( 0 ), vr1, true )
    f( Int32( 10 ), vr1, true )
    f( Int32( 5 ), vr1, true )
    vr1.MinClosed, vr1.MaxClosed = false, false
    f( Int32( 0 ), vr1, false )
    f( Int32( 10 ), vr1, false )
    f( Int32( -1 ), vr1, false )
    f( Int32( 11 ), vr1, false )
    vr2, err := NewRegexRestriction( "^a{1,4}$" )
    if err != nil { t.Fatal( err ) }
    f( String( "aa" ), vr2, true )
    f( String( "aaaaa" ), vr2, false )
}

func TestQuoteValue( t *testing.T ) {
    f := func( v Value, strs ...string ) {
        q := QuoteValue( v )
        for _, str := range strs { if str == q { return } }
        t.Fatalf( "No vals in %#v matched quoted val %q", strs, q )
    }
    f( Boolean( true ), "true" )
    f( Boolean( false ), "false" )
    f( Buffer( []byte{ 0, 1, 2 } ), "buf[000102]" )
    f( String( "s" ), `"s"` )
    f( Int32( 1 ), "1" )
    f( Int64( 1 ), "1" )
    f( Uint32( 1 ), "1" )
    f( Uint64( 1 ), "1" )
    f( Float32( 1.1 ), "1.1" )
    f( Float64( 1.1 ), "1.1" )
    tm := "2012-01-01T12:00:00Z"
    f( MustTimestamp( tm ), tm )
    en := MustEnum( "ns1@v1/E1", "v" )
    f( en, "ns1@v1/E1.v" )
    f( NullVal, "null" )
    f( MustList(), "[]" )
    f( MustList( String( "s" ), Int32( 1 ) ), `["s", 1]` )
    f( MustSymbolMap(), "{}" )
    f( MustSymbolMap( "k1", 1, "k2", "2" ),
        `{k1:1, k2:"2"}`, `{k2:"2", k1:1}` )
    map1 := MustSymbolMap( "k", 1 )
    expct := `ns1@v1/T1{k:1}`
    f( &Struct{ Type: qname( "ns1@v1/T1" ), Fields: map1 }, expct )
}

func TestIsNull( t *testing.T ) {
    assert.True( IsNull( &Null{} ) )
    assert.True( IsNull( NullVal ) )
    assert.False( IsNull( Int32( 1 ) ) )
}

func TestIdentifierCompare( t *testing.T ) {
    id1 := MustIdentifier( "a" )
    id2 := MustIdentifier( "a-b1" )
    id3 := MustIdentifier( "b" )
    id4 := MustIdentifier( "b-b1" )
    ids := []*Identifier{ id1, id2, id3, id4 }
    for i, e := 0, len( ids ) - 1; i < e; i++ {
        l, r := ids[ i ], ids[ i + 1 ]
        assert.True( l.Compare( r ) < 0 )
        assert.True( l.Compare( MustIdentifier( l.ExternalForm() ) ) == 0 )
        assert.True( r.Compare( l ) > 0 )
    }
}

func TestMustEnum( t *testing.T ) {
    assert.Equal(
        &Enum{ qname( "ns1@v1/E1" ), MustIdentifier( "val1" ) },
        MustEnum( "ns1@v1/E1", "val1" ),
    )
}

func assertMapTypedAccessor(
    v reflect.Value, 
    pref, typeName string, 
    fld interface{}, 
    expct interface{},
    t *testing.T ) {
    var methName string
    switch fld.( type ) {
    case string: methName = pref + typeName + "ByString"
    case *Identifier: methName = pref + typeName + "ById"
    default: t.Fatalf( "Bad field type: %v", fld )
    }
    if meth := v.MethodByName( methName ); meth.Kind() == reflect.Func {
        params := []reflect.Value{ reflect.ValueOf( fld ) }
        out := meth.Call( params )
        if pref == "Get" {
            if err := out[ 1 ].Interface(); err != nil { t.Fatal( err ) }
        }
        assert.Equal( expct, out[ 0 ].Interface() )
    } else { t.Fatalf( "Invalid kind for %s: %v", methName, meth.Kind() ) }
}

func assertMapTypedAccessors( 
    m *SymbolMapAccessor, 
    typeName, fld string, 
    expct interface{}, 
    t *testing.T ) {
    v := reflect.ValueOf( m )
    for _, pref := range []string { "Get", "Must" } {
        assertMapTypedAccessor( v, pref, typeName, fld, expct, t )
        assertMapTypedAccessor( 
            v, pref, typeName, MustIdentifier( fld ), expct, t )
    }
}

func TestSymbolMapAccessorTypes( t *testing.T ) {
    tm1 := Now()
    en1 := MustEnum( "ns1@v1/E1", "val" )
    map1 := MustSymbolMap()
    list1 := MustList()
    struct1 := MustStruct( "ns1@v1/S1" )
    m := MustSymbolMap(
        "string1", "s",
        "bool1", true,
        "buf1", []byte{ 1 },
        "int1", int32( 1 ),
        "int2", int64( 1 ),
        "int3", uint32( 1 ),
        "int4", uint64( 1 ),
        "dec1", float32( 1.1 ),
        "dec2", float64( 1.1 ),
        "time1", tm1,
        "enum1", en1,
        "struct1", struct1,
        "map1", map1,
        "list1", list1,
    )
    path := objpath.RootedAt( MustIdentifier( "map" ) )
    acc := NewSymbolMapAccessor( m, path )
    f := func( typeName, fld string, expct interface{} ) {
        assertMapTypedAccessors( acc, typeName, fld, expct, t )
    }
    f( "Value", "string1", String( "s" ) )
    f( "Boolean", "bool1", Boolean( true ) )
    f( "GoBool", "bool1", true )
    f( "Buffer", "buf1", Buffer( []byte{ 1 } ) )
    f( "GoBuffer", "buf1", []byte{ 1 } )
    f( "String", "string1", String( "s" ) )
    f( "GoString", "string1", "s" )
    f( "Int32", "int1", Int32( int32( 1 ) ) )
    f( "GoInt32", "int1", int32( 1 ) )
    f( "Int64", "int2", Int64( int64( 1 ) ) )
    f( "GoInt64", "int2", int64( 1 ) )
    f( "Uint32", "int3", Uint32( uint32( 1 ) ) )
    f( "GoUint32", "int3", uint32( 1 ) )
    f( "Uint64", "int4", Uint64( uint64( 1 ) ) )
    f( "GoUint64", "int4", uint64( 1 ) )
    f( "Float32", "dec1", Float32( float32( 1.1 ) ) )
    f( "GoFloat32", "dec1", float32( 1.1 ) )
    f( "Float64", "dec2", Float64( float64( 1.1 ) ) )
    f( "GoFloat64", "dec2", float64( 1.1 ) )
    f( "Timestamp", "time1", tm1 )
    f( "Enum", "enum1", en1 )
    f( "Struct", "struct1", struct1 )
    f( "SymbolMap", "map1", map1 )
    f( "List", "list1", list1 )
}

// Not re-testing all typed accessors exhaustively -- just one for a mingle/go
// type pair that was autogenerated and one that was handcoded. We also intermix
// coverage that paths are formed correctly when accessor is at the root and
// when it is created with a non-nil parent path
func TestSymbolMapAccessorExpectPanic( t *testing.T ) {
    f := func( path objpath.PathNode, call func() ) {
        defer func() {
            if err := recover(); err == nil {
                t.Fatal( "Expected error" )
            } else if ve, ok := err.( *ValueCastError ); ok {
                expct := ""
                if path != nil { expct += FormatIdPath( path ) + "." }
                expct += "f1: value is null"
                assert.Equal( expct, ve.Error() )
            } else { t.Fatal( err ) }
        }()
        call()
    }
    path := objpath.RootedAt( MustIdentifier( "o1" ) )
    acc1 := NewSymbolMapAccessor( MustSymbolMap(), path )
    acc2 := NewSymbolMapAccessor( MustSymbolMap(), nil )
    f( nil, func() { acc2.MustGoStringByString( "f1" ) } )
    f( path, func() { acc1.MustStringById( MustIdentifier( "f1" ) ) } )
    f( nil, func() { acc2.MustValueByString( "f1" ) } )
}

func TestSymbolMapAccessorCastErrorPath( t *testing.T ) {
    f := func( path objpath.PathNode, locStr string ) {
        acc := NewSymbolMapAccessor( MustSymbolMap( "f1", "s" ), path )
        if _, err := acc.GetStructByString( "f1" ); err == nil {
            t.Fatal( "Expected error" )
        } else {
            expct := 
                locStr + ": Expected *Struct but found mingle:core@v1/String"
            assert.Equal( expct, err.Error() )
        }
    }
    f( nil, "f1" )
    f( objpath.RootedAt( MustIdentifier( "o1" ) ), "o1.f1" )
}

func TestIdentifierFormatRegistry( t *testing.T ) {
    f := func( nm string, idFmt IdentifierFormat ) {
        idFmtAct := MustIdentifierFormatString( nm )
        assert.Equal( nm, idFmtAct.String() )
        assert.Equal( idFmt, idFmtAct )
    }
    f( "lc-underscore", LcUnderscore )
    f( "lc-hyphenated", LcHyphenated )
    f( "lc-camel-capped", LcCamelCapped )
    chk := func( err interface{} ) {
        msg := `Unrecognized id format: not-here`
        assert.Equal( msg, err.( error ).Error() )
    }
    assert.AssertPanic( 
        func() { MustIdentifierFormatString( "not-here" ) }, chk )
}

func TestMapImpl( t *testing.T ) {
    id1 := MustIdentifier( "id1" )
    id2 := MustIdentifier( "id2" )
    val1 := 1
    val2 := 2
    m := NewIdentifierMap()
    assert.Equal( 0, m.Len() )
    assert.False( m.HasKey( id1 ) )
    chkGet := func( id *Identifier, okExpct bool, expct interface{} ) {
        assert.Equal( expct, m.Get( id ) )
        act, ok := m.GetOk( id )
        assert.Equal( okExpct, ok )
        assert.Equal( act, expct )
    }
    chkGet( id1, false, nil )
    m.Put( id1, val1 )
    chkGet( id1, true, val1 )
    if err := m.PutSafe( id1, val2 ); err == nil {
        t.Fatalf( "Was able to put val2 at id1" )
    } else {
        assert.Equal( 
            "mingle: map already contains an entry for key: id1", err.Error() )
        chkGet( id1, true, val1 )
    }
    chkGet( id2, false, nil )
    m.Put( id1, val2 )
    chkGet( id1, true, val2 )
    m.Delete( id1 )
    chkGet( id1, false, nil )
    assert.Equal( 0, m.Len() )
}

func TestMapImplEachPair( t *testing.T ) {
    m := NewIdentifierMap()
    m.EachPair( func( id *Identifier, val interface{} ) {
        t.Fatal( "visit of empty map" )
    })
    m.Put( MustIdentifier( "id1" ), String( "s1" ) )
    m.Put( MustIdentifier( "id2" ), String( "s2" ) )
    out := &bytes.Buffer{}
    m.EachPair( func( id *Identifier, val interface{} ) {
        fmt.Fprintf( out, "%s=%s,", id, val.( String ) )
    })
    if outStr := out.String(); 
        ! ( outStr == "id1=s1,id2=s2," || outStr == "id2=s2,id1=s1," ) {
        t.Fatalf( "Unexpected outStr: %q", outStr )
    }
    err1 := fmt.Errorf( "test-error" )
    f := func( id *Identifier, val interface{} ) error {
        if id.ExternalForm() == "id1" { return err1 }
        return nil
    }
    if err := m.EachPairError( f ); err != err1 {
        t.Fatalf( "Expected err1 %s, got: %s", err1, err )
    }
}

func TestResolveIn( t *testing.T ) {
    dn1 := &DeclaredTypeName{ "T1" }
    ns := MustNamespace( "ns1@v1" )
    qn1 := &QualifiedTypeName{ Namespace: ns, Name: dn1 }
    assert.Equal( qn1, dn1.ResolveIn( ns ) )
}

func TestTypeNameIn( t *testing.T ) {
    nmStr := "mingle:core@v1/Int32"
    nm := MustQualifiedTypeName( nmStr )
    for _, typStr := range []string {
        nmStr, nmStr + "~[0,3]", nmStr + "?", nmStr + "*", nmStr + "*?+",
    } {
        typ := MustTypeReference( typStr )
        assert.Equal( nm, TypeNameIn( typ ) )
    }
}

func TestSortIds( t *testing.T ) {
    chk := func( in, expct []*Identifier ) {
        assert.Equal( expct, SortIds( in ) )
    }
    chk( []*Identifier{}, []*Identifier{} )
    ids := []*Identifier{ id( "i1" ), id( "i2" ), id( "i3" ) }
    for _, in := range [][]*Identifier{
        []*Identifier{ ids[ 0 ], ids[ 1 ], ids[ 2 ] },
        []*Identifier{ ids[ 2 ], ids[ 1 ], ids[ 0 ] },
        []*Identifier{ ids[ 2 ], ids[ 0 ], ids[ 1 ] },
    } {
        chk( in, ids )
    }
}

func TestMissingFieldsErrorFormatting( t *testing.T ) {
    chk := func( msg string, flds ...*Identifier ) {
        err := NewMissingFieldsError( nil, flds )
        assert.Equal( msg, err.Error() )
    }
    chk( "missing field(s): f1", id( "f1" ) )
    chk( "missing field(s): f1, f2", id( "f2" ), id( "f1" ) ) // check sorted
}

func TestUnrecognizedFieldErrorFormatting( t *testing.T ) {
    assert.Equal(
        "unrecognized field: f1",
        NewUnrecognizedFieldError( nil, id( "f1" ) ).Error(),
    )
}

//func TestServiceIdMap( t *testing.T ) {
//    m := NewServiceIdMap()
//    ns1 := MustNamespace( "ns1@v1" )
//    ns2 := MustNamespace( "ns1@v2" )
//    svc1 := id( "svc1" )
//    svc2 := id( "svc2" )
//    m.Put( ns1, svc1, 1 )
//    chkGetOk := func( 
//        expctVal interface{}, expctOk bool, ns *Namespace, svc *Identifier ) {
//        actVal, actOk := m.GetOk( ns, svc )
//        assert.Equal( expctOk, actOk )
//        if actOk { assert.Equal( expctVal, actVal ) }
//    }
//    chkGetOk( 1, true, ns1, svc1 )
//    chkGetOk( nil, false, ns1, svc2 )
//    chkGetOk( nil, false, ns2, svc2 )
//}
