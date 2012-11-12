package bincodec

import (
    "mingle/codec/testing"
    mg "mingle"
    mgio "mingle/io"
    "mingle/codec"
    "fmt"
    "bytes"
    "encoding/binary"
    "io"
)

var eng = testing.GetDefaultTestEngine()

var typeCodeFail = TypeCode( 100 )

func appendUtf8Input( s string, w io.Writer ) {
    appendInput( TypeCodeUtf8String, w )
    buf := []byte( s )
    appendBinInput( int32( len( buf ) ), w )
    if _, err := w.Write( buf ); err != nil { panic( err ) }
}

func appendBinInput( data interface{}, w io.Writer ) {
    if err := binary.Write( w, binary.LittleEndian, data ); err != nil {
        panic( err )
    }
}

func appendIdentifier( id *mg.Identifier, w io.Writer ) {
    mgWr := mg.NewWriter( w )
    if err := mgWr.WriteIdentifier( id ); err != nil { panic( err ) }
}

func appendTypeReference( typ mg.TypeReference, w io.Writer ) {
    mgWr := mg.NewWriter( w )
    if err := mgWr.WriteTypeReference( typ ); err != nil { panic( err ) }
}

func appendInput( data interface{}, w io.Writer ) {
    switch v := data.( type ) {
    case *mg.Identifier: appendIdentifier( v, w )
    case string: appendUtf8Input( v, w )
    case TypeCode: appendBinInput( uint8( v ), w )
    case int32: appendBinInput( v, w )
    case mg.TypeReference: appendTypeReference( v, w )
    default: panic( fmt.Errorf( "Unrecognized input elt: %T", v ) )
    }
}

func makeInput( data ...interface{} ) []byte {
    buf := &bytes.Buffer{}
    for _, val := range data { appendInput( val, buf ) }
    return buf.Bytes()
}

func failDecode( id, msg string, data ...interface{} ) *testing.TestSpec {
    return &testing.TestSpec{
        CodecId: CodecId,
        Id: mg.MustIdentifier( id ),
        Action: &testing.FailDecode{
            ErrorMessage: msg,
            Input: makeInput( data... ),
        },
    }
}

func init() {
    eng.PutCodecFactory( CodecId, func( hdrs *mgio.Headers ) codec.Codec {
        return New()
    })
    eng.MustPutSpecs(
        failDecode( 
            "unexpected-top-level-type-code",
            "[offset 0]: Saw type code 0x04 but expected 0x10",
            TypeCodeEnum,
        ),
        failDecode(
            "unexpected-symmap-val-type-code",
            `[offset 41]: Unexpected type code: 0x64`,
            TypeCodeStruct, int32( -1 ), mg.MustTypeReference( "ns@v1/S" ),
            TypeCodeField, mg.MustIdentifier( "f1" ), typeCodeFail,
        ),
        failDecode(
            "unexpected-list-val-type-code",
            `[offset 51]: Unexpected type code: 0x64`,
            TypeCodeStruct, int32( -1 ), mg.MustTypeReference( "ns@v1/S" ),
            TypeCodeField, mg.MustIdentifier( "f1" ),
            TypeCodeList, int32( -1 ),
                TypeCodeInt32, int32( 10 ), // an okay list val
                typeCodeFail,
        ),
        failDecode(
            "test-rfc3339-str-fail",
            "[offset 50]: Invalid timestamp: [<input>, line 1, col 1]: Invalid RFC3339 time: \"2009-23-22222\"",
            TypeCodeStruct, int32( -1 ), mg.MustTypeReference( "ns@v1/S" ),
            TypeCodeField, mg.MustIdentifier( "time1" ),
            TypeCodeRfc3339Str, "2009-23-22222",
        ),
    )
}
