package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

public
final
class MingleUint32
extends MingleNumber
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

        return Lang.compareUint32( num, o.num );
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
    public String toString() { return Lang.toUint32String( num ); }

    public long longValue() { return (long) num; }
    public int intValue() { return (int) num; }
    public short shortValue() { return (short) num; }
    public byte byteValue() { return (byte) num; }
    public double doubleValue() { return (double) num; }
    public float floatValue() { return (float) num; }

    public
    static
    MingleUint32
    parseUint( CharSequence s )
    {
        inputs.notNull( s, "s" );
        return new MingleUint32( Lang.parseUint32( s ) );
    }
}
