package com.bitgirder.lang;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
abstract
class TypedDouble< T extends TypedDouble< T > >
extends Number
implements Comparable< T >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static class SignatureForcer {}

    private final Double d;

    private 
    TypedDouble( Double d,
                 SignatureForcer sf ) 
    { 
        this.d = d; 
    }

    protected 
    TypedDouble( Double d,
                 String paramName ) 
    { 
        this( inputs.notNull( d, paramName ), (SignatureForcer) null ); 
    }

    protected TypedDouble( Double d ) { this( d, "d" ); }

    protected 
    TypedDouble( double d ) 
    { 
        this( Double.valueOf( d ), (SignatureForcer) null ); 
    }

    public
    final
    boolean
    equals( Object other )
    {
        if ( other == this ) return true;
        else if ( other != null && other.getClass().equals( getClass() ) )
        {
            TypedDouble< ? > o = (TypedDouble< ? >) other;
            return d.equals( o.d );
        }
        else return false;
    }

    public final int hashCode() { return d.hashCode(); }

    public 
    final 
    int compareTo( T other )
    {
        if ( other == null ) throw new NullPointerException();
        else return d.compareTo( other.d );
    }

    public final byte byteValue() { return d.byteValue(); }
    public final short shortValue() { return d.shortValue(); }
    public final int intValue() { return d.intValue(); }
    public final long longValue() { return d.longValue(); }
    public final float floatValue() { return d.floatValue(); }
    public final double doubleValue() { return d.doubleValue(); }

    @Override public final String toString() { return d.toString(); }
}
