package parser

import (
    "fmt"
    "strings"
    mg "mingle"
)

type NumericTokenTest struct { 
    Negative bool
    Token *NumericToken
}

type CoreParseTestType string

const (
    TestTypeString = CoreParseTestType( "string" )
    TestTypeNumber = CoreParseTestType( "number" )
    TestTypeIdentifier = CoreParseTestType( "identifier" )
    TestTypeNamespace = CoreParseTestType( "namespace" )
    TestTypeDeclaredTypeName = CoreParseTestType( "declared-type-name" )
    TestTypeQualifiedTypeName = CoreParseTestType( "qualified-type-name" )
    TestTypeTypeName = CoreParseTestType( "type-name" )
    TestTypeIdentifiedName = CoreParseTestType( "identified-name" )
    TestTypeTypeReference = CoreParseTestType( "type-reference" )
)

type ErrorExpect interface {}

type RestrictionErrorExpect string

type CoreParseTest struct {
    TestType CoreParseTestType
    In string
    ExternalForm string
    Expect interface{}
    Err ErrorExpect
}

var CoreParseTests = []*CoreParseTest{}

func init() {
    failPe := func( 
        in string, 
        col int, 
        msg string, 
        tt CoreParseTestType ) *CoreParseTest {
        return &CoreParseTest{
            In: in, TestType: tt, Err: &ParseErrorExpect{ col, msg } }
    }
    peFailBinder := func( 
        tt CoreParseTestType ) func( string, int, string ) *CoreParseTest {
        return func( in string, col int, msg string ) *CoreParseTest {
            return failPe( in, col, msg, tt )
        }
    }
    strSucc := func( in, extForm, expct string ) *CoreParseTest {
        return &CoreParseTest{
            In: in,
            ExternalForm: extForm,
            Expect: StringToken( expct ),
            TestType: TestTypeString,
        }
    }
    strFail := peFailBinder( TestTypeString )
    CoreParseTests = append( CoreParseTests,
        strSucc( `""`, `""`, "" ),
        strSucc( `"abc"`, `"abc"`, "abc" ),
        strSucc( `"\n\r\t\f\b\\\"\u01ff"`, `"\n\r\t\f\b\\\"ǿ"`, 
            "\n\r\t\f\b\\\"ǿ" ),
        strSucc( `"abǿcd"`, `"abǿcd"`, "abǿcd" ),
        strSucc( "\"a\U0001d11eb\"", "\"a\U0001d11eb\"", "a\U0001d11eb" ),
        strSucc( `"a\ud834\udd1eb"`, "\"a\U0001d11eb\"", "a\U0001d11eb" ),
        // independent of other escapes, just test that escapes are properly
        // consumed wherever in the input they are anchored (start/middle/end)
        strSucc( `"\u0061bc"`, `"abc"`, "abc" ),
        strSucc( `"a\u0062c"`, `"abc"`, "abc" ), 
        strSucc( `"ab\u0063"`, `"abc"`, "abc" ), 
        strFail( `"a`, 2, "Unterminated string literal" ),
        strFail( `"\"`, 3, "Unterminated string literal" ),
        strFail( `"\k"`, 3, `Unrecognized escape: \k (U+006B)` ),
        strFail( `"\u012k"`, 7, `Invalid hex char in escape: "k" (U+006B)` ),
        strFail( `"\u01k2"`, 6, `Invalid hex char in escape: "k" (U+006B)` ),
        strFail( `"\u012"`, 7, `Invalid hex char in escape: "\"" (U+0022)` ),
        strFail( `"\u01"`, 6, `Invalid hex char in escape: "\"" (U+0022)` ),
        strFail( `"\u0"`, 5, `Invalid hex char in escape: "\"" (U+0022)` ),
        strFail( `"\u"`, 4, `Invalid hex char in escape: "\"" (U+0022)` ),
        strFail( `"\U001f"`, 3, `Unrecognized escape: \U (U+0055)` ),
        strFail( "\"abc\u0001def\"", 5, 
            "Invalid control character in string literal: \"\\x01\" (U+0001)" ),
        strFail( "\"abc\rdef\"", 5,
            `Invalid control character in string literal: "\r" (U+000D)` ),
        strFail( "\"abc\ndef\"", 5,
            `Invalid control character in string literal: "\n" (U+000A)` ),
        strFail( "\"abc\tdef\"", 5,
            `Invalid control character in string literal: "\t" (U+0009)` ),
        strFail( "\"abc\fdef\"", 5,
            `Invalid control character in string literal: "\f" (U+000C)` ),
        strFail( "\"abc\bdef\"", 5,
            `Invalid control character in string literal: "\b" (U+0008)` ),
        strFail( `"a\ud834|\udd1e"`, 9,
            "Expected trailing surrogate, found: \"|\" (U+007C)" ),
        strFail( `"a\ud834\t\udd1e"`, 9, 
            "Expected trailing surrogate, found: \\t" ),
        strFail( `"a\ud834\u0061"`, 3, 
            "Invalid surrogate pair \\uD834\\u0061" ),
        strFail( `"a\udd1e\ud834"`, 3, 
            "Invalid surrogate pair \\uDD1E\\uD834" ),
    )
    numTokTest := func( neg bool, i, f, e string, c rune ) *NumericTokenTest {
        return &NumericTokenTest{
            Negative: neg,
            Token: &NumericToken{ Int: i, Frac: f, Exp: e, ExpRune: c },
        }
    }
    numSucc := func( in string, num *NumericTokenTest ) *CoreParseTest {
        return &CoreParseTest{ In: in, Expect: num, TestType: TestTypeNumber }
    }
    numFail := peFailBinder( TestTypeNumber )
    CoreParseTests = append( CoreParseTests,
        numSucc( "1", numTokTest( false, "1", "", "",  0 ) ),
        numSucc( "-1", numTokTest( true, "1", "", "",  0 ) ),
        numSucc( "1.1", numTokTest( false, "1", "1", "",  0 ) ),
        numSucc( "-1.1", numTokTest( true, "1", "1", "",  0 ) ),
        numSucc( "1.1e0", numTokTest( false, "1", "1", "0", 'e' ) ),
        numSucc( "1.1E3", numTokTest( false, "1", "1", "3", 'E' ) ),
        numSucc( "1.1e+1", numTokTest( false, "1", "1", "1", 'e' ) ),
        numSucc( "11e-1", numTokTest( false, "11", "", "-1", 'e' ) ),
        numSucc( "000000e0", numTokTest( false, "000000", "", "0", 'e' ) ),
        numSucc( "00001.100000", 
            numTokTest( false, "00001", "100000", "",  0 ) ),
        numFail( "0.", 2, "Number has empty or invalid fractional part" ),
        numFail( "0.x3", 3, 
            `Unexpected char in fractional part: "x" (U+0078)` ),
        numFail( "1.3f3", 4, 
            `Unexpected char in fractional part: "f" (U+0066)` ),
        numFail( "1.2.3", 4, 
            "Expected exponent start or num end, found: \".\" (U+002E)" ),
        numFail( "1L", 2, 
            `Unexpected char in integer part: "L" (U+004C)` ),
        numFail( "1eB", 3, `Unexpected char in exponent: "B" (U+0042)` ),
        numFail( "1e", 2, "Number has empty or invalid exponent" ),
    )
    id := func( parts ...string ) *mg.Identifier { 
        return mg.NewIdentifierUnsafe( parts )
    }
    idSucc := func( in, extForm string, parts ...string ) *CoreParseTest {
        return &CoreParseTest{
            In: in, 
            ExternalForm: extForm,
            Expect: id( parts... ), 
            TestType: TestTypeIdentifier,
        }
    }
    idFail := peFailBinder( TestTypeIdentifier )
    CoreParseTests = append( CoreParseTests,
        idSucc( "ident", "ident", "ident" ),
        idSucc( "test1", "test1", "test1" ),
        idSucc( "test_stuff", "test-stuff", "test", "stuff" ),
        idSucc( "test-stuff", "test-stuff", "test", "stuff" ),
        idSucc( "testStuff", "test-stuff", "test", "stuff" ),
        idSucc( "test-one-two", "test-one-two", "test", "one", "two" ),
        idSucc( "test_one_two", "test-one-two", "test", "one", "two" ),
        idSucc( "testOneTwo", "test-one-two", "test", "one", "two" ),
        idSucc( "test2-stuff2", "test2-stuff2", "test2", "stuff2" ),
        idSucc( "test2_stuff2", "test2-stuff2", "test2", "stuff2" ),
        idSucc( "test2Stuff2", "test2-stuff2", "test2", "stuff2" ),
        idSucc( "multiADJAcentCaps", "multi-a-d-j-acent-caps", 
            "multi", "a", "d", "j", "acent", "caps" ),
        idFail( "2bad", 1, "Illegal start of identifier part: \"2\" (U+0032)" ),
        idFail( "2", 1, "Illegal start of identifier part: \"2\" (U+0032)" ),
        idFail( "bad-2", 5, 
            "Illegal start of identifier part: \"2\" (U+0032)" ),
        idFail( "bad-2bad", 5, 
            "Illegal start of identifier part: \"2\" (U+0032)" ),
        idFail( "AcapCannotStart", 1, 
            "Illegal start of identifier part: \"A\" (U+0041)" ),
        idFail( "-leading-dash", 1, 
            "Illegal start of identifier part: \"-\" (U+002D)" ),
        idFail( "_leading_underscore", 1, 
            "Illegal start of identifier part: \"_\" (U+005F)" ),
        idFail( "a-bad-ch@r", 9, "Unexpected token: @" ),
        idFail( "bad-@", 5, 
            "Illegal start of identifier part: \"@\" (U+0040)" ),
        idFail( "bad-A", 5, 
            "Illegal start of identifier part: \"A\" (U+0041)" ),
        idFail( "giving-mixedMessages", 13, 
            "Invalid id rune: \"M\" (U+004D)" ),
        idFail( "too--many-dashes", 5, 
            "Illegal start of identifier part: \"-\" (U+002D)" ),
        idFail( "too__many_underscores", 5, 
            "Illegal start of identifier part: \"_\" (U+005F)" ),
        idFail( "trailing-dash-", 15, "Empty identifier part" ),
        idFail( "trailing_underscore_", 21, "Empty identifier part" ),
        idFail( "trailing-input/x", 15, "Unexpected token: /" ),
        idFail( "", 1, "Empty identifier" ),
    )
    ns := func( ver *mg.Identifier, parts ...*mg.Identifier ) *mg.Namespace {
        return &mg.Namespace{ parts, ver }
    }
    nsSucc := func( in, extForm string, expct *mg.Namespace ) *CoreParseTest {
        return &CoreParseTest{ 
            In: in, 
            ExternalForm: extForm,
            Expect: expct, 
            TestType: TestTypeNamespace,
        }
    }
    nsFail := peFailBinder( TestTypeNamespace )
    idV1 := id( "v1" )
    ns1V1T1 := ns( idV1, id( "ns1" ) )
    CoreParseTests = append( CoreParseTests,
        nsSucc( "ns@v1", "ns@v1", ns( idV1, id( "ns" ) ) ),
        nsSucc( "ns1:ns2:ns3@v1", 
            "ns1:ns2:ns3@v1",
            ns( idV1, id( "ns1" ), id( "ns2" ), id( "ns3" ) ) ),
        nsSucc( "nsIdent1:nsIdent2:ns3@v1", 
            "nsIdent1:nsIdent2:ns3@v1",
            ns( idV1, 
                id( "ns", "ident1" ), id( "ns", "ident2" ), id( "ns3" ) ) ),
        nsSucc( "ns-ident1:ns-ident2:ns3@v1", 
            "nsIdent1:nsIdent2:ns3@v1",
            ns( idV1, 
                id( "ns", "ident1" ), id( "ns", "ident2" ), id( "ns3" ) ) ),
        nsSucc( "ns_ident1:ns_ident2:ns3@v1", 
            "nsIdent1:nsIdent2:ns3@v1",
            ns( idV1, 
                id( "ns", "ident1" ), id( "ns", "ident2" ), id( "ns3" ) ) ),
        nsFail( "2bad:ns@v1", 1, 
            `Illegal start of identifier part: "2" (U+0032)` ),
        nsFail( "ns:2bad@v1", 4, 
            `Illegal start of identifier part: "2" (U+0032)` ), 
        nsFail( "ns1:ns2", 8, `Expected ':' or '@' but found: END` ), 
        // Arguably, a better error would be something like 'empty identifier',
        // but for our current algorithm the fact that there is no '@' will be
        // detected earlier
        nsFail( "ns1:ns2:", 9, "Empty identifier" ), 
        nsFail( "ns1:ns2:@v1", 9, 
            "Illegal start of identifier part: \"@\" (U+0040)" ), 
        nsFail( "ns1.ns2@v1", 4, "Expected ':' or '@' but found: ." ), 
        nsFail( "ns1 : ns2:ns3@v1", 4, `Expected ':' or '@' but found: " "` ), 
        nsFail( "ns1:ns2@v1/Stuff", 11, "Unexpected token: /" ), 
        nsFail( "@v1", 1, "Illegal start of identifier part: \"@\" (U+0040)" ), 
        nsFail( "ns1@V2", 5, `Illegal start of identifier part: "V" (U+0056)` ),
        nsFail( "ns1:ns2@v1:ns3", 11, "Unexpected token: :" ),
        nsFail( "ns1:ns2@v1@v2", 11, "Unexpected token: @" ),
        nsFail( "ns1@", 5, "Empty identifier" ), 
        nsFail( "ns1@ v1", 5, 
            "Illegal start of identifier part: \" \" (U+0020)" ), 
    )
    declNm := func( nm string ) *mg.DeclaredTypeName {
        return mg.NewDeclaredTypeNameUnsafe( nm )
    }
    declNmSucc := func( nm string ) *CoreParseTest {
        return &CoreParseTest{ 
            In: nm, 
            ExternalForm: nm,
            Expect: declNm( nm ),
            TestType: TestTypeDeclaredTypeName,
        }
    }
    declNmFail := peFailBinder( TestTypeDeclaredTypeName )
    CoreParseTests = append( CoreParseTests,
        declNmSucc( "T" ),
        declNmSucc( "T1" ),
        declNmSucc( "T1T2" ),
        declNmSucc( "BlahBlah3Blah" ),
        declNmSucc( "TUVWX" ),
        declNmFail( "a", 1, "Illegal type name start: \"a\" (U+0061)" ),
        declNmFail( "aBadName", 1, "Illegal type name start: \"a\" (U+0061)" ),
        declNmFail( "2", 1, "Illegal type name start: \"2\" (U+0032)" ),
        declNmFail( "2Bad", 1, "Illegal type name start: \"2\" (U+0032)" ),
        declNmFail( "Bad$Char", 4, "Illegal type name rune: \"$\" (U+0024)" ),
        declNmFail( "Bad_Char", 4, "Illegal type name rune: \"_\" (U+005F)" ),
        declNmFail( "Bad-Char", 4, "Unexpected token: -" ),
        declNmFail( "", 1, "Empty type name" ),
    )
    qn := func( ns *mg.Namespace,
                nm *mg.DeclaredTypeName ) *mg.QualifiedTypeName {
        return &mg.QualifiedTypeName{ ns, nm }
    }
    qnSucc := func( 
        in string, ns *mg.Namespace, nm *mg.DeclaredTypeName ) *CoreParseTest {
        return &CoreParseTest{
            In: in, 
            ExternalForm: in,
            Expect: qn( ns, nm ), 
            TestType: TestTypeQualifiedTypeName,
        }
    }
    qnFail := peFailBinder( TestTypeQualifiedTypeName )
    qnNs1V1T1 := qn( ns1V1T1, declNm( "T1" ) )
    CoreParseTests = append( CoreParseTests,
        qnSucc( "ns1@v1/T1", ns( idV1, id( "ns1" ) ), declNm( "T1" ) ),
        qnSucc( "ns1:ns2@v1/T1",
            ns( idV1, id( "ns1" ), id( "ns2" ) ), declNm( "T1" ),
        ),
        qnFail( "ns1@v1", 7, "Expected type path but found: END" ),
        qnFail( "ns1/T1", 4, "Expected ':' or '@' but found: /" ),
        qnFail( "ns1@v1/2Bad", 8, "Illegal type name start: \"2\" (U+0032)" ),
        qnFail( "ns1@v1/", 8, "Empty type name" ),
        qnFail( "ns1@v1/T1/", 10, "Unexpected token: /" ),
        qnFail( "ns1@v1//T1", 8, "Illegal type name start: \"/\" (U+002F)" ),
        qnFail( "ns1@v1/T1#T2", 10, "Illegal type name rune: \"#\" (U+0023)" ),
    )
    idNmSucc := func( 
        in, extForm string, 
        ns *mg.Namespace, 
        ids ...*mg.Identifier ) *CoreParseTest {

        return &CoreParseTest{
            In: in,
            ExternalForm: extForm,
            Expect: &mg.IdentifiedName{ ns, ids },
            TestType: TestTypeIdentifiedName,
        }
    }
    idNmFail := peFailBinder( TestTypeIdentifiedName )
    CoreParseTests = append( CoreParseTests,
        idNmSucc( "some:ns@v1/someId1",
            "some:ns@v1/some-id1",
            ns( idV1, id( "some" ), id( "ns" ) ), id( "some", "id1" ) ),
        idNmSucc( "some:ns@v1/someId1/someId2", 
            "some:ns@v1/some-id1/some-id2", 
            ns( idV1, id( "some" ), id( "ns" ) ), 
            id( "some", "id1" ), id( "some", "id2" ) ),
        idNmSucc( "some:ns@v1/someId1/some-id2/some_id3",
            "some:ns@v1/some-id1/some-id2/some-id3",
            ns( idV1, id( "some" ), id( "ns" ) ),
            id( "some", "id1" ), id( "some", "id2" ), id( "some", "id3" ) ),
        idNmSucc( "singleNs@v1/singleIdent", 
            "single-ns@v1/single-ident", 
            ns( idV1, id( "single", "ns" ) ), id( "single", "ident" ) ),
        idNmFail( "someNs@v1", 10, "Missing name" ),
        idNmFail( "some:ns@v1", 11, "Missing name" ),
        idNmFail( "some:ns@v1/", 12, "Empty identifier" ),
        idNmFail( "some:ns@v3/trailingSlash/", 26, "Empty identifier" ),
        idNmFail( "some:ns@v1/SomeId", 12, 
            `Illegal start of identifier part: "S" (U+0053)` ),
        idNmFail( "", 1, "Empty identifier" ),
        idNmFail( "/some:ns@v1/noGood/leadingSlash", 1, 
            "Illegal start of identifier part: \"/\" (U+002F)" ),
    )
    // We can't reference the predefined QnameString, QnameFloat32, etc, since
    // they are also defined in an init scope and we don't want to assume that
    // they are initialized before this block runs
    primQn := func( nm string ) *mg.QualifiedTypeName {
        return qn( ns( idV1, id( "mingle" ), id( "core" ) ), declNm( nm ) )
    }
    typRefSucc := func( in, ext string, expect interface{} ) *CoreParseTest {
        if ext == "" { ext = in }
        return &CoreParseTest{ 
            In: in, 
            ExternalForm: ext,
            Expect: expect, 
            TestType: TestTypeTypeReference,
        }
    }
    typRefFail := peFailBinder( TestTypeTypeReference )
    qnMgStr := primQn( "String" )
    atMgStr := &mg.AtomicTypeReference{ Name: qnMgStr }
    at2 := &mg.AtomicTypeReference{
        Name: qn( ns( idV1, id( "ns1" ), id( "ns2" ) ), declNm( "T1" ) ),
    }
    tm1 := mg.MustTimestamp( "2012-01-01T12:00:00Z" )
    tm2 := mg.MustTimestamp( "2012-01-02T12:00:00Z" )
    atTm1 := &mg.AtomicTypeReference{
        primQn( "Timestamp" ),
        &mg.RangeRestriction{ true, tm1, tm2, true },
    }
    rx := func( s string ) *mg.RegexRestriction { 
        res, err := mg.NewRegexRestriction( s )
        if err != nil { panic( err ) }
        return res
    }
    typRefRestrictFail := func( in, msg string ) *CoreParseTest {
        return &CoreParseTest{ 
            In: in, 
            Err: RestrictionErrorExpect( msg ),
            TestType: TestTypeTypeReference,
        }
    }
    atNs1V1T1 := &mg.AtomicTypeReference{ Name: qnNs1V1T1 }
    CoreParseTests = append( CoreParseTests,
        typRefSucc( "mingle:core@v1/String", "", atMgStr ),
        typRefSucc( "mingle:core@v1/String*", "", 
            &mg.ListTypeReference{ atMgStr, true } ),
        typRefSucc( "ns1:ns2@v1/T1", "", at2 ),
        typRefSucc( "ns1:ns2@v1/T1*+?", "", 
            mg.MustNullableTypeReference(
                &mg.ListTypeReference{ 
                    &mg.ListTypeReference{ at2, true }, false },
            ),
        ),
        typRefSucc( `mingle:core@v1/String~"^a+$"`, 
            "",
            &mg.AtomicTypeReference{ qnMgStr, rx( "^a+$" ) } ),
        typRefSucc( 
            `mingle:core@v1/String~"a\t\f\b\r\n\"\\\u0061\ud834\udd1e"`,
            "mingle:core@v1/String~\"a\\t\\f\\b\\r\\n\\\"\\\\a𝄞\"",
            &mg.AtomicTypeReference{
                qnMgStr, rx( "a\t\f\b\r\n\"\\a\U0001d11e" ) },
        ),
        typRefSucc( `mingle:core@v1/String~"^a+$"*+`,
            "",
            &mg.ListTypeReference{
                &mg.ListTypeReference{
                    &mg.AtomicTypeReference{ qnMgStr, rx( "^a+$" ) },
                    true,
                },
                false,
            },
        ),
        typRefSucc( `&ns1@v1/T1`, "", mg.NewPointerTypeReference( atNs1V1T1 ) ),
        typRefSucc( `&&ns1@v1/T1`, "",
            mg.NewPointerTypeReference( 
                mg.NewPointerTypeReference( atNs1V1T1 ) ) ),
        typRefSucc( `&String`, "&mingle:core@v1/String",
            mg.NewPointerTypeReference( 
                &mg.AtomicTypeReference{ Name: qnMgStr } ) ),
        typRefSucc( `&&String~".*"`, `&&mingle:core@v1/String~".*"`,
            mg.NewPointerTypeReference(
                mg.NewPointerTypeReference(
                    &mg.AtomicTypeReference{ 
                        Name: qnMgStr,
                        Restriction: rx( ".*" ),
                    },
                ),
            ),
        ),
        typRefSucc( `&ns1@v1/T1*+`, "",
            &mg.ListTypeReference{
                &mg.ListTypeReference{ 
                    mg.NewPointerTypeReference( atNs1V1T1 ), true },
                false,
            },
        ),
        typRefSucc( `&ns1@v1/T1?*+`, "",
            &mg.ListTypeReference{
                &mg.ListTypeReference{
                    mg.MustNullableTypeReference( 
                        mg.NewPointerTypeReference( atNs1V1T1 ), 
                    ),
                    true,
                },
                false,
            },
        ),
        typRefSucc( `&ns1@v1/T1?*?+`, "",
            &mg.ListTypeReference{
                mg.MustNullableTypeReference(
                    &mg.ListTypeReference{
                        mg.MustNullableTypeReference( 
                            mg.NewPointerTypeReference( atNs1V1T1 ), 
                        ),
                        true,
                    },
                ),
                false,
            },
        ),
        typRefSucc( `&ns1@v1/T1?`, "",
            mg.MustNullableTypeReference( 
                mg.NewPointerTypeReference( atNs1V1T1 ) ) ),
    )
    // Now just basic coverage of core type resolution and restrictions.
    addCoreTypRefSucc := func( inPref, inBase string, expct interface{} ) {
        fqForm := inPref + "mingle:core@v1/" + inBase
        CoreParseTests = append( CoreParseTests,
            typRefSucc( fqForm, fqForm, expct ),
            typRefSucc( inPref + inBase, fqForm, expct ),
        )
    }
    addCoreTypRefSuccWithPtr := func( inBase string, expct mg.TypeReference ) {
        addCoreTypRefSucc( "", inBase, expct )
        addCoreTypRefSucc( "&", inBase, mg.NewPointerTypeReference( expct ) )
    }
    primNames := []string{
        "Boolean", "String", "Int32", "Uint32", "Int64", "Uint64", "Float32",
        "Float64", "Buffer", "Timestamp", "Null",
    }
    func() {
        for _, s := range primNames {
            primTyp := &mg.AtomicTypeReference{ primQn( s ), nil }
            addCoreTypRefSuccWithPtr( s, primTyp )
            in := s + "?"
            if ( mg.IsNullableType( primTyp ) ) {
                nullTyp := mg.MustNullableTypeReference( primTyp )
                typRefSucc( in, "mingle:core@v1" + in, nullTyp )
            } else {
                typRefFail( in, 1, "not a nullable type" )
            }
        }
    }()
    addCoreTypRefSuccWithPtr( `String~"^a+$"`, 
        &mg.AtomicTypeReference{ qnMgStr, rx( "^a+$" ) },
    )
    addCoreTypRefSuccWithPtr( `String~["aaa","bbb"]`,
        &mg.AtomicTypeReference{ 
            qnMgStr, 
            &mg.RangeRestriction{ 
                true, 
                mg.String( "aaa" ), 
                mg.String( "bbb" ), 
                true,
            },
        },
    )
    // We simultaneously permute primitive num types and interval combinations
    // with the next 4
    addCoreTypRefSuccWithPtr( `Int32~(0,1]`,
        &mg.AtomicTypeReference{ 
            primQn( "Int32" ),
            &mg.RangeRestriction{ 
                false, 
                mg.Int32( int32( 0 ) ), 
                mg.Int32( int32( 1 ) ), 
                true,
            },
        },
    )
    addCoreTypRefSuccWithPtr( `Uint32~[0,1]`,
        &mg.AtomicTypeReference{ 
            primQn( "Uint32" ),
            &mg.RangeRestriction{ 
                true, 
                mg.Uint32( uint32( 0 ) ), 
                mg.Uint32( uint32( 1 ) ),
                true,
            },
        },
    )
    addCoreTypRefSuccWithPtr( `Int64~[0,1)`,
        &mg.AtomicTypeReference{ 
            primQn( "Int64" ),
            &mg.RangeRestriction{ 
                true, 
                mg.Int64( int64( 0 ) ), 
                mg.Int64( int64( 1 ) ), 
                false,
            },
        },
    )
    addCoreTypRefSuccWithPtr( `Uint64~(0,2)`,
        &mg.AtomicTypeReference{ 
            primQn( "Uint64" ),
            &mg.RangeRestriction{ 
                false, 
                mg.Uint64( uint64( 0 ) ), 
                mg.Uint64( uint64( 2 ) ), 
                false,
            },
        },
    )
    // For floating points we use numbers that can convert without loss of
    // precision between machine and string form, to simplify testing of
    // external forms.
    addCoreTypRefSuccWithPtr( `Float32~[1,2)`,
        &mg.AtomicTypeReference{
            primQn( "Float32" ),
            &mg.RangeRestriction{ 
                true, 
                mg.Float32( float32( 1.0 ) ), 
                mg.Float32( float32( 2.0 ) ), 
                false,
            },
        },
    )
    addCoreTypRefSuccWithPtr( `Float64~[0.1,2.1)`,
        &mg.AtomicTypeReference{
            primQn( "Float64" ),
            &mg.RangeRestriction{ 
                true, 
                mg.Float64( float64( 0.1 ) ), 
                mg.Float64( float64( 2.1 ) ), 
                false,
            },
        },
    )
    addCoreTypRefSuccWithPtr( 
        fmt.Sprintf( "Timestamp~[%q,%q]", tm1, tm2 ), atTm1 )
    CoreParseTests = append( CoreParseTests,
        typRefFail( "/T1", 1, 
            "Expected identifier or declared type name but found: /" ),
        // don't need to exhaustively retest all name parse errors, but we do
        // want to make sure they happen for qn and declNm types and are
        // reported in the correct location
        typRefFail( "ns1", 4, "Expected ':' or '@' but found: END" ),
        typRefFail( "ns1@", 5, "Empty identifier" ),
        typRefFail( "ns1:@v1", 5, 
            "Illegal start of identifier part: \"@\" (U+0040)" ),
        typRefFail( "ns1@v1", 7, "Expected type path but found: END" ),
        typRefFail( "ns1@v1/bad", 8, 
            "Illegal type name start: \"b\" (U+0062)" ),
        typRefFail( "ns1@v1/T1*?-+", 12, "Unexpected token: -" ),
        typRefFail( "ns1@v1/T1*? +", 12, `Unexpected token: " "` ),
        typRefFail( "&", 2, "Unexpected end of input" ),
        typRefFail( "&&&&", 5, "Unexpected end of input" ),
        typRefFail( "&ns1", 5, "Expected ':' or '@' but found: END" ),
        typRefFail( "&&+", 3, 
            "Expected identifier or declared type name but found: +" ),
        typRefFail( "mingle:core@v1/String~", 23, "Unexpected end of input" ),
        typRefFail( "mingle:core@v1/~\"s*\"", 16, 
            "Illegal type name start: \"~\" (U+007E)" ),
        typRefFail( "mingle:core@v1~\"s*\"", 15, 
            "Expected type path but found: ~" ),
        typRefFail( "mingle:core@v1/String~=\"sdf\"", 23, 
            `Unexpected char: "=" (U+003D)` ),
        typRefFail( "Int32~(1:2)", 9, "Expected , but found: :" ),
        typRefFail( "Int32~[1,3}", 11, 
            `Expected one of [ ")", "]" ] but found: }` ),
        typRefFail( "Int32~[abc,2)", 8, `Expected range value but found: abc` ),
        typRefFail( `Int32~[-"abc",2)`, 9, "Expected number but found: abc" ),
        typRefFail( "Int32~[--3,4)", 9, "Expected range value but found: -" ),
        typRefFail( "Int32~[,]", 7, "Infinite range must be open" ),
        typRefFail( "Int32~[8,]", 10, "Infinite high range must be open" ),
        typRefFail( "Int32~[,8]", 7, "Infinite low range must be open" ),
        typRefFail( "Stuff", 1, `cannot resolve as a standard type: Stuff` ),
        typRefFail( `Stuff~(,)`, 1,
            `cannot resolve as a standard type: Stuff` ),
        typRefFail( "&ns1@v1/T1???", 12, 
            "a nullable type cannot itself be made nullable" ),
        typRefFail( "ns1@v1/T1?", 1, "not a nullable type" ),
        typRefFail( "S1~12.1", 4, "Expected type restriction but found: 12.1" ),
        typRefRestrictFail( `ns1@v1/T~"a"`, 
            "Invalid target type for regex restriction: ns1@v1/T" ),
        typRefRestrictFail( `ns1@v1/T~[0,1]`, 
            "Invalid target type for range restriction: ns1@v1/T" ),
        typRefRestrictFail( `String~[0,"1")`, 
            "Got number as min value for range" ),
        typRefRestrictFail( `String~["0",1)`, 
            "Got number as max value for range" ),
        typRefRestrictFail( `Timestamp~(,1)`,
            "Got number as max value for range" ),
        typRefRestrictFail( `Int32~["a",2)`,
            "Got string as min value for range" ),
        typRefRestrictFail( `Int32~(1,"20")`,
            "Got string as max value for range" ),
        typRefRestrictFail( `Int32~"a"`, 
            "Invalid target type for regex restriction: " + 
            "mingle:core@v1/Int32" ),
        typRefRestrictFail( `mingle:core@v1/Buffer~[0,1]`, 
            "Invalid target type for range restriction: " + 
            "mingle:core@v1/Buffer" ),
        typRefRestrictFail( fmt.Sprintf( "Timestamp~[%q,%q]", tm2, tm1 ), 
            "Unsatisfiable range" ),
        typRefRestrictFail( `Timestamp~["2001-0x-22",)`,
            "Invalid min value in range restriction: val: Invalid timestamp: " +
            "[<input>, line 1, col 1]: Invalid RFC3339 time: \"2001-0x-22\"" ),
        typRefRestrictFail( `String~"ab[a-z"`, 
            "error parsing regexp: missing closing ]: `[a-z`" ),
    )
    // Now add various base range coverage tests for all types; 
    // we range below over arrays which are similar to rangeValTypes defined in
    // model_parser.go and NumericTypes defined in model.go since we can't
    // assume that the originals have yet been initialized
    rangeTypeNames := []string { 
        "String", "Int32", "Int64", "Uint32", "Uint64", "Float32", "Float64", 
        "Timestamp", 
    }
    for _, str := range rangeTypeNames {
        CoreParseTests = append( CoreParseTests,
            typRefSucc( fmt.Sprintf( "mingle:core@v1/%s~(,)", str ),
                "",
                &mg.AtomicTypeReference{ 
                    primQn( str ), &mg.RangeRestriction{} },
            ),
        )
    }
    for _, iType := range []string { "Int32", "Int64", "Uint32", "Uint64" } {
        for _, s := range []string { 
            "[0,-1]", "(0,0)", "(0,0]", "[0,0)", "(0,1)" } {
            if iType[ 0 ] == 'I' || strings.Index( s, "-" ) < 0 {
                CoreParseTests = 
                    append( 
                        CoreParseTests, 
                        typRefRestrictFail( 
                            iType + "~" + s, 
                            "Unsatisfiable range",
                        ),
                    )
            }
        }
        CoreParseTests = append( CoreParseTests,
            typRefRestrictFail(
                iType + `~[1.0,2]`, "Got decimal as min value for range" ),
            typRefRestrictFail(
                iType + `~[1,2.0]`, "Got decimal as max value for range" ),
        )
    }
    for _, dType := range []string { "Float32", "Float64" } {
        for _, s := range []string { "(1.0,1.0)", "(0.0,-1.0)" } {
            CoreParseTests = append( CoreParseTests,
                typRefRestrictFail( dType + "~" + s, "Unsatisfiable range" ) )
        }
    }
    for _, typ := range []string { 
        "Int32", "Int64", "Uint32", "Uint64", "Float32", "Float64",
    } {
        base := "mingle:core@v1/" + typ
        CoreParseTests = append( CoreParseTests,
            typRefRestrictFail(
                base + `~("0",12]`, "Got string as min value for range" ),
            typRefRestrictFail(
                base + `~(0,"12"]`, "Got string as max value for range" ),
        )
    }
}
