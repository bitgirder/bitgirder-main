package mingle

import (
    "testing"
    "fmt"
    "bitgirder/assert"
    "bitgirder/objpath"
    "bytes"
    "time"
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
    m := MustSymbolMap( mkId( "key1" ), "val1" )
    assert.Equal( m, MustValue( m ) )
    typ := mkQn( mkNs( mkId( "v1" ), mkId( "ns1" ) ), mkDeclNm( "T1" ) )
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
    qn := mkQn( mkNs( mkId( "v1" ), mkId( "ns1" ) ), mkDeclNm( "E1" ) )
    en := &Enum{ Type: qn, Value: mkId( "e1" ) }
    assert.Equal( en, MustValue( en ) )
}

func assertAsListValues( t *testing.T ) {
    assert.Equal( MustList(), MustValue( []interface{}{} ) )
    assert.Equal(
        MustList( String( "s1" ), String( "s2" ), Int32( 3 ) ),
        MustValue( []interface{} { "s1", String( "s2" ), int32( 3 ) } ),
    )
}

// Also gives coverage for MustList()
func TestCreateListWithType( t *testing.T ) {
    a := &assert.Asserter{ t }
    t1 := &ListTypeReference{ ElementType: TypeInt32 }
    l := MustList( t1, int32( 1 ), int32( 2 ) )
    a.Equal( 2, l.Len() )
    a.Equal( Int32( 1 ), l.Get( 0 ) )
    a.Equal( Int32( 2 ), l.Get( 1 ) )
    a.Truef( l.Type.Equals( t1 ), "bad type: %s", l.Type )
}

func TestasValueLocatedErrorFormatting( t *testing.T ) {
    loc := objpath.RootedAt( "f1" )
    assert.Equal( "f1: Blah", (&asValueLocatedError{ loc, "Blah" }).Error() )
    assert.Equal( "Blah", (&asValueLocatedError{ nil, "Blah" }).Error() )
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
                err.( *asValueLocatedError ).Error(), 
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
                err.( *asValueLocatedError ).Error(),
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
                err.( *asValueLocatedError ).Error(),
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

func TestCreateSymbolMapErrorBadVal( t *testing.T ) {
    assertMapLiteralError(
        t,
        "error in map literal pairs at index 1: " +
            "inVal: Unhandled mingle value &{} (*mingle.notAMingleValue)",
        func() ( interface{}, error ) {
            return CreateSymbolMap( mkId( "k" ), &notAMingleValue{} )
        },
    )
}

func TestCreateSymbolMapOddPairLen( t *testing.T ) {
    assertMapLiteralError(
        t,
        "invalid pairs len: 3",
        func() ( interface{}, error ) {
            return CreateSymbolMap( mkId( "f1" ), "v1", mkId( "f2" ) )
        },
    )
}

func TestCreateSymbolMapDuplicateKeyError( t *testing.T ) {
    assertMapLiteralError(
        t,
        "duplicate entry for 'f1' starting at index 4",
        func() ( interface{}, error ) {
            return CreateSymbolMap( 
                mkId( "f1" ), "v1", 
                mkId( "f2" ), 1, 
                mkId( "f1" ), "v2",
            )
        },
    )
}

func TestExpectSymbolMapPanic( t *testing.T ) {
    assert.AssertPanic(
        func() { MustSymbolMap( 1, "bad" ) },
        func( err interface{} ) {
            msg := 
                "error in map literal pairs at index 0: invalid key type: int"
            assert.Equal( msg, err.( *MapLiteralError ).Error() )
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
            err.( *asValueLocatedError ).Error(),
        )
    } else { t.Fatalf( "No error returned" ) }
}

func TestExpectListPanic( t *testing.T ) {
    assert.AssertPanic(
        func() { MustList( 1, notAMingleValue{}, "3" ) },
        func( err interface{} ) {
            assert.Equal( 
                "inVal: Unhandled mingle value {} (mingle.notAMingleValue)",
                err.( *asValueLocatedError ).Error(),
            )
        },
    )
}

func TestEmptySymbolMap( t *testing.T ) {
    m := MustSymbolMap()
    assert.Equal( 0, m.Len() )
    assert.Equal( nil, m.Get( mkId( "f1" ) ) )
}

func TestEmptySymbolMapEachPair( t *testing.T ) {
    MustSymbolMap().EachPair( func( k *Identifier, v Value ) {
        t.Fatalf( "Visitor called on empty map" )
    })
}

func TestNonEmptySymbolMapEachPair( t *testing.T ) {
    m := MustSymbolMap( mkId( "k1" ), Int32( 1 ), mkId( "k2" ), Int64( 2 ) )
    vals := []Value{ Int32( 1 ), Int64( 2 ) }
    set := func( k *Identifier, v Value, kStr string, i int ) {
        if k.Equals( mkId( kStr ) ) {
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
    m := MustSymbolMap( 
        mkId( "k1" ), "v1", 
        mkId( "k2" ), "v2", 
        mkId( "k3" ), "v3",
    )
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
        "invalid pairs len: 1",
        func() ( interface{}, error ) {
            return CreateStruct( ns1V1Qn( "T1" ), "missingVal" )
        },
    )
}

func TestExpectStructError( t *testing.T ) {
    assert.AssertPanic(
        func() { MustStruct( ns1V1Qn( "T1" ), "missingVal" ) },
        func( err interface{} ) {
            assert.Equal(
                "invalid pairs len: 1", 
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
            objpath.RootedAt( mkId( "f1" ) ).
                Descend( mkId( "some", "fld1" ) ).
                StartList().
                Next().
                StartList().
                Descend( mkId( "some", "fld2" ) ).
                StartList().
                Next().
                Next().
                Descend( mkId( "some", "fld3" ) ),
        )
    assert.Equal( "f1.some-fld1[ 1 ][ 0 ].some-fld2[ 2 ].some-fld3", str )
}

func TestTypeCastFormatting( t *testing.T ) {
    path := objpath.RootedAt( mkId( "f1" ) ).Descend( mkId( "f2" ) )
    t1 := ns1V1At( "T1" )
    t2 := ns1V1At( "T2" )
    suff := "Expected value of type ns1@v1/T1 but found ns1@v1/T2"
    err := NewTypeInputError( t1, t2, nil )
    assert.Equal( suff, err.Error() )
    err = NewTypeInputError( t1, t2, path )
    assert.Equal( FormatIdPath( path ) + ": " + suff, err.Error() )
}

func TestTypeOf( t *testing.T ) {
    a := &assert.Asserter{ t }
    a.Equal( TypeBoolean, TypeOf( Boolean( true ) ) )
    a.Equal( TypeBuffer, TypeOf( Buffer( []byte{} ) ) )
    a.Equal( TypeString, TypeOf( String( "" ) ) )
    a.Equal( TypeInt32, TypeOf( Int32( 1 ) ) )
    a.Equal( TypeInt64, TypeOf( Int64( 1 ) ) )
    a.Equal( TypeUint32, TypeOf( Uint32( 1 ) ) )
    a.Equal( TypeUint64, TypeOf( Uint64( 1 ) ) )
    a.Equal( TypeFloat32, TypeOf( Float32( 1.0 ) ) )
    a.Equal( TypeFloat64, TypeOf( Float64( 1.0 ) ) )
    a.Equal( TypeTimestamp, TypeOf( Now() ) )
    a.Equal( TypeSymbolMap, TypeOf( MustSymbolMap() ) )
    a.Equal( TypeOpaqueList, TypeOf( MustList() ) )
    qn := ns1V1Qn( "T1" )
    typ := NewAtomicTypeReference( qn, nil )
    a.Equal( typ, TypeOf( &Enum{ Type: qn } ) )
    a.Equal( typ, TypeOf( &Struct{ Type: qn } ) )
}

func TestAtomicTypeIn( t *testing.T ) {
    at := ns1V1At( "T1" )
    chk := func( typ TypeReference ) {
        assert.True( at.Equals( AtomicTypeIn( typ ) ) )
        assert.True( 
            at.Equals( AtomicTypeIn( NewPointerTypeReference( typ ) ) ) )
    }
    chk( at )
    chk( &ListTypeReference{ at, true } )
    chk( &ListTypeReference{ at, false } )
    chk( 
        &ListTypeReference{ 
            ElementType: &ListTypeReference{ 
                ElementType: &ListTypeReference{
                    ElementType: at,
                    AllowsEmpty: true,
                },
                AllowsEmpty: true,
            },
            AllowsEmpty: false,
        },
    )
    chk( 
        &ListTypeReference{ 
            ElementType: &ListTypeReference{ 
                ElementType: &NullableTypeReference{
                    &ListTypeReference{
                        ElementType: at,
                        AllowsEmpty: true,
                    },
                },
                AllowsEmpty: true,
            },
            AllowsEmpty: false,
        },
    )
    chk( NewPointerTypeReference( at ) )
    chk( &NullableTypeReference{ NewPointerTypeReference( at ) } )
}

func TestResolveInCore( t *testing.T ) {
    f := func( nm string, expct *QualifiedTypeName ) {
        decl := &DeclaredTypeName{ nm }
        if qn, ok := ResolveInCore( decl ); ok {
            assert.True( qn.Equals( expct ) )
        } else { t.Fatalf( "Couldn't resolve: %s", nm ) }
    }
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
    for _, qn := range NumericTypeNames {
        mkNum := func( s string ) Comparer {
            val, err := ParseNumber( s, qn )
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
    vr1 := MustRangeRestriction( 
        QnameInt32, true, Int32( 0 ), Int32( 10 ), true )
    f( Int32( 0 ), vr1, true )
    f( Int32( 10 ), vr1, true )
    f( Int32( 5 ), vr1, true )
    vr2 := MustRangeRestriction( 
        QnameInt32, false, Int32( 0 ), Int32( 10 ), false )
    f( Int32( 0 ), vr2, false )
    f( Int32( 10 ), vr2, false )
    f( Int32( -1 ), vr2, false )
    f( Int32( 11 ), vr2, false )
    vr3, err := CreateRegexRestriction( "^a{1,4}$" )
    if err != nil { t.Fatal( err ) }
    f( String( "aa" ), vr3, true )
    f( String( "aaaaa" ), vr3, false )
}

func testAtomicRestrictionError( 
    t *AtomicRestrictionErrorTest, a *assert.PathAsserter ) {

    vr, err := t.getRestriction()
    if err != nil { 
        a.EqualErrors( t.Error, err ) 
        return
    }
    _, err = CreateAtomicTypeReference( t.Name, vr )
    a.EqualErrors( t.Error, err )
}

func TestAtomicRestrictionError( t *testing.T ) {
    la := assert.NewListPathAsserter( t )
    for _, test := range GetAtomicRestrictionErrorTests() {
        testAtomicRestrictionError( test, la )
        la = la.Next()
    }
}  

func TestCanAssign( t *testing.T ) {
    la := assert.NewListPathAsserter( t )
    int32Rng := NewAtomicTypeReference(
        QnameInt32,
        MustRangeRestriction( QnameInt32, true, Int32( 0 ), Int32( 1 ), true ),
    )
    mkList := func( typ *ListTypeReference, vals ...interface{} ) *List {
        res := MustList( vals... )
        res.Type = typ
        return res
    }
    for _, s := range []struct { 
        typ TypeReference
        val Value
        ignoreRestriction bool
        expctFail bool
    } {
        { typ: TypeNull, val: NullVal },
        { typ: TypeInt32, val: Int32( int32( 1 ) ) },
        { typ: TypeInt32, val: Int64( int64( 1 ) ), expctFail: true },
        { typ: int32Rng, val: Int32( 0 ) },
        { typ: int32Rng, val: Int32( 2 ), expctFail: true },
        { typ: int32Rng,
          val: Int32( 2 ),
          ignoreRestriction: true,
          expctFail: false,
        },
        { typ: TypeInt32,
          val: mkList( &ListTypeReference{ TypeInt32, true } ),
          expctFail: true,
        },
        { typ: &ListTypeReference{ TypeInt32, true }, 
          val: mkList( &ListTypeReference{ TypeInt32, true } ),
        },
        { typ: &ListTypeReference{ 
            &ListTypeReference{ TypeInt32, true }, 
            true,
          },
          val: mkList( 
            &ListTypeReference{ 
                &ListTypeReference{ TypeInt32, true }, 
                true,
            },
          ),
        },
        { typ: &ListTypeReference{ 
            &ListTypeReference{ TypeInt32, true }, 
            true,
          },
          val: mkList( &ListTypeReference{ TypeInt32, true } ),
          expctFail: true,
        },
        { typ: &ListTypeReference{ TypeInt32, true },
          val: mkList( 
            &ListTypeReference{ 
                &ListTypeReference{ TypeInt32, true }, 
                true,
            },
          ),
          expctFail: true,
        },
        { typ: &NullableTypeReference{ TypeString }, val: String( "s" ) },
        { typ: &NullableTypeReference{ TypeString }, val: NullVal },
        { typ: &NullableTypeReference{ TypeString }, val: NullVal },
        { typ: ns1V1At( "S1" ), val: MustStruct( ns1V1Qn( "S1" ) ) },
        { typ: ns1V1At( "S1" ), val: &Enum{ ns1V1Qn( "S1" ), mkId( "c1" ) },
        },
        { typ: ns1V1At( "S1" ), val: String( "s" ), expctFail: true },
        { typ: ns1V1At( "S1" ),
          val: MustStruct( ns1V1Qn( "S2" ) ),
          expctFail: true,
        },
        { typ: ns1V1At( "S1" ),
          val: &Enum{ ns1V1Qn( "S2" ), mkId( "c1" ) },
          expctFail: true,
        },
        { typ: TypeValue, val: String( "s" ) },
        { typ: TypeValue, val: NullVal, expctFail: true },
        { typ: TypeValue, val: MustStruct( ns1V1Qn( "S1" ) ) },
        { typ: TypeNullableValue, val: NullVal },
        { typ: TypeNullableValue, val: String( "s" ) },
    } {
//        la.Logf( 
//            "checking typ %s, val: %s, expctFail: %t, ignoreRestriction: %t",
//            s.typ, QuoteValue( s.val ), s.expctFail, s.ignoreRestriction )
        actFail := ! CanAssign( s.val, s.typ, ! s.ignoreRestriction )
        la.Descend( "expctFail" ).Equal( s.expctFail, actFail )
        la = la.Next()
    }
}

func TestCanAssignType( t *testing.T ) {
    a := assert.Asserter{ t }
    chkBase := func( from, to TypeReference, expct bool ) {
        act := CanAssignType( from, to )
        a.Equalf( expct, act, "assignment from %s --> %s, expct: %t, act: %t",
            from, to, expct, act )
    }
    chk := func( from, to TypeReference, expct bool ) {
        chkBase( from, to, expct )
        chkBase( from, TypeValue, true )
        chkBase( from, &NullableTypeReference{ TypeValue }, true )
    }
    mkLt := func( typ TypeReference, bools ...bool ) TypeReference {
        for _, allowsEmpty := range bools {
            typ = &ListTypeReference{ typ, allowsEmpty }
        }
        return typ
    }
    ltInt32 := func( bools ...bool ) TypeReference {
        return mkLt( TypeInt32, bools... )
    }
    int32Ptr := NewPointerTypeReference( TypeInt32 )
    chk( TypeInt32, TypeInt32, true )
    chk( TypeInt32, int32Ptr, false )
    chk( TypeInt32, TypeInt64, false )
    chk( int32Ptr, int32Ptr, true )
    chk( int32Ptr, &NullableTypeReference{ int32Ptr }, true )
    chk( 
        &NullableTypeReference{ int32Ptr }, 
        &NullableTypeReference{ int32Ptr }, 
        true,
    )
    chk( ltInt32( true ), TypeInt32, false )
    chk( ltInt32( true ), ltInt32( true ), true )
    chk( ltInt32( false ), ltInt32( false ), true )
    chk( ltInt32( false ), ltInt32( true ), false )
    chk( ltInt32( true ), ltInt32( false ), false )
    chk( ltInt32( true ), &NullableTypeReference{ ltInt32( true ) }, true )
    chk( &NullableTypeReference{ ltInt32( true ) }, ltInt32( true ), false )
    chk( TypeInt32, TypeValue, true )
    chk( ltInt32( true ), TypeValue, true )
    int32Rng := NewAtomicTypeReference(
        QnameInt32,
        MustRangeRestriction( QnameInt32, true, Int32( 0 ), Int32( 1 ), true ),
    )
    int32RngPtr := NewPointerTypeReference( int32Rng )
    chk( int32Rng, TypeInt32, true )
    chk( TypeInt32, int32Rng, false )
    chk( int32RngPtr, int32Ptr, false )
    chk( int32RngPtr, int32RngPtr, true )
    chk( int32RngPtr, &NullableTypeReference{ int32RngPtr }, true )
    chk( 
        &NullableTypeReference{ int32RngPtr },
        &NullableTypeReference{ int32RngPtr },
        true,
    )
    chk( &NullableTypeReference{ int32RngPtr }, int32RngPtr, false )
    chk( TypeValue, TypeValue, true )
    chk( ltInt32( true ), mkLt( TypeValue, true ), false )
    chk( ltInt32( false ), mkLt( TypeValue, false ), false )
}

type quoteValueAsserter struct {
    *assert.Asserter
}

func ( a *quoteValueAsserter ) call( v Value, strs ...string ) {
    q := QuoteValue( v )
    for _, str := range strs { if str == q { return } }
    a.Fatalf( "No vals in %#v matched quoted val %q", strs, q )
}

func TestQuoteValue( t *testing.T ) {
    a := &quoteValueAsserter{ &assert.Asserter{ t } }
    a.call( Boolean( true ), "true" )
    a.call( Boolean( false ), "false" )
    a.call( Buffer( []byte{ 0, 1, 2 } ), "buf[3]" )
    a.call( String( "s" ), `"s"` )
    a.call( Int32( 1 ), "1" )
    a.call( Int64( 1 ), "1" )
    a.call( Uint32( 1 ), "1" )
    a.call( Uint64( 1 ), "1" )
    a.call( Float32( 1.1 ), "1.1" )
    a.call( Float64( 1.1 ), "1.1" )
    tm := "2012-01-01T12:00:00Z"
    a.call( MustTimestamp( tm ), tm )
    en := &Enum{ ns1V1Qn( "E1" ), mkId( "v" ) }
    a.call( en, "ns1@v1/E1.v" )
    a.call( NullVal, "null" )
    a.call( MustList(), "[]" )
    a.call( MustList( String( "s" ), Int32( 1 ) ), `["s", 1]` )
    a.call( MustSymbolMap(), "{}" )
    a.call( MustSymbolMap( mkId( "k1" ), 1, mkId( "k2" ), "2" ),
        `{k1:1, k2:"2"}`, `{k2:"2", k1:1}` )
    fldK := mkId( "k" )
    map1 := MustSymbolMap( fldK, 1 )
    qn1 := ns1V1Qn( "S1" )
    s1 := &Struct{ Type: qn1, Fields: map1 }
    s1Str := `ns1@v1/S1{k:1}`
    a.call( s1, s1Str )
}

func TestIsNull( t *testing.T ) {
    assert.True( IsNull( &Null{} ) )
    assert.True( IsNull( NullVal ) )
    assert.False( IsNull( Int32( 1 ) ) )
}

func TestIdentifierCompare( t *testing.T ) {
    id1 := mkId( "a" )
    id2 := mkId( "a", "b1" )
    id3 := mkId( "b" )
    id4 := mkId( "b", "b1" )
    ids := []*Identifier{ id1, id2, id3, id4 }
    for i, e := 0, len( ids ) - 1; i < e; i++ {
        l, r := ids[ i ], ids[ i + 1 ]
        assert.True( l.Compare( r ) < 0 )
        assert.True( l.Compare( l.dup() ) == 0 )
        assert.True( r.Compare( l ) > 0 )
    }
}

func TestMustEnum( t *testing.T ) {
    assert.Equal(
        &Enum{ ns1V1Qn( "E1" ), mkId( "val1" ) },
        &Enum{ ns1V1Qn( "E1" ), mkId( "val1" ) },
    )
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
    id1 := mkId( "id1" )
    id2 := mkId( "id2" )
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
    m.Put( mkId( "id1" ), String( "s1" ) )
    m.Put( mkId( "id2" ), String( "s2" ) )
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
    ns := mkNs( mkId( "v1" ), mkId( "ns1" ) )
    qn1 := &QualifiedTypeName{ Namespace: ns, Name: dn1 }
    assert.Equal( qn1, dn1.ResolveIn( ns ) )
}

func TestSortIds( t *testing.T ) {
    chk := func( in, expct []*Identifier ) {
        assert.Equal( expct, SortIds( in ) )
    }
    chk( []*Identifier{}, []*Identifier{} )
    ids := []*Identifier{ mkId( "i1" ), mkId( "i2" ), mkId( "i3" ) }
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
    chk( "missing field(s): f1", mkId( "f1" ) )
    // check sorted
    chk( "missing field(s): f1, f2", mkId( "f2" ), mkId( "f1" ) ) 
}

func TestUnrecognizedFieldErrorFormatting( t *testing.T ) {
    assert.Equal(
        "unrecognized field: f1",
        NewUnrecognizedFieldError( nil, mkId( "f1" ) ).Error(),
    )
}

func TestTypeReferenceEquals( t *testing.T ) {
    chk0 := func( t1, t2 TypeReference, eq bool ) {
        if t1.Equals( t2 ) {
            if ! eq { t.Fatalf( "%s == %s", t1, t2 ) }
        } else {
            if eq { t.Fatalf( "%s != %s", t1, t2 ) }
        }
    }
    chk := func( t1, t2 TypeReference, eq bool ) {
        chk0( t1, t2, eq )
        chk0( t2, t1, eq )
    }
    qn1, qn2 := ns1V1Qn( "T1" ), ns1V1Qn( "T2" )
    at1 := NewAtomicTypeReference( qn1, nil )
    rgx := func( s string ) *RegexRestriction {
        res, err := CreateRegexRestriction( s )
        if err != nil { panic( err ) }
        return res
    }
    at1Rgx := NewAtomicTypeReference( qn1, rgx( ".*" ) )
    rng := func( i int32 ) *RangeRestriction {
        return MustRangeRestriction( 
            QnameInt32, true, Int32( i ), Int32( i + 1 ), true )
    }
    at1Rng := NewAtomicTypeReference( qn1, rng( 1 ) )
    at2 := NewAtomicTypeReference( qn2, nil )
    lt1Empty := &ListTypeReference{ ElementType: at1, AllowsEmpty: true }
    lt1NonEmpty := &ListTypeReference{ ElementType: at1, AllowsEmpty: false }
    pt1 := NewPointerTypeReference( at1 )
    pt2 := NewPointerTypeReference( at2 )
    nt1 := MustNullableTypeReference( pt1 )
    nt2 := MustNullableTypeReference( pt2 )
    chk( at1, at1, true )
    chk( at1, NewAtomicTypeReference( qn1, nil ), true )
    chk( at1, at2, false )
    chk( at1Rgx, at1Rgx, true )
    chk( at1Rgx, NewAtomicTypeReference( qn1, rgx( ".*" ) ), true )
    chk( at1Rgx, at1, false )
    chk( at1Rgx, NewAtomicTypeReference( qn1, rgx( "a.*" ) ), false )
    chk( at1Rgx, NewAtomicTypeReference( qn2, rgx( ".*" ) ), false )
    chk( at1Rgx, at1Rng, false )
    chk( at1Rng, NewAtomicTypeReference( qn1, rng( 1 ) ), true )
    chk( at1Rng, at1, false )
    chk( at1Rng, NewAtomicTypeReference( qn1, rng( 2 ) ), false )
    chk( at1Rng, NewAtomicTypeReference( qn2, rng( 1 ) ), false )
    chk( lt1Empty, lt1Empty, true )
    chk( lt1Empty, lt1NonEmpty, false )
    chk( lt1Empty, 
        &ListTypeReference{ ElementType: at2, AllowsEmpty: true }, false )
    chk( lt1Empty, 
        &ListTypeReference{ ElementType: lt1Empty, AllowsEmpty: true }, false )
    chk( pt1, pt1, true )
    chk( pt1, NewPointerTypeReference( at1 ), true )
    chk( pt1, pt2, false )
    chk( pt1, NewPointerTypeReference( pt1 ), false )
    chk( pt1, NewPointerTypeReference( lt1NonEmpty ), false )
    chk( pt1, NewPointerTypeReference( nt1 ), false )
    chk( NewPointerTypeReference( lt1Empty ), 
        NewPointerTypeReference( lt1Empty ), true )
    chk( nt1, nt1, true )
    chk( nt1, MustNullableTypeReference( pt1 ), true )
    chk( nt1, nt2, false )
    chk( nt1, lt1Empty, false )
}

type numberParseTest struct {
    *assert.PathAsserter
    in string
    out Value
    typ *QualifiedTypeName
    err error
}

func ( t *numberParseTest ) call() {
//    t.Logf( "parsing %q as %s", t.in, t.typ )
    if act, err := ParseNumber( t.in, t.typ ); err == nil {
        AssertEqualValues( t.out, act, t )
    } else { t.EqualErrors( t.err, err ) }
}

func TestNumberParsers( t *testing.T ) {
    la := assert.NewListPathAsserter( t )
    tests := make( []*numberParseTest, 0, 16 )
    tests = append( tests,
        &numberParseTest{ 
            in: "1", out: Int32( int32( 1 ) ), typ: QnameInt32 },
        &numberParseTest{ 
            in: "1", out: Uint32( uint32( 1 ) ), typ: QnameUint32 },
        &numberParseTest{ 
            in: "1", out: Int64( int64( 1 ) ), typ: QnameInt64 },
        &numberParseTest{ 
            in: "1", out: Uint64( uint64( 1 ) ), typ: QnameUint64 },
        &numberParseTest{ 
            in: "1.1", out: Float32( float32( 1.1 ) ), typ: QnameFloat32 },
        &numberParseTest{ 
            in: "1.1", out: Float64( float64( 1.1 ) ), typ: QnameFloat64 },
    )
    rngErr := func( val string, typ *QualifiedTypeName ) {
        err := newNumberRangeError( val )
        tests = append( tests, &numberParseTest{ in: val, typ: typ, err: err } )
    }
    rngErr( "2147483648", QnameInt32 )
    rngErr( "-2147483649", QnameInt32 )
    rngErr( "4294967296", QnameUint32 )
    rngErr( "-1", QnameUint32 )
    rngErr( "9223372036854775808", QnameInt64 )
    rngErr( "-9223372036854775809", QnameInt64 )
    rngErr( "18446744073709551616", QnameUint64 )
    rngErr( "-1", QnameUint64 )
    sxErr := func( in string, typ *QualifiedTypeName ) {
        err := newNumberSyntaxError( typ, in )
        test := &numberParseTest{ in: in, typ: typ, err: err }
        tests = append( tests, test )
    }
    sxErr( "1.1", QnameInt32 )
    sxErr( "1.1", QnameUint32 )
    sxErr( "1.1", QnameInt64 )
    sxErr( "1.1", QnameUint64 )
    for _, qn := range NumericTypeNames { sxErr( "badNum", qn ) }
    for _, t := range tests {
        t.PathAsserter = la
        t.call()
        la = la.Next()
    }
}

func assertEquality( c1, c2 interface{}, eq, expctEq bool, t *testing.T ) {
    if expctEq && ( ! eq ) { t.Fatalf( "%v != %v", c1, c2 ) }
    if eq && ( ! expctEq ) { t.Fatalf( "%v == %v", c1, c2 ) }
}

type equalityTest struct {
    lhs, rhs interface{}
    expctEq bool
}

func assertEqualityTests( 
    f func( a, b interface{} ) bool, t *testing.T, tests ...equalityTest ) {
    for _, test := range tests {
        eq := f( test.lhs, test.rhs )
        assertEquality( test.lhs, test.rhs, eq, test.expctEq, t )
    }
}

func assertIdComp( id1, id2 *Identifier, expctCmp int, t *testing.T ) {
    if id2 != nil {
        cmp := id1.Compare( id2 )
        assert.Equal( expctCmp, cmp )
    }
    assertEquality( id1, id2, id1.Equals( id2 ), expctCmp == 0, t )
}

func TestIdentifierComparison( t *testing.T ) {
    idA := mkId( "id", "a" )
    idB := mkId( "id", "b" )
    assertIdComp( idA, nil, -1, t )
    assertIdComp( idA, idA, 0, t )
    assertIdComp( idA, idB, -1, t )
    assertIdComp( idB, idA, 1, t )
}

func assertIdentFormat(
    id *Identifier, fmt IdentifierFormat, expct string, t *testing.T ) {
    act := id.Format( fmt )
    assert.Equal( expct, act )
}

func TestIdentifierFormatting( t *testing.T ) {
    id := mkId( "test", "id", "a", "b", "c" )
    assertIdentFormat( id, LcHyphenated, "test-id-a-b-c", t )
    assertIdentFormat( id, LcUnderscore, "test_id_a_b_c", t )
    assertIdentFormat( id, LcCamelCapped, "testIdABC", t )
    id2 := mkId( "test" )
    for _, idFmt := range IdentifierFormats {
        assertIdentFormat( id2, idFmt, "test", t )
    }
}

func TestNamespaceStringer( t *testing.T ) {
    ns := mkNs( mkId( "v1" ), mkId( "ns1" ), mkId( "ns2" ) )
    assert.Equal( "ns1:ns2@v1", ns.String() )
}

func TestNamespaceEquality( t *testing.T ) {
    ns1A := mkNs( mkId( "v1" ), mkId( "ns1" ) )
    ns1B := mkNs( mkId( "v1" ), mkId( "ns1" ) )
    ns2 := mkNs( mkId( "v1" ), mkId( "ns1" ), mkId( "ns2" ) )
    tests := []equalityTest{
        equalityTest{ ns1A, ns1A, true },
        equalityTest{ ns1A, ns1B, true },
        equalityTest{ ns1A, ns2, false },
        equalityTest{ ns2, ns1A, false },
    }
    f := func( a, b interface{} ) bool {
        return a.( *Namespace ).Equals( b.( *Namespace ) )
    }
    assertEqualityTests( f, t, tests... )
    assertEquality( ns1A, nil, ns1A.Equals( nil ), false, t )
}

func TestDeclaredTypeNameStringer( t *testing.T ) {
    str := "TypeName"
    dt := mkDeclNm( str )
    assert.Equal( str, dt.ExternalForm() )
    assert.Equal( str, dt.String() )
}

func TestDeclaredTypeNameEquality( t *testing.T ) {
    nm1A := mkDeclNm( "T1" )
    nm1B := mkDeclNm( "T1" )
    nm2 := mkDeclNm( "T2" )
    assertEqualityTests(
        func( a, b interface{} ) bool {
            return a.( *DeclaredTypeName ).Equals( b.( *DeclaredTypeName ) )
        },
        t,
        equalityTest{ nm1A, nm1A, true },
        equalityTest{ nm1A, nm1B, true },
        equalityTest{ nm1B, nm1A, true },
        equalityTest{ nm1A, nm2, false },
        equalityTest{ nm2, nm1A, false },
    )
    assertEquality( nm1A, nil, nm1A.Equals( nil ), false, t )
}

// Also gives coverage of MustQualifiedTypeName()
func TestQualifiedTypeNameStringer( t *testing.T ) {
    qn := mkQn( mkNs( mkId( "v1" ), mkId( "ns1" ) ), mkDeclNm( "T1" ) )
    str := "ns1@v1/T1"
    assert.Equal( str, qn.String() )
    assert.Equal( str, qn.ExternalForm() )
}

func TestQualifiedTypeNameEquality( t *testing.T ) {
    nm1A := mkQn( mkNs( mkId( "v1" ), mkId( "ns1" ) ), mkDeclNm( "T1" ) )
    nm1B := mkQn( mkNs( mkId( "v1" ), mkId( "ns1" ) ), mkDeclNm( "T1" ) )
    assertEquality( nm1A, nil, nm1A.Equals( nil ), false, t )
    assertEqualityTests(
        func( a, b interface{} ) bool {
            return a.( *QualifiedTypeName ).Equals( b.( *QualifiedTypeName ) )
        },
        t,
        equalityTest{ nm1A, nm1A, true },
        equalityTest{ nm1A, nm1B, true },
        equalityTest{ 
            nm1A, 
            mkQn( mkNs( mkId( "v1" ), mkId( "ns1" ) ), mkDeclNm( "T2" ) ),
            false,
        },
        equalityTest{ 
            nm1A, 
            mkQn( mkNs( mkId( "v2" ), mkId( "ns1" ) ), mkDeclNm( "T1" ) ),
            false,
        },
        equalityTest{ 
            nm1A, 
            mkQn( mkNs( mkId( "v1" ), mkId( "ns2" ) ), mkDeclNm( "T1" ) ),
            false,
        },
        equalityTest{ 
            nm1A, 
            mkQn( 
                mkNs( mkId( "v1" ), mkId( "ns1" ), mkId( "ns2" ) ), 
                mkDeclNm( "T1" ),
            ),
            false,
        },
        equalityTest{ 
            mkQn( 
                mkNs( mkId( "v1" ), mkId( "ns1" ), mkId( "ns2" ) ), 
                mkDeclNm( "T1" ),
            ),
            mkQn( 
                mkNs( mkId( "v1" ), mkId( "ns1" ), mkId( "ns2" ) ), 
                mkDeclNm( "T1" ),
            ),
            true,
        },
    )
}

func TestTypeReferenceStringer( t *testing.T ) {
    chk := func( extForm string, typ TypeReference ) {
        assert.Equal( extForm, typ.String() )
        assert.Equal( extForm, typ.ExternalForm() )
    }
    qn := mkQn( mkNs( mkId( "v1" ), mkId( "ns1" ) ), mkDeclNm( "T1" ) )
    at := NewAtomicTypeReference( qn, nil )
    ptr := NewPointerTypeReference( at )
    chk( "ns1@v1/T1", at )
    chk( 
        "ns1@v1/T1*", 
        &ListTypeReference{ ElementType: at, AllowsEmpty: true },
    )
    chk( 
        "ns1@v1/T1+",
        &ListTypeReference{ ElementType: at, AllowsEmpty: false },
    )
    chk( 
        "ns1@v1/T1*++",
        &ListTypeReference{ 
            ElementType: &ListTypeReference{
                ElementType: &ListTypeReference{
                    ElementType: at,
                    AllowsEmpty: true,
                },
                AllowsEmpty: false,
            },
            AllowsEmpty: false,
        },
    )
    chk( "&(ns1@v1/T1)", ptr )
    chk( "&(&(ns1@v1/T1))", NewPointerTypeReference( ptr ) )
    chk( "&(ns1@v1/T1)?", &NullableTypeReference{ ptr } )
    chk(
        "&(ns1@v1/T1)?**+",
        &ListTypeReference{
            ElementType: &ListTypeReference{
                ElementType: &ListTypeReference{
                    ElementType: &NullableTypeReference{ ptr },
                    AllowsEmpty: true,
                },
                AllowsEmpty: true,
            },
            AllowsEmpty: false,
        },
    )
    chk(
        "&(&(ns1@v1/T1)*)",
        NewPointerTypeReference(
            &ListTypeReference{ ElementType: ptr, AllowsEmpty: true },
        ),
    )
}

func TestTimestampStrings( t *testing.T ) {
    tm := time.Now()
    mgTm := Timestamp( tm )
    assert.Equal( tm.Format( time.RFC3339Nano ), mgTm.String() )
    assert.Equal( tm.Format( time.RFC3339Nano ), mgTm.Rfc3339Nano() )
}

func TestCastValueErrorFormatting( t *testing.T ) {
    path := objpath.RootedAt( mkId( "f1" ) )
    err := NewInputErrorf( path, "Blah %s", "X" )
    assert.Equal( "f1: Blah X", err.Error() )
}

func TestEqualValues( t *testing.T ) {
    m1 := MustSymbolMap( mkId( "f1" ), int32( 1 ) )
    m2 := MustSymbolMap( mkId( "f1" ), int32( 2 ) )
    m3 := MustSymbolMap( mkId( "f1" ), int32( 1 ), mkId( "f2" ), int32( 2 ) )
    la := assert.NewListPathAsserter( t )
    for _, t := range []struct{ v1, v2 interface{}; eq bool } {
        { true, true, true },
        { true, false, false },
        { NullVal, NullVal, true },
        { int32( 0 ), int32( 0 ), true },
        { int32( 0 ), int32( 1 ), false },
        { int64( 0 ), int64( 0 ), true },
        { int64( 0 ), int64( 1 ), false },
        { uint32( 0 ), uint32( 0 ), true },
        { uint32( 0 ), uint32( 1 ), false },
        { uint64( 0 ), uint64( 0 ), true },
        { uint64( 0 ), uint64( 1 ), false },
        { float32( 0 ), float32( 0 ), true },
        { float32( 0 ), float32( 1 ), false },
        { float64( 0 ), float64( 0 ), true },
        { float64( 0 ), float64( 1 ), false },
        { int32( 0 ), int64( 0 ), false }, // not re-checking all mismatches
        { []byte{}, []byte{}, true },
        { []byte{ 0 }, []byte{ 0 }, true },
        { []byte{ 1 }, []byte{ 0 }, false },
        { "a", "a", true },
        { "a", "b", false },
        { "a", int32( 1 ), false },
        {
            MustTimestamp( "2014-01-01T00:00:00Z" ),
            MustTimestamp( "2014-01-01T00:00:00Z" ),
            true,
        },
        {
            MustTimestamp( "2014-01-01T00:00:00Z" ),
            MustTimestamp( "2014-01-02T00:00:00Z" ),
            false,
        },
        { 
            &Enum{ ns1V1Qn( "E1" ), mkId( "v1" ) },
            &Enum{ ns1V1Qn( "E1" ), mkId( "v1" ) },
            true,
        },
        { 
            &Enum{ ns1V1Qn( "E2" ), mkId( "v1" ) },
            &Enum{ ns1V1Qn( "E1" ), mkId( "v1" ) },
            false,
        },
        { 
            &Enum{ ns1V1Qn( "E1" ), mkId( "v1" ) },
            &Enum{ ns1V1Qn( "E1" ), mkId( "v2" ) },
            false,
        },
        { EmptySymbolMap(), EmptySymbolMap(), true },
        { m1, EmptySymbolMap(), false },
        { m1, m1, true },
        { m1, m2, false },
        { m1, m3, false },
        { 
            &Struct{ ns1V1Qn( "S1" ), m1 },
            &Struct{ ns1V1Qn( "S1" ), m1 },
            true,
        },
        { 
            &Struct{ ns1V1Qn( "S1" ), m1 },
            &Struct{ ns1V1Qn( "S2" ), m1 },
            false,
        },
        { 
            &Struct{ ns1V1Qn( "S1" ), m1 },
            &Struct{ ns1V1Qn( "S1" ), m2 },
            false,
        },
        { EmptyList(), EmptyList(), true },
        {
            EmptyList(),
            &List{ 
                &ListTypeReference{ TypeInt32, true },
                []Value{ Int32( 0 ), Int32( 1 ) },
            },
            false,
        },
        {
            &List{ 
                &ListTypeReference{ TypeInt32, true },
                []Value{ Int32( 0 ), Int32( 1 ) },
            },
            &List{ 
                &ListTypeReference{ TypeInt32, true },
                []Value{ Int32( 0 ), Int32( 1 ) },
            },
            true,
        },
        {
            &List{ 
                &ListTypeReference{ TypeInt32, true },
                []Value{ Int32( 0 ), Int32( 1 ) },
            },
            &List{ 
                &ListTypeReference{ TypeInt32, true },
                []Value{ Int32( 0 ), Int32( 2 ) },
            },
            false,
        },
        {
            &List{ 
                &ListTypeReference{ TypeInt64, true },
                []Value{ Int64( 0 ), Int64( 1 ) },
            },
            &List{ 
                &ListTypeReference{ TypeInt32, true },
                []Value{ Int32( 0 ), Int32( 1 ) },
            },
            false,
        },
        {
            &List{ 
                &ListTypeReference{ TypeValue, true },
                []Value{ Int32( 0 ), Int32( 1 ) },
            },
            &List{ 
                &ListTypeReference{ TypeInt32, true },
                []Value{ Int32( 0 ), Int32( 1 ) },
            },
            false,
        },
    } {
        chk := func( v1, v2 Value ) {
            act := EqualValues( v1, v2 )
            la.Equalf( t.eq, act, "%s == %s: %t, expct: %t",
                QuoteValue( v1 ), QuoteValue( v2 ), act, t.eq )
        }
        v1, v2 := MustValue( t.v1 ), MustValue( t.v2 )
        chk( v1, v2 )
        chk( v2, v1 )
        la = la.Next()
    }
}

func TestCreateDeclaredTypeNameError( t *testing.T ) {
    type dnTest struct { in string; msg string }
    la := assert.NewListPathAsserter( t )
    for _, test := range []dnTest{
        { in: "", msg: `invalid type name: ""` },
        { in: "A Name", msg: `invalid type name: "A Name"` },
        { in: "A$Bad", msg: `invalid type name: "A$Bad"` },
        { in: " WhitespaceBad ", msg: `invalid type name: " WhitespaceBad "` },
    } {
        expct := newDeclaredTypeNameError( test.msg )
        _, err := CreateDeclaredTypeName( test.in )
        la.EqualErrors( expct, err )
        la = la.Next()
    }
}
