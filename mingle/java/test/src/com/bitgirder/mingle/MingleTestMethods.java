package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;

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
    AtomicTypeReference
    atomic( CharSequence s )
    {
        inputs.notNull( s, "s" );
        return (AtomicTypeReference) MingleTypeReference.create( s );
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
}
