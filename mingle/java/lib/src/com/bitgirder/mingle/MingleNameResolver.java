package com.bitgirder.mingle;

// pkg-level only for now
interface MingleNameResolver
{
    // returns null to indicate that this instance cannot resolve nm.
    public
    QualifiedTypeName
    resolve( DeclaredTypeName nm );
}
