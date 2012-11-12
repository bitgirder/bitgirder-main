package com.bitgirder.mingle.codegen.java;

import static com.bitgirder.mingle.codegen.java.CodegenConstants.*;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.mingle.model.StructureDefinition;
import com.bitgirder.mingle.model.ExceptionDefinition;
import com.bitgirder.mingle.model.FieldDefinition;
import com.bitgirder.mingle.model.TypeDefinition;
import com.bitgirder.mingle.model.AtomicTypeReference;
import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.QualifiedTypeName;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleIdentifier;

import java.util.List;
import java.util.Deque;
import java.util.Map;

final
class StructureDefinitionBuilder
extends TypeDefinitionBuilder< StructureDefinition >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static JvType JV_TYPE_MINGLE_BINDER =
        JvQname.create( "com.bitgirder.mingle.bind", "MingleBinder" );

    private final static JvType JV_TYPE_JAVA_OBJ =
        JvQname.create( "java.lang", "Object" );

    private final static JvId JV_ID_CAST_THIS = new JvId( "castThis" );

    private final static JvTypeName JV_TYPE_B = new JvTypeName( "B" );

    private final static JvTypeName JV_TYPE_NAME_BUILDER =
        new JvTypeName( "Builder" );

    private final static JvType JV_TYPE_BINDER_BUILDER =
        JvTypeExpression.
            dotTypeOf( JV_TYPE_MINGLE_BINDER, JV_TYPE_NAME_BUILDER );
    
    private final static JvTypeName JV_TYPE_ABSTRACT_BUILDER =
        new JvTypeName( "AbstractBuilder" );

    private final static JvType JV_TYPE_BOUND_STRUCT =  
        JvQname.create( "com.bitgirder.mingle.bind", "BoundStruct" );

    private final static JvType JV_TYPE_BOUND_EXCEPTION =
        JvQname.create( "com.bitgirder.mingle.bind", "BoundException" );

    private final static JvTypeName JV_TYPE_ABSTRACT_BIND_IMPLEMENTATION =
        new JvTypeName( "AbstractBindImplementation" );

    private final static JvId JV_ID_INITIALIZE = new JvId( "initialize" );

    private final static JvId JV_ID_CAUSE = new JvId( "_cause" );

    private final static JvType JV_TYPE_THROWABLE =
        JvQname.create( "java.lang", "Throwable" );

    private final static JvId JV_ID_ADD_BINDING =
        new JvId( "addBinding" );

    private final static JvId JV_ID_CREATE = new JvId( "create" );

    private final static JvId JV_ID_GET_ROOT = new JvId( "getRoot" );

    private final static JvId JV_ID_IMPL_SET_TYPE_DEF =
        new JvId( "implSetTypeDef" );

    private final static JvType JV_TYPE_STRUCTURE_DEF =
        JvQname.create( "com.bitgirder.mingle.model", "StructureDefinition" );

    private final static JvType JV_TYPE_OBJ_PATH =
        JvQname.create( "com.bitgirder.lang.path", "ObjectPath" );

    private final static JvType JV_TYPE_OBJ_PATH_STRING =
        JvTypeExpression.
            withParams( 
                JV_TYPE_OBJ_PATH, JvQname.create( "java.lang", "String" ) );
        
    private final static JvType JV_TYPE_MINGLE_IDENTIFIER =
        JvQname.create( "com.bitgirder.mingle.model", "MingleIdentifier" );
    
    private final static JvType JV_TYPE_OBJ_PATH_IDENT =
        JvTypeExpression.
            withParams( JV_TYPE_OBJ_PATH, JV_TYPE_MINGLE_IDENTIFIER );

    private final static JvType JV_TYPE_MINGLE_VALUE =
        JvQname.create( "com.bitgirder.mingle.model", "MingleValue" );

    private final static JvType JV_TYPE_MINGLE_STRUCTURE =
        JvQname.create( "com.bitgirder.mingle.model", "MingleStructure" );

    private final static JvId JV_ID_IMPL_FROM_MINGLE_STRUCTURE =
        new JvId( "implFromMingleStructure" );

    private final static JvId JV_ID_IMPL_SET_FIELDS =
        new JvId( "implSetFields" );

    private final static JvId JV_ID_IMPL_SET_FIELD = new JvId( "implSetField" );

    private final static JvType JV_TYPE_QNAME =
        JvQname.create( "com.bitgirder.mingle.model", "QualifiedTypeName" );
    
    private final static JvType JV_TYPE_SYMBOL_MAP_BUILDER =
        JvQname.
            create( "com.bitgirder.mingle.model", "MingleSymbolMapBuilder" );

    private final static JvType JV_TYPE_MINGLE_TYPE_REF =
        JvQname.create( "com.bitgirder.mingle.model", "MingleTypeReference" );

    private final static JvType JV_TYPE_ATOMIC_TYPE_REF =
        JvQname.create( "com.bitgirder.mingle.model", "AtomicTypeReference" );

    // Initialized early and in the order listed here
    private JvIdMapper idMap;
    private final Map< MingleIdentifier, JvId > jvFldDeflIds = Lang.newMap();

    // These are set early in the build process and used often afterwards and
    // are in layout order
    private final List< JvField > sprFields = Lang.newList();
    private final List< JvField > instFields = Lang.newList();

    // non-null after addSuperFields() for instances having isException()
    // returning true
    private JvField causeFld;

    private
    boolean
    isException()
    {
        return typeDef() instanceof ExceptionDefinition;
    }

    private
    JvType
    getGenClassSupertype()
    {
        MingleTypeReference sprRef = typeDef().getSuperType();
        
        if ( sprRef == null ) 
        {
            if ( isException() ) return JV_TYPE_BOUND_EXCEPTION;
            else return JV_TYPE_BOUND_STRUCT;
        }
        else return jvTypeOf( sprRef );
    }
    
    private List< JvField > instanceFields() { return instFields; }
    private List< JvField > sprFields() { return sprFields; }

    private
    Map< MingleIdentifier, FieldGeneratorParameters >
    getFieldGeneratorParamMap( StructureDefinition sd )
    {
        StructureGeneratorParameters p =
            getGeneratorParameters( sd, StructureGeneratorParameters.class );

        if ( p == null ) return Lang.emptyMap();
        else return fieldParamsByName( p.fldParams );
    }

    // Util method for use in initializing fields or field templates
    private
    List< JvField >
    getJvFields( StructureDefinition sd )
    {
        List< JvField > res = Lang.newList();

        Map< MingleIdentifier, FieldGeneratorParameters > m =
            getFieldGeneratorParamMap( sd );

        for ( FieldDefinition fd : sd.getFieldSet() )
        {
            res.add( asJvField( fd, m.get( fd.getName() ) ) );
        }

        return res;
    }

    private 
    List< JvField >
    allFields()
    {
        List< JvField > res = Lang.newList();
        res.addAll( sprFields() );
        res.addAll( instanceFields() );
        if ( causeFld != null ) res.add( causeFld );

        return res;
    }

    private
    Iterable< StructureDefinition >
    getSuperTypes()
    {
        Deque< StructureDefinition > deq = Lang.newDeque();

        for ( TypeDefinition td = context().getSuperDef( typeDef() ); 
              td != null; 
              td = context().getSuperDef( td ) )
        {
            deq.push( (StructureDefinition) td );
        }

        return deq;
    }

    private
    JvField
    createCauseField( JvClass cls )
    {
        JvField cause = new JvField();
        cause.name = JV_ID_CAUSE;
        cause.type = JV_TYPE_THROWABLE;

        return cause;
    }

    private
    void
    addSuperFields( JvClass cls )
    {
        Iterable< StructureDefinition > deq = getSuperTypes();

        for ( StructureDefinition sd : deq )
        {
            for ( JvField f : getJvFields( sd ) ) sprFields.add( f );
        }

        // set it here but don't add it to sprFields: we always want it as the
        // last field listed in allFields()
        if ( isException() ) causeFld = createCauseField( cls );
    }

    private
    void
    addFieldGetter( JvField f,
                    JvClass cls )
    {
        JvMethod m = new JvMethod();
        m.mods.vis = JvVisibility.PUBLIC;
        m.mods.isFinal = true;
        m.name = f.name;
        m.retType = f.type;
        
        m.body.add( new JvReturn( f.name ) );

        cls.methods.add( m );
    }

    private
    void
    addFields( StructureDefinition sd,
               JvClass cls )
    {
        for ( JvField jf : getJvFields( sd ) )
        {
            addJvConstField( jf.mgField, jf.type, jvFldDeflIds, cls );
            jf.mods.vis = JvVisibility.PRIVATE;
            jf.mods.isFinal = true;

            cls.fields.add( jf );
            addFieldGetter( jf, cls );
            instFields.add( jf );
        }
    }

    private
    void
    addBoundStructSuperCall( JvConstructor c )
    {
        JvFuncCall call = new JvFuncCall();
        call.target = JvId.SUPER;

        for ( JvField jf : sprFields() ) call.params.add( jf.name );
        if ( causeFld != null ) call.params.add( causeFld.name );

        c.body.add( new JvStatement( call ) );
    }

    private
    void
    addConstructor( JvClass cls )
    {
        JvConstructor c = new JvConstructor();
        c.vis = JvVisibility.PROTECTED;
        addBoundStructSuperCall( c );

        for ( JvField jf : sprFields() ) c.params.add( JvParam.forField( jf ) );

        for ( JvField jf : instanceFields() )
        {
            c.params.add( JvParam.create( jf.name, jf.type ) );

            c.body.add(
                new JvStatement(
                    new JvAssign( 
                        new JvAccess( JvId.THIS, jf.name ), 
                        expValidateFieldValue( jf, idMap ) 
                    ) 
                )
            );
        }

        if ( causeFld != null ) c.params.add( JvParam.forField( causeFld ) );

        cls.constructors.add( c );
    }

    private
    void
    addFieldValidation( JvMethod m,
                        JvField f )
    {
        m.body.add( new JvStatement( expValidateFieldValue( f, idMap ) ) );
    }

    private
    void
    addStaticFactoryBody( JvMethod m,
                          int paramLen )
    {
        JvInstantiate inst = new JvInstantiate();
        inst.target = m.retType;

        // this loop adds the call parameters to the final return statement, but
        // as a side effect also adds method parameters as appropriate 
        int i = 0;
        for ( JvField f : allFields() )
        {
            if ( i < paramLen ) m.params.add( JvParam.forField( f ) );
            inst.params.add( i < paramLen ? f.name : JvLiteral.NULL ); 
            ++i;
        }

        m.body.add( new JvReturn( inst ) );
    }

    private
    void
    addStaticFactory( JvPackage pkg,
                      JvClass cls,
                      int paramLen )
    {
        JvQname typ = JvQname.create( pkg, cls.name );

        JvMethod m = new JvMethod();
        m.mods.vis = JvVisibility.PUBLIC;
        m.mods.isStatic = true;
        m.mods.isFinal = false;
        m.name = JV_ID_CREATE;
        m.retType = typ;

        addStaticFactoryBody( m, paramLen );

        cls.methods.add( m );
    }

    private
    void
    addStaticFactories( JvPackage pkg,
                        JvClass cls )
    {
        int allLen = allFields().size();

        addStaticFactory( pkg, cls, allLen );

        // If this is an exception add a static factory that doesn't take a
        // cause
        if ( isException() ) addStaticFactory( pkg, cls, allLen - 1 );
    }

    private
    void
    setBuilderTypeInfo( JvClass bldr,
                        JvClass encl )
    {
        JvTypeExpression selfTyp =
            JvTypeExpression.withParams( bldr.name, JV_TYPE_B );

        bldr.typeParams.add( new JvTypeExtendParameter( JV_TYPE_B, selfTyp ) );

        JvTypeExpression sprTyp = new JvTypeExpression();

        sprTyp.type = 
            JvTypeExpression.dotTypeOf( encl.sprTyp, JV_TYPE_ABSTRACT_BUILDER );
        
        sprTyp.args.add( JV_TYPE_B );

        bldr.sprTyp = sprTyp;
    }

    private
    JvMethod
    createBuilderFieldSetter( JvField f )
    {
        JvMethod m = new JvMethod();
        m.mods.isFinal = true;
        m.mods.vis = JvVisibility.PUBLIC;
        m.retType = JV_TYPE_B;
        m.name = setterNameFor( f );
        m.params.add( JvParam.create( f.name, f.type ) );

        m.body.add(
            new JvStatement(
                new JvAssign( new JvAccess( JvId.THIS, f.name ), f.name ) ) );
 
        m.body.add( new JvReturn( JvFuncCall.create( JV_ID_CAST_THIS ) ) );
 
        return m;
    }

    private
    void
    addBuilderFieldSetters( JvClass bldr,
                            JvClass cls )
    {
        for ( JvField f : instanceFields() )
        {
            f = f.copyOf();
            f.mods.isFinal = false;
            f.mods.vis = JvVisibility.PROTECTED;

            JvId jvDeflId = jvFldDeflIds.get( f.mgField.getName() );
            if ( jvDeflId != null ) f.assign = jvDeflId;

            bldr.fields.add( f );

            bldr.methods.add( createBuilderFieldSetter( f ) );
        }
    }

    private
    JvMethod
    createBuilderBuildMethod( JvClass cls )
    {
        JvMethod m = new JvMethod();
        m.mods.vis = JvVisibility.PUBLIC;
        m.mods.isFinal = true;
        m.retType = cls.name;
        m.name = new JvId( "build" );

        JvFuncCall ret = new JvFuncCall();
        ret.target = new JvAccess( m.retType, JV_ID_CREATE );

        for ( JvField f : allFields() ) ret.params.add( f.name );

        m.body.add( new JvReturn( ret ) );

        return m;
    }

    private
    JvClass
    addAbstractBuilder( JvClass cls )
    {
        JvClass bldr = new JvClass();
        bldr.mods.vis = JvVisibility.PUBLIC; 
        bldr.mods.isStatic = true;
        bldr.mods.isFinal = false;
        bldr.name = JV_TYPE_ABSTRACT_BUILDER;
        setBuilderTypeInfo( bldr, cls );

        addBuilderFieldSetters( bldr, cls );

        cls.nestedTypes.add( bldr );

        return bldr;
    }

    private
    void
    addBoundBuilder( JvClass abstractBldr,
                     JvClass cls )
    {
        JvClass bb = new JvClass();
        bb.mods.vis = JvVisibility.PUBLIC;
        bb.mods.isStatic = true;
        bb.mods.isFinal = true;
        bb.name = JV_TYPE_NAME_BUILDER;

        bb.sprTyp = JvTypeExpression.withParams( abstractBldr.name, bb.name );
        
        bb.methods.add( createBuilderBuildMethod( cls ) );

        cls.nestedTypes.add( bb );
    }

    private
    void
    addBinderConstructor( JvClass init )
    {
        JvConstructor c = new JvConstructor();
        c.vis = JvVisibility.PROTECTED;

        init.constructors.add( c );
    }

    private
    void
    addBinderCallSetTypeDef( JvMethod m,
                             JvId typesId )
    {
        m.body.add(
            new JvStatement(
                JvFuncCall.create(
                    JV_ID_IMPL_SET_TYPE_DEF,
                    typesId,
                    idMap.idFor( typeDef().getName() )
                )
            )
        );
    }

    private
    void
    addBindingInstallCall( JvMethod m,
                           JvClass cls,
                           JvId bldrId )
    {
        m.body.add(
            new JvStatement(
                JvFuncCall.create(
                    new JvAccess( bldrId, JV_ID_ADD_BINDING ),
                    idMap.idFor( typeDef().getName() ),
                    JvId.THIS,
                    new JvAccess( cls.name, JvId.CLASS )
                )
            )
        );
    }

    private
    void
    addBinderInitMethod( JvClass initCls,
                         JvClass cls )
    {
        JvMethod m = new JvMethod();
        m.mods.vis = JvVisibility.PUBLIC;
        m.mods.isFinal = false;
        m.retType = JvPrimitiveType.VOID;
        m.name = JV_ID_INITIALIZE;
        m.anns.add( JvAnnotation.OVERRIDE );

        JvId bldrId = nextJvId();
        JvId typesId = nextJvId();

        m.params.add( JvParam.create( bldrId, JV_TYPE_BINDER_BUILDER ) );
        m.params.add( JvParam.create( typesId, JV_QNAME_TYPE_DEF_LOOKUP ) );

        addBinderCallSetTypeDef( m, typesId );
        addBindingInstallCall( m, cls, bldrId );
                        
        initCls.methods.add( m );
    }

    private
    void
    addAsMingleValueImplSetFields( JvMethod m,
                                   JvId typedObjId,
                                   JvId smBldrId,
                                   JvId mbId,
                                   JvId pathId )
    {
        for ( JvField f : instanceFields() )
        {
            m.body.add(
                new JvStatement(
                    JvFuncCall.create(
                        JV_ID_IMPL_SET_FIELD,
                        idMap.idFor( f.mgField ),
                        new JvString( f.name ),
                        JvFuncCall.create( new JvAccess( typedObjId, f.name ) ),
                        smBldrId,
                        mbId,
                        pathId,
                        expBoolLiteral( useOpaqueJavaType( f ) )
                    )
                )
            );
        }
    }

    private
    void
    addBinderImplSetFieldsBody( JvMethod m,
                                JvClass cls,
                                JvId objId,
                                JvId smBldrId,
                                JvId mbId,
                                JvId pathId )
    {
        JvFuncCall spr = new JvFuncCall();
        spr.target = new JvAccess( JvId.SUPER, m.name );
        for ( JvParam p : m.params ) spr.params.add( p.id );
        m.body.add( new JvStatement( spr ) );

        JvLocalVar typed = new JvLocalVar();
        typed.name = nextJvId();
        typed.type = cls.name;
        typed.assign = new JvCast( cls.name, objId );

        m.body.add( typed );
    
        addAsMingleValueImplSetFields( m, typed.name, smBldrId, mbId, pathId );
    }

    private
    void
    addBinderImplSetFields( JvClass init,
                            JvClass cls )
    {
        JvMethod m = new JvMethod();
        m.mods.vis = JvVisibility.PROTECTED;
        m.mods.isFinal = false;
        m.anns.add( JvAnnotation.OVERRIDE );
        m.retType = JvPrimitiveType.VOID;
        m.name = JV_ID_IMPL_SET_FIELDS;

        JvId objId = nextJvId();
        JvId smBldrId = nextJvId();
        JvId bndrId = nextJvId();
        JvId pathId = nextJvId();

        m.params.add( JvParam.create( objId, JV_TYPE_JAVA_OBJ ) );
        m.params.add( JvParam.create( smBldrId, JV_TYPE_SYMBOL_MAP_BUILDER ) );
        m.params.add( JvParam.create( bndrId, JV_TYPE_MINGLE_BINDER ) );
        m.params.add( JvParam.create( pathId, JV_TYPE_OBJ_PATH_STRING ) );

        addBinderImplSetFieldsBody( m, cls, objId, smBldrId, bndrId, pathId );

        init.methods.add( m );
    }

    private
    JvLocalVar
    getImplAsJavaValueVar( AsJavaValCtx ctx )
    {
        JvLocalVar res = new JvLocalVar();

        res.name = nextJvId();
        res.type = ctx.f.type;

        if ( ctx.f.mgField == null ) res.assign = JvLiteral.NULL;
        else res.assign = optUnbox( ctx.f.type, expAsJavaValue( ctx, idMap ) );
        
        return res;
    }

    private
    void
    addBinderAsJavaValueBodyReturn( JvMethod m,
                                    JvClass cls,
                                    List< JvLocalVar > params )
    {
        JvFuncCall ret = new JvFuncCall();
        ret.target = new JvAccess( cls.name, JV_ID_CREATE );

        for ( JvLocalVar p : params ) ret.params.add( p.name );

        m.body.add( new JvReturn( ret ) );
    }

    private
    void
    addBinderAsJavaValueBody( JvMethod m,
                              JvClass cls,
                              JvId smId,
                              JvId mbId,
                              JvId pathId )
    {
        List< JvLocalVar > flds = Lang.newList();

        for ( JvField f : allFields() )
        {
            AsJavaValCtx ctx = new AsJavaValCtx( f, smId, mbId, pathId );
            JvLocalVar lv = getImplAsJavaValueVar( ctx );

            flds.add( lv );
            m.body.add( lv );
        }

        addBinderAsJavaValueBodyReturn( m, cls, flds );
    }

    private
    void
    addBinderAsJavaValue( JvClass init,
                          JvClass cls )
    {
        JvMethod m = new JvMethod();
        m.mods.vis = JvVisibility.PROTECTED;
        m.mods.isFinal = false;
        m.anns.add( JvAnnotation.OVERRIDE );
        m.retType = JV_TYPE_JAVA_OBJ;
        m.name = JV_ID_IMPL_FROM_MINGLE_STRUCTURE;

        JvId smId = nextJvId();
        JvId mbId = nextJvId();
        JvId pathId = nextJvId();

        m.params.add( JvParam.create( smId, JV_QNAME_SYM_MAP ) );
        m.params.add( JvParam.create( mbId, JV_TYPE_MINGLE_BINDER ) );
        m.params.add( JvParam.create( pathId, JV_TYPE_OBJ_PATH_IDENT ) );

        addBinderAsJavaValueBody( m, cls, smId, mbId, pathId );
        
        init.methods.add( m );
    }

    private
    void
    addBinderInitializer( JvClass cls )
    {
        JvClass init = new JvClass();
        init.name = JV_TYPE_BIND_IMPLEMENTATION;
        init.mods.vis = JvVisibility.PROTECTED;
        init.mods.isFinal = false;
        init.mods.isStatic = true;

        init.sprTyp = 
            JvTypeExpression.
                dotTypeOf( cls.sprTyp, JV_TYPE_ABSTRACT_BIND_IMPLEMENTATION );

        addBinderConstructor( init );
        addBinderInitMethod( init, cls );
        addBinderImplSetFields( init, cls );
        addBinderAsJavaValue( init, cls );

        cls.nestedTypes.add( init );
    }

    private
    void
    addBindingFacilities( JvClass cls )
    {
        addBinderInitializer( cls );
    }

    private
    JvVisibility
    getTopLevelVisibility()
    {
        StructureGeneratorParameters p =
            getGeneratorParameters( StructureGeneratorParameters.class );

        return p == null ? JvVisibility.PUBLIC : p.vis;
    }

    private
    JvClass
    buildClass( JvPackage pkg )
    {
        JvClass cls = new JvClass(); 
        idMap = new JvIdMapper( cls );

        cls.mods.vis = getTopLevelVisibility();
        cls.mods.isFinal = false;
        cls.name = jvTypeNameOf( typeDef() );
        cls.sprTyp = getGenClassSupertype();

        addSuperFields( cls );
        addFields( typeDef(), cls );
        addConstructor( cls );
        addStaticFactories( pkg, cls );
        JvClass abstractBldr = addAbstractBuilder( cls );
        addBoundBuilder( abstractBldr, cls );
        addBindingFacilities( cls );

        return cls;
    }

    BuildResult
    buildImpl()
    {
        BuildResult res = new BuildResult();

        JvCompilationUnit u = new JvCompilationUnit();
        u.pkg = jvPackageOf( typeDef() );
        u.decl = buildClass( u.pkg );

        res.units.add( u );
        res.addInitializerEntry( u, JV_TYPE_BIND_IMPLEMENTATION );

        return res;
    }
}
