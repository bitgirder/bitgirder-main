package json

import (
    "testing"
    "bitgirder/assert"
    codecTesting "mingle/codec/testing"
    mg "mingle"
    "mingle/codec"
    "bytes"
    "io"
)

func toJsonStr( s *mg.Struct, c *JsonCodec, t *testing.T ) string {
    buf := &bytes.Buffer{}
    if err := codec.Encode( s, c, buf ); err != nil { t.Fatal( err ) }
    return buf.String()
}

func fromJsonStr( s string, c *JsonCodec ) ( *mg.Struct, error ) {
    return codec.Decode( c, bytes.NewBufferString( s ) )
}

func TestStandardSpecs( t *testing.T ) {
    codecTesting.TestCodecSpecs( mg.MustIdentifier( "json" ), t )
}

func TestOmitTypeFieldsAndExpandEnumsFails( t *testing.T ) {
    opts := &JsonCodecOpts{ OmitTypeFields: true, ExpandEnums: true }
    if _, err := CreateJsonCodec( opts ); err == nil {
        t.Fatalf( "Expected err" )
    } else {
        if jce, ok := err.( *JsonCodecInitializerError ); ok {
            assert.Equal( 
                "Invalid combination: ExpandEnums and OmitTypeFields", 
                jce.Error() )
        } else { t.Fatal( err ) }
    }
}

func TestJsonEmptyStreamFails( t *testing.T ) {
    if _, err := fromJsonStr( "", NewJsonCodec() ); err == nil {
        t.Fatalf( "Got decode" )
    } else if err != io.EOF { t.Fatal( err ) }
}

func TestCodecRegistration( t *testing.T ) {
    codecTesting.TestCodecRegistration( CodecId, t, func( cdc codec.Codec ) {
        _ = cdc.( *JsonCodec )
    })
}
