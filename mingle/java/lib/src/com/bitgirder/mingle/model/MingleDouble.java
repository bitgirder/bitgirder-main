package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class MingleDouble
implements MingleValue,
           Comparable< MingleDouble >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final double num;

    MingleDouble( double num ) { this.num = num; }

    public
    int
    compareTo( MingleDouble o )
    {
        if ( o == null ) throw new NullPointerException();
        return Double.compare( num, o.num );
    }

    // See java.lang.Double.hashCode()
    public 
    int 
    hashCode() 
    { 
        long l = Double.doubleToLongBits( num );
        return (int) ( l ^ ( l >>> 32 ) );
    }

    public
    boolean
    equals( Object other )
    {
        return 
            other == this ||
            ( other instanceof MingleDouble &&
              ( (MingleDouble) other ).num == num );
    }

    @Override public String toString() { return Double.toString( num ); }

    public long longValue() { return (long) num; }
    public int intValue() { return (int) num; }
    public short shortValue() { return (short) num; }
    public byte byteValue() { return (byte) num; }
    public double doubleValue() { return (double) num; }
    public float floatValue() { return (float) num; }
}
