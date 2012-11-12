package com.bitgirder.lang;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.io.Serializable;

// Note that although we're unlikely to change it, the fact that this class is
// implemented on top of java.lang.String should be an implementation detail.
// However, the fact that instances are immutable *is* a part of the public
// contract.
//
// We have the actual string comparison function as an abstract method.
// Alternatively, we could keep the isCaseSensitive flag as an instance field
// and use it to implements the string comparison. The current decision reflects
// the desire to trade additional function call overhead for reduced memory size
// of the object. In any event, this decision is also an implementation detail,
// and the isEqualString method and constructor are package visible only.
abstract
class AbstractTypedString< T extends AbstractTypedString >
implements CharSequence,
           Serializable
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final String s;
    private final int hashCode;

    AbstractTypedString( String s,
                         boolean isCaseSensitive )
    {
        this.s = state.notNull( s, "s" );
        this.hashCode = ( isCaseSensitive ? s : s.toLowerCase() ).hashCode();
    }

    // isEqualString and compareString are the main workhorse methods of equals
    // and compareTo and should return values that lead to implementations that
    // are consistent-with-equals (see javadocs in java.util.Comparator).

    abstract 
    boolean 
    isEqualString( String s1,
                   String s2 );

    abstract
    int
    compareString( String s1,
                   String s2 );

    @Override public final String toString() { return s; }

    public final char charAt( int indx ) { return s.charAt( indx ); }
    public final int length() { return s.length(); }
    
    public 
    final 
    CharSequence
    subSequence( int start,
                 int end )
    {
        return s.subSequence( start, end );
    }

    public final int hashCode() { return hashCode; }

    public
    final
    boolean
    equals( Object other )
    {
        return other == this ||
               ( other != null &&
                 other.getClass().equals( getClass() ) &&
                 isEqualString( s, ( (AbstractTypedString< ? >) other ).s ) );
    }

    public
    final
    int
    compareTo( T other )
    {
        // throw NPE to conform to contract of Comparable
        if ( other == null ) throw new NullPointerException();
        else if ( other == this ) return 0; // okay since consistent-with-equals
        else return compareString( s, other.s );
    }
}
