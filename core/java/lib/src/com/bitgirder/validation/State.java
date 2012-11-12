package com.bitgirder.validation;

import com.bitgirder.lang.Lang;

import java.util.Collection;
import java.util.Map;

import java.nio.ByteBuffer;

public
class State
extends Validator
{
    // Can be overridden if needed. Generally only useful to do so in
    // specialized test code
    public
    IllegalStateException
    createException( CharSequence inputName,
                     CharSequence msg )
    {
        return new IllegalStateException( getDefaultMessage( inputName, msg ) );
    }

    // Abbreviated, less user-friendly form of notNull, for use in inline and
    // test assertions. Only use this in a place where a user couldn't be
    // expected to understand, recover from, or diagnose the error anyway.
    public
    final
    < T >
    T
    notNull( T obj )
    {
        isFalse( obj == null, (Object[]) null );
        return obj;
    }

    // Fails if one is null and the other isn't. Otherwise, returns true if both
    // are non-null, false if both are null. The return value is useful when
    // using this initial check to guard further comparisons on the two objects
    // (where the comparisons will be accessing methods/fields on the objects
    // and need to know them to be not-null):
    //
    //      if ( sameNullity( o1, o2 ) ) doSomeDeeperCheck( o1, o2 );
    //
    public
    final
    boolean
    sameNullity( Object expct,
                 Object actual )
    {
        if ( ( expct == null && actual != null ) || 
             ( expct != null && actual == null ) )
        {
            if ( expct == null ) 
            {
                fail( "'expct' is null, but 'actual' is:", actual );
            }
            else fail( "'actual' is null, but 'expct' is:", expct );
        }
        
        return expct != null; // could also have used actual != null here
    }

    // The externalIndx param is meant for error message construction, since
    // this method is used both to compare to byte arrays passed in directly by
    // the user, but also to check to byte buffers by the method which checks
    // streams. The latter will pass in externalIndx to be the offset in the
    // streams being compared, allowing this method to correctly identify by
    // absolute position any offending bytes.
    //
    // checkLen is the number of bytes that are to be compared. Setting a
    // negative value indicates that the entirety of both arrays should be
    // checked.
    // 
    // The param msg may be null.
    //
    private
    void
    equal( ByteBuffer expct,
           ByteBuffer actual,
           long externalIndx,
           int checkLen,
           String msg )
    {
        expct = expct.slice();
        actual = actual.slice();

        equalInt(
            (int) expct.remaining(), (int) actual.remaining(), 
            "Byte sequence lengths differ" );

        if ( checkLen < 0 ) checkLen = actual.remaining();

        for ( int i = 0; i < checkLen; ++i )
        {
            if ( actual.get( i ) != expct.get( i ) )
            {
                if ( msg == null ) 
                {
                    msg = "Byte sequences differ at index " + 
                        ( externalIndx + i ) + ", expct=" + expct.get( i ) + 
                        ", actual=" + actual.get( i );
                }
                fail( msg );
            }
        }
    }

    public
    final
    void
    equal( ByteBuffer expct,
           ByteBuffer actual,
           String msg )
    {
        if ( sameNullity( expct, actual ) ) equal( expct, actual, 0, -1, msg );
    }

    public
    final
    void
    equal( ByteBuffer expct,
           ByteBuffer actual )
    {
        equal( expct, actual, "ByteBuffers differ" );
    }

    public
    final
    void
    equal( byte[] expct,
           byte[] actual,
           String msg )
    {
        equal( ByteBuffer.wrap( expct ), ByteBuffer.wrap( actual ), msg );
    }

    public
    final
    void
    equal( byte[] expct,
           byte[] actual )
    {
        equal( actual, expct, "Byte arrays differ" );
    }

    public
    final
    void
    equal( Collection< ? > expct,
           Collection< ? > actual,
           String msg )
    {
        if ( sameNullity( expct, actual ) )
        {
            equalInt( (int) expct.size(), (int) actual.size(), msg );
            isTrue( expct.containsAll( actual ), msg );
        }
    }
    
    // If explicit nulls is true we check for exact key set equality; if not we
    // check only that the actual keys are a subset of the expected mappings,
    // with missing keys taking on the value null implicitly during the value
    // check loop.
    private
    void
    equal( Map< ?, ? > expct,
           Map< ?, ? > actual,
           boolean explicitNulls )
    {
        if ( sameNullity( expct, actual ) )
        {
            if ( explicitNulls ) equal( expct.keySet(), actual.keySet() );
            else isTrue( expct.keySet().containsAll( actual.keySet() ) );

            for ( Map.Entry< ?, ? > e : expct.entrySet() )
            {
                equal( e.getValue(), actual.get( e.getKey() ) );
            }
        }
    }

    public
    final
    void
    equal( Map< ?, ? > expct,
           Map< ?, ? > actual )
    {
        equal( expct, actual, true );
    }

    public
    final
    void
    equalImplicitNull( Map< ?, ? > expct,
                       Map< ?, ? > actual )
    {
        equal( expct, actual, false );
    }

    public
    final
    void
    equal( Collection< ? > expct,
           Collection< ? > actual )
    {
        equal( expct, actual, createNotEqualMessage( expct, actual ) );
    }

    private
    Object
    format( Object obj )
    {
        if ( obj instanceof CharSequence )
        {
            return Lang.getRfc4627String( (CharSequence) obj );
        }
        else return obj;
    }

    private
    Object[]
    createNotEqualMessage( Object expct,
                           Object actual )
    {
        return new Object[] {
            "Arguments are not equal (got", 
            format( actual ), 
            "but expected", 
            format( expct ), 
            ")"
        };
    }

    public
    final
    void
    equal( Object expct,
           Object actual,
           Object... msg )
    {
        if ( sameNullity( expct, actual ) )
        {
            isTrue( expct.equals( actual ), msg );
        }
    }

    public
    final
    void
    equal( Object expct,
           Object actual )
    {
        equal( expct, actual, createNotEqualMessage( expct, actual ) );
    }

    public
    final
    void
    equalf( Object expct,
            Object actual,
            String fmt,
            Object... args )
    {
        if ( sameNullity( expct, actual ) )
        {
            isTruef( expct.equals( actual ), fmt, args );
        }
    }

    public
    final
    void
    equalString( CharSequence s1,
                 CharSequence s2 )
    {
        // If needed we can do this without the toString() later
        if ( sameNullity( s1, s2 ) ) equal( s1.toString(), s2.toString() );
    }

    public
    final
    void
    equal( long expct,
           long actual,
           Object... msg )
    {
        isTrue( expct == actual, msg );
    }

    public
    final
    void
    equal( long expct,
           long actual )
    {
        equal( expct, actual, createNotEqualMessage( expct, actual ) );
    }

    public
    final
    void
    equalInt( int expct,
              int actual,
              Object... msg )
    {
        equal( (long) expct, (long) actual, msg );
    }

    public
    final
    void
    equalInt( int expct,
              int actual )
    {
        equal( (long) expct, (long) actual );
    }

    public
    final
    void
    within( double expct,
            double actual,
            double tol )
    {
        positiveD( tol, "tol" );

        isTrue( 
            Math.abs( expct - actual ) <= tol,
            "Actual value", actual, "is not within", tol, "of expected value",
            expct );
    }

    // Asserts that the ratio of actual / expct is withing the given tolerance
    public
    final
    void
    withinRatio( double expct,
                 double actual,
                 double tol )
    {
        positiveD( tol, "tol" );

        double ratio = 1 - Math.abs( actual / expct );

        isTrue( 
            ratio < tol, 
            "Ratio of", actual, "to", expct, "is not within", tol );
    }

    // Asserts that o is non-null and an instance of T (non-null by virtue of
    // the behavior of Class.isInstance() -- see javadocs for more)
    public
    final
    < T >
    T
    cast( Class< T > cls,
          Object o )
    {
        T res = null;

        notNull( cls, "cls" );

        if ( ! cls.isInstance( o ) )
        {
            fail( 
                "Expected an instance of", cls, "but got",
                ( o == null ? "null object" : "instance of " + o.getClass() ) );
        }
        else res = cls.cast( o );

        return res;
    }
}
