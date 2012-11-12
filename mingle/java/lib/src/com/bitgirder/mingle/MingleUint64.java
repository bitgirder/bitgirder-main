package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.math.BigInteger;

public
final
class MingleUint64
implements MingleValue,
           Comparable< MingleUint64 >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static BigInteger HIGH_UINT_BIT =
        BigInteger.ONE.shiftLeft( 63 );

    private final long num;

    MingleUint64( long num ) { this.num = num; }

    public
    int
    compareTo( MingleUint64 o )
    {
        if ( o == null ) throw new NullPointerException();

        if ( num == o.num ) return 0;

        if ( num < 0L && o.num >= 0L ) return 1;
        if ( num >= 0L && o.num < 0L ) return -1;

        // whatever the result at this point, it doesn't concern the high bit
        long msk1 = num & Long.MAX_VALUE;
        long msk2 = o.num & Long.MAX_VALUE;

        return msk1 < msk2 ? -1 : 1;
    }

    public int hashCode() { return (int) ( num ^ ( num >>> 32 ) ); }

    public
    boolean
    equals( Object other )
    {
        return 
            other == this ||
            ( other instanceof MingleUint64 &&
              ( (MingleUint64) other ).num == num );
    }

    @Override 
    public 
    String 
    toString() 
    { 
        // For now just unconditionally use BigInteger
        BigInteger bi = BigInteger.valueOf( num & Long.MAX_VALUE );
        if ( num < 0L ) bi = bi.and( HIGH_UINT_BIT );

        return bi.toString( 10 );
    }

    public long longValue() { return (long) num; }
    public int intValue() { return (int) num; }
    public short shortValue() { return (short) num; }
    public byte byteValue() { return (byte) num; }
    public double doubleValue() { return (double) num; }
    public float floatValue() { return (float) num; }
}
