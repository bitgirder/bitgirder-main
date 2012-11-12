package json

import (
    "mingle/codec/testing"
    "mingle/codec"
    "encoding/json"
    "strings"
    "fmt"
    "bitgirder/assert"
    mg "mingle"
    mgio "mingle/io"
)

func trimJson( s string ) []byte { 
    lines := strings.Split( s, "\n" )
    for i, line := range lines { lines[ i ] = strings.TrimSpace( line ) }
    return []byte( strings.Join( lines, " " ) )
}

var testSpecInitEng = testing.GetDefaultTestEngine()
var testSpecInitCodecId = mg.MustIdentifier( "json" )

func initFailDecode( id, input, errMsg string ) *testing.TestSpec {
    return &testing.TestSpec{
        CodecId: testSpecInitCodecId,
        Id: mg.MustIdentifier( id ),
        Action: &testing.FailDecode{
            Input: []byte( trimJson( input ) ),
            ErrorMessage: errMsg,
        },
    }
}

func initDecodeInput( id, input string, expct *mg.Struct ) *testing.TestSpec {
    return &testing.TestSpec{
        CodecId: testSpecInitCodecId,
        Id: mg.MustIdentifier( id ),
        Action: &testing.DecodeInput{
            Input: trimJson( input ),
            Expect: expct,
        },
    }
}

func initEncodeValue( 
    id string, 
    val *mg.Struct, 
    opts *mgio.Headers,
    chk testing.EncodeCheck ) *testing.TestSpec {
    return &testing.TestSpec{
        CodecId: testSpecInitCodecId,
        CodecOpts: opts,
        Id: mg.MustIdentifier( id ),
        Action: &testing.EncodeValue{ Value: val, Check: chk },
    }
}

func mustGoJsonMap( pairs ...interface{} ) map[ string ]interface{} {
    res := make( map[ string ]interface{}, len( pairs ) / 2 )
    for i, e := 0, len( pairs ); i < e; i += 2 {
        res[ pairs[ i ].( string ) ] = pairs[ i + 1 ]
    }
    return res
}

// This is a better alternative than comparing jsonAct to an expected json
// string, since we have no control (for the moment) over the order in which
// fields are serialized, meaning there is not, in the general case, a single
// valid expected string. By using a map and assert.Equal, we avoid that issue
func assertJson( 
    jsonAct []byte, expct map[ string ]interface{}, a *assert.Asserter ) {
    goAct := make( map[ string ]interface{} )
    if err := json.Unmarshal( []byte( jsonAct ), &goAct ); err != nil { 
        a.Fatal( err ) 
    }
    a.Equal( expct, goAct )
}

func initIdFormatTests() {
    id := mg.MustIdentifier( "a-field" )
    expct := mg.MustStruct( "ns1@v1/S1", id, "A val" )
    for _, idFmt := range mg.IdentifierFormats {
        idStr := id.Format( idFmt )
        jsonTmpl := `{ "$type": "ns1@v1/S1", "%s": "A val" }`
        jsonStr := fmt.Sprintf( jsonTmpl, idStr )
        nameBase := "key-id-fmt-" + idFmt.String()
        encodeCheck := func( ce *testing.CheckableEncode ) {
            m := mustGoJsonMap( "$type", "ns1@v1/S1", idStr, "A val" )
            assertJson( ce.Buffer, m, ce.Asserter )
        }
        opts := mgio.MustHeadersPairs( "id-format", idFmt.String() )
        testSpecInitEng.MustPutSpecs(
            initDecodeInput( "decode-" + nameBase, jsonStr, expct ),
            initEncodeValue( "encode-" + nameBase, expct, opts, encodeCheck ),
        )
    }
}

func initEnumExpandTests() {
    en := mg.MustEnum( "ns1@v1/E1", "val1" )
    ms := mg.MustStruct( "ns1@v1/S1", "an-enum", en )
    add := func( id string, enVal interface{}, opts *mgio.Headers ) {
        m := mustGoJsonMap( "$type", "ns1@v1/S1", "an-enum", enVal )
        chk := func( ce *testing.CheckableEncode ) {
            assertJson( ce.Buffer, m, ce.Asserter )
        }
        testSpecInitEng.MustPutSpecs( initEncodeValue( id, ms, opts, chk ) )
    }
    add( "short-enums-by-default", "val1", nil )
    add( 
        "expanded-enum", 
        mustGoJsonMap( "$type", "ns1@v1/E1", "$constant", "val1" ),
        mgio.MustHeadersPairs( "expand-enums", true ),
    )
}

func initOmitTypeFieldTests() {
    ev := initEncodeValue(
        "omit-type-fields",
        mg.MustStruct( "ns1@v1/S1", 
            "f1", mg.MustEnum( "ns1@v1/E1", "val1" ),
            "f2", "val2",
        ),
        mgio.MustHeadersPairs( "omit-type-fields", true ),
        func( ce *testing.CheckableEncode ) {
            assertJson(
                ce.Buffer,
                mustGoJsonMap( "f1", "val1", "f2", "val2" ),
                ce.Asserter,
            )
        },
    )
    ev.OutboundCodec = NewJsonCodec()
    testSpecInitEng.MustPutSpecs( ev )
}

func init() {
    initIdFormatTests()
    initEnumExpandTests()
    initOmitTypeFieldTests()
    testSpecInitEng.MustPutSpecs(
        initDecodeInput( 
            "explicit-null-field-val",
            `{"$type":"ns1@v1/S1","f1":null}`,
            mg.MustStruct( "ns1@v1/S1" ),
        ),
        initEncodeValue(
            "default-codec-encoding",
            mg.MustStruct( "ns1@v1/S1", 
                "an-enum", mg.MustEnum( "ns1@v1/E1", "val1" ) ),
            nil,
            func( ce *testing.CheckableEncode ) {
                assertJson(
                    ce.Buffer,
                    mustGoJsonMap( "$type", "ns1@v1/S1", "an-enum", "val1" ),
                    ce.Asserter,
                )
            },
        ),
        initFailDecode(
            "unrecognized-keys-in-enum",
            `{ 
                "$type": "ns1@v1/S1",
                "f1": {
                    "$type": "ns1@v1/E1",
                    "$constant": "val1",
                    "stuff": "bad"
                }
            }`,
            "f1: Enum has one or more unrecognized keys",
        ),
        initFailDecode(
            "unrecognized-control-key-toplevel",
            `{ "$type": "ns1@v1/S1", "f1": 1, "$f2": "bad" }`,
            `Unrecognized control key: "$f2"`,
        ),
        initFailDecode(
            "unrecognized-control-key-nested",
            `{
                "$type": "ns1@v1/S1",
                "f1": { "$type": "ns1@v1/S2", "$f2": "bad" }
            }`,
            `f1: Unrecognized control key: "$f2"`,
        ),
        initFailDecode(
            "invalid-enum-const",
            `{ 
                "$type": "ns1@v1/S1", 
                "f1": { "$type": "ns1@v1/E1", "$constant": [] }
            }`,
            "f1.$constant: Invalid constant value",
        ),
        initFailDecode(
            "enum-const-without-type",
            `{ "$constant": "val1" }`,
            `Missing type key ("$type")`,
        ),
        initFailDecode(
            "incomplete-type-name",
            `{ "$type": "ns1@v1" }`,
             `$type: Expected type path but found: END`,
        ),
        initFailDecode(
            "invalid-identifier-key",
            `{ "$type": "ns1@v1/S1", "f1": { "f2": { "2bad": false } } }`,
            `f1.f2.2bad: Invalid field name "2bad": Illegal start of ` +
                `identifier part: "2" (U+0032)`,
        ),
        initFailDecode(
            "invalid-identifier-enum-val",
            `{
                "$type": "ns1@v1/S1", 
                "f1": { "$type": "ns1@v1/E1", "$constant": "2bad" }
            }`,
            `f1.$constant: Invalid enum value "2bad": Illegal start of ` +
                `identifier part: "2" (U+0032)`,
        ),
        initFailDecode( "empty-document", "{}", `Missing type key ("$type")` ),
        initFailDecode( 
            "toplevel-array", "[]", "Unexpected top level JSON value" ),
    )
}

func testCodecFor( h *mgio.Headers ) codec.Codec {
    if h.Fields().Len() == 0 { return NewJsonCodec() }
    opts := &JsonCodecOpts{}
    acc := h.FieldsAccessor()
    if s, err := acc.GetGoStringByString( "id-format" ); err == nil {
        opts.IdFormat = mg.MustIdentifierFormatString( s )
    }
    if b, err := acc.GetGoBoolByString( "expand-enums" ); err == nil {
        opts.ExpandEnums = b
    }
    if b, err := acc.GetGoBoolByString( "omit-type-fields" ); err == nil {
        opts.OmitTypeFields = b
    }
    return MustJsonCodec( opts )
}

func init() {
    testSpecInitEng.PutCodecFactory( testSpecInitCodecId, testCodecFor )
}
