package com.bitgirder.mingle.http;

import com.bitgirder.mingle.codec.MingleCodec;

public
interface MingleHttpCodecContext
{
    public MingleCodec codec();
    public CharSequence contentType();
}
