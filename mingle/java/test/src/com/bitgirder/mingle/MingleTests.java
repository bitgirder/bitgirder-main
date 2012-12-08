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
 
    private final static long TS1_SECS = 1187990143;
    private final static int TS1_NS = 123450000;
 
    private final static MingleTimestamp TS1 =
        MingleTimestamp.fromUnixNanos( TS1_SECS, TS1_NS );
 
    private final static String TS1_RFC3339 =
        "2007-08-24T21:15:43.123450000Z";

//    private final static String TS2_STRING =
//        "2007-08-24T13:15:43.000000000-08:00";

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
        assertInspection( TS1, TS1_RFC3339 );

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

    @Test
    private
    void
    testTimestampRfc3339()
        throws Exception
    {
        state.equalString( TS1_RFC3339, TS1.getRfc3339String() );
        state.equal( TS1, MingleTimestamp.parse( TS1_RFC3339 ) );
    }

    @Test
    private
    void
    testTimestampFromMillis()
    {
        long ms = System.currentTimeMillis();
        MingleTimestamp ts = MingleTimestamp.fromMillis( ms );

        state.equal( ms / 1000L, ts.seconds() );
        state.equalInt( ( (int) ( ms % 1000 ) ) * 1000000, ts.nanos() );
        state.equal( ms, ts.getTimeInMillis() );
    }

    private
    void
    assertTsComp( int expct,
                  int secs1,
                  int ns1,
                  int secs2,
                  int ns2 )
    {
        MingleTimestamp t1 = MingleTimestamp.fromUnixNanos( (long) secs1, ns1 );
        MingleTimestamp t2 = MingleTimestamp.fromUnixNanos( (long) secs2, ns2 );

        state.equalInt( expct, t1.compareTo( t2 ) );
        state.equalInt( -expct, t2.compareTo( t1 ) );
    }

    @Test
    private
    void
    testTimestampCompare()
    {
        assertTsComp( -1, 1, 0, 1, 1 );
        assertTsComp( -1, 1, 0, 2, 0 );
        assertTsComp( 0, 1, 1, 1, 1 );
        assertTsComp( 0, -1, 0, -1, 0 );
        assertTsComp( 1, -1, 1, -1, 2 );
        assertTsComp( 1, -1, 0, -2, 0 );
    }
}
