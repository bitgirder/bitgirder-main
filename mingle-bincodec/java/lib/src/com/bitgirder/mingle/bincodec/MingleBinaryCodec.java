package com.bitgirder.mingle.bincodec;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.codec.MingleCodec;
import com.bitgirder.mingle.codec.MingleDecoder;
import com.bitgirder.mingle.codec.MingleEncoder;

public
final
class MingleBinaryCodec
implements MingleCodec
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    MingleBinaryCodec() {}

    public
    < E >
    MingleDecoder< E >
    createDecoder( Class< E > cls )
    {
        inputs.notNull( cls, "cls" );

        return MingleBinaryDecoder.create( cls );
    }

    public
    MingleEncoder
    createEncoder( Object me )
    {
        inputs.notNull( me, "me" );

        return new MingleBinaryEncoder( me );
    }
}
