package com.bitgirder.mingle.codec;

// Codecs must be threadsafe, the individual MingleEncoder or MingleDecoder
// instances which they return need not be (almost certainly: will not be)
public
interface MingleCodec
{
    public
    MingleEncoder
    createEncoder( Object me )
        throws MingleCodecException;

    public
    < E >
    MingleDecoder< E >
    createDecoder( Class< E > cls )
        throws MingleCodecException;
}
