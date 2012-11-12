package com.bitgirder.mingle.http;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.http.HttpHeaderName;

import com.bitgirder.mingle.model.MingleIdentifierFormat;

public
final
class MingleHttpConstants
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public final static HttpHeaderName HEADER_ID_STYLE =
        HttpHeaderName.forString( "x-service-id-style" );

    public final static MingleIdentifierFormat DEFAULT_ID_STYLE =
        MingleIdentifierFormat.LC_HYPHENATED;

    private MingleHttpConstants() {}
}
