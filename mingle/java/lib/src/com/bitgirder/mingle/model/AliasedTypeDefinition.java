package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class AliasedTypeDefinition
extends TypeDefinition
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleTypeReference aliasedType;

    private
    AliasedTypeDefinition( Builder b )
    {
        super( b );

        this.aliasedType = inputs.notNull( b.aliasedType, "aliasedType" );
    }

    public MingleTypeReference getAliasedType() { return aliasedType; }

    public
    final
    static
    class Builder
    extends TypeDefinition.Builder< AliasedTypeDefinition, Builder >
    {
        private MingleTypeReference aliasedType;

        public
        Builder
        setAliasedType( MingleTypeReference aliasedType )
        {
            this.aliasedType = inputs.notNull( aliasedType, "aliasedType" );
            return this;
        }

        public
        AliasedTypeDefinition
        build()
        {
            return new AliasedTypeDefinition( this );
        }
    }
}
