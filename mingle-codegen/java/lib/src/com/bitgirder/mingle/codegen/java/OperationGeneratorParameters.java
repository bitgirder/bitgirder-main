package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.mingle.model.MingleIdentifier;

import java.util.List;
import java.util.Map;

final
class OperationGeneratorParameters
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final MingleIdentifier name;
    final List< FieldGeneratorParameters > fldParams;
    final boolean useOpaqueJavaReturnType;

    private
    OperationGeneratorParameters( Builder b )
    {
        this.name = inputs.notNull( b.name, "name" );
        this.fldParams = Lang.unmodifiableCopy( b.fldParams );
        this.useOpaqueJavaReturnType = b.useOpaqueJavaReturnType;
    }

    final
    static
    class Builder
    {
        private MingleIdentifier name;
        private List< FieldGeneratorParameters > fldParams = Lang.emptyList();
        private boolean useOpaqueJavaReturnType;

        public
        Builder
        setName( MingleIdentifier name )
        {
            this.name = inputs.notNull( name, "name" );
            return this;
        }

        public
        Builder
        setFieldParams( List< FieldGeneratorParameters > fldParams )
        {
            this.fldParams = inputs.noneNull( fldParams, "fldParams" );
            return this;
        }

        public
        Builder
        setUseOpaqueJavaReturnType( boolean useOpaqueJavaReturnType )
        {
            this.useOpaqueJavaReturnType = useOpaqueJavaReturnType;
            return this;
        }

        OperationGeneratorParameters
        build()
        {
            return new OperationGeneratorParameters( this );
        }
    }
}
