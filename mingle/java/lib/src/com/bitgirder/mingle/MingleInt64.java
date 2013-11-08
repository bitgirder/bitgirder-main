package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class MingleInt64
extends MingleNumber
implements MingleValue,
           Comparable< MingleInt64 >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final long num;

    public MingleInt64( long num ) { this.num = num; }

    public
    int
    compareTo( MingleInt64 o )
    {
        if ( o == null ) throw new NullPointerException();
        return num < o.num ? -1 : num == o.num ? 0 : 1;
    }

    // same as java.lang.Long.hashCode()
    public int hashCode() { return (int) ( num ^ ( num >>> 32 ) ); }

    public
    boolean
    equals( Object other )
    {
        return 
            other == this ||
            ( other instanceof MingleInt64 &&
              ( (MingleInt64) other ).num == num );
    }

    @Override public String toString() { return Long.toString( num ); }

    public long longValue() { return (long) num; }
    public int intValue() { return (int) num; }
    public short shortValue() { return (short) num; }
    public byte byteValue() { return (byte) num; }
    public double doubleValue() { return (double) num; }
    public float floatValue() { return (float) num; }

    public
    static
    MingleInt64
    parseInt( CharSequence s )
    {
        inputs.notNull( s, "s" );
        
        try { return new MingleInt64( Long.parseLong( s.toString() ) ); }
        catch ( NumberFormatException nfe )
        {
            throw asNumberFormatException( nfe, s );
        }
    }
}
