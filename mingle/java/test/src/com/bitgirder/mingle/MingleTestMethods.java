package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ObjectPaths;

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

    // ev1.type known to equal ev2.type
    private
    static
    void
    assertEventDataEqual( MingleValueReactorEvent ev1,
                          String ev1Name,
                          MingleValueReactorEvent ev2,
                          String ev2Name )
    {
        Object o1 = null;
        Object o2 = null;
        String desc = null;

        switch ( ev1.type() ) {
        case VALUE: o1 = ev1.value(); o2 = ev2.value(); desc = "value"; break;
        case FIELD_START:
            o1 = ev1.field(); o2 = ev2.field(); desc = "field"; break;
        case STRUCT_START:
            o1 = ev1.structType(); o2 = ev2.structType(); desc = "structType";
            break;
        default: return;
        }

        state.equalf( o1, o2, "%s.%s != %s.%s (%s != %s)",
            ev1Name, desc, ev2Name, desc, o1, o2 );
    }

    public
    static
    void
    assertEventsEqual( MingleValueReactorEvent ev1,
                       String ev1Name,
                       MingleValueReactorEvent ev2,
                       String ev2Name )
    {
        if ( state.sameNullity( ev1, ev2 ) ) 
        {
            state.equalf( ev1.type(), ev2.type(), 
                "%s.type != %s.type (%s != %s)", ev1Name, ev2Name, ev1.type(),
                ev2.type() );
        }

        assertEventDataEqual( ev1, ev1Name, ev2, ev2Name );

        if ( state.sameNullity( ev1.path(), ev2.path() ) ) 
        {
            assertIdPathsEqual( ev1.path(), ev1Name + ".path()", ev2.path(), 
                ev2Name + ".path()" );
        }
    }
}
