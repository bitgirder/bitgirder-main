package bincodec

import (
    mg "mingle"
    "mingle/codec"
    "io"
//    "log"
)

var CodecId = mg.MustIdentifier( "binary" )

type BinCodec struct {}

func New() *BinCodec { return &BinCodec{} }

func ( bc *BinCodec ) EncoderTo( w io.Writer ) mg.EventProcessor {
    return mg.NewWriter( w ).AsReactor()
}

func ( bc *BinCodec ) DecodeFrom( 
    r io.Reader, rep mg.EventProcessor ) error {
    err := mg.NewReader( r ).ReadReactorValue( rep )
    if ioe, ok := err.( *mg.BinIoError ); ok {
        err = codec.Error( ioe.Error() )
    }
    return err
}

func init() {
    codec.RegisterCodec(
        &codec.CodecRegistration{
            Codec: New(),
            Id: CodecId,
            Source: "mingle/codec/bincodec",
        },
    )
}
