package com.bitgirder.mingle;

import static com.bitgirder.mingle.Mingle.*;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.PathWiseAsserter;

import com.bitgirder.test.Test;

import java.util.Iterator;

@Test
final
class MingleTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }
 
    private final static MingleTimestamp TEST_TIMESTAMP1;
 
    private final static MingleTimestamp TEST_TIMESTAMP2;

    private final static String TEST_TIMESTAMP1_STRING =
        "2007-08-24T13:15:43.123450000-08:00";

    private final static String TEST_TIMESTAMP2_STRING =
        "2007-08-24T13:15:43.000000000-08:00";

    private
    static
    MingleIdentifier
    id( CharSequence s )
    {
        return MingleIdentifier.create( s );
    }

    private
    static
    AtomicTypeReference
    at( CharSequence s )
    {
        MingleTypeReference t = MingleTypeReference.create( s );
        return state.cast( AtomicTypeReference.class, t );
    }

    private
    void
    assertValueClassFor( MingleTypeReference typ,
                         Class< ? extends MingleValue > expct )
    {
        state.equal( Mingle.valueClassFor( typ ), expct );
    }

    @Test
    private
    void
    testJavaClassFor()
    {
        assertValueClassFor( Mingle.TYPE_BOOLEAN, MingleBoolean.class );
        assertValueClassFor( Mingle.TYPE_INT32, MingleInt32.class );
        assertValueClassFor( Mingle.TYPE_INT64, MingleInt64.class );
        assertValueClassFor( Mingle.TYPE_UINT32, MingleUint32.class );
        assertValueClassFor( Mingle.TYPE_UINT64, MingleUint64.class );
        assertValueClassFor( Mingle.TYPE_FLOAT32, MingleFloat32.class );
        assertValueClassFor( Mingle.TYPE_FLOAT64, MingleFloat64.class );
        assertValueClassFor( Mingle.TYPE_STRING, MingleString.class );
        assertValueClassFor( Mingle.TYPE_BUFFER, MingleBuffer.class );
        assertValueClassFor( Mingle.TYPE_TIMESTAMP, MingleTimestamp.class );
        assertValueClassFor( Mingle.TYPE_VALUE, MingleValue.class );
        assertValueClassFor( Mingle.TYPE_NULL, MingleNull.class );
        assertValueClassFor( Mingle.TYPE_SYMBOL_MAP, MingleSymbolMap.class );

        assertValueClassFor(
            new AtomicTypeReference( new DeclaredTypeName( "Blah" ), null ),
            null
        );
    }

    private
    void
    assertInspection( MingleValue mv,
                      CharSequence... possibles )
    {
        String s = Mingle.inspect( mv ).toString();

        for ( CharSequence possible : possibles )
        {
            if ( s.equals( possible.toString() ) ) return;
        }

        state.failf( "Unmatched inspection: %s", Lang.quoteString( s ) );
    }

    @Test
    private
    void
    testInspection()
    {
        assertInspection( MingleBoolean.TRUE, "true" );
        assertInspection( MingleBoolean.FALSE, "false" );
        assertInspection( new MingleInt32( 1 ), "1" );
        assertInspection( new MingleUint32( 2 ), "2" );
        assertInspection( new MingleInt64( -1 ), "-1" );
        assertInspection( new MingleUint64( 1 ), "1" );
        assertInspection( new MingleFloat32( 1.1f ), "1.1" );
        assertInspection( new MingleFloat64( 1.1 ), "1.1" );
        assertInspection( new MingleString( "" ), "\"\"" );
        assertInspection( new MingleString( "abc\t\rd" ), "\"abc\\t\\rd\"" );
        assertInspection( TEST_TIMESTAMP1, TEST_TIMESTAMP1_STRING );

        assertInspection( new MingleBuffer( new byte[] {} ), "buffer:[]" );
        assertInspection( 
            new MingleBuffer( new byte[] { (byte) 0, (byte) 1 } ), 
            "buffer:[0001]" 
        );

        assertInspection( MingleNull.getInstance(), "null" );

        assertInspection( MingleList.empty(), "[]" );

        assertInspection( MingleList.asList( new MingleInt32( 1 ) ), "[1]" );

        assertInspection( 
            MingleList.asList(
                MingleNull.getInstance(),
                new MingleString( "s" ),
                MingleList.asList( 
                    new MingleInt32( 1 ),
                    new MingleFloat32( 1.1f )
                )
            ),
            "[null, \"s\", [1, 1.1]]"
        );

        assertInspection( MingleSymbolMap.empty(), "{}" );

        assertInspection(
            new MingleSymbolMap.Builder().
                setInt32( "id1", 1 ).
                set( "id2", MingleList.asList( new MingleInt32( 1 ) ) ).
                build(),
            "{id1:1, id2:[1]}", "{id2:[1], id1:1}"
        );

        assertInspection(
            new MingleEnum( at( "ns1@v1/E1" ), id( "val1" ) ),
            "ns1@v1/E1.val1"
        );

        assertInspection(
            new MingleStruct.Builder().setType( "ns1@v1/S1" ).build(),
            "ns1@v1/S1{}"
        );

        assertInspection(
            new MingleStruct.Builder().
                setType( "ns1@v1/S1" ).
                set( "id1", new MingleUint32( 1 ) ).
                build(),
            "ns1@v1/S1{id1:1}"
        );
    }
    
    static
    {
        // base builder for timestamps 1 and 2;
        MingleTimestamp.Builder b =
            new MingleTimestamp.Builder().
                setYear( 2007 ).
                setMonth( 8 ).
                setDate( 24 ).
                setHour( 13 ).
                setMinute( 15 ).
                setSeconds( 43 ).
                setTimeZone( "GMT-08:00" );
        
        // build ts2 with no frac part
        TEST_TIMESTAMP2 = b.build();

        // build ts1 (which came first in our code) with the frac part
        b.setFraction( "12345" );
        TEST_TIMESTAMP1 = b.build();
    }

    private
    void
    assertFormat( MingleIdentifier id,
                  CharSequence expct,
                  MingleIdentifierFormat fmt )
    {
        state.equalString( expct, id.format( fmt ) );
    }

    @Test
    private
    void
    testIdentifierFormatters()
    {
        MingleIdentifier id = 
            new MingleIdentifier( new String[] { "test", "ident" } );

        assertFormat( id, "test-ident", MingleIdentifierFormat.LC_HYPHENATED );
        assertFormat( id, "test_ident", MingleIdentifierFormat.LC_UNDERSCORE );
        assertFormat( id, "testIdent", MingleIdentifierFormat.LC_CAMEL_CAPPED );
    }

    private
    void
    assertFormat( CharSequence expct,
                  ObjectPath< MingleIdentifier > p )
    {
        state.equalString( expct, Mingle.formatIdPath( p ) );
    }

    @Test
    private
    void
    testIdPathFormat()
    {
        ObjectPath< MingleIdentifier > p = ObjectPath.getRoot();

        assertFormat( "a", p = p.descend( id( "a" ) ) );
        assertFormat( "a.b", p = p.descend( id( "b" ) ) );
        assertFormat( "a.b[ 3 ]", p = p.startImmutableList( 3 ) );
        assertFormat( "a.b[ 3 ].c", p = p.descend( id( "c" ) ) );
    }

    private
    static
    void
    assertEqualMap( MingleValue mv1,
                    MingleValue mv2,
                    PathWiseAsserter< MingleIdentifier > a )
    {
        MingleSymbolMap mp1 = a.cast( MingleSymbolMap.class, mv1 );
        MingleSymbolMap mp2 = a.cast( MingleSymbolMap.class, mv2 );

        a.equal( mp1.getKeySet(), mp2.getKeySet() );

        for ( MingleIdentifier fld : mp1.getFields() )
        {
            assertEqual( mp1.get( fld ), mp2.get( fld ), a.descend( fld ) );
        }
    }

    private
    static
    void
    assertEqualStruct( MingleValue mv1,
                       MingleValue mv2,
                       PathWiseAsserter< MingleIdentifier > a )
    {
        MingleStruct ms1 = a.cast( MingleStruct.class, mv1 );
        MingleStruct ms2 = a.cast( MingleStruct.class, mv2 );

        a.equal( ms1.getType(), ms2.getType() );
        assertEqualMap( ms1.getFields(), ms2.getFields(), a );
    }

    private
    static
    void
    assertEqualEnum( MingleValue mv1,
                     MingleValue mv2,
                     PathWiseAsserter< MingleIdentifier > a )
    {
        MingleEnum e1 = a.cast( MingleEnum.class, mv1 );
        MingleEnum e2 = a.cast( MingleEnum.class, mv2 );

        a.equal( e1.getType(), e2.getType() );
        a.equal( e1.getValue(), e2.getValue() );
    }

    private
    static
    void
    assertEqualList( MingleValue mv1,
                     MingleValue mv2,
                     PathWiseAsserter< MingleIdentifier > a )
    {
        MingleList l1 = a.cast( MingleList.class, mv1 );
        MingleList l2 = a.cast( MingleList.class, mv2 );

        Iterator< MingleValue > it1 = l1.iterator();
        Iterator< MingleValue > it2 = l2.iterator();

        PathWiseAsserter< MingleIdentifier > la = a.startImmutableList();

        while ( it1.hasNext() )
        {
            la.isTrue( it2.hasNext(), "list lengths differ" );

            assertEqual( it1.next(), it2.next(), la );
            la = la.next();
        }

        la.isFalse( it2.hasNext(), "list lengths differ" );
    }

    public
    static
    void
    assertEqual( MingleValue mv1,
                 MingleValue mv2,
                 PathWiseAsserter< MingleIdentifier > a )
    {
        inputs.notNull( a, "a" );
        
        if ( ! a.sameNullity( mv1, mv2 ) ) return;

        if ( mv1 instanceof MingleStruct ) assertEqualStruct( mv1, mv2, a );
        else if ( mv1 instanceof MingleEnum ) assertEqualEnum( mv1, mv2, a );
        else if ( mv1 instanceof MingleSymbolMap ) 
        {
            assertEqualMap( mv1, mv2, a );
        }
        else if ( mv1 instanceof MingleList ) assertEqualList( mv1, mv2, a );
        else a.equal( mv1, mv2 );
    }

    public
    static
    void
    assertEqual( MingleValue mv1,
                 MingleValue mv2 )
    {
        PathWiseAsserter< MingleIdentifier > a =
            new PathWiseAsserter< MingleIdentifier >(
                ObjectPath.< MingleIdentifier >getRoot(),
                Mingle.getIdPathFormatter()
            );
        
        assertEqual( mv1, mv2, a );
    }

    @Test
    private
    void
    testIsNumberType()
    {
        for ( MingleTypeReference t : new MingleTypeReference[] {
                TYPE_INT32, TYPE_INT64, TYPE_UINT32, TYPE_UINT64,
                TYPE_FLOAT32, TYPE_FLOAT64 } )
        {
            state.isTrue( Mingle.isNumberType( t ) );

            boolean dec = t == TYPE_FLOAT32 || t == TYPE_FLOAT64;
            state.equal( dec, Mingle.isDecimalType( t ) );
            state.equal( ! dec, Mingle.isIntegralType( t ) );
        }

        MingleRangeRestriction dummy = 
            MingleRangeRestriction.
                create( false, null, null, false, MingleInt32.class );

        state.isTrue(
            Mingle.isIntegralType( 
                new AtomicTypeReference( QNAME_INT32, dummy ) ) );
        
        state.isTrue(
            Mingle.isDecimalType(
                new AtomicTypeReference( QNAME_FLOAT32, dummy ) ) );
    }

    private
    void
    assertTypeNameIn( QualifiedTypeName qn,
                      MingleTypeReference ref )
    {
        state.equal( qn, Mingle.typeNameIn( ref ) );
    }

    @Test
    private
    void
    testTypeNameIn()
    {
        QualifiedTypeName qn = QNAME_STRING;
        MingleTypeReference ref = TYPE_STRING;
        
        assertTypeNameIn( qn, ref );
        assertTypeNameIn( qn, ref = new NullableTypeReference( ref ) );
        assertTypeNameIn( qn, ref = new NullableTypeReference( ref ) );
        assertTypeNameIn( qn, ref = new ListTypeReference( ref, true ) );
        assertTypeNameIn( qn, ref = new ListTypeReference( ref, false ) );
        assertTypeNameIn( qn, ref = new NullableTypeReference( ref ) );
        assertTypeNameIn( qn, ref = new ListTypeReference( ref, false ) );
    }
}
