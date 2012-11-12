package com.bitgirder.lang.reflect;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class StandaloneFooBean
implements ReflectionTests.Foo
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private Object markerVal;
    private String returnVal;

    public
    void
    assertValue( String expct )
    {
        state.equalString( expct, returnVal );
    }
}
