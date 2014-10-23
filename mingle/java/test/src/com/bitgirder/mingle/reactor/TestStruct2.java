package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.util.Arrays;

public
final
class TestStruct2
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    public int f1;
    public TestStruct2 f2;

    public int hashCode() { return 1; }

    public 
    boolean 
    equals( Object o ) 
    {
        if ( o == this ) return true;
        if ( ! ( o instanceof TestStruct2 ) ) return false;

        TestStruct2 ts2 = (TestStruct2) o;

        return Arrays.equals(
            new Object[]{ f1, f2 },
            new Object[]{ ts2.f1, ts2.f2 }
        );
    }
}
