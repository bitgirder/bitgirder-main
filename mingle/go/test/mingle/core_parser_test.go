package mingle

import (
    "testing"
    "bitgirder/assert"
    "time"
//    "log"
)

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
    at := &AtomicTypeReference{ Name: qn }
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
    chk( "&ns1@v1/T1", ptr )
    chk( "&ns1@v1/T1?", &NullableTypeReference{ ptr } )
    chk(
        "&ns1@v1/T1?**+",
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
}

func TestTimestampStrings( t *testing.T ) {
    tm := time.Now()
    mgTm := Timestamp( tm )
    assert.Equal( tm.Format( time.RFC3339Nano ), mgTm.String() )
    assert.Equal( tm.Format( time.RFC3339Nano ), mgTm.Rfc3339Nano() )
}
