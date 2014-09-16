package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Strings;
import com.bitgirder.lang.Lang;

import java.util.Arrays;

public
final
class TestStruct1
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    public int f1;
    public Int32List f2;
    public TestStruct1 f3;

    // correct, don't care about speed
    public int hashCode() { return 0; }

    public
    boolean
    equals( Object o )
    {
        if ( o == this ) return true;
        if ( ! ( o instanceof TestStruct1 ) ) return false;

        TestStruct1 s = (TestStruct1) o;

        return Arrays.equals(
            new Object[] { f1, f2, f3 },
            new Object[] { s.f1, s.f2, s.f3 }
        );
    }

    @Override
    public
    String
    toString()
    {
        return Strings.inspect( this, true,
            "f1", f1,
            "f2", f2,
            "f3", f3
        ).
        toString();
    }
}
