package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

public
final
class MingleUint64
extends MingleNumber
implements MingleValue,
           Comparable< MingleUint64 >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final long num;

    public MingleUint64( long num ) { this.num = num; }

    public
    int
    compareTo( MingleUint64 o )
    {
        if ( o == null ) throw new NullPointerException();

        return Lang.compareUint64( num, o.num );
    }

    public int hashCode() { return (int) ( num ^ ( num >>> 32 ) ); }

    public
    boolean
    equals( Object other )
    {
        if ( other == this ) return true;
        if ( ! ( other instanceof MingleUint64 ) ) return false;

        MingleUint64 o = (MingleUint64) other;
        return o.num == num;
    }

    @Override 
    public String toString() { return Lang.toUint64String( num ); }

    public long longValue() { return (long) num; }
    public int intValue() { return (int) num; }
    public short shortValue() { return (short) num; }
    public byte byteValue() { return (byte) num; }
    public double doubleValue() { return (double) num; }
    public float floatValue() { return (float) num; }

    public
    static
    MingleUint64
    parseUint( CharSequence s )
    {
        inputs.notNull( s, "s" );
        return new MingleUint64( Lang.parseUint64( s ) );
    }
}
