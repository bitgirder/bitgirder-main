package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

public
final
class MingleUint32
implements MingleValue,
           Comparable< MingleUint32 >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final int num;

    MingleUint32( int num ) { this.num = num; }

    public
    int
    compareTo( MingleUint32 o )
    {
        if ( o == null ) throw new NullPointerException();

        long u1 = Lang.asUnsignedInt( num );
        long u2 = Lang.asUnsignedInt( o.num );

        return u1 < u2 ? -1 : u1 == u2 ? 0 : 1;
    }

    public int hashCode() { return num; }

    public
    boolean
    equals( Object other )
    {
        return 
            other == this ||
            ( other instanceof MingleUint32 &&
              ( (MingleUint32) other ).num == num );
    }

    @Override 
    public 
    String 
    toString() 
    { 
        return Long.toString( Lang.asUnsignedInt( num ) ); 
    }

    public long longValue() { return (long) num; }
    public int intValue() { return (int) num; }
    public short shortValue() { return (short) num; }
    public byte byteValue() { return (byte) num; }
    public double doubleValue() { return (double) num; }
    public float floatValue() { return (float) num; }
}
