package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.TypedString;

public
final
class MingleString
extends TypedString< MingleString >
implements MingleValue
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    MingleString( CharSequence cs ) { super( cs, "cs" ); }
}
