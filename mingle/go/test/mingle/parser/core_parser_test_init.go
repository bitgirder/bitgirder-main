package parser

import (
    mg "mingle"
)

type CoreParseTestType string

const (
    TestTypeString = CoreParseTestType( "string" )
    TestTypeNumber = CoreParseTestType( "number" )
    TestTypeIdentifier = CoreParseTestType( "identifier" )
    TestTypeNamespace = CoreParseTestType( "namespace" )
    TestTypeDeclaredTypeName = CoreParseTestType( "declared-type-name" )
    TestTypeQualifiedTypeName = CoreParseTestType( "qualified-type-name" )
    TestTypeTypeName = CoreParseTestType( "type-name" )
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

func strToQuants( s string ) quantList {
    res := make( []SpecialToken, len( s ) )
    for i, c := range s {
        switch c {
        case '*': res[ i ] = SpecialTokenAsterisk
        case '+': res[ i ] = SpecialTokenPlus
        case '?': res[ i ] = SpecialTokenQuestionMark
        default: panic( libErrorf( "unhandled quant: %s", c ) )
        }
    }
    return res
}

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
    numSucc := func( in string, num *NumericToken ) *CoreParseTest {
        return &CoreParseTest{ In: in, Expect: num, TestType: TestTypeNumber }
    }
    numFail := peFailBinder( TestTypeNumber )
    CoreParseTests = append( CoreParseTests,
        numSucc( "1", &NumericToken{ "1", "", "", 0 } ),
        numSucc( "1.1", &NumericToken{ "1", "1", "",  0 } ),
        numSucc( "1.1e0", &NumericToken{ "1", "1", "0", 'e' } ),
        numSucc( "1.1E3", &NumericToken{ "1", "1", "3", 'E' } ),
        numSucc( "1.1e+1", &NumericToken{ "1", "1", "1", 'e' } ),
        numSucc( "11e-1", &NumericToken{ "11", "", "-1", 'e' } ),
        numSucc( "000000e0", &NumericToken{ "000000", "", "0", 'e' } ),
        numSucc( "00001.100000", &NumericToken{ "00001", "100000", "",  0 } ),
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
    loc := func( col int ) *Location {
        return &Location{ Source: ParseSourceInput, Line: 1, Col: col }
    }
    typRefSucc := func( 
        in, ext string, expect *CompletableTypeReference ) *CoreParseTest {

        if expect.quants == nil { expect.quants = []SpecialToken{} }
        if expect.ErrLoc == nil { expect.ErrLoc = loc( 1 ) }
        if ext == "" { ext = in }
        return &CoreParseTest{ 
            In: in, 
            ExternalForm: ext,
            Expect: expect, 
            TestType: TestTypeTypeReference,
        }
    }
    typRefFail := peFailBinder( TestTypeTypeReference )
    rx := func( s string, col int ) *RegexRestrictionSyntax { 
        return &RegexRestrictionSyntax{ Pat: s, Loc: loc( col ) }
    }
    CoreParseTests = append( CoreParseTests,
        typRefSucc( "ns1@v1/T1", "", 
            &CompletableTypeReference{ Name: qnNs1V1T1 },
        ),
        typRefSucc( "T1", "", 
            &CompletableTypeReference{ Name: qnNs1V1T1.Name },
        ),
        typRefSucc( "ns1@v1/T1*", "", 
            &CompletableTypeReference{
                Name: qnNs1V1T1,
                quants: strToQuants( "*" ),
            },
        ),
        typRefSucc( "ns1@v1/T1*+?", "", 
            &CompletableTypeReference{
                Name: qnNs1V1T1,
                quants: strToQuants( "*+?" ),
           },
        ),
        typRefSucc( `ns1@v1/T1~"^a+$"`, "",
            &CompletableTypeReference{
                Name: qnNs1V1T1,
                Restriction: rx( "^a+$", 11 ),
            },
        ),
        typRefSucc(
            `ns1@v1/T1~"a\t\f\b\r\n\"\\\u0061\ud834\udd1e"`,
            "ns1@v1/T1~\"a\\t\\f\\b\\r\\n\\\"\\\\a𝄞\"",
            &CompletableTypeReference{
                Name: qnNs1V1T1,
                Restriction: rx( "a\t\f\b\r\n\"\\a\U0001d11e", 11 ),
            },
        ),
        typRefSucc( `ns1@v1/T1~"^a+$"*+`, "",
            &CompletableTypeReference{
                Name: qnNs1V1T1,
                Restriction: rx( "^a+$", 11 ),
                quants: strToQuants( "*+" ),
            },
        ),
        typRefSucc( "&ns1@v1/T1", "", 
            &CompletableTypeReference{ Name: qnNs1V1T1, ptrDepth: 1 } ),
        typRefSucc( "&&ns1@v1/T1", "",
            &CompletableTypeReference{ Name: qnNs1V1T1, ptrDepth: 2 } ),
        typRefSucc( "&T1", "",
            &CompletableTypeReference{ Name: qnNs1V1T1.Name, ptrDepth: 1 } ),
        typRefSucc( `&&ns1@v1/T1~".*"`, "",
            &CompletableTypeReference{
                Name: qnNs1V1T1,
                Restriction: rx( ".*", 13 ),
                ptrDepth: 2,
            },
        ),
        typRefSucc( "&ns1@v1/T1*+", "", 
            &CompletableTypeReference{ 
                Name: qnNs1V1T1, 
                quants: strToQuants( "*+" ),
                ptrDepth: 1,
            },
        ),
    )
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
        typRefFail( "ns1@v1/T1~", 11, "Unexpected end of input" ),
        typRefFail( `ns1@v1/~"s*"`, 8, 
            `Illegal type name start: "~" (U+007E)` ),
        typRefFail( `ns1@v1~"s*"`, 7, "Expected type path but found: ~" ),
        typRefFail( `ns1@v1/T1~="sdf"`, 11, `Unexpected char: "=" (U+003D)` ),
        typRefFail( "T1~(1:2)", 6, "Expected , but found: :" ),
        typRefFail( "T1~[1,3}", 8, 
            `Expected one of [ ")", "]" ] but found: }` ),
        typRefFail( "T1~[abc,2)", 5, `Expected range value but found: abc` ),
        typRefFail( `T1~[-"abc",2)`, 6, "Expected number but found: abc" ),
        typRefFail( "T1~[--3,4)", 6, "Expected range value but found: -" ),
        typRefFail( "T1~[,]", 4, "Infinite range must be open" ),
        typRefFail( "T1~[8,]", 7, "Infinite high range must be open" ),
        typRefFail( "T1~[,8]", 4, "Infinite low range must be open" ),
        typRefFail( "&ns1@v1/T1???", 12, 
            "a nullable type cannot itself be made nullable" ),
        typRefFail( "T1~12.1", 4, "Expected type restriction but found: 12.1" ),
    )
}
