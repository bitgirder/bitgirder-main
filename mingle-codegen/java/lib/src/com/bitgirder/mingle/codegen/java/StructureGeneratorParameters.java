package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.mingle.model.MingleIdentifier;

import java.util.List;
import java.util.Map;

abstract
class StructureGeneratorParameters
extends TypeMaskedGeneratorParameters
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final List< FieldGeneratorParameters > fldParams;
    final JvVisibility vis;

    StructureGeneratorParameters( Builder< ? > b )
    {
        super( b );

        this.fldParams = Lang.unmodifiableCopy( b.fldParams );
        this.vis = b.vis;
    }

    static
    abstract
    class Builder< B extends Builder< B > >
    extends TypeMaskedGeneratorParameters.Builder< B >
    {
        private List< FieldGeneratorParameters > fldParams = Lang.emptyList();
        private JvVisibility vis = JvVisibility.PUBLIC;

        final
        B
        setFieldParameters( List< FieldGeneratorParameters > fldParams )
        {
            this.fldParams = inputs.noneNull( fldParams, "fldParams" );
            return castThis();
        }

        final
        B
        setVisibility( JvVisibility vis )
        {
            this.vis = inputs.notNull( vis, "vis" );
            return castThis();
        }
    }
}
