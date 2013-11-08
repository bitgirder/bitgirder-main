package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class MingleInt32
extends MingleNumber
implements MingleValue,
           Comparable< MingleInt32 >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final int num;

    public MingleInt32( int num ) { this.num = num; }

    public
    int
    compareTo( MingleInt32 o )
    {
        if ( o == null ) throw new NullPointerException();
        return num < o.num ? -1 : num == o.num ? 0 : 1;
    }

    public int hashCode() { return num; }

    public
    boolean
    equals( Object other )
    {
        return 
            other == this ||
            ( other instanceof MingleInt32 &&
              ( (MingleInt32) other ).num == num );
    }

    @Override public String toString() { return Integer.toString( num ); }

    public long longValue() { return (long) num; }
    public int intValue() { return (int) num; }
    public short shortValue() { return (short) num; }
    public byte byteValue() { return (byte) num; }
    public double doubleValue() { return (double) num; }
    public float floatValue() { return (float) num; }

    public
    static
    MingleInt32
    parseInt( CharSequence s )
    {
        inputs.notNull( s, "s" );

        try { return new MingleInt32( Integer.parseInt( s.toString() ) ); }
        catch ( NumberFormatException nfe )
        {
            throw asNumberFormatException( nfe, s );
        }
    }
}
