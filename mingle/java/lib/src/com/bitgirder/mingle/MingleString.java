package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.TypedString;
import com.bitgirder.lang.Lang;

public
final
class MingleString
extends TypedString< MingleString >
implements MingleValue
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    MingleString( CharSequence cs ) { super( cs, "cs" ); }

    public
    CharSequence
    getExternalForm()
    {
        return Lang.getRfc4627String( this );
    }
}
