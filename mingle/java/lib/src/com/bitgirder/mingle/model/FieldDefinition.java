package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;

public
final
class FieldDefinition
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleIdentifier name;
    private final MingleTypeReference typeRef;
    private final MingleValue defl;

    private
    FieldDefinition( Builder b )
    {
        this.name = inputs.notNull( b.name, "name" );
        this.typeRef = inputs.notNull( b.typeRef, "typeRef" );
        this.defl = b.coerceDefault();
    }

    public MingleIdentifier getName() { return name; }
    public MingleTypeReference getType() { return typeRef; }
    public MingleValue getDefault() { return defl; }

    public
    final
    static
    class Builder
    {
        private final static ObjectPath< MingleIdentifier > ROOT =
            ObjectPath.getRoot();

        private MingleIdentifier name;
        private MingleTypeReference typeRef;
        private MingleValue defl;

        public
        Builder
        setName( MingleIdentifier name )
        {
            this.name = inputs.notNull( name, "name" );
            return this;
        }

        public
        Builder
        setType( MingleTypeReference typeRef )
        {
            this.typeRef = inputs.notNull( typeRef, "typeRef" );
            return this;
        }

        public
        Builder
        setDefault( MingleValue defl )
        {
            this.defl = inputs.notNull( defl, "defl" );
            return this;
        }

        private
        MingleValue
        coerceDefault()
        {
            inputs.notNull( typeRef, "typeRef" );

            if ( defl == null ) return null;
            else return MingleModels.asMingleInstance( typeRef, defl, ROOT );
        }

        public FieldDefinition build() { return new FieldDefinition( this ); }
    }
}
