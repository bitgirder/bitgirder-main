package com.bitgirder.lang;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.Validator;
import com.bitgirder.validation.State;

import java.lang.reflect.Array;

import java.util.AbstractList;
import java.util.ArrayDeque;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collection;
import java.util.Collections;
import java.util.Deque;
import java.util.HashMap;
import java.util.HashSet;
import java.util.HashSet;
import java.util.LinkedList;
import java.util.List;
import java.util.Map;
import java.util.RandomAccess;
import java.util.Queue;
import java.util.Set;
import java.util.SortedMap;
import java.util.SortedSet;
import java.util.TreeMap;
import java.util.TreeSet;
import java.util.UUID;
import java.util.Comparator;

import java.util.concurrent.BlockingQueue;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;
import java.util.concurrent.ConcurrentSkipListSet;
import java.util.concurrent.ConcurrentLinkedQueue;
import java.util.concurrent.ArrayBlockingQueue;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.LinkedBlockingDeque;

import java.io.UnsupportedEncodingException;
import java.io.ByteArrayOutputStream;
import java.io.PrintStream;

import java.math.BigInteger;

public
final
class Lang
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static Runnable NO_OP =
        new Runnable() { public void run() {} };

    public final static String TRACE_ORIGIN_PROP_NAME =
        "com.bitgirder.lang.Lang.traceOrigin";

    private final static Casts casts = new Casts();

    private final static BigInteger BI_MAX_UINT32 =
        BigInteger.ONE.shiftLeft( 32 ).subtract( BigInteger.ONE );

    private final static BigInteger BI_MAX_UINT64 =
        BigInteger.ONE.shiftLeft( 64 ).subtract( BigInteger.ONE );

    private final static Range< Integer > OCTET_RANGE = Range.closed( 0, 255 );

    // since CompletionImpl instances are immutable we can reuse this with any
    // parameterization V of Completion< V > when it is meant to represent a
    // successful return value of null
    private final static Completion< Object > COMPLETION_SUCCESS_NULL =
        new CompletionImpl< Object >( null, null );

    private Lang() {}

    public static Runnable getNoOpRunnable() { return NO_OP; }

    public
    static
    < V >
    V
    castUnchecked( Object obj )
    {
        @SuppressWarnings( "unchecked" )
        V res = (V) obj;

        return res;
    }

    public
    static
    String
    randomUuid()
    {
        return UUID.randomUUID().toString();
    }

    // Performs an integer division version of ceil. Casting to double/float and
    // attempting to use Math.ceil can cause rounding errors, so we do it here
    // without loss of precision
    public
    static
    int
    ceilI( int numerator,
           int denominator )
    {
        if ( denominator == 0 ) 
        {
            throw new ArithmeticException( "denominator is 0" );
        }

        int res = numerator / denominator;
        if ( numerator % denominator != 0 ) ++res;

        return res;
    }

    // get b as an unsigned byte, stored in an int
    public 
    static 
    int 
    asOctet( byte b ) 
    { 
        int res = b & 127;
        return b < 0 ? res + 128 : res;
    }

    // Gets an unsigned byte, passed in an int, as a signed byte. This
    // implementation is essentially just a checked narrowing conversion from
    // int --> byte
    public
    static
    byte
    fromOctet( int octet )
    {
        inputs.inRange( octet, "octet", OCTET_RANGE );
        return (byte) octet;
    }

    public
    static
    int
    compareUint32( int i1,
                   int i2 )
    {
        if ( i1 == i2 ) return 0;

        boolean less = ( i1 + Integer.MIN_VALUE ) < ( i2 + Integer.MIN_VALUE );
        return less ? -1 : 1;
    }

    public
    static
    String
    toUint32String( int i )
    {
        if ( i >= 0 ) return Integer.toString( i );

        long l = ( 1L << 31 ) | (long) ( i & Integer.MAX_VALUE );

        return Long.toString( l );
    }

    public
    static
    int
    compareUint64( long l1,
                   long l2 )
    {
        if ( l1 == l2 ) return 0;

        boolean less = ( l1 + Long.MIN_VALUE ) < ( l2 + Long.MIN_VALUE );

        return less ? -1 : 1;
    }

    // null-checks param on behalf of public frontends
    private
    static
    BigInteger
    parseUint( CharSequence s,
               BigInteger max,
               String errNm )
    {
        String numStr = inputs.notNull( s, "s" ).toString().trim();

        if ( numStr.length() == 0 ) 
        {
            throw new NumberFormatException( "Empty number" );
        }

        BigInteger bi = new BigInteger( s.toString() );

        if ( bi.signum() < 0 ) 
        {
            throw new NumberFormatException( "Number is negative: " + s );
        }

        if ( bi.compareTo( max ) > 0 )
        {
            throw new NumberFormatException( 
                "Number is too large for " + errNm + ": " + s );
        }

        return bi;
    }

    public
    static
    int
    parseUint32( CharSequence s )
    {
        return parseUint( s, BI_MAX_UINT32, "uint32" ).intValue();
    }

    public
    static
    String
    toUint64String( long l )
    {
        if ( l >= 0 ) return Long.toString( l );

        BigInteger bi = BigInteger.valueOf( l & Long.MAX_VALUE );
        bi = bi.flipBit( 63 );

        return bi.toString();
    }

    public
    static
    long
    parseUint64( CharSequence s )
    {
        return parseUint( s, BI_MAX_UINT64, "uint64" ).longValue();
    }

    public
    static
    < T >
    T[]
    toArray( Collection< T > coll,
             Class< T > clsToken,
             boolean allowNull,
             String paramName )
    {
        inputs.notNull( coll, paramName );
        
        // Because T is the target of clsToken, we know it is not a
        // parameterized type
        @SuppressWarnings( "unchecked" )
        T[] res = (T[]) Array.newInstance( clsToken, coll.size() );

        int i = 0;
        for ( T elt : coll )
        {
            if ( elt == null && ! allowNull )
            {
                inputs.fail( "Collection element at index", i, "is null" );
            }

            res[ i++ ] = elt;
        }

        return res;
    }

    private
    final
    static
    class TypedStringComparator< S extends TypedString< S > >
    implements Comparator< S >
    {
        public
        int
        compare( S o1,
                 S o2 )
        {
            if ( o1 == null || o2 == null ) throw new NullPointerException();
            else return o1.toString().compareTo( o2.toString() );
        }
    }

    public
    static
    < S extends TypedString< S > >
    Comparator< S >
    getComparator()
    {
        return new TypedStringComparator< S >();
    }

    public
    static
    < V >
    void
    addAll( Collection< ? super V > coll,
            Iterable< ? extends V > it )
    {
        inputs.notNull( coll, "coll" );
        inputs.notNull( it, "it" );

        if ( it instanceof Collection )
        {
            @SuppressWarnings( "unchecked" )
            Collection< ? extends V > c = (Collection< ? extends V >) it;

            coll.addAll( c );
        }
        else for ( V v : it ) coll.add( v );
    }

    public static < V > List< V > newList() { return new ArrayList< V >(); }

    @SafeVarargs
    public
    static
    < V >
    List< V >
    asList( V... arr )
    {
        inputs.notNull( arr, "arr" );
        return Arrays.asList( arr );
    }

    public 
    static
    < V >
    List< V >
    singletonList( V elt )
    {
        return Collections.singletonList( elt );
    }

    public
    static
    < V >
    Set< V >
    singletonSet( V elt )
    {
        Set< V > res = newSet();
        res.add( elt );

        return unmodifiableSet( res );
    }

    public
    static
    < V >
    Set< V >
    collectSet( Iterable< V > vals )
    {
        inputs.notNull( vals, "vals" );

        Set< V > res = newSet();

        for ( V val : vals ) res.add( val );

        return res;
    }

    public
    static
    < V >
    Collection< V >
    copyOf( Collection< V > c )
    {
        return new ArrayList< V >( inputs.notNull( c, "c" ) );
    }

    public
    static
    < V >
    List< V >
    copyOf( List< V > l )
    {
        inputs.notNull( l, "l" );
        return new ArrayList< V >( l );
    }

    public 
    static 
    < V > 
    List< V > 
    emptyList() 
    { 
        return Collections.emptyList();
    }

    public static < V > Queue< V > newQueue() { return new LinkedList< V >(); }

    public
    static
    < V >
    Deque< V >
    newDeque() 
    { 
        return new LinkedList< V >(); 
    }

    public
    static
    < V >
    Deque< V >
    newDeque( int expctMaxCap )
    {
        return new ArrayDeque< V >( expctMaxCap );
    }

    public
    static
    < V >
    List< V >
    newList( int expctMaxCap )
    {
        inputs.nonnegativeI( expctMaxCap, "expctMaxCap" );
        return new ArrayList< V >( expctMaxCap );
    }

    public
    static
    < V >
    List< V >
    newList( Collection< V > coll )
    {
        return new ArrayList< V >( inputs.notNull( coll, "coll" ) );
    }

    public
    static
    < V >
    List< V >
    newSynchronizedList()
    {
        return Collections.synchronizedList( Lang.< V >newList() );
    }

    // Returns a set which is not guaranteed to be safe for concurrent
    // operations, but which is guaranteed to provide for safe publishing of
    // non-overlapping operations made from different threads.
    //
    // Currently returns a synchronized set, but that should be considered only
    // an implementation detail subject to change
    public
    static
    < V >
    Set< V >
    publishSafeSet()
    {
        return Collections.synchronizedSet( new HashSet< V >() );
    }
    
    public
    static
    < V >
    Set< V >
    newConcurrentSet()
    {
        // Ultimately we'll want to replace this with some better impls,
        // probably a non-blocking one such as backed by a concurrent hash
        // map, possibly even adding a variant of this method which takes the
        // type key to determine whether V is comparable, possibly returning
        // different impls based on that fact
        return Collections.synchronizedSet( new HashSet< V >() );
    }

    public
    static
    < K, V >
    Map< K, V >
    publishSafeMap()
    {
        return new ConcurrentHashMap< K, V >();
    }

    public
    static
    < K, V >
    ConcurrentMap< K, V >
    newConcurrentMap()
    {
        return new ConcurrentHashMap< K, V >();
    }

    public
    static
    < V >
    Collection< V >
    publishSafeCollection()
    {
        return new ConcurrentLinkedQueue< V >();
    }

    public
    static
    < V >
    Collection< V >
    newConcurrentCollection()
    {
        return new LinkedBlockingQueue< V >();
    }

    public
    static
    < V >
    BlockingQueue< V >
    newBlockingQueue()
    {
        return new LinkedBlockingQueue< V >();
    }

    public
    static
    < V >
    Queue< V >
    newConcurrentQueue()
    {
        return new ConcurrentLinkedQueue< V >();
    }

    public
    static
    < V >
    Queue< V >
    newBoundedConcurrentQueue( int sz )
    {
        inputs.positiveI( sz, "sz" );
        return new ArrayBlockingQueue< V >( sz );
    }

    public
    static
    < V >
    Deque< V >
    newConcurrentDeque()
    {
        return new LinkedBlockingDeque< V >();
    }

    public static < V > Set< V > newSet() { return new HashSet< V >(); }

    public
    static
    < V >
    Set< V >
    newSet( int initialCap )
    {
        inputs.nonnegativeI( initialCap, "initialCap" );
        return new HashSet< V >( initialCap );
    }

    // Coll may contain null, but all null elements will be coalesced into a
    // single entry
    public
    static
    < V >
    Set< V >
    newSet( Collection< V > coll )
    {
        return new HashSet< V >( inputs.notNull( coll, "coll" ) );
    }

    public
    static
    < V >
    Set< V >
    newSynchronizedSet()
    {
        return Collections.synchronizedSet( Lang.< V >newSet() );
    }

    public
    static
    < V >
    Set< V >
    copyOf( Set< V > other )
    {
        return newSet( inputs.notNull( other, "other" ) );
    }

    // arr itself may not be null, but elts may be null according to newSet(
    // Collection )
    @SafeVarargs
    public
    static
    < V >
    Set< V >
    newSet( V... arr )
    {
        return newSet( Arrays.asList( inputs.notNull( arr, "arr" ) ) );
    }

    public
    static
    < V extends Comparable< V > >
    SortedSet< V >
    newSortedSet()
    {
        return new TreeSet< V >();
    }

    public
    static
    < V >
    SortedSet< V >
    newSortedSet( Comparator< V > comp )
    {
        return new TreeSet< V >( inputs.notNull( comp, "comp" ) );
    }

    public
    static
    < K, V >
    Map< K, V >
    putAll( Map< K, V > map,
            Class< ? extends K > keyCls,
            Class< ? extends V > valCls,
            Object... pairs )
    {
        inputs.notNull( map, "map" );
        inputs.notNull( keyCls, "keyCls" );
        inputs.notNull( valCls, "valCls" );
        inputs.notNull( pairs, "pairs" ); // may contain nulls as map allows

        for ( int i = 0, e = pairs.length; i < e; )
        {
            Object keyObj = pairs[ i++ ];
            state.isTrue( i < e, "Missing final map value" );
            Object valObj = pairs[ i++ ];

            if ( keyObj == null || keyCls.isInstance( keyObj ) )
            {
                if ( valObj == null || valCls.isInstance( valObj ) )
                {
                    map.put( keyCls.cast( keyObj ), valCls.cast( valObj ) );
                }
                else casts.fail( "Invalid map value type:", valObj.getClass() );
            }
            else casts.fail( "Invalid map key type:", keyObj.getClass() );
        }

        return map;
    }

    // Returns a map that accepts null values, but is not guaranteed to accept
    // the null key. The map returned by this method will be such that
    // concurrent reads may be safely executed without any extra
    // synchronization (impl note: we currently rely on the fact that
    // java.util.HashMap satisfies this property. If that later turns out to not
    // be the case we can create our own map impl that does support this
    // property)
    public
    static
    < K, V >
    Map< K, V >
    newMap() 
    {
        return new HashMap< K, V >(); 
    }

    // Returns a map that accepts null values, but is not guaranteed to accept
    // the null key
    public
    static
    < K, V >
    SortedMap< K, V >
    newSortedMap()
    {
        return new TreeMap< K, V >();
    }

    public
    static
    < K, V >
    Map< K, V >
    emptyMap()
    {
        return Collections.emptyMap(); 
    }

    public
    static
    < K, V >
    Map< K, V >
    newMap( int initialCapacity )
    {
        inputs.nonnegativeI( initialCapacity, "initialCapacity" );
        return new HashMap< K, V >( initialCapacity );
    }

    public
    static
    < K, V >
    Map< K, V >
    unmodifiableMap( Map< K, V > m )
    {
        inputs.notNull( m, "m" );
        return Collections.unmodifiableMap( m );
    }

    // l may contain null elts.
    public
    static
    < V >
    List< V >
    unmodifiableCopy( List< V > l )
    {
        return unmodifiableCopy( l, "l", true );
    }

    private
    final
    static
    class ImmutableList< V, L extends List< V > & RandomAccess >
    extends AbstractList< V >
    {
        private final L list;
        private final boolean allowNullElts;

        private 
        ImmutableList( L list,
                       boolean allowNullElts ) 
        { 
            this.list = list; 
            this.allowNullElts = allowNullElts;
        }

        public int size() { return list.size(); }

        public V get( int i ) { return list.get( i ); }
    }

    private
    static
    < V >
    ImmutableList< V, ? >
    makeImmutableList( List< V > l,
                       String listName,
                       boolean allowNullElts )
    {
        ArrayList< V > l2 = new ArrayList< V >( l.size() );

        int indx = 0;

        for ( V elt : l )
        {
            if ( elt == null && ! allowNullElts )
            {
                inputs.fail( 
                    "List '" + listName + "' contains null at index", 
                    indx );
            }
            else l2.add( elt );

            ++indx;
        }

        return new ImmutableList< V, ArrayList< V > >( l2, allowNullElts );
    }

    // unmodifiableCopy is kind of a misnomer, since if l is already the result
    // of a previous call to this method, l itself is returned, not a copy.
    // TODO: rename this method something like immutableList(...)
    public
    static
    < V >
    List< V >
    unmodifiableCopy( List< V > l,
                      String listName,
                      boolean allowNullElts )
    {
        inputs.notNull( l, "l" );

        if ( l instanceof ImmutableList && 
             ( (ImmutableList) l ).allowNullElts == allowNullElts ) 
        {
            return l;
        }
        else return makeImmutableList( l, listName, allowNullElts );
    }

    public
    static
    < V >
    List< V >
    unmodifiableCopy( List< V > l,
                      String listName )
    {
        return unmodifiableCopy( l, listName, false );
    }

    // Added as a convenience in this class so callers can access this method
    // via this class instead of via java.util.Collections
    public
    static
    < V >
    List< V >
    unmodifiableList( List< V > l )
    {
        inputs.notNull( l, "l" );
        return Collections.unmodifiableList( l );
    }

    public
    static
    < V >
    Set< V >
    unmodifiableSet( Set< V > s )
    {
        inputs.notNull( s, "s" );
        return Collections.unmodifiableSet( s );
    }

    public static < V > Set< V > emptySet() { return Collections.emptySet(); }

    public
    static
    < V >
    Set< V >
    unmodifiableCopy( Set< V > s )
    {
        inputs.notNull( s, "s" );
        return unmodifiableSet( new HashSet< V >( s ) );
    }

    public
    static
    < K, V >
    Map< K, V >
    unmodifiableCopy( Map< K, V > m )
    {
        inputs.notNull( m, "m" );
        return unmodifiableMap( new HashMap< K, V >( m ) );
    }

    public
    static
    < K, V >
    Map< K, V >
    copyOf( Map< K, V > m )
    {
        inputs.notNull( m, "m" );
        return new HashMap< K, V >( m );
    }

    public
    static
    < T >
    T[]
    copyArray( T[] arr )
    {
        inputs.notNull( arr, "arr" );
        return Arrays.copyOf( arr, arr.length );
    }

    public
    static
    < K, V >
    Map< K, V >
    newMap( Class< ? extends K > keyCls,
            Class< ? extends V > valCls,
            Object... pairs )
    {
        return putAll( new HashMap< K, V >(), keyCls, valCls, pairs );
    }

    private
    final
    static
    class Casts
    extends Validator
    {
        public
        ClassCastException
        createException( CharSequence inputName,
                         CharSequence msg )
        {
            return 
                new ClassCastException( getDefaultMessage( inputName, msg ) );
        }
    }

    // Does a put only if m does not already contain key; val/key may be null if
    // m allows it
    public
    static
    < K, V >
    void
    putUnique( Map< K, V > m,
               K key,
               V val )
    {
        inputs.notNull( m, "m" );
        inputs.notNull( key, "key" );

        if ( m.containsKey( key ) )
        {
            state.fail( 
                "Map already contains value", m.get( key ), "for key", key,
                "(Attempt to set value", val + ")" );
        }
        else m.put( key, val );
    }

    public
    static
    < K, V >
    void
    putAppend( Map< K, List< V > > map,
               K key,
               V val )
    {
        inputs.notNull( map, "map" );
        inputs.notNull( key, "key" );
        inputs.notNull( val, "val" );

        List< V > appendTarg = map.get( key );

        if ( appendTarg == null )
        {
            appendTarg = newList();
            map.put( key, appendTarg );
        }

        appendTarg.add( val );
    }

    // unmodifiableDeep(List|Map)MapCopy are so named since leaving out the
    // List|Map element would lead to two methods with the same erasure; the
    // final "Map" preceding copy is to leave room for methods later that handle
    // lists of lists or lists of maps (unmodifiableDeep(List|Map)ListCopy)

    public
    static
    < K, V >
    Map< K, List< V > >
    unmodifiableDeepListMapCopy( Map< K, List< V > > map )
    {
        inputs.noneNull( map, "map" );

        Map< K, List< V > > res = newMap( map.size() );

        for ( Map.Entry< K, List< V > > e : map.entrySet() )
        {
            res.put( e.getKey(), unmodifiableCopy( e.getValue() ) );
        }

        return unmodifiableMap( res );
    }

    public
    static
    < K, L, V >
    Map< K, Map< L, V > >
    unmodifiableDeepMapMapCopy( Map< K, Map< L, V > > map )
    {
        inputs.noneNull( map, "map" );

        Map< K, Map< L, V > > res = newMap( map.size() );

        for ( Map.Entry< K, Map< L, V > > e : map.entrySet() )
        {
            res.put( e.getKey(), unmodifiableCopy( e.getValue() ) );
        }

        return unmodifiableMap( res );
    }

    private
    final
    static
    class CompletionImpl< V >
    extends AbstractCompletion< V >
    {
        private 
        CompletionImpl( V res,
                        Throwable th )
        {
            super( res, th );
        }
    }

    // res may be null in keeping consistent with other completion initializers
    public
    static
    < V >
    Completion< V >
    successCompletion( V comp )
    {
        if ( comp == null ) 
        {
            @SuppressWarnings( "unchecked" )
            Completion< V > res = (Completion< V >) COMPLETION_SUCCESS_NULL;
            return res;
        }
        else return new CompletionImpl< V >( comp, null );
    }

    public
    static
    < V >
    Completion< V >
    successCompletion()
    {
        return Lang.< V >successCompletion( null );
    }

    public
    static
    < V >
    Completion< V >
    failureCompletion( Throwable th )
    {
        return new CompletionImpl< V >( null, inputs.notNull( th, "th" ) );
    }

    // The methods which follow are support methods for converting java strings
    // to json strings. The method is included here instead of in the JSON libs
    // since JSON strings are also a very convenient, safe, and meaningful way
    // to serialize unicode strings in general

    private
    static
    void
    appendUnicodeEscape( char ch,
                         StringBuilder sb )
    {
        sb.append( "\\u" );

        String numStr = Integer.toString( ch, 16 );

        if ( ch <= '\u000F' ) sb.append( "000" );
        else if ( ch <= '\u00FF' ) sb.append( "00" );
        else if ( ch <= '\u0FFF' ) sb.append( "0" );

        sb.append( Integer.toString( ch, 16 ) );
    }

    private
    static
    void
    verifyAndAppendLowSurrogate( CharSequence cs,
                                 int indx,
                                 StringBuilder sb )
    {
        if ( indx < cs.length() )
        {
            char ch = cs.charAt( indx );

            if ( Character.isLowSurrogate( ch ) ) sb.append( ch );
            else
            {
                inputs.fail(
                    "Character at index", indx, "is not a low surrogate " +
                    "but preceding character was a high surrogate" );
            }
        }
        else
        {
            inputs.fail(
                "Unexpected end of string while expecting low surrogate" );
        }
    }

    private
    static
    void
    appendOrdinaryChar( char ch,
                        StringBuilder sb )
    {

        if ( ch == '\u0020' || ch == '\u0021' ||
             ( ch >= '\u0023' && ch <= '\u005b' ) || ch >= '\u005d' )
        {
            sb.append( ch );
        }
        else appendUnicodeEscape( ch, sb );
    }

    private
    static
    void
    appendChar( char ch,
                StringBuilder sb )
    {
        switch ( ch )
        {
            case '"': sb.append( "\\\"" ); break;
            case '\\': sb.append( "\\\\" ); break;
            case '\b': sb.append( "\\b" ); break;
            case '\f': sb.append( "\\f" ); break;
            case '\n': sb.append( "\\n" ); break;
            case '\r': sb.append( "\\r" ); break;
            case '\t': sb.append( "\\t" ); break;

            default: appendOrdinaryChar( ch, sb );
        }
    }

    public
    static
    StringBuilder
    appendRfc4627String( StringBuilder sb,
                         CharSequence str )
    {
        inputs.notNull( sb, "sb" );
        inputs.notNull( str, "str" );

        sb.append( '"' );

        for ( int i = 0, e = str.length(); i < e; )
        {
            char ch = str.charAt( i++ );

            if ( Character.isHighSurrogate( ch ) )
            {
                sb.append( ch );
                verifyAndAppendLowSurrogate( str, i++, sb );
            }
            else appendChar( ch, sb );
        }

        sb.append( '"' );

        return sb;
    }

    public
    static
    CharSequence
    getRfc4627String( CharSequence str )
    {
        return appendRfc4627String( new StringBuilder(), str );
    }

    public
    static
    CharSequence
    quoteString( CharSequence str )
    {
        return getRfc4627String( str );
    }

    public
    static
    CharSequence
    stackTraceToString( Throwable th )
    {
        inputs.notNull( th, "th" );

        String charset = "utf-16";

        try 
        { 
            ByteArrayOutputStream bos = new ByteArrayOutputStream();
            PrintStream ps = new PrintStream( bos, false, charset );
    
            th.printStackTrace( ps );
            
            return new String( bos.toByteArray(), charset );
        } 
        catch ( UnsupportedEncodingException uee )
        {
            throw state.createFail( "No character set:", charset );
        }
    }

    public
    static
    List< StackTraceElement >
    getCallerTrace( int callerOffset )
    {
        inputs.nonnegativeI( callerOffset, "callerOffset" );

        List< StackTraceElement > trace = 
            asList( Thread.currentThread().getStackTrace() );

        // remove Thread.getStackTrace() and this method itself from
        // consideration
        callerOffset += 2; 
        
        if ( callerOffset >= trace.size() ) trace = emptyList();
        else 
        {
            trace = 
                unmodifiableList( trace.subList( callerOffset, trace.size() ) );
        }

        return trace;
    }

    public
    static
    void
    setOriginTrace( OriginTraceSettable obj,
                    int callerOffset )
    {
        inputs.notNull( obj, "obj" );
        inputs.nonnegativeI( callerOffset, "callerOffset" );

        // The +1 is for this method itself
        List< StackTraceElement > trace = getCallerTrace( callerOffset + 1 );
        
        if ( ! trace.isEmpty() ) obj.setOriginStackTrace( trace );
    }

    public
    static
    boolean
    isJavaPrimitive( Class< ? > cls )
    {
        inputs.notNull( cls, "cls" );

        return
            cls.equals( Boolean.TYPE ) ||
            cls.equals( Byte.TYPE ) ||
            cls.equals( Short.TYPE ) ||
            cls.equals( Character.TYPE ) ||
            cls.equals( Integer.TYPE ) ||
            cls.equals( Long.TYPE ) ||
            cls.equals( Float.TYPE ) ||
            cls.equals( Double.TYPE ) ||
            cls.equals( Void.TYPE );
    }

    public
    final
    static
    class ImmutableListBuilder< V >
    {
        private final List< V > l = newList();

        public
        ImmutableListBuilder< V >
        add( V elt )
        {
            l.add( elt );
            return this;
        }

        public List< V > build() { return unmodifiableList( newList( l ) ); }
    }

    public
    static
    < V >
    ImmutableListBuilder< V >
    immutableListBuilder()
    {
        return new ImmutableListBuilder< V >();
    }

    public
    static
    boolean
    startsWith( CharSequence s1,
                CharSequence s2 )
    {
        inputs.notNull( s1, "s1" );
        inputs.notNull( s2, "s2" );

        if ( s2.length() > s1.length() ) return false;
        else
        {
            for ( int i = 0, e = s2.length(); i < e; ++i )
            {
                if ( s1.charAt( i ) != s2.charAt( i ) ) return false;
            }

            return true;
        }
    }
}
