package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class TestStruct2
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    public int hashCode() { return 1; }

    public boolean equals( Object o ) { return o instanceof TestStruct2; }
}
