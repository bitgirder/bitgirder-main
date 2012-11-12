package bincodec

import (
    "testing"
    "bitgirder/assert"
    "mingle/codec"
    mg "mingle"
    codecTesting "mingle/codec/testing"
    "bytes"
)

func getBinCodec() *BinCodec {
    return New()
}

func TestStandardSpecs( t *testing.T ) {
    codecTesting.TestCodecSpecs( CodecId, t )
}

func TestCodecLeavesTrailingInput( t *testing.T ) {
    cdc := getBinCodec()
    buf := &bytes.Buffer{}
    val := mg.MustStruct( "ns1@v1/S1" )
    if err := codec.Encode( val, cdc, buf ); err != nil { t.Fatal( err ) }
    buf.WriteString( "more-stuff" )
    if val2, err := codec.Decode( cdc, buf ); err == nil {
        assert.Equal( val, val2 )
        assert.Equal( "more-stuff", buf.String() )
    } else { t.Fatal( err ) }
}

func TestCodecRegistration( t *testing.T ) {
    codecTesting.TestCodecRegistration( CodecId, t, func( cdc codec.Codec ) {
        _ = cdc.( *BinCodec )
    })
}

func TestWriteReactorErrors( t *testing.T ) {
    codecTesting.TestCodecErrorSequences( New(), t )
}
