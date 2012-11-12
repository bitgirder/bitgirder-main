package com.bitgirder.lang.reflect;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class UnmatchedParameterKeyException
extends RuntimeException
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    // may also decide to save key in a field here as well
    UnmatchedParameterKeyException( Object key )
    {
        super( "Unmatched parameter key: " + state.notNull( key, "key" ) );
    }
}
