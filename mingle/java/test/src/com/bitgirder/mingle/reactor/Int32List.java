package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import java.util.Arrays;
import java.util.List;

// used to wrap int[] and provide toString() and equals()
public
final
class Int32List
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final int[] arr;

    public
    Int32List( int[] arr )
    {
        this.arr = inputs.notNull( arr, "arr" );
    }

    public int hashCode() { return Arrays.hashCode( arr ); }

    public
    boolean
    equals( Object o )
    {
        if ( o == this ) return true;
        if ( ! ( o instanceof Int32List ) ) return false;

        Int32List l = (Int32List) o;

        return Arrays.equals( arr, l.arr );
    }

    @Override
    public
    String
    toString()
    {
        List< Integer > l = Lang.newList( arr.length );
        for ( int i : arr ) l.add( i );

        return "Int32List" + l;
    }
}
