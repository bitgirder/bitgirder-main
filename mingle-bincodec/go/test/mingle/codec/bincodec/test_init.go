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
    for _, fi := range mg.BinWriterFailureInputs {
        eng.MustPutSpecs(
            &testing.TestSpec{
                CodecId: CodecId,
                Id: fi.Id,
                Action: &testing.FailDecode{
                    ErrorMessage: fi.ErrMsg,
                    Input: fi.Input,
                },
            },
        )
    }
}
