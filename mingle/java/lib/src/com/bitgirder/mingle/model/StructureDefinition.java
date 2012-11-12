package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

import java.util.List;

public
abstract
class StructureDefinition
extends TypeDefinition
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final FieldSet fields;
    private final List< MingleTypeReference > constructors;

    StructureDefinition( Builder< ?, ? > b )
    {
        super( b );
        this.fields = inputs.notNull( b.fields, "fields" );
        this.constructors = Lang.unmodifiableCopy( b.constructors );
    }

    public final FieldSet getFieldSet() { return fields; }

    public 
    final 
    List< MingleTypeReference > 
    getConstructors() 
    {
        return constructors;
    }

    public
    static
    abstract
    class Builder< S extends StructureDefinition, B extends Builder< S, B > >
    extends TypeDefinition.Builder< S, B >
    {
        private FieldSet fields;
        private List< MingleTypeReference > constructors = Lang.emptyList();

        Builder() {}

        public
        final
        B
        setFields( FieldSet fields )
        {
            this.fields = inputs.notNull( fields, "fields" );
            return castThis();
        }

        public
        final
        B
        setConstructors( List< MingleTypeReference > constructors )
        {
            this.constructors = inputs.noneNull( constructors, "constructors" );
            return castThis();
        }
    }
}
