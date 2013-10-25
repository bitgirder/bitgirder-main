package com.bitgirder.validation;

import com.bitgirder.lang.Range;

import java.util.Map;
import java.util.Set;
import java.util.Arrays;
import java.util.Properties;

import java.util.regex.Pattern;
import java.util.regex.Matcher;

public
abstract
class Validator
{
    private final static Object[] EMPTY_MESSAGE = new Object[] {};

    private final static Inputs inputs = new Inputs();
 
    protected
    final
    String
    getDefaultMessage( CharSequence inputName,
                       CharSequence msg )
    {
        String res;

        if ( inputName == null && msg == null ) res = null;
        else
        {
            StringBuilder sb = new StringBuilder();
            if ( inputName != null ) 
            {
                sb.append( "Input '" ).append( inputName ).append( "' " );
            }

            res = sb.append( msg ).toString();
        }

        return res;
    }

    // Must throw a runtime exception using the given message (it may add to or
    // further embellish it as desired, such as by adding location information
    // based on inputName). Either or both may be null, although the latter
    // should be warned or logged by the subclass, since this likely represents
    // a programming error.
    //
    // Most impls will simply throw an exception with a message created by
    // getDefaultMessage
    public
    abstract
    RuntimeException
    createException( CharSequence inputName,
                     CharSequence msg );

    private
    void
    throwException( CharSequence inputName,
                    CharSequence msg )
    {
        throw createException( inputName, msg );
    }

    private
    CharSequence
    makeMessage( Object... message )
    {
        // Note: we just use a StringBuilder here and don't call into
        // Strings.join, to avoid possible circularity problems where there is
        // some bug in Strings.join which leads to a call from within that
        // method back to this one, ultimately blowing out the stack
        StringBuilder sb = new StringBuilder();

        if ( message != null )
        {
            for ( int i = 0, e = message.length; i < e; )
            {
                sb.append( message[ i ] );
                if ( ++i < e ) sb.append( ' ' );
            }
        }

        return sb;
    }
 
    // Does not actually return a value, but is typed to so that callers can use
    // this method either as a standalone call:
    //
    //      v.fail( "..." );
    //
    // as well as as part of a 'throw' statement, which is particularly useful
    // when needing to convince the compiler that a method will not return a
    // value:
    //
    //      throw v.fail( "..." );
    //
    public
    final
    RuntimeException
    fail( Object... message )
    { 
        throw createException( null, makeMessage( message ) );
    }

    public
    final
    RuntimeException
    failf( String fmt,
           Object... args )
    {
        return fail( String.format( fmt, args ) );
    }

    public
    final
    RuntimeException
    failInput( String input,
               Object... message )
    {
        throw createException( input, makeMessage( message ) );
    }

    public
    final
    RuntimeException
    createFail( Object... message )
    {
        return createException( (CharSequence) null, makeMessage( message ) );
    }

    public
    final
    RuntimeException
    createFailf( String fmt,
                 Object... args )
    {
        return createFail( String.format( fmt, args ) );
    }

    public
    final
    void
    isTrue( boolean b,
            Object... message )
    {
        if ( ! b ) fail( message );
    }

    public
    final
    void
    isTrue( String inputName,
            boolean b,
            Object... message )
    {
        if ( ! b ) failInput( inputName, message );
    }

    public
    final
    void
    isTruef( boolean b,
             String fmt,
             Object... args )
    {
        if ( ! b ) failf( fmt, args );
    }

    public
    final
    void
    isFalse( boolean b,
             Object... message )
    {
        isTrue( ! b, message );
    }

    public
    final
    void
    isFalsef( boolean b,
              String fmt,
              Object... args )
    {
        isTruef( ! b, fmt, args );
    }

    public
    final
    < T >
    T
    notNull( T val,
             String inputName )
    {
        if ( val == null ) throw createException( inputName, "cannot be null" );
 
        return val;
    }

    public
    final
    < T, I extends Iterable< T > >
    I
    noneNull( I vals,
              String inputName )
    {
        notNull( vals, inputName ); 

        int i = 0;
        for ( T val : vals )
        {
            isTrue( val != null, "Element ", i, " in iteration order is null" );
            i++;
        }

        return vals;
    }
 
    public
    final
    < T >
    T[]
    noneNull( T[] vals,
              String inputName )
    {
        noneNull( Arrays.asList( notNull( vals, inputName ) ), inputName );
        return vals;
    }

    private
    void
    failNoKey( String mapName,
               String mapType,
               Object key )
    {
        fail( mapType + " '" + mapName + "' has no value for key " + key );
    }

    // This is the only version now, but we could extend it to allow for a fourt
    // parameter, a boolean specifying whether a null value is allowed. This
    // method will then be a default call into that one with a value of false
    // (do not allow null keys)
    public
    final
    < K, V >
    V
    get( Map< K, V > map,
         K key,
         String mapName )
    {
        inputs.notNull( map, "map" );

        V res = map.get( key );

        if ( res == null ) failNoKey( mapName, "Map", key );

        return res;
    }

    public
    final
    String
    getProperty( Properties props,
                 String prop,
                 String propsName )
    {
        inputs.notNull( props, "props" );
        inputs.notNull( prop, "prop" );
        inputs.notNull( propsName, "propsName" );
        
        String res = props.getProperty( prop );

        if ( res == null ) failNoKey( propsName, "Properties", prop );

        return res;
    }

    public
    final
    < K, V >
    V
    remove( Map< K, V > map,
            K key,
            String mapName )
    {
        inputs.notNull( map, "map" );

        V res = map.remove( key );

        if ( res == null ) failNoKey( mapName, "Map", key );

        return res;
    }

    public
    final
    boolean
    remove( Set< ? > s,
            Object o,
            String setName )
    {
        inputs.notNull( s, "s" );
        inputs.notNull( setName, "setName" );

        boolean res = s.remove( o );

        isTrue( res, "Set '" + setName + "' did not contain element:", o );

        return res;
    }

    // Implies also a call to notNull( map, mapName )
    public
    final
    < K, V, M extends Map< K, V > >
    M
    noneNull( M map,
              String mapName,
              boolean allowNullValues )
    {
        notNull( map, mapName );

        for ( K key : map.keySet() )
        {
            isFalse( key == null, "Map '", mapName, "' has null key" );

            isFalse( 
                map.get( key ) == null && ! allowNullValues,
                "Map", mapName, "has null value for key", key );
        }

        return map;
    }
 
    public
    final
    < K, V, M extends Map< K, V > >
    M
    noneNull( M map,
              String mapName )
    {
        return noneNull( map, mapName, false );
    }

    // For now the only Properties we use are System properties, so this method
    // just goes straight to System.getProperty. If we later need to access
    // arbitrary Properties entries via this class, we should convert the
    // getSystemProperty method to be just a callthrough to whatever we add for
    // arbitrary Properties instances, and have this method pass in the return
    // from System.getProperties()
    public
    String
    getSystemProperty( String name )
    {
        inputs.notNull( name, "name" );
        
        String res = System.getProperty( name );
        if ( res == null ) fail( "No system property is set for key:", name );

        return res;
    }

    public
    final
    Matcher
    matches( CharSequence s,
             String inputName,
             Pattern pat )
    {
        // if pat is null, this is a programming error and is handled separately
        inputs.notNull( pat, "pat" );

        notNull( s, inputName );
 
        Matcher res = pat.matcher( s );

        if ( ! res.matches() )
        {
            throwException( 
                inputName, 
                "does not match pattern " + pat + " (got " + s + ")" 
            );
        }

        return res;
    }

    private
    long
    numeric( long val,
             String inputName,
             NumericAssertion assertion )
    {
        if ( assertion == null ) assertion = NumericAssertion.NONE;

        boolean res = true; // we have to assign to satisfy javac

        switch ( assertion )
        {
            case NONE: res = true; break;
            case POSITIVE: res = val > 0; break;
            case NEGATIVE: res = val < 0; break;
            case NONNEGATIVE: res = val >= 0; break;
        }

        if ( ! res )
        {
            fail( inputName, "must be", assertion, "(got", val, ")" );
        }

        return val;
    }

    public
    final
    long
    positiveL( long val,
               String inputName )
    {
        return numeric( val, inputName, NumericAssertion.POSITIVE );
    }

    public
    final
    int 
    positiveI( int val,
               String inputName )
    {
        return (int) positiveL( val, inputName );
    }

    public
    final
    long
    nonnegativeL( long val,
                  String inputName )
    {
        return numeric( val, inputName, NumericAssertion.NONNEGATIVE );
    }

    public
    final
    int
    nonnegativeI( int val,
                  String inputName )
    {
        return (int) nonnegativeL( val, inputName );
    }

    public
    final
    double
    positiveD( double val,
               String inputName )
    {
        if ( val < Double.MIN_VALUE )
        {
            fail( inputName, "is not a representable positive value:", val );
        }

        return val;
    }

    public
    final
    float
    positiveF( float val,
               String inputName )
    {
        if ( val < Float.MIN_VALUE )
        {
            fail( inputName, "is not a representable positive value:", val );
        }

        return val;
    }

    // Implies a not-null check too for val using inputName
    public
    < V extends Comparable< V > >
    V
    inRange( V val,
             String inputName,
             Range< V > range )
    {
        notNull( val, inputName );
        inputs.notNull( range, "range" );

        if ( ! range.includes( val ) ) 
        {
            fail( inputName, "is not in range", range, "(value was", val, ")" );
        }

        return val;
    }

    private
    static
    enum NumericAssertion
    {
        NONE,
        POSITIVE,
        NEGATIVE,
        NONNEGATIVE;
 
        public
        String
        toString()
        {
            return name().toLowerCase();
        }
    }
}
