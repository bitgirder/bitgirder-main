package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class Mingle
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final static MingleNameResolver CORE_NAME_RESOLVER =
        new MingleNameResolver() {
            public QualifiedTypeName resolve( DeclaredTypeName nm ) {
                return Mingle.resolveInCore( nm );
            }
        };
    
    public final static MingleNamespace NS_CORE;
    public final static QualifiedTypeName QNAME_STRING;
    public final static MingleTypeReference TYPE_STRING;

    private Mingle() {}

    static
    QualifiedTypeName
    resolveInCore( DeclaredTypeName nm )
    {
        return null;
    }

    private
    static
    MingleIdentifier
    initId( String... parts )
    {
        return new MingleIdentifier( parts );
    }

    private
    static
    MingleNamespace
    initNs( MingleIdentifier ver,
            MingleIdentifier... parts )
    {
        return new MingleNamespace( parts, ver );
    }

    private
    static
    QualifiedTypeName
    initCoreQname( String nm )
    {
        state.notNull( NS_CORE, "NS_CORE" );
        return new QualifiedTypeName( NS_CORE, new DeclaredTypeName( nm ) );
    }

    private
    static
    AtomicTypeReference
    initCoreType( QualifiedTypeName qn )
    {
        return new AtomicTypeReference( qn, null );
    }

    static
    {
        NS_CORE = 
            initNs( initId( "v1" ), initId( "mingle" ), initId( "core" ) );
        
        QNAME_STRING = initCoreQname( "String" );
        TYPE_STRING = initCoreType( QNAME_STRING );
    }
}
