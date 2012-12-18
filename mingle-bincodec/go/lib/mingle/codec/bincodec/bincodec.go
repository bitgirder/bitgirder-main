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

func ( bc *BinCodec ) EncoderTo( w io.Writer ) mg.Reactor {
    return mg.NewWriter( w ).AsReactor()
}

func ( bc *BinCodec ) DecodeFrom( r io.Reader, rct mg.Reactor ) error {
    err := mg.NewReader( r ).ReadReactorValue( rct )
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
