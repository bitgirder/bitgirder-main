package com.bitgirder.lang.reflect;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class StandaloneFoo
extends ReflectionTests.AbstractFoo
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private
    StandaloneFoo( Object markerVal,
                   String returnVal )
        throws Exception
    {
        super( markerVal, returnVal );
    }
}
