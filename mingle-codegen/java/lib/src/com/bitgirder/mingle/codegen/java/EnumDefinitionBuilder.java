package com.bitgirder.mingle.codegen.java;

import static com.bitgirder.mingle.codegen.java.CodegenConstants.*;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.mingle.model.EnumDefinition;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.AtomicTypeReference;

import java.util.List;
import java.util.Map;
import java.util.Iterator;

final
class EnumDefinitionBuilder
extends TypeDefinitionBuilder< EnumDefinition >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static JvTypeName JV_TYPE_NAME_ENUM_BINDING_IMPL =
        new JvTypeName( "EnumBindingImpl" );

    private final static JvType JV_TYPE_ABSTRACT_ENUM_BINDING =
        JvQname.create( "com.bitgirder.mingle.bind", "AbstractEnumBinding" );

    private final static JvId JV_ID_GET_JAVA_ENUM = new JvId( "getJavaEnum" );

    private final static JvId JV_ID_GET_MINGLE_ENUM = 
        new JvId( "getMingleEnum" );

    private final static JvId JV_ID_EQUALS = new JvId( "equals" );

    private final static JvId JV_ID_NULL = new JvId( "null" );

    private final static JvId JV_ID_DEFAULT = new JvId( "default" );

    private final static JvId JV_ID_INITIALIZE = new JvId( "initialize" );

    private final static JvId JV_ID_ADD_BINDING = new JvId( "addBinding" );

    private final static JvTypeName JV_TYPE_NAME_BINDER_INITIALIZER =
        new JvTypeName( "BinderInitializer" );

    private JvIdMapper idMap;
    private final List< EnMapping > enMappings = Lang.newList();
    private final Map< MingleIdentifier, JvId > mgEnumConstants = Lang.newMap();

    private
    final
    static
    class EnMapping
    {
        private final JvId jvId;
        private final MingleIdentifier mgId;

        private
        EnMapping( JvId jvId,
                   MingleIdentifier mgId )
        {
            this.jvId = jvId;
            this.mgId = mgId;
        }
    }

    private
    void
    addMgConstant( MingleIdentifier id,
                   JvEnumDecl en )
    {
        JvField f = JvField.createConstField();
        f.type = JV_QNAME_MG_ENUM;
        f.name = nextJvId();

        f.assign =
            JvFuncCall.create(
                new JvAccess( JV_QNAME_MG_ENUM, JV_ID_CREATE ),
                new JvCast( 
                    JV_QNAME_ATOMIC_TYPE_REF,
                    idMap.idFor( 
                        AtomicTypeReference.create( typeDef().getName() ) )
                ),
                idMap.idFor( id )
            );
        
        en.fields.add( f );
        mgEnumConstants.put( id, f.name );
    }

    private
    void
    initMappings( JvEnumDecl en )
    {
        for ( MingleIdentifier nm : typeDef().getNames() )
        {
            enMappings.add(
                new EnMapping(
                    new JvId( MingleModels.asJavaEnumName( nm ) ),
                    nm
                )
            );

            addMgConstant( nm, en );
        }
    }

    private
    JvExpression
    jvEnValExpression( JvExpression left,
                       EnMapping em )
    {
        return new JvAccess( left, em.jvId );
    }

    private
    void
    addBindingConstructor( JvClass cls,
                           JvEnumDecl en )
    {
        JvConstructor c = new JvConstructor();
        c.vis = JvVisibility.PRIVATE;

        c.body.add(
            new JvStatement(
                JvFuncCall.create(
                    JvId.SUPER,
                    new JvAccess( en.name, JvId.CLASS )
                )
            )
        );

        cls.constructors.add( c );
    }

    private
    void
    addGetJavaEnumBody( JvEnumDecl en,
                        JvMethod m,
                        JvId mgValId )
    {
        JvBranch b = new JvBranch();
        m.body.add( b );

        for ( Iterator< EnMapping > it = enMappings.iterator(); it.hasNext(); )
        {
            EnMapping em = it.next();

            b.test = 
                JvFuncCall.create(
                    new JvAccess( idMap.idFor( em.mgId ), JV_ID_EQUALS ),
                    mgValId
                );
            
            b.onTrue.add( new JvReturn( jvEnValExpression( en.name, em ) ) );

            if ( it.hasNext() ) b.onFalse.add( b = new JvBranch() );
            else b.onFalse.add( new JvReturn( JV_ID_NULL ) );
        }
    }

    private
    void
    addBindingGetJavaEnum( JvClass cls,
                           JvEnumDecl en )
    {
        JvMethod m = new JvMethod();
        m.mods.vis = JvVisibility.PROTECTED;
        m.retType = en.name;
        m.name = JV_ID_GET_JAVA_ENUM;

        JvId mgValId = nextJvId();
        m.params.add( JvParam.create( mgValId, JV_QNAME_MG_IDENTIFIER ) );

        addGetJavaEnumBody( en, m, mgValId );

        cls.methods.add( m );
    }

    private
    JvId
    mgEnumIdRefFor( MingleIdentifier mgId )
    {
        return state.get( mgEnumConstants, mgId, "mgEnumConstants" );
    }

    private
    void
    addGetMingleEnumBody( JvMethod m,
                          JvId jvValId )
    {
        JvSwitch s = new JvSwitch();
        s.target = jvValId;

        for ( EnMapping em : enMappings )
        {
            s.cases.add(
                JvCase.create( 
                    em.jvId, new JvReturn( mgEnumIdRefFor( em.mgId ) ) ) );
        }

        s.defl.add( new JvReturn( JV_ID_NULL ) );

        m.body.add( s );
    }

    private
    void
    addBindingGetMingleEnum( JvClass cls,
                             JvEnumDecl en )
    { 
        JvMethod m = new JvMethod();
        m.mods.vis = JvVisibility.PROTECTED;
        m.retType = JV_QNAME_MG_ENUM;
        m.name = JV_ID_GET_MINGLE_ENUM;

        JvId jvValId = nextJvId();
        m.params.add( JvParam.create( jvValId, en.name ) );
        
        addGetMingleEnumBody( m, jvValId );

        cls.methods.add( m );
    }

    private
    JvClass
    addEnumBinding( JvEnumDecl en )
    {
        JvClass cls = new JvClass();
        cls.mods.vis = JvVisibility.PRIVATE;
        cls.mods.isFinal = true;
        cls.mods.isStatic = true;
        cls.name = JV_TYPE_NAME_ENUM_BINDING_IMPL;
        cls.sprTyp = 
            JvTypeExpression.withParams(
                JV_TYPE_ABSTRACT_ENUM_BINDING, en.name );

        addBindingConstructor( cls, en );
        addBindingGetJavaEnum( cls, en );
        addBindingGetMingleEnum( cls, en );

        en.nestedTypes.add( cls );

        return cls;
    }

    private
    void
    addInitInitialize( JvEnumDecl en,
                       JvClass init,
                       JvClass binding )
    {
        JvMethod m = new JvMethod();
        m.mods.vis = JvVisibility.PUBLIC;
        m.retType = JvPrimitiveType.VOID;
        m.name = JV_ID_INITIALIZE;

        JvId bldrId = nextJvId();
        m.params.add( JvParam.create( bldrId, JV_TYPE_BINDER_BUILDER ) );
        m.params.add( JvParam.create( nextJvId(), JV_QNAME_TYPE_DEF_LOOKUP ) );

        m.body.add(
            new JvStatement(
                JvFuncCall.create(
                    new JvAccess( bldrId, JV_ID_ADD_BINDING ),
                    idMap.idFor( typeDef().getName() ),
                    JvInstantiate.create( binding.name ),
                    new JvAccess( en.name, JvId.CLASS )
                )
            )
        );

        init.methods.add( m );
    }

    private
    void
    addInitializer( JvClass binding,
                    JvEnumDecl en,
                    JvPackage jvPkg,
                    BuildResult bldRes )
    {
        JvClass init = new JvClass();
        init.mods.vis = JvVisibility.PRIVATE;
        init.mods.isStatic = true;
        init.mods.isFinal = true;
        init.name = JV_TYPE_NAME_BINDER_INITIALIZER;
        init.implemented.add( JV_TYPE_BINDERS_INITIALIZER );

        addInitInitialize( en, init, binding );

        en.nestedTypes.add( init );
        bldRes.addInitializerEntry( jvPkg, en.name, init.name );
    }
    
    private
    JvEnumDecl
    createJvEnumDecl( BuildResult bldRes,
                      JvPackage jvPkg )
    {
        JvEnumDecl res = new JvEnumDecl();
        res.name = jvTypeNameOf( typeDef() );
        res.vis = JvVisibility.PUBLIC;

        idMap = new JvIdMapper( res.fields );
        initMappings( res );

        for ( EnMapping em : enMappings ) res.constants.add( em.jvId );

        JvClass binding = addEnumBinding( res );
        addInitializer( binding, res, jvPkg, bldRes );

        return res;
    }

    BuildResult
    buildImpl()
    {
        BuildResult res = new BuildResult();

        JvCompilationUnit u = new JvCompilationUnit();
        u.pkg = jvPackageOf( typeDef() );
        u.decl = createJvEnumDecl( res, u.pkg );
        
        res.units.add( u );

        return res;
    }
}
