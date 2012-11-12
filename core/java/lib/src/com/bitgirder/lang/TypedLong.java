package com.bitgirder.lang;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
abstract
class TypedLong< I extends TypedLong< I > >
extends Number
implements Comparable< I >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    // Because longVal is final and immutable, we allow subclasses to access it
    // directly
    protected final long longVal;

    protected TypedLong( long longVal ) { this.longVal = longVal; }

    public final byte byteValue() { return (byte) longVal; }
    public final short shortValue() { return (short) longVal; }
    public final int intValue() { return (int) longVal; }
    public final long longValue() { return longVal; }
    public final float floatValue() { return (float) longVal; }
    public final double doubleValue() { return (double) longVal; }

    @Override 
    public final String toString() { return Long.toString( longVal ); }

    public
    final
    boolean
    equals( Object other )
    {
        if ( other == this ) return true;
        else if ( other != null && other.getClass().equals( getClass() ) )
        {
            return longVal == ( (TypedLong) other ).longVal;
        }
        else return false;
    }

    // see Long.hashCode()
    public
    final
    int
    hashCode()
    {
        return (int) ( longVal ^ ( longVal >>> 32 ) );
    }

    public
    final
    int
    compareTo( I other )
    {
        if ( other == null ) throw new NullPointerException();
        else
        {
            return 
                longVal == other.longVal ?
                    0 : ( longVal < other.longVal ? -1 : 1 );
        }
    }
}
