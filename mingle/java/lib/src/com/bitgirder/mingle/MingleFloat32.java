package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class MingleFloat32
extends MingleNumber
implements MingleValue,
           Comparable< MingleFloat32 >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final float num;

    public MingleFloat32( float num ) { this.num = num; }

    public
    int
    compareTo( MingleFloat32 o )
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
            ( other instanceof MingleFloat32 &&
              ( (MingleFloat32) other ).num == num );
    }

    @Override public String toString() { return Float.toString( num ); }

    public long longValue() { return (long) num; }
    public int intValue() { return (int) num; }
    public short shortValue() { return (short) num; }
    public byte byteValue() { return (byte) num; }
    public double doubleValue() { return (double) num; }
    public float floatValue() { return (float) num; }

    public
    static
    MingleFloat32
    parseFloat( CharSequence s )
    {
        inputs.notNull( s, "s" );
        return new MingleFloat32( Float.parseFloat( s.toString() ) );
    }
}
