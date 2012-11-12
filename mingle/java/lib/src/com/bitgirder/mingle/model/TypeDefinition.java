package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

public
abstract
class TypeDefinition
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final QualifiedTypeName nm;
    private final MingleTypeReference superTypeRef;

    TypeDefinition( Builder< ?, ? > b )
    {
        this.nm = inputs.notNull( b.nm, "nm" );
        this.superTypeRef = b.superTypeRef;
    }

    public QualifiedTypeName getName() { return nm; }
    public MingleTypeReference getSuperType() { return superTypeRef; }

    public
    abstract
    static
    class Builder< T extends TypeDefinition, B extends Builder< T, B > >
    {
        private QualifiedTypeName nm;
        private MingleTypeReference superTypeRef;

        Builder() {}

        final B castThis() { return Lang.< B >castUnchecked( this ); }

        public
        final
        B
        setName( QualifiedTypeName nm )
        {
            this.nm = inputs.notNull( nm, "nm" );
            return castThis();
        }

        public
        final
        B
        setSuperType( MingleTypeReference superTypeRef )
        {
            this.superTypeRef = inputs.notNull( superTypeRef, "superTypeRef" );
            return castThis();
        }
 
        public
        abstract
        T
        build();
    }
}
