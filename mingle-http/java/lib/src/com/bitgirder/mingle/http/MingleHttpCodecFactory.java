package com.bitgirder.mingle.http;

import com.bitgirder.http.HttpRequestMessage;

import com.bitgirder.mingle.codec.MingleCodecException;

public
interface MingleHttpCodecFactory
{
    public
    MingleHttpCodecContext
    codecContextFor( HttpRequestMessage msg )
        throws MingleCodecException;
}
