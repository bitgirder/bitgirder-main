package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;
import java.util.Map;

public
final
class TypeDefinitionLookup
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Map< QualifiedTypeName, TypeDefinition > types;

    private
    TypeDefinitionLookup( Builder b )
    {
        this.types = Lang.unmodifiableMap( b.types );
    }

    public Iterable< TypeDefinition > getTypes() { return types.values(); }

    public 
    TypeDefinition 
    expectType( QualifiedTypeName qn )
    {
        inputs.notNull( qn, "qn" );

        TypeDefinition res = types.get( qn );

        if ( res == null ) throw new NoSuchTypeDefinitionException( qn );
        else return res;
    }

//    MingleStruct
//    asMingleStruct()
//    {
//        MingleStructBuilder b = MingleModels.structBuilder();
//        b.setType( TYPE );
//
//        b.f().set( ID_TYPES, TypeDefinitions.asMingleList( getTypes() ) );
//
//        return b.build();
//    }

    public
    final
    static
    class DuplicateDefinitionException
    extends RuntimeException
    {
        private
        DuplicateDefinitionException( QualifiedTypeName qn )
        {
            super( "Duplicate definitions of type " + qn );
        }
    }

    public
    final
    static
    class Builder
    {
        private final Map< QualifiedTypeName, TypeDefinition > types =
            Lang.newMap();

        public
        Builder
        addTypes( Iterable< ? extends TypeDefinition > l )
        {
            inputs.noneNull( l, "l" );

            for ( TypeDefinition td : l )
            {
                QualifiedTypeName nm = td.getName();

                if ( types.put( nm, td ) != null )
                {
                    throw new DuplicateDefinitionException( nm );
                }
            }

            return this;
        }

        public
        Builder
        addTypes( TypeDefinitionCollection coll )
        {
            inputs.notNull( coll, "coll" );
            return addTypes( coll.getTypes() );
        }

        public
        Builder
        addTypes( TypeDefinitionLookup types )
        {
            inputs.notNull( types, "types" );
            return addTypes( types.getTypes() );
        }

        public
        Builder
        addType( TypeDefinition td )
        {
            inputs.notNull( td, "td" );
            return addTypes( Lang.singletonList( td ) );
        }

        public 
        TypeDefinitionLookup 
        build() 
        { 
            return new TypeDefinitionLookup( this ); 
        }
    }

//    static
//    TypeDefinitionLookup
//    fromMingleStruct( MingleStruct ms )
//    {
//        state.notNull( ms, "ms" );
//
//        ObjectPath< MingleIdentifier > path = ObjectPath.getRoot();
//
//        MingleSymbolMapAccessor acc =
//            MingleModels.expectStruct( ms, path, TYPE );
//
//        Builder b = new Builder();
//
//        MingleList mgTyps = acc.expectMingleList( ID_TYPES );
//        b.addTypes( TypeDefinitions.asTypeDefinitions( mgTyps ) );
//
//        return b.build();
//    }
}
