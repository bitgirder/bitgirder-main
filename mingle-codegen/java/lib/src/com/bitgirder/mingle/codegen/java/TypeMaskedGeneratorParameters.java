package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.mingle.model.MingleTypeReference;

import java.util.regex.Pattern;

abstract
class TypeMaskedGeneratorParameters
implements GeneratorParameters
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Pattern typeMask;

    TypeMaskedGeneratorParameters( Builder< ? > b )
    {
        inputs.notNull( b, "b" );
        this.typeMask = inputs.notNull( b.typeMask, "typeMask" );
    }

    final
    boolean
    matches( MingleTypeReference typ )
    {
        inputs.notNull( typ, "typ" );

        return typeMask.matcher( typ.getExternalForm() ).matches();
    }

    static
    abstract
    class Builder< B extends Builder< B > >
    {
        private Pattern typeMask;

        final B castThis() { return Lang.< B >castUnchecked( this ); }

        public
        final
        B
        setTypeMask( Pattern typeMask )
        {
            this.typeMask = inputs.notNull( typeMask, "typeMask" );
            return castThis();
        }
    }
}
