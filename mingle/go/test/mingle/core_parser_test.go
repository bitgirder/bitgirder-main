package mingle

import (
    "testing"
    "bitgirder/assert"
    "fmt"
    "time"
    "mingle/parser/loc"
    "mingle/parser/lexer"
    "mingle/parser/syntax"
    pt "mingle/parser/testing"
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

func parseCore( cpt *pt.CoreParseTest ) ( interface{}, error ) {
    switch cpt.TestType {
    case pt.TestTypeIdentifier: return ParseIdentifier( cpt.In )
    case pt.TestTypeNamespace: return ParseNamespace( cpt.In )
    case pt.TestTypeDeclaredTypeName: return ParseDeclaredTypeName( cpt.In )
    case pt.TestTypeQualifiedTypeName: return ParseQualifiedTypeName( cpt.In )
    case pt.TestTypeIdentifiedName: return ParseIdentifiedName( cpt.In )
    case pt.TestTypeTypeReference: return ParseTypeReference( cpt.In )
    }
    panic( fmt.Errorf( "Unhandled TestType: %s", cpt.TestType ) )
}

func assertCoreParseError( 
    cpt *pt.CoreParseTest, err error, a *assert.PathAsserter ) {
    switch v := cpt.Err.( type ) {
    case nil: a.Fatal( err )
    case *pt.ParseErrorExpect:
        if pErr, ok := err.( *loc.ParseError ); ok {
            a.Descend( "Col" ).Equal( v.Col, pErr.Loc.Col )
            a.Descend( "Message" ).Equal( v.Message, pErr.Message )
        } else { a.Fatal( err ) }
    case pt.RestrictionErrorExpect:
        a.Equal( &RestrictionTypeError{ string( v ) }, err )
    default: panic( fmt.Errorf( "Unhandled error expectation type %T", v ) )
    }
}

func convertPtIdentifier( id pt.Identifier ) *Identifier {
    return &Identifier{ []string( id ) }
}

func convertPtIdentifiers( ids []pt.Identifier ) []*Identifier {
    res := make( []*Identifier, len( ids ) )
    for i, part := range ids { res[ i ] = convertPtIdentifier( part ) }
    return res
}

func convertPtNamespace( ns *pt.Namespace ) *Namespace {
    return &Namespace{
        Version: convertPtIdentifier( ns.Version ),
        Parts: convertPtIdentifiers( ns.Parts ),
    }
}

func convertPtDeclTypeName( nm pt.DeclaredTypeName ) *DeclaredTypeName {
    return &DeclaredTypeName{ string( nm ) }
}

func convertPtQualifiedTypeName( nm *pt.QualifiedTypeName ) *QualifiedTypeName {
    return &QualifiedTypeName{
        Namespace: convertPtNamespace( nm.Namespace ),
        Name: convertPtDeclTypeName( nm.Name ),
    }
}

func convertPtIdentifiedName( nm *pt.IdentifiedName ) *IdentifiedName {
    return &IdentifiedName{
        Namespace: convertPtNamespace( nm.Namespace ),
        Names: convertPtIdentifiers( nm.Names ),
    }
}

func convertPtRangeVal( data interface{} ) Value {
    switch v := data.( type ) {
    case nil: return Value( nil )
    case int32: return Int32( v )
    case int64: return Int64( v )
    case uint32: return Uint32( v )
    case uint64: return Uint64( v )
    case float32: return Float32( v )
    case float64: return Float64( v )
    case string: return String( v )
    case pt.Timestamp: return MustTimestamp( string( v ) )
    }
    panic( fmt.Errorf( "Unhandled pt range val type %T", data ) )
}

func convertPtRangeRestriction( rr *pt.RangeRestriction ) *RangeRestriction {
    return &RangeRestriction{
        MinClosed: rr.MinClosed,
        Min: convertPtRangeVal( rr.Min ),
        Max: convertPtRangeVal( rr.Max ),
        MaxClosed: rr.MaxClosed,
    }
}

func convertPtRestriction( rVal interface{} ) ValueRestriction {
    switch r := rVal.( type ) {
    case pt.RegexRestriction: 
        res, err := NewRegexRestriction( string( r ) )
        if err != nil { panic( err ) }
        return res
    case *pt.RangeRestriction: return convertPtRangeRestriction( r )
    case nil: return nil
    }
    panic( fmt.Errorf( "Unhandled pt restriction type %T", rVal ) )
}

func convertPtTypeReference( tVal interface{} ) TypeReference {
    switch t := tVal.( type ) {
    case *pt.AtomicTypeReference: 
        return &AtomicTypeReference{
            Name: convertPtVal( t.Name ).( *QualifiedTypeName ),
            Restriction: convertPtRestriction( t.Restriction ),
        }
    case *pt.ListTypeReference:
        return &ListTypeReference{ 
            ElementType: convertPtTypeReference( t.ElementType ),
            AllowsEmpty: t.AllowsEmpty,
        }
    case *pt.NullableTypeReference:
        return NewNullableTypeReference( convertPtTypeReference( t.Type ) )
    }
    panic( fmt.Errorf( "Unhandled pt type reference type %T", tVal ) )
}

func convertPtVal( val interface{} ) interface{} {
    switch v := val.( type ) {
    case pt.Identifier: return convertPtIdentifier( v )
    case *pt.Namespace: return convertPtNamespace( v )
    case pt.DeclaredTypeName: return convertPtDeclTypeName( v )
    case *pt.QualifiedTypeName: return convertPtQualifiedTypeName( v )
    case *pt.IdentifiedName: return convertPtIdentifiedName( v )
    case *pt.AtomicTypeReference, 
         *pt.ListTypeReference,
         *pt.NullableTypeReference: 
        return convertPtTypeReference( v )
    }
    panic( fmt.Errorf( "Unhandled pt val type (%T)", val ) )
}

func assertExternalForm( ext string, val interface{}, a *assert.PathAsserter ) {
    if v, ok := val.( extFormer ); ok {
        a.Equal( ext, v.ExternalForm() )
    } else { a.Fatalf( "Not an external former: %T", val ) }
}

func assertCoreParse( cpt *pt.CoreParseTest, a *assert.PathAsserter ) {
    if cpt.TestType == pt.TestTypeString || cpt.TestType == pt.TestTypeNumber { 
        return 
    }
    if val, err := parseCore( cpt ); err == nil {
        if cpt.Err == nil {
            conv := convertPtVal( cpt.Expect )
            a.Equal( conv, val )
            if ext := cpt.ExternalForm; ext != "" { 
                assertExternalForm( ext, conv, a.Descend( "ExternalForm" ) )
            }
        } else { a.Fatalf( "Expected parse failure: %#v", cpt.Err ) }
    } else { assertCoreParseError( cpt, err, a.Descend( "Err" ) ) }
}

func TestCoreParser( t *testing.T ) {
    a := assert.NewPathAsserter( t ).StartList()
    for _, cpt := range pt.CoreParseTests {
        assertCoreParse( cpt, a )
        a = a.Next()
    }
}

func TestIdentifierMustPanic( t *testing.T ) {
    errExpct := &pt.ParseErrorExpect{ 
        5, "Illegal start of identifier part: \"$\" (U+0024)" }
    pt.AssertParsePanic( 
        errExpct, t, func() { MustIdentifier( "bad-$ident" ) } )
}

func TestIdentifierStringer( t *testing.T ) {
    if id, err := ParseIdentifier( "test-id" ); err == nil {
        assert.Equal( "test-id", id.String() )
    } else { t.Fatal( err ) }
}

func assertIdComp( id1, id2 *Identifier, expctCmp int, t *testing.T ) {
    if id2 != nil {
        cmp := id1.Compare( id2 )
        assert.Equal( expctCmp, cmp )
    }
    assertEquality( id1, id2, id1.Equals( id2 ), expctCmp == 0, t )
}

func TestIdentifierComparison( t *testing.T ) {
    id1A := MustIdentifier( "id-a" )
    id1B := MustIdentifier( "idA" )
    id2 := MustIdentifier( "idB" )
    assertIdComp( id1A, nil, -1, t )
    assertIdComp( id1A, id1B, 0, t )
    assertIdComp( id1A, id1A, 0, t )
    assertIdComp( id1B, id1A, 0, t )
    assertIdComp( id1A, id2, -1, t )
    assertIdComp( id2, id1A, 1, t )
}

func assertIdentFormat(
    id *Identifier, fmt IdentifierFormat, expct string, t *testing.T ) {
    act := id.Format( fmt )
    assert.Equal( expct, act )
}

func TestIdentifierFormatting( t *testing.T ) {
    id := MustIdentifier( "test-id-a-b-c" )
    assertIdentFormat( id, LcHyphenated, "test-id-a-b-c", t )
    assertIdentFormat( id, LcUnderscore, "test_id_a_b_c", t )
    assertIdentFormat( id, LcCamelCapped, "testIdABC", t )
    id2 := MustIdentifier( "test" )
    for _, idFmt := range IdentifierFormats {
        assertIdentFormat( id2, idFmt, "test", t )
    }
}

func TestNamespaceStringer( t *testing.T ) {
    ns := MustNamespace( "ns1:ns2@v1" )
    assert.Equal( "ns1:ns2@v1", ns.String() )
}

func TestNamespaceMustPanic( t *testing.T ) {
    errExpct := &pt.ParseErrorExpect{ 9, `Expected ':' or '@' but found: END` }
    pt.AssertParsePanic( errExpct, t, func() { MustNamespace( "not-a-ns" ) } )
}

func TestNamespaceEquality( t *testing.T ) {
    ns1A := MustNamespace( "ns1@v1" )
    ns1B := MustNamespace( "ns1@v1" )
    tests := []equalityTest{
        equalityTest{ ns1A, ns1A, true },
        equalityTest{ ns1A, ns1B, true },
    }
    ns2 := MustNamespace( "ns1:ns2@v1" )
    strs := []string { "ns1@v2", "ns2@v1", "ns1:ns3@v1", "ns1:ns2@v2" }
    for _, str := range strs {
        ns := MustNamespace( str )
        tests = append( tests, equalityTest{ ns1A, ns, false } )
        tests = append( tests, equalityTest{ ns2, ns, false } )
    }
    f := func( a, b interface{} ) bool {
            return a.( *Namespace ).Equals( b.( *Namespace ) )
         }
    assertEqualityTests( f, t, tests... )
    assertEquality( ns1A, nil, ns1A.Equals( nil ), false, t )
}

func TestDeclaredTypeNameMustPanic( t *testing.T ) {
    errExpct := &pt.ParseErrorExpect{ 
        1, "Illegal type name start: \"a\" (U+0061)" }
    pt.AssertParsePanic( 
        errExpct, t, func() { MustDeclaredTypeName( "aBad" ) } )
}

func TestDeclaredTypeNameStringer( t *testing.T ) {
    str := "TypeName"
    dt := MustDeclaredTypeName( str )
    assert.Equal( str, dt.ExternalForm() )
    assert.Equal( str, dt.String() )
}

func TestDeclaredTypeNameEquality( t *testing.T ) {
    nm1A := MustDeclaredTypeName( "T1" )
    nm1B := MustDeclaredTypeName( "T1" )
    nm2 := MustDeclaredTypeName( "T2" )
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

func TestQualifiedTypeNameMustPanic( t *testing.T ) {
    errExpct := &pt.ParseErrorExpect{ 
        7, "Expected type path but found: END" }
    pt.AssertParsePanic( 
        errExpct, t, func() { MustQualifiedTypeName( "ns1@v1" ) } )
}

// Also gives coverage of MustQualifiedTypeName()
func TestQualifiedTypeNameStringer( t *testing.T ) {
    str := "ns1@v1/T1"
    qn := MustQualifiedTypeName( str )
    assert.Equal( str, qn.String() )
    assert.Equal( str, qn.ExternalForm() )
}

func TestQualifiedTypeNameEquality( t *testing.T ) {
    nm1A := MustQualifiedTypeName( "ns1@v1/T1" )
    nm1B := MustQualifiedTypeName( "ns1@v1/T1" )
    assertEquality( nm1A, nil, nm1A.Equals( nil ), false, t )
    assertEqualityTests(
        func( a, b interface{} ) bool {
            return a.( *QualifiedTypeName ).Equals( b.( *QualifiedTypeName ) )
        },
        t,
        equalityTest{ nm1A, nm1A, true },
        equalityTest{ nm1A, nm1B, true },
        equalityTest{ nm1A, MustQualifiedTypeName( "ns1@v1/T2" ), false },
        equalityTest{ nm1A, MustQualifiedTypeName( "ns1@v2/T1" ), false },
        equalityTest{ nm1A, MustQualifiedTypeName( "ns2@v1/T1" ), false },
        equalityTest{ nm1A, MustQualifiedTypeName( "ns1:ns2@v1/T1" ), false },
        equalityTest{ 
            MustQualifiedTypeName( "ns1:ns2@v1/T1" ),
            MustQualifiedTypeName( "ns1:ns2@v1/T1" ),
            true,
        },
    )
}

// Used to test just the syntax (not semantics) of parsing type restrictions.
// Technically this could be in mingle/parser/syntax tests, not in here, but the
// assertions themselves are simplified by being able to easily parse and assert
// equality of things like the base type reference. Rather than duplicate that
// in mingle/parser/syntax just to aid in testing though, we carry out the tests
// here in this package.
type restrictionSyntaxTest struct {
    str string
    typ string
    restriction syntax.RestrictionSyntax
    *testing.T // set before call()
}

func ( t *restrictionSyntaxTest ) call() {
    if ctr, _, err := parseCompletableTypeReference( t.str ); err == nil {
        assert.Equal( t.restriction, ctr.Restriction )
    } else { t.Fatal( err ) }
}

// This is meant to cover things specific to the parser itself, namely the
// placement of parse locations; things like parse errors and interpretation of
// valid/invalid ranges of standard types is covered by the type reference tests
// in modelParseTests
func TestTypeReferenceRestrictionSyntax( t *testing.T ) {
    // quick func to create an expected (l)ocation by (c)olumn
    lc := func( col int ) *loc.Location {
        return &loc.Location{ 1, col, loc.ParseSourceInput }
    }
    tests := []*restrictionSyntaxTest{
        &restrictionSyntaxTest{
            str: "String1~\"^a+$\"",
            typ: "String1",
            restriction: &syntax.RegexRestrictionSyntax{ "^a+$", lc( 9 ) },
        },
        &restrictionSyntaxTest{
            str: "String1 ~\n\t\"a*\"?",
            typ: "String1?",
            restriction: &syntax.RegexRestrictionSyntax{ 
                Pat: "a*", 
                Loc: &loc.Location{ 2, 2, loc.ParseSourceInput },
            },
        },
        &restrictionSyntaxTest{
            str: "ns1:ns2@v1/String~\"B*\"",
            typ: "ns1:ns2@v1/String",
            restriction: &syntax.RegexRestrictionSyntax{ "B*", lc( 19 ) },
        },
        &restrictionSyntaxTest{
            str: "ns1:ns2@v1/String~\"a|b*\"*+",
            typ: "ns1:ns2@v1/String*+",
            restriction: &syntax.RegexRestrictionSyntax{ "a|b*", lc( 19 ) },
        },
        &restrictionSyntaxTest{
            str: "mingle:core@v1/String~\"a$\"",
            typ: "mingle:core@v1/String",
            restriction: &syntax.RegexRestrictionSyntax{ "a$", lc( 23 ) },
        },
        &restrictionSyntaxTest{
            str: "Num~[1,2]",
            typ: "Num",
            restriction: &syntax.RangeRestrictionSyntax{
                true,
                &syntax.NumRestrictionSyntax{ 
                    false, &lexer.NumericToken{ Int: "1" }, lc( 6 ) },
                &syntax.NumRestrictionSyntax{ 
                    false, &lexer.NumericToken{ Int: "2" }, lc( 8 ) },
                true,
            },
        },
        &restrictionSyntaxTest{
            str: "Num~[-1,+2]",
            typ: "Num",
            restriction: &syntax.RangeRestrictionSyntax{
                true,
                &syntax.NumRestrictionSyntax{ 
                    true, &lexer.NumericToken{ Int: "1" }, lc( 6 ) },
                &syntax.NumRestrictionSyntax{ 
                    false, &lexer.NumericToken{ Int: "2" }, lc( 9 ) },
                true,
            },
        },
        &restrictionSyntaxTest{
            str: "Num~(,8.0e3]?*",
            typ: "Num?*",
            restriction: &syntax.RangeRestrictionSyntax{
                false,
                nil,
                &syntax.NumRestrictionSyntax{ 
                    false, &lexer.NumericToken{ "8", "0", "3", 'e' }, lc( 7 ) },
                true,
            },
        },
        &restrictionSyntaxTest{
            str: "Num~( 8, )",
            typ: "Num",
            restriction: &syntax.RangeRestrictionSyntax{
                false,
                &syntax.NumRestrictionSyntax{ 
                    false, &lexer.NumericToken{ "8", "", "", 0 }, lc( 7 ) },
                nil,
                false,
            },
        },
        &restrictionSyntaxTest{
            str: "Num~(-100,100)?",
            typ: "Num?",
            restriction: &syntax.RangeRestrictionSyntax{
                false,
                &syntax.NumRestrictionSyntax{ 
                    true, &lexer.NumericToken{ Int: "100" }, lc( 6 ) },
                &syntax.NumRestrictionSyntax{ 
                    false, &lexer.NumericToken{ Int: "100" }, lc( 11 ) },
                false,
            },
        },
        &restrictionSyntaxTest{
            str: "Num~(,)",
            typ: "Num",
            restriction: 
                &syntax.RangeRestrictionSyntax{ false, nil, nil, false },
        },
        &restrictionSyntaxTest{
            str: "Str~[\"a\",\"aaaa\")",
            typ: "Str",
            restriction: &syntax.RangeRestrictionSyntax{
                true,
                &syntax.StringRestrictionSyntax{ "a", lc( 6 ) },
                &syntax.StringRestrictionSyntax{ "aaaa", lc( 10 ) },
                false,
            },
        },
    }
    for _, test := range tests { 
        test.T = t
        test.call()
    }
}

func TestTypeReferenceStringer( t *testing.T ) {
    for _, quant := range []string{ "", "*", "+", "?", "*++" } {
        expct := "ns1@v1/T1" + quant
        ref := MustTypeReference( expct )
        assert.Equal( expct, ref.String() )
        assert.Equal( expct, ref.ExternalForm() )
    }
}

func TestTypeReferenceMustPanic( t *testing.T ) {
    errExpct := 
        &pt.ParseErrorExpect{ 7, "Expected type path but found: END" }
    pt.AssertParsePanic( errExpct, t, func() { MustTypeReference( "ns1@v1" ) } )
}

func typeEqFunc ( a, b interface{} ) bool {
    return a.( TypeReference ).Equals( b.( TypeReference) )
}

func assertTypeRefBaseEquality( str string, t *testing.T ) TypeReference {
    res := MustTypeReference( str )
    assertEquality( res, nil, res.Equals( nil ), false, t )
    assertEquality( res, res, res.Equals( res ), true, t )
    // Check equality with a different instance
    ref2 := MustTypeReference( res.ExternalForm() )
    assertEquality( res, ref2, res.Equals( ref2 ), true, t )
    return res
}

func TestAtomicTypeReferenceEquality( t *testing.T ) {
    at1 := assertTypeRefBaseEquality( "ns1@v1/T1", t )
    tests := []equalityTest{}
    for _, str := range []string{ 
            "ns1@v1/T2", "ns2@v1/T1", "ns1@v1/T1?", "ns1@v1/T1*",
            "ns1@v1/T1+", "ns1@v1/T1**+", 
    } {
        ref := MustTypeReference( str )
        tests = append( tests, equalityTest{ at1, ref, false } )
    }
    assertEqualityTests( typeEqFunc, t, tests... )
}

func TestListTypeReferenceEquality( t *testing.T ) {
    lt1 := assertTypeRefBaseEquality( "ns1@v1/T1*", t )
    tests := make( []equalityTest, 0, 16 )
    for _, quant := range []string { "*", "+", "**", "*+*", "*?**+" } {
        str := "ns1@v1/T1" + quant
        ref1 := MustTypeReference( str )
        ref2 := MustTypeReference( str )
        tests = append( tests, equalityTest{ ref1, ref2, true } )
    }
    for _, str := range []string {
        "ns1@v1/T1+", "ns1@v1/T1**", "ns1@v1/T1?", "ns1@v1/T1?*",
    } {
        ref := MustTypeReference( str )
        tests = append( tests, equalityTest{ lt1, ref, false } )
    }
    assertEqualityTests( typeEqFunc, t, tests... )
}

func TestNullableTypeReferenceEquality( t *testing.T ) {
    nt1 := assertTypeRefBaseEquality( "ns1@v1/T1?", t )
    tests := []equalityTest{}
    for _, quant := range []string { "?", "??", "*?", "+?", "*?+?" } {
        str := "ns1@v1/T1" + quant
        ref1 := MustTypeReference( str )
        ref2 := MustTypeReference( str )
        tests = append( tests, equalityTest{ ref1, ref2, true } )
    }
    for _, str := range []string {
        "ns1@v1/T1*?", "ns1@v1/T1??", "ns1@v1/T1?*", "ns1@v1/T1+",
    } {
        ref := MustTypeReference( str )
        tests = append( tests, equalityTest{ nt1, ref, false } )
    }
    assertEqualityTests( typeEqFunc, t, tests... )
}

func TestAtomicTypeRestrictionEquality( t *testing.T ) {
    f := func( s1, s2 string ) {
        t1 := MustTypeReference( s1 )
        wantEq := false
        if s2 == "" { s2, wantEq = s1, true }
        t2 := MustTypeReference( s2 )
        assert.Equal( wantEq, t1.Equals( t2 ) )
        assert.Equal( wantEq, t2.Equals( t1 ) )
    }
    f( "Int32~[0,2]", "" )
    f( "Int32~(0,2]", "" )
    f( "Int32~(0,2)", "" )
    f( "Int32~[0,2)", "" )
    f( "Int32~[0,2)", "Int32~[0,3)" )
    f( "Int32~[0,2)", "Int32~(0,2)" )
    f( "Int32~[0,2)", "Int64~[0,2)" )
    f( `String~"a"`, "" )
    f( `String~"a"`, `String~"b"` )
}

func TestIdentifiedNameStringer( t *testing.T ) {
    str := "ns1@v1/n1/n2"
    assert.Equal( str, MustIdentifiedName( str ).String() )
}

func TestIdentifiedNameMustPanic( t *testing.T ) {
    errExpct := &pt.ParseErrorExpect{ 10, "Missing name" }
    pt.AssertParsePanic( errExpct, t, 
        func() { MustIdentifiedName( "someNs@v1" ) } )
}

func TestIdentifiedNameEquality( t *testing.T ) {
    nm1A := MustIdentifiedName( "ns1@v1/id1" )
    assertEquality( nm1A, nil, nm1A.Equals( nil ), false, t )
    assertEqualityTests(
        func( a, b interface{} ) bool {
            return a.( *IdentifiedName ).Equals( b.( *IdentifiedName ) )
        },
        t,
        equalityTest{ nm1A, nm1A, true },
        equalityTest{ nm1A, MustIdentifiedName( "ns1@v1/id1" ), true },
        equalityTest{ nm1A, MustIdentifiedName( "ns2@v1/id1" ), false },
        equalityTest{ nm1A, MustIdentifiedName( "ns1@v2/id1" ), false },
        equalityTest{ nm1A, MustIdentifiedName( "ns1@v1/id2" ), false },
        equalityTest{ nm1A, MustIdentifiedName( "ns1@v1/id1/id2" ), false },
        equalityTest{ nm1A, MustIdentifiedName( "ns1:ns2@v1/id1" ), false },
        equalityTest{
            MustIdentifiedName( "ns1:ns2@v1/id1/id2/id3" ),
            MustIdentifiedName( "ns1:ns2@v1/id1/id2/id3" ),
            true,
        },
    )
}

func TestTimestampStrings( t *testing.T ) {
    tm := time.Now()
    mgTm := Timestamp( tm )
    assert.Equal( tm.Format( time.RFC3339Nano ), mgTm.String() )
    assert.Equal( tm.Format( time.RFC3339Nano ), mgTm.Rfc3339Nano() )
}

func TestTimestampParse( t *testing.T ) {
    f := func( src, expct string ) {
        tm := MustTimestamp( src )
        if expct == "" { expct = src }
        assert.Equal( expct, tm.Rfc3339Nano() )
    }
    f( Now().Rfc3339Nano(), "" )
    f( "2012-01-01T12:00:00Z", "" )
    f( "2012-01-01T12:00:00+07:00", "" )
    f( "2012-01-01T12:00:00-07:00", "" )
    f( "2012-01-01T12:00:00+00:00", "2012-01-01T12:00:00Z" )
}

func TestTimestampParseFail( t *testing.T ) {
    str := "2012-01-10X12:00:00"
    if _, err := ParseTimestamp( str ); err == nil {
        t.Fatalf( "Was able to parse: %q", str )
    } else {
        pe := err.( *loc.ParseError )
        assert.Equal( 
            fmt.Sprintf( "Invalid RFC3339 time: %q", str ), pe.Message )
        assert.Equal( 1, pe.Loc.Line )
        assert.Equal( 1, pe.Loc.Col )
    }
}
