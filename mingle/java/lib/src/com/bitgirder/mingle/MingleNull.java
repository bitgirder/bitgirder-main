package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class MingleNull
implements MingleValue
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static MingleNull INSTANCE = new MingleNull();

    private MingleNull() {}

    public static MingleNull getInstance() { return INSTANCE; }

    // Make sure our string form is different from java null
    @Override public String toString() { return "[mingle null]"; }
}
