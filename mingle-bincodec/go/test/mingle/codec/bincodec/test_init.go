package bincodec

import (
    "mingle/codec/testing"
    mg "mingle"
    mgio "mingle/io"
    "mingle/codec"
)

var eng = testing.GetDefaultTestEngine()

func init() {
    eng.PutCodecFactory( CodecId, func( hdrs *mgio.Headers ) codec.Codec {
        return New()
    })
    for _, test := range mg.CreateCoreIoTests() {
        idt, ok := test.( *mg.BinIoInvalidDataTest )
        if ! ok { continue }
        eng.MustPutSpecs(
            &testing.TestSpec{
                CodecId: CodecId,
                Id: mg.MustIdentifier( idt.Name ),
                Action: &testing.FailDecode{
                    ErrorMessage: idt.ErrMsg,
                    Input: idt.Input,
                },
            },
        )
    }
}
