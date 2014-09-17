package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import static com.bitgirder.log.CodeLoggers.Statics.*;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ObjectPaths;
import com.bitgirder.lang.path.PathWiseAsserter;

import java.util.Iterator;

public
final
class MingleTestMethods
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private MingleTestMethods() {}

    public
    static
    MingleIdentifier
    id( CharSequence s )
    {
        inputs.notNull( s, "s" );
        return MingleIdentifier.create( s );
    }

    public
    static
    QualifiedTypeName
    qname( CharSequence s )
    {
        inputs.notNull( s, "s" );
        return QualifiedTypeName.create( s );
    }

    public
    static
    DeclaredTypeName
    declaredName( CharSequence s )
    {
        inputs.notNull( s, "s" );
        return DeclaredTypeName.create( s );
    }

    public
    static
    AtomicTypeReference
    atomic( QualifiedTypeName name )
    {
        inputs.notNull( name, "name" );
        return new AtomicTypeReference( name, null );
    }

    public
    static
    AtomicTypeReference
    atomic( CharSequence name )
    {
        inputs.notNull( name, "name" );
        return atomic( qname( name ) );
    }

    public
    static
    AtomicTypeReference
    atomic( QualifiedTypeName name,
            MingleValueRestriction restriction )
    {
        inputs.notNull( name, "name" );
        inputs.notNull( restriction, "restriction" );

        return new AtomicTypeReference( name, restriction );
    }

    public
    static
    ListTypeReference
    listType( MingleTypeReference typ,
              boolean allowsEmpty )
    {
        inputs.notNull( typ, "typ" );
        return new ListTypeReference( typ, allowsEmpty );
    }

    public
    static
    NullableTypeReference
    nullableType( MingleTypeReference typ )
    {
        inputs.notNull( typ, "typ" );
        return new NullableTypeReference( typ );
    }

    public
    static
    PointerTypeReference
    ptrType( MingleTypeReference typ )
    {
        inputs.notNull( typ, "typ" );
        return new PointerTypeReference( typ );
    }

    public
    static
    ObjectPath< MingleIdentifier >
    idPathRoot( MingleIdentifier id )
    {
        inputs.notNull( id, "id" );
        return ObjectPath.getRoot( id );
    }

    public
    static
    ObjectPath< MingleIdentifier >
    idPathRoot( CharSequence id )
    {
        inputs.notNull( id, "id" );
        return idPathRoot( MingleIdentifier.create( id ) );
    }

    public
    static
    MingleList
    emptyList( ListTypeReference typ )
    {
        return MingleList.createLive( typ, Lang.< MingleValue >emptyList() );
    }

    public
    static
    MingleList
    emptyList()
    {
        return emptyList( Mingle.TYPE_OPAQUE_LIST );
    }

    public
    static
    < V >
    V
    mapGet( MingleSymbolMap m,
            MingleIdentifier id,
            Class< V > cls )
    {
        inputs.notNull( m, "m" );
        inputs.notNull( id, "id" );
        inputs.notNull( cls, "cls" );

        MingleValue mv = m.get( id );
        if ( mv == null || mv instanceof MingleNull ) return null;

        if ( cls.equals( String.class ) ) {
            return cls.cast( ( (MingleString) mv ).toString() );
        } else if ( cls.equals( Integer.class ) ) {
            return cls.cast( ( (MingleNumber) mv ).intValue() );
        } else if ( cls.equals( Boolean.class ) ) {
            return cls.cast( ( (MingleBoolean) mv ).booleanValue() );
        } else if ( cls.equals( byte[].class ) ) {
            return cls.cast( ( (MingleBuffer) mv ).array() );
        }

        return cls.cast( mv );
    }

    public
    static
    < V >
    V
    mapGet( MingleSymbolMap m,
            CharSequence id,
            Class< V > cls )
    {
        inputs.notNull( id, "id" );
        return mapGet( m, MingleIdentifier.create( id ), cls );
    }

    public
    static
    < V >
    V
    mapExpect( MingleSymbolMap m,
               MingleIdentifier id,
               Class< V > cls )
    {
        V res = mapGet( m, id, cls );
        if ( res != null ) return res;

        throw state.failf( "no value for field: %s", id );
    }
    
    public
    static
    < V >
    V
    mapExpect( MingleSymbolMap m,
               CharSequence id,
               Class< V > cls )
    {
        inputs.notNull( id, "id" );
        return mapExpect( m, MingleIdentifier.create( id ), cls );
    }
    
    public
    static
    MingleValue
    mapGetValue( MingleSymbolMap m,
                 CharSequence id )
    {
        return mapGet( m, id, MingleValue.class );
    }
    
    public
    static
    MingleValue
    mapExpectValue( MingleSymbolMap m,
                    CharSequence id )
    {
        return mapExpect( m, id, MingleValue.class );
    }
    
    public
    static
    String
    mapGetString( MingleSymbolMap m,
                  CharSequence id )
    {
        return mapGet( m, id, String.class );
    }
    
    public
    static
    String
    mapExpectString( MingleSymbolMap m,
                     CharSequence id )
    {
        return mapExpect( m, id, String.class );
    }

    public
    static
    void
    assertIdPathsEqual( ObjectPath< MingleIdentifier > p1,
                        String p1Name,
                        ObjectPath< MingleIdentifier > p2,
                        String p2Name )
    {
        state.isTruef( ObjectPaths.areEqual( p1, p2 ),
            "%s != %s (%s != %s)",
            p1Name,
            p2Name,
            p1 == null ? null : Mingle.formatIdPath( p1 ),
            p2 == null ? null : Mingle.formatIdPath( p2 )
        );
    }

    public
    static
    void
    assertIdPathsEqual( ObjectPath< MingleIdentifier > expct,
                        ObjectPath< MingleIdentifier > act )
    {
        assertIdPathsEqual( expct, "expct", act, "act" );
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

    private
    static
    void
    assertEqualNull( MingleValue mv1,
                     MingleValue mv2,
                     PathWiseAsserter< MingleIdentifier > a )
    {
        if ( mv1 == null ) mv1 = MingleNull.getInstance();
        if ( mv2 == null ) mv2 = MingleNull.getInstance();
        a.equal( mv1, mv2 );
    }

    public
    static
    void
    assertEqual( MingleValue mv1,
                 MingleValue mv2,
                 PathWiseAsserter< MingleIdentifier > a )
    {
        inputs.notNull( a, "a" );
        codef( "checking mv1 %s, mv2 %s at %s", mv1, mv2, a.formatPath() );
        
        if ( ! a.sameNullity( mv1, mv2 ) ) return;

        if ( mv1 instanceof MingleStruct ) {
            assertEqualStruct( mv1, mv2, a );
        } else if ( mv1 instanceof MingleEnum ) {
            assertEqualEnum( mv1, mv2, a );
        } else if ( mv1 instanceof MingleSymbolMap ) {
            assertEqualMap( mv1, mv2, a );
        } else if ( mv1 instanceof MingleList ) {
            assertEqualList( mv1, mv2, a );
        } else if ( mv1 instanceof MingleNull || mv2 instanceof MingleNull ) {
            assertEqualNull( mv1, mv2, a );
        } else {
            a.equal( mv1, mv2 );
        }
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

    public
    static
    CharSequence
    optInspect( MingleValue mv )
    {
        return Mingle.inspect( mv == null ? MingleNull.getInstance(): mv );
    }
}
