package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;

public
final
class TypeDefinitionCollection
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Iterable< TypeDefinition > types;

    private
    TypeDefinitionCollection( Iterable< TypeDefinition > types )
    {
        this.types = types;
    }

    public Iterable< TypeDefinition > getTypes() { return types; }

    public
    static
    TypeDefinitionCollection
    wrap( Iterable< TypeDefinition > types )
    {
        inputs.noneNull( types, "types" );
        return new TypeDefinitionCollection( types );
    }

    public
    static
    TypeDefinitionCollection
    create( Iterable< TypeDefinition > types )
    {
        inputs.notNull( types, "types" );

        List< TypeDefinition > l = Lang.newList();
        for ( TypeDefinition td : types ) l.add( td );

        return wrap( l );
    }
}
