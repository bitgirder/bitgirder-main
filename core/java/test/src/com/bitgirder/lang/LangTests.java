package com.bitgirder.lang;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.test.Test;

import java.util.Map;
import java.util.Collection;
import java.util.TreeMap;
import java.util.Random;
import java.util.UUID;
import java.util.LinkedList;
import java.util.List;
import java.util.Arrays;
import java.util.Iterator;

import java.io.ObjectInputStream;
import java.io.ObjectOutputStream;
import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;

@Test
public
final
class LangTests
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final Random rand = new Random();

    public
    static
    void
    code( Object... msg )
    {
        System.out.println( Strings.join( "", msg ) );
    }

    public
    static
    void
    codef( String tmpl,
           Object... args )
    {
        code( String.format( tmpl, args ) );
    }

    @Test
    private
    void
    testNewMap()
    {
        Map< String, Integer > m =
            Lang.newMap( String.class, Integer.class, 
                "one", 1, 
                "two", 2, 
                null, 3,
                "four", null );

        state.equalInt( 1, m.get( "one" ) );
        state.equalInt( 2, m.get( "two" ) );
        state.equalInt( 3, m.get( null ) );
        state.isTrue( m.get( "four" ) == null );
        state.equalInt( 4, m.size() );

        // test that returned map is still mutable
        m.put( "five", 5 );
        state.equalInt( 5, m.get( "five" ) );
        m.put( "three", 7 );
        state.equalInt( 7, m.get( "three" ) );
    }

    @Test
    private
    void
    testNewMapEmpty()
    {
        state.isTrue( Lang.newMap( String.class, Integer.class ).isEmpty() );
    }

    @Test( expected = IllegalStateException.class,
           expectedPattern = "Missing final map value" )
    private
    void
    testNewMapDetectsMissingVal()
    {
        Lang.newMap( String.class, Integer.class, "one", 1, "two" );
    }

    @Test( expected = ClassCastException.class,
           expectedPattern = "Invalid map key type: class java.lang.Boolean" )
    private
    void
    testNewMapFailsKeyCast()
    {
        Lang.newMap( String.class, Integer.class, "one", 1, true, 2 );
    }

    @Test( expected = ClassCastException.class,
           expectedPattern = "Invalid map value type: class java.lang.Boolean" )
    private
    void
    testNewMapFailsValCast()
    {
        Lang.newMap( String.class, Integer.class, "one", 1, "two", true );
    }

    private
    < V extends Comparable< V > >
    void
    assertRange( Range< V > r,
                 boolean expctInclude,
                 V... vals )
    {
        for ( V val : vals ) state.equal( expctInclude, r.includes( val ) );
    }

    @Test
    private
    void
    testRanges()
    {
        Range< Integer > r;
        
        r = Range.of( 5 );
        assertRange( r, true, 5 );
        assertRange( r, false, 4, 6 );

        r = Range.closed( 2, 6 );
        assertRange( r, true, 2, 3, 4, 5, 6 );
        assertRange( r, false, 1, 7, 10010 );

        r = Range.open( 2, 6 );
        assertRange( r, true, 3, 4, 5 );
        assertRange( r, false, 1, 2, 6, 7 );

        r = Range.openMin( 2, 6 );
        assertRange( r, true, 3, 4, 5, 6 );
        assertRange( r, false, 1, 2, 7 );

        r = Range.openMax( 2, 6 );
        assertRange( r, true, 2, 3, 4, 5 );
        assertRange( r, false, 1, 6, 7 );

        r = Range.openMin( null, 12 );
        assertRange( r, true, Integer.MIN_VALUE, 0, 11, 12 );
        assertRange( r, false, 13 );

        r = Range.open( 12, null );
        assertRange( r, true, 13, Integer.MAX_VALUE );
        assertRange( r, false, -1, 12 );

        r = Range.open();
        assertRange( r, true, Integer.MIN_VALUE, -1, 0, 1, Integer.MAX_VALUE );
    }

    @Test
    private
    void
    assertRangeCreateFailure()
    {
        try { Range.< Integer >openMax( null, 12 ); state.fail(); }
        catch ( IllegalArgumentException ok ) {}

        try { Range.< Integer >openMin( 12, null ); state.fail(); }
        catch ( IllegalArgumentException ok ) {}
        
        try { Range.< Integer >closed( 12, null ); state.fail(); }
        catch ( IllegalArgumentException ok ) {}

        try { Range.< Integer >closed( null, 12 ); state.fail(); }
        catch ( IllegalArgumentException ok ) {}
    }

    @Test( expected = IllegalArgumentException.class,
           expectedPattern = "\\Qmax < min ( 10 < 12 )\\E" )
    private
    void
    testBadRangeBounds()
    {
        Range.< Integer >closed( 12, 10 );
    }

    private
    final
    static
    < V >
    void
    assertRandomizerCounts( Map< V, ? extends Number > counts,
                            Map< V, ? extends Number > freqsExpct,
                            double tol )
    {
        inputs.notNull( counts, "counts" );
        inputs.notNull( freqsExpct, "freqsExpct" );
        inputs.isTrue( tol >= 0d );

        inputs.isTrue( 
            counts.keySet().containsAll( freqsExpct.keySet() ) &&
            counts.size() == freqsExpct.size(),
            "Map key sets are not equal" );

        double sum = 0d;
        for ( Number n : counts.values() ) sum += n.doubleValue();

        for ( Map.Entry< V, ? extends Number > e : freqsExpct.entrySet() )
        {
            double freqExpct = e.getValue().doubleValue();
            double freqActual = counts.get( e.getKey() ).doubleValue() / sum;

            state.inRange( freqActual, "freqActual", 
                Range.closed( freqExpct - tol, freqExpct + tol ) );
        }
    }

    // Note about the tolerance used: no idea how it corresponds to what should
    // be expected from the implementation of java.util.Random, but it seems
    // pretty reasonable in light of the fact that what we're really after is
    // testing that the randomizer is behaving reasonably well and not going
    // totally haywire (returning a fixed value, values in the wrong order,
    // etc).
    @Test
    private
    void
    testRandomizer()
    {
        // use tree map to preserve the ordering of assignment by the
        // randomizer, so we can test forFloat()
        Map< String, Float > freqs = 
            Lang.putAll( 
                new TreeMap< String, Float >(), String.class, Float.class, 
                "1/7", 0.5f, "2/7", 1.0f, "4/7", 2.0f );

        Map< String, Float > freqsExpct =
            Lang.newMap( String.class, Float.class,
                "1/7", 1f/7f, "2/7", 2f/7f, "4/7", 4f/7f );

        Map< String, Integer > counts = 
            Lang.newMap( String.class, Integer.class,
                "1/7", 0, "2/7", 0, "4/7", 0 );

        Randomizer< String > r = Randomizer.create( freqs, new Random() );

        state.equal( "1/7", r.forFloat( 0f ) );
        state.equal( "1/7", r.forFloat( 1f/7f - 0.0001f ) );
        state.equal( "2/7", r.forFloat( 1f/7f ) );
        state.equal( "2/7", r.forFloat( 1f/7f + 0.0001f ) );
        state.equal( "4/7", r.forFloat( 0.99999999f ) );

        for ( int i = 0; i < 700000; ++i ) 
        {
            String next = r.next();
            counts.put( next, counts.get( next ) + 1 );
        }

        assertRandomizerCounts( counts, freqsExpct, 0.01 );
    }

    @Test
    private
    void
    testSingleElementRandomizer()
    {
        Map< String, Float > freqs =
            Lang.newMap( String.class, Float.class, "the val", 17f );
 
        Randomizer< String > r = Randomizer.create( freqs, new Random() );

        for ( int i = 0; i < 10000; ++i ) state.equal( r.next(), "the val" );
    }

    private
    final
    static
    class TypedStringImpl
    extends TypedString< TypedStringImpl >
    {
        private TypedStringImpl() { super( UUID.randomUUID().toString() ); } 
        private TypedStringImpl( String foo ) { super( foo, "foo" ); }
    }

    private
    final
    static
    class CaseInsensitiveTypedStringImpl
    extends CaseInsensitiveTypedString< CaseInsensitiveTypedStringImpl >
    {
        private
        CaseInsensitiveTypedStringImpl( CharSequence cs )
        {
            super( cs );
        }
    }

    // For use in testing impl of TypedString.equals
    private
    final
    static
    class TypedStringOtherImpl
    extends TypedString< TypedStringOtherImpl >
    {
        private TypedStringOtherImpl( String foo ) { super( foo, "foo" ); }
    }

    @Test
    private
    void
    testTypedStringBasic()
    {
        new TypedStringImpl(); // just to cover default superclass constructor

        TypedStringImpl s = new TypedStringImpl( "hello" );
        state.equal( "hello", s.toString() );
        state.equal( "\"hello\"", s.inspect() );
        state.isTrue( 'h' == s.charAt( 0 ) );
        state.isTrue( 'o' == s.charAt( 4 ) );
        state.isTrue( 'e' == s.charAt( 1 ) );
        state.equalInt( 5, s.length() );
        state.equal( "ell", s.subSequence( 1, 4 ).toString() );
    }

    @Test( expectedPattern = "Input 'foo' cannot be null",
           expected = IllegalArgumentException.class )
    private
    void
    testTypedStringInputParamUsage()
    {
        new TypedStringImpl( null );
    }

    @Test( expected = NullPointerException.class )
    private
    void
    testAbstractTypedStringCompareToOnNullInput()
    {
        new TypedStringImpl( "foo" ).compareTo( null );
    }

    @Test
    private
    void
    testTypedStringComparisons()
    {
        TypedStringImpl foo1 = new TypedStringImpl( "foo" );
        TypedStringImpl foo2 = new TypedStringImpl( "foo" );
        TypedStringImpl bar1 = new TypedStringImpl( "bar" );

        state.isTrue( foo1.equals( foo2 ) );
        state.isTrue( foo1.equals( foo1 ) );
        state.isTrue( foo1.compareTo( foo2 ) == 0 );
        state.equalInt( foo1.hashCode(), foo2.hashCode() );
        state.isFalse( foo1.equals( null ) );
        state.isFalse( foo1.equals( bar1 ) );
        state.isTrue( foo1.compareTo( bar1 ) > 0 );
        state.isTrue( bar1.compareTo( foo1 ) < 0 );

        TypedStringOtherImpl otherFoo = new TypedStringOtherImpl( "foo" );
        state.isTrue( foo1.toString().equals( otherFoo.toString() ) );
        state.isFalse( foo1.equals( otherFoo ) );
    }

    @Test
    private
    void
    testCaseInsensitiveTypedString()
    {
        CaseInsensitiveTypedStringImpl s1 = 
            new CaseInsensitiveTypedStringImpl( "HellO" );
        state.isTrue( s1.toString().equals( "HellO" ) );

        CaseInsensitiveTypedStringImpl s2 = 
            new CaseInsensitiveTypedStringImpl( "hello" );

        CaseInsensitiveTypedStringImpl s3 =
            new CaseInsensitiveTypedStringImpl( "goodbye" );

        state.isFalse( s1.equals( s3 ) );
        state.isTrue( s1.compareTo( s3 ) > 0 );
        state.isTrue( s3.compareTo( s1 ) < 0 );
        state.isTrue( s1.compareTo( s2 ) == 0 );
        state.equal( s1, s2 );
        state.isFalse( s1.toString().equals( s2.toString() ) );
    }

    private
    final
    static
    class TypedLongImpl
    extends TypedLong< TypedLongImpl >
    {
        private TypedLongImpl( long l ) { super( l ); }
    }

    @Test
    private
    void
    testTypedLong()
    {
        TypedLongImpl l1 = new TypedLongImpl( 1 );
        TypedLongImpl l2 = new TypedLongImpl( 2 );
        TypedLongImpl l3 = new TypedLongImpl( 2 );

        state.equal( 1L, l1.longValue() );
        state.isTrue( l1.equals( l1 ) );
        state.isFalse( l1.equals( l2 ) );
        state.isFalse( l1.equals( null ) );
        state.isTrue( l2.equals( l3 ) );

        state.isTrue( l1.compareTo( l2 ) < 0 );
        state.isTrue( l2.compareTo( l1 ) > 0 );
        state.isTrue( l2.compareTo( l3 ) == 0 );

        state.equal( 
            Long.MIN_VALUE, new TypedLongImpl( Long.MIN_VALUE ).longValue() );
        
        state.equal(
            Long.MAX_VALUE, new TypedLongImpl( Long.MAX_VALUE ).longValue() );
    }

    private
    final
    static
    class TypedDoubleImpl
    extends TypedDouble< TypedDoubleImpl >
    {
        private TypedDoubleImpl( double d ) { super( d ); }
    }

    @Test
    private
    void
    testTypedDouble()
    {
        TypedDoubleImpl d1 = new TypedDoubleImpl( 1.0d );
        TypedDoubleImpl d2 = new TypedDoubleImpl( 2.0d );
        TypedDoubleImpl d3 = new TypedDoubleImpl( 2.0d );

        state.equal( 
            Double.valueOf( 1.0d ), Double.valueOf( d1.doubleValue() ) );

        state.isTrue( d1.equals( d1 ) );
        state.isFalse( d1.equals( d2 ) );
        state.isFalse( d1.equals( null ) );
        state.isTrue( d2.equals( d3 ) );

        state.isTrue( d1.compareTo( d2 ) < 0 );
        state.isTrue( d2.compareTo( d1 ) > 0 );
        state.isTrue( d2.compareTo( d3 ) == 0 );
    }

    // Does a unique put and then one that should fail
    @Test( expected = IllegalStateException.class,
           expectedPattern = "Map already contains value 3 for key.*" )
    private
    void
    testPutUnique()
        throws Exception
    {
        Map< String, Integer > m =
            Lang.newMap( String.class, Integer.class,
                "one", 1, "two", 2 );
        
        Lang.putUnique( m, "three", 3 );
        state.equalInt( 1, m.get( "one" ) );
        state.equalInt( 3, m.get( "three" ) );

        Lang.putUnique( m, "three", -3 ); 
    }

    @Test( expected = IllegalStateException.class,
           expectedPattern = "Map already contains value null for key.*" )
    private
    void
    testPutUniqueWithNullVals()
        throws Exception
    {
        Map< String, Integer > m = Lang.newMap();

        Lang.putUnique( m, "Hello", null );
        state.isTrue( m.containsKey( "Hello" ) );

        Lang.putUnique( m, "Hello", null );
    }

    @Test
    private
    void
    testUnmodifiableCopyMap()
    {
        Map< String, Integer > orig =
            Lang.newMap( String.class, Integer.class,
                "one", 1, "two", 2 );
        
        Map< String, Integer > copy = Lang.unmodifiableCopy( orig );
        state.equal( orig, copy );

        // We currently rely on the implementation detail that the returned map
        // throws UnsupportedOperationException on put. That's not necessarily
        // part of the public contract of Lang.unmodifiableCopy (though it could
        // be if we do the work to make it so), but is what we expect for
        // testing.
        try 
        { 
            copy.put( "three", 3 ); 
            state.fail( "Put succeeded" );
        }
        catch ( UnsupportedOperationException okay ) {}
    }

    // regression for bug which allowed an immutable list with null elts to be
    // returned from a call to unmodifiableList with allowNullElts set to false
    @Test( expected = IllegalArgumentException.class,
           expectedPattern = "^List 'l2' contains null at index 1$" )
    private
    void
    testImmutableWithoutNullsRejectsImmutableWithNulls()
    {
        List< String > l1 = 
            Lang.unmodifiableCopy( 
                Arrays.< String >asList( "hi", null, "there" ), "init", true );
 
        List< String > l2 = Lang.unmodifiableCopy( l1, "l2", false );
    }

    @Test
    private
    void
    testRfc4627Strings()
        throws Exception
    {
        state.equalString(
            "\"\\n\\t\\f\\r\\b\\\\\\\"\\u0003\"",
            Lang.getRfc4627String( "\n\t\f\r\b\\\"\u0003" ) );

        state.equalString(
            "\"a gclef: \ud834\uDD1e\"",
            Lang.getRfc4627String( "a gclef: \ud834\uDD1e" ) );
        
        // just cover the quoteString() frontend as well
        state.equalString( "\"a\"", Lang.quoteString( "a" ) );
    }

    @Test
    private
    void
    testStackTraceToString()
    {
        String msg1 = UUID.randomUUID().toString();
        String msg2 = UUID.randomUUID().toString();

        Exception ex1 = new Exception( msg1 );
        Exception ex2 = new Exception( msg2, ex1 );

        String trace = Lang.stackTraceToString( ex2 ).toString();

        state.isTrue( trace.indexOf( msg1 ) > 0 );
        state.isTrue( trace.indexOf( msg2 ) > 0 );

        String here = getClass().getName() + ".testStackTraceToString";
        state.isTrue( trace.indexOf( here ) > 0 );
    }

    private
    String
    nextPropName()
    {
        return "test-property-" + UUID.randomUUID();
    }

    private
    void
    assertBooleanSystemProperty( String val,
                                 boolean expct )
    {
        String name = nextPropName();

        System.setProperty( name, val );

        state.equal( expct, inputs.hasBooleanSystemProperty( name ) );
    }

    @Test
    private
    void
    testHasBooleanSystemPropertySuccess()
    {
        assertBooleanSystemProperty( "yes", true );
        assertBooleanSystemProperty( "true", true );
        assertBooleanSystemProperty( "no", false );
        assertBooleanSystemProperty( "false", false );
        
        state.isFalse( inputs.hasBooleanSystemProperty( nextPropName() ) );
    }

    @Test( expected = IllegalArgumentException.class,
           expectedPattern = 
            "Invalid property value for property " +
            "com\\.bitgirder\\.validation\\.Inputs\\.malformedBooleanProperty" +
            ": fffff" )
    private
    void
    testHasBooleanSystemPropertyParseFailure()
    {
        String propName =
            "com.bitgirder.validation.Inputs.malformedBooleanProperty";

        System.setProperty( propName, "fffff" );
        inputs.hasBooleanSystemProperty( propName );
    }

    @Test
    private
    void
    testAddAllCoverage()
    {
        final List< String > it1 = Lang.asList( "hello", "there" );

        Collection< String > c = Lang.newList();

        // test when Iterable is a Collection
        Lang.addAll( c, it1 );
        state.equalString( "hello|there", Strings.join( "|", c ) );

        c = Lang.newList();

        Iterable< String > it2 = 
            new Iterable< String >() {
                public Iterator< String > iterator() { return it1.iterator(); }
            };

        // test when it is just an Iterable
        Lang.addAll( c, it2 );
        state.equalString( "hello|there", Strings.join( "|", c ) );
    }

    private
    void
    assertImmutable( List< String > l )
    {
        try
        {
            l.add( "shouldn't be added" );
            state.fail( "Was able to add an element" );
        }
        catch ( UnsupportedOperationException okay ) {}
    }

    @Test
    private
    void
    testImmutableList()
    {
        Lang.ImmutableListBuilder< String > b = Lang.immutableListBuilder();

        List< String > l = b.build();
        state.isTrue( l.isEmpty() );
        assertImmutable( l );

        l = b.add( "s1" ).add( "s2" ).build();
        state.equalString( "s1|s2", Strings.join( "|", l ) );
        assertImmutable( l );
    }

    private
    void
    assertStartsWith( CharSequence s1,
                      CharSequence s2,
                      boolean expct )
    {
        state.equal( expct, Lang.startsWith( s1, s2 ) );
    }

    @Test
    private
    void
    testStartsWith()
    {
        assertStartsWith( "", "", true );
        assertStartsWith( "a", "", true );
        assertStartsWith( "a", "a", true );
        assertStartsWith( "abc", "a", true );
        assertStartsWith( "abc", "ab", true );
        assertStartsWith( "abc", "abc", true );
        assertStartsWith( "", "a", false );
        assertStartsWith( "a", "ab", false );
        assertStartsWith( "abc", "abd", false );
    }

    @Test
    private
    void
    testTypedStringSerializable()
        throws Exception
    {
        TypedStringImpl t1 = new TypedStringImpl( "hello" );

        ByteArrayOutputStream bos = new ByteArrayOutputStream();
        ObjectOutputStream oos = new ObjectOutputStream( bos );
        oos.writeObject( t1 );

        ByteArrayInputStream bis = 
            new ByteArrayInputStream( bos.toByteArray() );
        
        ObjectInputStream ois = new ObjectInputStream( bis );

        TypedStringImpl t2 = (TypedStringImpl) ois.readObject();
        state.equalString( t1, t2 );
    }

    @Test
    private
    void
    testOctetToUnsignedByteConversions()
    {
        state.equalInt( 0, Lang.asOctet( (byte) 0 ) );
        state.equalInt( 1, Lang.asOctet( (byte) 1 ) );
        state.equalInt( 127, Lang.asOctet( (byte) 127 ) );
        state.equalInt( 128, Lang.asOctet( (byte) -128 ) );
        state.equalInt( 129, Lang.asOctet( (byte) -127 ) );
        state.equalInt( 255, Lang.asOctet( (byte) -1 ) );

        state.isTrue( (byte) 0 == Lang.fromOctet( 0 ) );
        state.isTrue( (byte) 1 == Lang.fromOctet( 1 ) );
        state.isTrue( (byte) 2 == Lang.fromOctet( 2 ) );
        state.isTrue( (byte) 127 == Lang.fromOctet( 127 ) );
        state.isTrue( (byte) -128 == Lang.fromOctet( 128 ) );
        state.isTrue( (byte) -127 == Lang.fromOctet( 129 ) );
        state.isTrue( (byte) -1 == Lang.fromOctet( 255 ) );
    }

    private
    void
    assertUintString( String s1,
                      Number expct )
    {
        Number n;

        if ( expct instanceof Long ) n = Lang.parseUint64( s1 );
        else n = Lang.parseUint32( s1 );

        state.equal( expct, n );

        String s2;
        if ( expct instanceof Long ) s2 = Lang.toUint64String( (Long) n );
        else s2 = Lang.toUint32String( (Integer) n );

        state.equal( s1, s2 );
    }

    private
    void
    assertUintParseFail( int sz,
                         String in,
                         String tmpl )
    {
        state.isTrue( sz == 32 || sz == 64 );

        try
        {
            if ( sz == 32 ) Lang.parseUint32( in ); else Lang.parseUint64( in );
            state.fail( "Parsed:", in );
        }
        catch ( NumberFormatException nfe )
        {
            state.equal( nfe.getMessage(), String.format( tmpl, in ) );
        }
    }

    private
    void
    assertUintComp( int sign,
                    Number n1,
                    Number n2 )
    {
        int res;

        if ( n1 instanceof Long ) 
        {
            res = Lang.compareUint64( (Long) n1, (Long) n2 );
        }
        else res = Lang.compareUint32( (Integer) n1, (Integer) n2 );
 
        state.equalInt( sign, res );
    }

    @Test
    private
    void
    testUint32()
    {
        assertUintComp( -1, 0, 1 );
        assertUintComp( -1, 1, -1 );
        assertUintComp( -1, 1, Integer.MIN_VALUE );
        assertUintComp( 0, 0, 0 );
        assertUintComp( 0, 1, 1 );
        assertUintComp( 0, -1, -1 );
        assertUintComp( 1, -1, 1 );
        assertUintComp( 1, 2, 1 );
        assertUintComp( 1, -1, Integer.MIN_VALUE );
   
        assertUintString( "0", 0 );
        assertUintString( "2147483648", Integer.MIN_VALUE );
        assertUintString( "4294967295", -1 );

        assertUintParseFail( 32, "-1", "Number is negative: %s" );

        assertUintParseFail( 
            32, "4294967296", "Number is too large for uint32: %s" );

        assertUintParseFail( 32, "", "Empty number" );
    }

    @Test
    private
    void
    testUint64()
    {
        assertUintComp( -1, 0L, 1L );
        assertUintComp( -1, 1L, -1L );
        assertUintComp( -1, 1L, Long.MIN_VALUE );
        assertUintComp( 0, 0L, 0L );
        assertUintComp( 0, 1L, 1L );
        assertUintComp( 0, -1L, -1L );
        assertUintComp( 1, -1L, 1L );
        assertUintComp( 1, 2L, 1L );
        assertUintComp( 1, -1L, Long.MIN_VALUE );
   
        assertUintString( "0", 0L );
        assertUintString( "9223372036854775808", Long.MIN_VALUE );
        assertUintString( "18446744073709551615", -1L );

        assertUintParseFail( 64, "-1", "Number is negative: %s" );

        assertUintParseFail( 
            64, "18446744073709551616", "Number is too large for uint64: %s" );

        assertUintParseFail( 64, "", "Empty number" );
    }
}
