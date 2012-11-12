package com.bitgirder.mingle.bind;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.mingle.model.MingleStructure;
import com.bitgirder.mingle.model.MingleStructureBuilder;
import com.bitgirder.mingle.model.MingleSymbolMap;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleNull;
import com.bitgirder.mingle.model.MingleSymbolMapBuilder;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.MingleTypeCastException;
import com.bitgirder.mingle.model.AtomicTypeReference;
import com.bitgirder.mingle.model.QualifiedTypeName;
import com.bitgirder.mingle.model.TypeDefinition;
import com.bitgirder.mingle.model.ExceptionDefinition;
import com.bitgirder.mingle.model.FieldDefinition;
import com.bitgirder.mingle.model.TypeDefinitionLookup;

public
abstract
class AbstractBindImplementation
implements MingleBinders.Initializer,
           MingleBinding
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private boolean isException;
    private AtomicTypeReference mgType;

    protected AbstractBindImplementation() {}

    protected
    final
    void
    implSetTypeDef( TypeDefinitionLookup types,
                    QualifiedTypeName nm )
    {
        TypeDefinition typeDef = types.expectType( nm );
        isException = typeDef instanceof ExceptionDefinition;

        mgType = AtomicTypeReference.create( typeDef.getName() );
    }

    protected
    abstract
    Object
    implFromMingleStructure( MingleSymbolMap m,
                             MingleBinder mb,
                             ObjectPath< MingleIdentifier > path );

    // overridable
    void
    completeJavaValue( Object jVal,
                       AtomicTypeReference typ,
                       MingleValue mv,
                       MingleBinder mb,
                       ObjectPath< MingleIdentifier > path )
    {}

    private
    MingleSymbolMap
    expectSymbolMap( MingleValue mv,
                     ObjectPath< MingleIdentifier > path )
    {
        if ( mv instanceof MingleSymbolMap ) return (MingleSymbolMap) mv;
        else 
        {
            MingleTypeReference act = MingleModels.typeReferenceOf( mv );

            if ( mv instanceof MingleStructure )
            {
                MingleStructure ms = (MingleStructure) mv;
 
                if ( act.equals( mgType ) ) return ms.getFields();
                else throw new MingleTypeCastException( mgType, act, path );
            }
            else throw new MingleTypeCastException( mgType, act, path );
        }
    }

    public
    final
    Object
    asJavaValue( AtomicTypeReference typ,
                 MingleValue mv,
                 MingleBinder mb,
                 ObjectPath< MingleIdentifier > path )
    {
        state.equal( mgType, typ );
        MingleSymbolMap m = expectSymbolMap( mv, path );

        Object res = implFromMingleStructure( m, mb, path );
        completeJavaValue( res, typ, mv, mb, path );

        return res;
    }

    protected
    final
    void
    implSetField( FieldDefinition fd,
                  String jvFldId,
                  Object val,
                  MingleSymbolMapBuilder b,
                  MingleBinder mb,
                  ObjectPath< String > path,
                  boolean useOpaque )
    {
        MingleBinders.setField( fd, jvFldId, val, b, mb, path, useOpaque );
    }

    protected
    void
    implSetFields( Object obj,
                   MingleSymbolMapBuilder b,
                   MingleBinder mb,
                   ObjectPath< String > path )
    {}

    public
    final
    MingleValue
    asMingleValue( Object obj,
                   MingleBinder mb,
                   ObjectPath< String > path )
    {
        state.notNull( obj, "obj" );
        state.notNull( mb, "mb" );
        state.notNull( path, "path" );

        MingleStructureBuilder b = 
            isException 
                ? MingleModels.exceptionBuilder()
                : MingleModels.structBuilder();

        b.setType( mgType );

        implSetFields( obj, b.fields(), mb, path );

        return b.build();
    }
}
