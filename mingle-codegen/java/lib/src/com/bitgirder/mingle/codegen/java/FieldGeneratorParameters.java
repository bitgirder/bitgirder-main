package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.model.MingleIdentifier;

final
class FieldGeneratorParameters
implements GeneratorParameters
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleIdentifier name;
    private final boolean useOpaqueJavaType;

    private
    FieldGeneratorParameters( Builder b )
    {
        this.name = inputs.notNull( b.name, "name" );
        this.useOpaqueJavaType = b.useOpaqueJavaType;
    }

    MingleIdentifier name() { return name; }
    boolean useOpaqueJavaType() { return useOpaqueJavaType; }

    final
    static
    class Builder
    {
        private MingleIdentifier name;
        private boolean useOpaqueJavaType;

        public
        Builder
        setName( MingleIdentifier name )
        {
            this.name = inputs.notNull( name, "name" );
            return this;
        }

        public
        Builder
        setUseOpaqueJavaType( boolean useOpaqueJavaType )
        {
            this.useOpaqueJavaType = useOpaqueJavaType;
            return this;
        }
        
        FieldGeneratorParameters
        build()
        {
            return new FieldGeneratorParameters( this );
        }
    }
}
