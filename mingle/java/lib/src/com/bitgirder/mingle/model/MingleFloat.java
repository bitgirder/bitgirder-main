package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class MingleFloat
implements MingleValue,
           Comparable< MingleFloat >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final float num;

    MingleFloat( float num ) { this.num = num; }

    public
    int
    compareTo( MingleFloat o )
    {
        if ( o == null ) throw new NullPointerException();
        return Float.compare( num, o.num );
    }

    // See java.lang.Float.hashCode()
    public int hashCode() { return Float.floatToIntBits( num ); }

    public
    boolean
    equals( Object other )
    {
        return 
            other == this ||
            ( other instanceof MingleFloat &&
              ( (MingleFloat) other ).num == num );
    }

    @Override public String toString() { return Float.toString( num ); }

    public long longValue() { return (long) num; }
    public int intValue() { return (int) num; }
    public short shortValue() { return (short) num; }
    public byte byteValue() { return (byte) num; }
    public double doubleValue() { return (double) num; }
    public float floatValue() { return (float) num; }
}
