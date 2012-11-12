package mingle

import (
    "testing"
    "fmt"
    "bitgirder/assert"
    "bitgirder/objpath"
    "reflect"
    "encoding/base64"
    "bytes"
//    "log"
)

func typeRef( s string ) TypeReference { return MustTypeReference( s ) }
func id( s string ) *Identifier { return MustIdentifier( s ) }

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
    typ := MustTypeReference( "ns1@v1/T1" )
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
            assert.Equal( expctStr, err.( *MapLiteralError ).Error() )
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
            msg := "Unhandled type ref initializer: 12"
            assert.Equal( msg, err.( string ) )
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
            return CreateStruct( "T1", "missingVal" )
        },
    )
}

func TestExpectStructError( t *testing.T ) {
    assert.AssertPanic(
        func() { MustStruct( "T1", "missingVal" ) },
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
    err := &TypeCastError{ expected: t1, actual: t2 }
    assert.Equal( suff, err.Error() )
    err.path = path
    assert.Equal( FormatIdPath( path ) + ": " + suff, err.Error() )
}

func TestCastValueErrorFormatting( t *testing.T ) {
    t1 := typeRef( "ns1@v1/T1" )
    path := objpath.RootedAt( id( "f1" ) )
    err := asValueCastError( path, t1, "Blah %s", "X" )
    assert.Equal( "f1: Error converting to ns1@v1/T1: Blah X", err.Error() )
}

type castValueTest struct {

    // set at point of instantiation
    val interface{}
    expct interface{}
    typ TypeReferenceInitializer
    vcErrMsg string // error message expected from a value cast
    isTcErr bool
    isVcErr bool
    errLoc string // default is meaningful; nested errors locs can override
    errMsg string // manually check for this string literal

    // set during test run
    *testing.T
    mgVal Value
}

func ( t *castValueTest ) expectError() bool {
    return t.vcErrMsg != "" || t.isTcErr
}

func ( t *castValueTest ) getErrLoc() string {
    if t.errLoc == "" { return "mg-val" }
    return t.errLoc
}

func ( t *castValueTest ) assertVcError( err error ) {
    if vcErr, ok := err.( *ValidationError ); ok {
        if t.errMsg == "" {
            act := fmt.Sprintf( "%s: Error converting to %s: %s",
                    t.getErrLoc(), t.typ, t.vcErrMsg )
            assert.Equal( act, vcErr.Error() )
        } else { assert.Equal( t.errMsg, err.Error() ) }
    } else { t.Fatal( err ) } 
}

func ( t *castValueTest ) assertTcError( err error ) {
    if tcErr, ok := err.( *TypeCastError ); ok {
        if t.errMsg == "" {
            assert.Equal( 
                fmt.Sprintf( "%s: Expected value of type %s but found %s",
                    t.getErrLoc(), t.typ, TypeOf( t.mgVal ) ),
                tcErr.Error(),
            )
        } else { assert.Equal( t.errMsg, err.Error() ) }
    } else { t.Fatal( err ) }
}

func ( t *castValueTest ) assertError( err error ) {
    switch {
    case t.vcErrMsg != "" || t.isVcErr:
        t.assertVcError( err )
        return
    case t.isTcErr:
        t.assertTcError( err )
        return
    }
    t.Fatal( err )
}

func ( test *castValueTest ) call( t *testing.T ) {
    test.T = t
    test.mgVal = MustValue( test.val )
    mgTyp := asTypeReference( test.typ )
    path := objpath.RootedAt( MustIdentifier( "mg-val"  ) )
    if mgAct, err := CastValue( test.mgVal, mgTyp, path ); err == nil {
        var mgExpct Value
        if test.expct == nil { 
            mgExpct = test.mgVal 
        } else { mgExpct = MustValue( test.expct ) }
        if test.expectError() { t.Fatalf( "Expected err, got: %v", mgAct ) }
        if comp, ok := mgExpct.( Comparer ); ok {
            if reflect.TypeOf( comp ) == reflect.TypeOf( mgAct ) {
                assert.Equal( 0, comp.Compare( mgAct ) )
            }
        } else { assert.Equal( mgExpct, mgAct ) }
    } else { test.assertError( err ) }
}

type numTestContext struct {
    val Value 
    str string
    typ TypeReference
}

// All number coercions to to each other num type and to/from string:
func appendIdentityNumCastTests( tests []*castValueTest ) []*castValueTest {
    numTests := []*numTestContext {
        &numTestContext{ val: Int32( 1 ), str: "1", typ: TypeInt32 },
        &numTestContext{ val: Int64( 1 ), str: "1", typ: TypeInt64 },
        &numTestContext{ val: Uint32( 1 ), str: "1", typ: TypeUint32 },
        &numTestContext{ val: Uint64( 1 ), str: "1", typ: TypeUint64 },
        &numTestContext{ val: Float32( 1.0 ), str: "1", typ: TypeFloat32 },
        &numTestContext{ val: Float64( 1.0 ), str: "1", typ: TypeFloat64 },
    }
    for _, numCtx := range numTests {
    for _, valCtx := range numTests {
        tests = append( tests,
            &castValueTest{ 
                val: numCtx.val, expct: numCtx.str, typ: TypeString },
            &castValueTest{ 
                val: numCtx.str, expct: numCtx.val, typ: numCtx.typ },
            &castValueTest{ 
                val: valCtx.val, expct: numCtx.val, typ: numCtx.typ },
        )
    }}
    return tests
}

// Trunc tests for float/double/dec/string --> int32/int64/integral
func appendTruncateNumCastTests( tests []*castValueTest ) []*castValueTest {
    decVals := []Value{
        Float32( 1.1 ),
        Float64( 1.1 ),
        String( "1.1" ),
    }
    for _, val := range decVals {
        tests = append( tests,
            &castValueTest{ val: val, expct: Int32( 1 ), typ: TypeInt32 },
            &castValueTest{ val: val, expct: Int64( 1 ), typ: TypeInt64 },
            &castValueTest{ val: val, expct: Uint32( 1 ), typ: TypeUint32 },
            &castValueTest{ val: val, expct: Uint64( 1 ), typ: TypeUint64 },
        )
    }
    return tests
}

func getNumValueCastTests() []*castValueTest {
    numTests := make( []*castValueTest, 0 )
    numTests = appendIdentityNumCastTests( numTests )
    numTests = appendTruncateNumCastTests( numTests )
    return numTests
}

func getIdentityStringListCastTests() []*castValueTest {
    res := make( []*castValueTest, 0 )
    for _, quant := range []string { "*", "+" } {
        res = append( res,
            &castValueTest{
                val: []interface{}{ "s1", "s2" },
                expct: MustList( "s1", "s2" ),
                typ: typeRef( "mingle:core@v1/String" + quant ),
            },
        )
        for _, quant2 := range []string { "*", "+" } {
            res = append( res,
                &castValueTest{
                    val: []interface{}{
                        []interface{}{ "s1", "s2" },
                        []interface{}{ "s3", "s4" },
                    },
                    expct: MustList(
                        MustList( "s1", "s2" ),
                        MustList( "s3", "s4" ),
                    ),
                    typ: typeRef( "mingle:core@v1/String" + quant2 + quant ),
                },
            )
        }
    }
    return res
}

func getCastNullToNullableTypeTests() []*castValueTest {
    res := make( []*castValueTest, 0 )
    typs := make( []TypeReference, 0 )
    for _, prim := range PrimitiveTypes {
        typs = append( typs, &NullableTypeReference{ prim } )
    }
    typs = append( typs,
        typeRef( "mingle:core@v1/String??" ),
        typeRef( "mingle:core@v1/String*?" ),
        typeRef( "mingle:core@v1/String+?" ),
        typeRef( "ns1@v1/T?" ),
        typeRef( "ns1@v1/T*?" ),
    )
    for _, typ := range typs {
        res = append( res, &castValueTest{ val: nil, expct: nil, typ: typ } )
    }
    return res
}

func TestCastValue( t *testing.T ) {
    tests := make( []*castValueTest, 0, 100 ) 
    buf1 := Buffer( []byte{ byte( 0 ), byte( 1 ), byte( 2 ) } )
    tm1 := Now()
    map1 := MustSymbolMap( "key1", 1, "key2", "val2" )
    struct1 := MustStruct( "ns1@v1/S1", "key1", "val1" )
    en1 := MustEnum( "ns1@v1/En1", "en-val1" )
    tests = append( tests, 
        &castValueTest{ 
            val: Boolean( true ), expct: Boolean( true ), typ: TypeBoolean },
        &castValueTest{ val: buf1, expct: buf1, typ: TypeBuffer },
        &castValueTest{ val: "s", expct: "s", typ: TypeString },
        &castValueTest{ val: Int32( 1 ), expct: Int32( 1 ), typ: TypeInt32 },
        &castValueTest{ val: Int64( 1 ), expct: Int64( 1 ), typ: TypeInt64 },
        &castValueTest{ val: Uint32( 1 ), expct: Uint32( 1 ), typ: TypeUint32 },
        &castValueTest{ val: Uint64( 1 ), expct: Uint64( 1 ), typ: TypeUint64 },
        &castValueTest{ 
            val: Int32( -1 ), 
            expct: Uint32( uint32( 4294967295 ) ),
            typ: TypeUint32,
        },
        &castValueTest{
            val: Int64( -1 ),
            expct: Uint32( uint32( 4294967295 ) ),
            typ: TypeUint32,
        },
        &castValueTest{
            val: Int32( -1 ),
            expct: Uint64( uint64( 18446744073709551615 ) ),
            typ: TypeUint64,
        },
        &castValueTest{
            val: Int64( -1 ),
            expct: Uint64( uint64( 18446744073709551615 ) ),
            typ: TypeUint64,
        },
        &castValueTest{ 
            val: Float32( 1.0 ), expct: Float32( 1.0 ), typ: TypeFloat32 },
        &castValueTest{ 
            val: Float64( 1.0 ), expct: Float64( 1.0 ), typ: TypeFloat64 },
        &castValueTest{ val: tm1, expct: tm1, typ: TypeTimestamp },
        &castValueTest{ val: en1, expct: en1, typ: en1.Type },
        &castValueTest{ val: en1, expct: en1, typ: TypeEnum },
        &castValueTest{ val: map1, expct: map1, typ: TypeSymbolMap },
        &castValueTest{ val: struct1, expct: struct1, typ: TypeStruct },
        &castValueTest{ val: struct1, expct: struct1, typ: struct1.Type },
        &castValueTest{ val: 1, expct: 1, typ: TypeValue },
        &castValueTest{ val: nil, expct: nil, typ: TypeNull },
        &castValueTest{ val: "true", expct: true, typ: TypeBoolean },
        &castValueTest{ val: "TRUE", expct: true, typ: TypeBoolean },
        &castValueTest{ val: "TruE", expct: true, typ: TypeBoolean },
        &castValueTest{ val: "false", expct: false, typ: TypeBoolean },
        &castValueTest{ val: true, expct: "true", typ: TypeString },
        &castValueTest{ val: false, expct: "false", typ: TypeString },
        &castValueTest{ 
            val: "s", expct: "s", typ: typeRef( "mingle:core@v1/String?" ) },
        &castValueTest{
            val: nil, expct: nil, typ: typeRef( "mingle:core@v1/String?" ) },
        &castValueTest{ val: en1, typ: typeRef( "ns1@v1/Bad" ), isTcErr: true },
        &castValueTest{ 
            val: struct1, typ: typeRef( "ns1@v1/Bad" ), isTcErr: true },
    )
    tests = append( tests, getNumValueCastTests()... )
    tests = append( tests, getIdentityStringListCastTests()... )
    buf1B64 := base64.StdEncoding.EncodeToString( buf1 )
    tests = append( tests,
        &castValueTest{ val: buf1, expct: buf1B64, typ: TypeString },
        &castValueTest{ val: buf1B64, expct: buf1, typ: TypeBuffer },
    )
    tests = append( tests,
        &castValueTest{ val: tm1, expct: tm1.Rfc3339Nano(), typ: TypeString },
        &castValueTest{ 
            val: tm1.Rfc3339Nano(), expct: tm1, typ: TypeTimestamp },
    )
    tests = append( tests,
        &castValueTest{ val: en1, expct: "en-val1", typ: TypeString } )
    tests = append( tests, getCastNullToNullableTypeTests()... )
    tests = append( tests,
        // test conversions in a deeply nested list
        &castValueTest{
            val: []interface{}{
                []interface{}{ "1", int32( 1 ), int64( 1 ) },
                []interface{}{ float32( 1.0 ), float64( 1.0 ) },
                []interface{}{},
            },
            expct: MustList(
                MustList( Int64( 1 ), Int64( 1 ), Int64( 1 ) ),
                MustList( Int64( 1 ), Int64( 1 ) ),
                MustList(),
            ),
            typ: typeRef( "mingle:core@v1/Int64**" ),
        },
        &castValueTest{
            val: []interface{}{ int64( 1 ), nil, "hi" },
            expct: MustList( "1", nil, "hi" ),
            typ: typeRef( "mingle:core@v1/String?*" ),
        },
    )
    tests = append( tests,
        &castValueTest{ val: "abbbc", typ: `String~"^ab+c$"` },
        &castValueTest{ val: "abbbc", typ: `String~"^ab+c$"?` },
        &castValueTest{ val: nil, typ: `String~"^ab+c$"?` },
        &castValueTest{ val: "", typ: `String~"^a*"?` },
        &castValueTest{ 
            val: MustList( "123", 129 ), 
            expct: MustList( "123", "129" ),
            typ: `String~"^\\d+$"*`, 
        },
    )
    for _, quant := range []string { "*", "+", "?*", "*?" } {
        tests = append( tests,
            &castValueTest{ 
                val: MustList( "a", "aaaaaa" ),
                typ: `String~"^a+$"` + quant,
            },
        )
    }
    tests = append( tests,
        &castValueTest{ val: int64( 1 ), typ: "Int64~[-1,1]" },
        &castValueTest{ val: "1", expct: int64( 1 ), typ: "Int64~[-1,1]" },
        &castValueTest{ val: int64( 1 ), typ: "Int64~(,2)" },
        &castValueTest{ val: int64( 1 ), typ: "Int64~[1,1]" },
        &castValueTest{ val: int64( 1 ), typ: "Int64~[-2, 32)" },
        &castValueTest{ val: int32( 1 ), typ: "Int32~[-2, 32)" },
        &castValueTest{ val: uint32( 3 ), typ: "Uint32~[2,32)" },
        &castValueTest{ val: uint64( 3 ), typ: "Uint64~[2,32)" },
        &castValueTest{ val: Float32( -1.1 ), typ: "Float32~[-2.0,32)" },
        &castValueTest{ val: Float64( -1.1 ), typ: "Float64~[-2.0,32)" },
        &castValueTest{
            val: Now(),
            typ: `Timestamp~["1970-01-01T00:00:00Z","2200-01-01T00:00:00Z"]`,
        },
    )
    for _, quant := range []string{ "*", "**", "***" } {
        tests = append( tests,
            &castValueTest{
                val: []interface{}{},
                expct: MustList(),
                typ: typeRef( "mingle:core@v1/Int64" + quant ),
            },
        )
    }
    for _, quant := range []string{ "*", "+" } {
        tests = append( tests,
            &castValueTest{
                val: []interface{}{ []interface{}{}, []interface{}{} },
                expct: MustList( MustList(), MustList() ),
                typ: typeRef( "mingle:core@v1/Int64*" + quant ),
            },
        )
    }
    tests = append( tests,
        &castValueTest{ val: "s", typ: TypeNull, isTcErr: true },
        &castValueTest{ 
            val: "s", typ: TypeBoolean, vcErrMsg: "Invalid boolean value: s", },
        &castValueTest{ 
            val: MustList( 1, 2 ), typ: TypeString, isTcErr: true },
        &castValueTest{
            val: MustList(), 
            typ: typeRef( "mingle:core@v1/String?" ),
            isTcErr: true,
            errMsg: "mg-val: Expected value of type mingle:core@v1/String " +
                "but found mingle:core@v1/Value*",
        },
        &castValueTest{ 
            val: nil, typ: TypeString, vcErrMsg: "value is null" },
        &castValueTest{
            val: nil, 
            typ: `mingle:core@v1/String~"a"`, 
            vcErrMsg: "value is null",
        },
        &castValueTest{
            val: "s", typ: typeRef( "mingle:core@v1/String*" ), isTcErr: true },
        &castValueTest{
            val: MustList( 1, struct1 ),
            typ: typeRef( "mingle:core@v1/Int32*" ),
            isTcErr: true,
            errMsg: "mg-val[ 1 ]: Expected value of type " +
                "mingle:core@v1/Int32 but found ns1@v1/S1",
        },
        &castValueTest{
            val: struct1,
            typ: typeRef( "mingle:core@v1/Int32?" ),
            isTcErr: true,
            errMsg: "mg-val: Expected value of type mingle:core@v1/Int32 " +
                "but found ns1@v1/S1",
        },
        &castValueTest{
            val: MustList(),
            typ: typeRef( "mingle:core@v1/String+" ),
            vcErrMsg: "list is empty",
        },
        &castValueTest{
            val: MustList( MustList( 1, 2 ), MustList() ),
            typ: typeRef( "mingle:core@v1/Int32+*" ),
            isVcErr: true,
            errMsg: "mg-val[ 1 ]: Error converting to mingle:core@v1/Int32+: " +
                "list is empty",
        },
        &castValueTest{
            val: 12,
            typ: typeRef( "ns1@v1/T1" ),
            isTcErr: true,
        },
    )
    for _, typ := range NumericTypes {
        msg := "invalid syntax: not-a-num"
        tests = append( tests,
            &castValueTest{ val: "not-a-num", typ: typ, vcErrMsg: msg } )
    }
    for _, prim := range PrimitiveTypes {
        if prim != TypeValue { // Value would actually be valid cast
            tests = append( tests,
                &castValueTest{ val: struct1, typ: prim, isTcErr: true } )
        }
    }
    tests = append( tests,
        &castValueTest{
            val: "ac", 
            typ: `mingle:core@v1/String~"^ab+c$"`,
            vcErrMsg: `Value "ac" does not satisfy restriction "^ab+c$"`,
        },
        &castValueTest{
            val: "ab",
            typ: `mingle:core@v1/String~"^a*$"?`,
            isVcErr: true,
            errMsg: "mg-val: Error converting to " +
                "mingle:core@v1/String~\"^a*$\": Value \"ab\" does not " +
                "satisfy restriction \"^a*$\"",
        },
        &castValueTest{
            val: MustList( "a", "b" ),
            typ: `String~"^a+$"*`,
            isVcErr: true,
            errMsg: "mg-val[ 1 ]: Error converting to " +
                "mingle:core@v1/String~\"^a+$\": Value \"b\" does not " +
                "satisfy restriction \"^a+$\"",
        },
        &castValueTest{
            val: 12, 
            typ: "mingle:core@v1/Int32~[0,10)", 
            vcErrMsg: "Value 12 does not satisfy restriction [0,10)",
        },
        &castValueTest{
            val: MustTimestamp( "2012-01-01T00:00:00Z" ),
            typ: `mingle:core@v1/Timestamp~` +
                `["2000-01-01T00:00:00Z","2001-01-01T00:00:00Z"]`,
            vcErrMsg: "Value 2012-01-01T00:00:00Z does not satisfy " +
                "restriction " +
                "[\"2000-01-01T00:00:00Z\",\"2001-01-01T00:00:00Z\"]",
        },
    )
    for _, test := range tests { test.call( t ) }
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
    typ := typeRef( "ns1@v1/T1" )
    assert.Equal( typ, TypeOf( &Enum{ Type: typ } ) )
    assert.Equal( typ, TypeOf( &Struct{ Type: typ } ) )
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
        decl := &DeclaredTypeName{ []byte( nm ) }
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
    f( "Enum", QnameEnum )
    f( "SymbolMap", QnameSymbolMap )
    f( "Struct", QnameStruct )
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
    f( Buffer( []byte{ 0, 1, 2 } ), "000102" )
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
    f( &Struct{ Type: typeRef( "ns1@v1/T1" ), Fields: map1 }, expct )
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
        &Enum{ typeRef( "ns1@v1/E1" ), MustIdentifier( "val1" ) },
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
            } else if ve, ok := err.( *ValidationError ); ok {
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
            expct := locStr + ": Expected value of type " +
                "mingle:core@v1/Struct but found mingle:core@v1/String"
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
    chk := func( id *Identifier, okExpct bool, expct interface{} ) {
        assert.Equal( expct, m.Get( id ) )
    }
    assert.Equal( 0, m.Len() )
    chk( id1, false, nil )
    m.Put( id1, val1 )
    chk( id1, true, val1 )
    if err := m.PutSafe( id1, val2 ); err == nil {
        t.Fatalf( "Was able to put val2 at id1" )
    } else {
        assert.Equal( 
            "mingle: map already contains an entry for key: id1", err.Error() )
        chk( id1, true, val1 )
    }
    chk( id2, false, nil )
    m.Put( id1, val2 )
    chk( id1, true, val2 )
    m.Delete( id1 )
    chk( id1, false, nil )
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
    dn1 := &DeclaredTypeName{ []byte( "T1" ) }
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

