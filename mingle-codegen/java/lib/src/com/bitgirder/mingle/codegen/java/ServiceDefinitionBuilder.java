package com.bitgirder.mingle.codegen.java;

import static com.bitgirder.mingle.codegen.java.CodegenConstants.*;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.mingle.model.ServiceDefinition;
import com.bitgirder.mingle.model.OperationDefinition;
import com.bitgirder.mingle.model.OperationSignature;
import com.bitgirder.mingle.model.PrototypeDefinition;
import com.bitgirder.mingle.model.FieldDefinition;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleIdentifierFormat;
import com.bitgirder.mingle.model.MingleTypeName;
import com.bitgirder.mingle.model.QualifiedTypeName;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.AtomicTypeReference;
import com.bitgirder.mingle.model.TypeDefinitions;

import java.util.List;
import java.util.Set;
import java.util.Collection;
import java.util.Iterator;
import java.util.Map;

final
class ServiceDefinitionBuilder
extends TypeDefinitionBuilder< ServiceDefinition >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static JvId JV_ID_OP = new JvId( "op" );

    private final static JvId JV_ID_B = new JvId( "b" );
    private final static JvTypeName JV_TYPE_NAME_B = new JvTypeName( "B" );

    private final static JvQname JV_QNAME_MG_CALL_CTX =
        JvQname.create( 
            "com.bitgirder.mingle.service", "MingleServiceCallContext" );

    private final static JvQname JV_QNAME_BOUND_SERVICE =
        JvQname.create( "com.bitgirder.mingle.bind", "BoundService" );

    private final static JvQname JV_QNAME_BOUND_CLIENT =
        JvQname.create( "com.bitgirder.mingle.bind", "BoundServiceClient" );

    private final static JvTypeName JV_TYPE_NAME_BUILDER =
        new JvTypeName( "Builder" );

    private final static JvTypeName JV_TYPE_NAME_ABSTRACT_OPERATION =
        new JvTypeName( "AbstractOperation" );

    private final static JvTypeName JV_TYPE_NAME_OPERATION =
        new JvTypeName( "Operation" );

    private final static JvTypeName JV_TYPE_NAME_ABSTRACT_CALL =
        new JvTypeName( "AbstractCall" );

    private final static JvId JV_ID_FAIL = new JvId( "fail" );

    private final static JvId JV_ID_NEW_MAP = new JvId( "newMap" );

    private final static JvId JV_ID_IMPL_SET_TYPE_NAME =
        new JvId( "implSetTypeName" );

    private final static JvId JV_ID_IMPL_START = 
        new JvId( "implStart" );

    private final static JvId JV_ID_IMPL_GET_OPERATION =
        new JvId( "implGetOperation" );

    private final static JvTypeName JV_TYPE_NAME_INPUT =
        new JvTypeName( "Input" );

    private final static JvId JV_ID_AUTH_EXECUTOR = new JvId( "authExecutor" );

    private final static JvId JV_ID_SET_AUTH_EXECUTOR =
        new JvId( "setAuthExecutor" );

    private final static JvId JV_ID_IMPL_SET_AUTH_EXECUTOR =
        new JvId( "implSetAuthExecutor" );
    
    private final static JvId JV_ID_IMPL_VALIDATE_AUTH_EXECUTOR =
        new JvId( "implValidateAuthExecutor" );

    private final static JvId JV_ID_IMPL_CREATE_AUTH_CONTEXT =
        new JvId( "implCreateAuthenticationContext" );

    private final static JvTypeName JV_TYPE_NAME_AUTH_CTX =
        new JvTypeName( "AuthenticationContext" );

    private final static JvId JV_ID_IMPL_CREATE_AUTH_INPUT =
        new JvId( "implCreateAuthInput" );

    private final static JvId JV_ID_IMPL_GET_JAVA_AUTH_VALUE =
        new JvId( "implGetJavaAuthValue" );

    private final static JvId JV_ID_AUTH_INPUT = new JvId( "authInput" );

    private final static JvId JV_ID_IMPL_AUTH_INPUT = 
        new JvId( "implAuthInput" );

    private final static JvId JV_ID_AUTH_RESULT = new JvId( "authResult" );

    private final static JvId JV_ID_IMPL_AUTH_RESULT =
        new JvId( "implAuthResult" );

    private final static JvQname JV_QNAME_RPC_SERVER =
        JvQname.create( "com.bitgirder.process", "ProcessRpcServer" );

    private final static JvQname JV_QNAME_MG_SVC_RESP =
        JvQname.create( "com.bitgirder.mingle.model", "MingleServiceResponse" );

    private final static JvTypeName JV_TYPE_NAME_RESP_CTX =
        new JvTypeName( "ResponderContext" );
    
    private final static JvType JV_TYPE_RPC_RESP_CTX =
        JvTypeExpression.dotTypeOf(
            JV_QNAME_RPC_SERVER, JV_TYPE_NAME_RESP_CTX );

    private final static JvType JV_TYPE_MG_RESP_RPC_RESP_CTX =
        JvTypeExpression.withParams(
            JV_TYPE_RPC_RESP_CTX, JV_QNAME_MG_SVC_RESP );

    private final static JvId JV_ID_GET_OPERATION = new JvId( "getOperation" );

    private final static JvId JV_ID_SET_OPERATION = new JvId( "setOperation" );
    
    private final static JvId JV_ID_GET_REQUEST = new JvId( "getRequest" );

    private final static JvId JV_ID_IS_ROUTE_MATCH = new JvId( "isRouteMatch" );

    private final static JvId JV_ID_AS_JAVA_VALUE = new JvId( "asJavaValue" );

    private final static JvId JV_ID_BUILD = new JvId( "build" );

    private final static JvId JV_ID_SET_NAMESPACE = new JvId( "setNamespace" );

    private final static JvId JV_ID_SET_SVC_ID = new JvId( "setServiceId" );

    private final static JvId JV_ID_IMPL_SET_AUTH_INPUT_TYPE =
        new JvId( "implSetAuthInputType" );

    private final static JvId JV_ID_SET_AUTHENTICATION =
        new JvId( "setAuthentication" );

    private final static JvId JV_ID_IMPL_SET_AUTHENTICATION =
        new JvId( "implSetAuthentication" );

    private final static JvId JV_ID_IMPL_SET_PARAMETERS =
        new JvId( "implSetParameters" );

    private final static JvId JV_ID_SET_PARAMETER =
        new JvId( "setParameter" );

    private AuthTypeContext authTypCtx;
    private final List< JvOpContext > opCtxs = Lang.newList();

    private
    final
    static
    class JvOpContext
    {
        JvId name;
        JvType retType;
        final List< JvField > jvFields = Lang.newList();
        OperationDefinition opDef;
        OperationSignature sig; // stored standalone for conciseness elsewhere
        OperationGeneratorParameters opGenParams; // maybe null

        private
        CharSequence
        upcaseName()
        {
            String s = name.toString();

            StringBuilder res = new StringBuilder();
            res.append( Character.toUpperCase( s.charAt( 0 ) ) );

            if ( s.length() > 0 ) res.append( s, 1, s.length() );
            
            return res;
        }

        JvTypeName
        getOpClassName()
        {
            return new JvTypeName( upcaseName() );
        }

        JvId
        getOpImplStartMethodName()
        {
            return new JvId( "start" );
        }

        boolean
        useOpaqueReturnType()
        {
            return opGenParams != null && opGenParams.useOpaqueJavaReturnType;
        }
    }

    private
    JvType
    jvOpReturnType( OperationSignature sig,
                    OperationGeneratorParameters opGenParams )
    {
        if ( opGenParams != null && opGenParams.useOpaqueJavaReturnType )
        {
            return jvOpaqueType();
        }
        else return jvTypeOf( sig.getReturnType(), false );
    }

    private
    JvOpContext
    createJvOpContext( OperationDefinition opDef,
                       OperationGeneratorParameters opGenParams )
    {
        JvOpContext res = new JvOpContext();
        res.name = asJvId( opDef.getName() );

        res.sig = opDef.getSignature();
        res.retType = jvOpReturnType( res.sig, opGenParams );
        res.opDef = opDef;
        res.opGenParams = opGenParams;

        Map< MingleIdentifier, FieldGeneratorParameters > m = 
            opGenParams == null 
                ? null : fieldParamsByName( opGenParams.fldParams );

        for ( FieldDefinition fd : res.sig.getFieldSet().getFields() )
        {
            FieldGeneratorParameters fgp = 
                m == null ? null : m.get( fd.getName() );

            res.jvFields.add( asJvField( fd, fgp ) );
        }
        
        return res;
    }

    private
    final
    static
    class AuthTypeContext
    {
        private final MingleTypeReference authInType;
        private final JvType authOutType;
        private final PrototypeDefinition protoDef;

        private
        AuthTypeContext( MingleTypeReference authInType,
                         JvType authOutType,
                         PrototypeDefinition protoDef )
        {
            this.authInType = authInType;
            this.authOutType = authOutType;
            this.protoDef = protoDef;
        }

        private OperationSignature sig() { return protoDef.getSignature(); }
    }

    private
    PrototypeDefinition
    getSecurity()
    {
        QualifiedTypeName secRef = typeDef().getSecurity();

        if ( secRef == null ) return null;
        else
        {
            return
                (PrototypeDefinition) 
                    context().runtime().getTypes().expectType( secRef );
        }
    }

    private
    void
    setAuthTypeContext()
    {
         PrototypeDefinition protoDef = getSecurity();
        
        if ( protoDef != null )
        {
            OperationSignature sig = protoDef.getSignature();

            MingleTypeReference authInType = 
                TypeDefinitions.expectAuthInputType( sig );

            // set isOuterMost to false since authOutType will actually be a
            // type parameter, not a standalone type (aka, we don't want
            // primitives but rather their boxed versions)
            JvType authOutType = jvTypeOf( sig.getReturnType(), false );

            authTypCtx = 
                new AuthTypeContext( authInType, authOutType, protoDef );
        }
    }

    private
    Map< MingleIdentifier, OperationGeneratorParameters >
    getOpGenParamMap()
    {
        ServiceGeneratorParameters sp = genParams();

        if ( sp == null ) return Lang.emptyMap();
        else return sp.opParamsByName();
    }

    private
    void
    init()
    {
        setAuthTypeContext();

        Map< MingleIdentifier, OperationGeneratorParameters > m =
            getOpGenParamMap();

        for ( OperationDefinition opDef : typeDef().getOperations() )
        {
            opCtxs.add( createJvOpContext( opDef, m.get( opDef.getName() ) ) );
        }
    }

    private
    abstract
    class AbstractBuilder
    {
        final JvClass cls = new JvClass();
        final JvIdMapper idMap = new JvIdMapper( cls );

        final
        JvId
        nsIdFieldId()
        {
            return idMap.idFor( typeDef().getName().getNamespace() );
        }

        final
        JvId
        svcIdFieldId()
        {
            CharSequence nm = 
                typeDef().getName().getName().get( 0 ).getExternalForm();

            StringBuilder sb = new StringBuilder( nm.length() );
            sb.append( Character.toLowerCase( nm.charAt( 0 ) ) );
            if ( nm.length() > 0 ) sb.append( nm, 1, nm.length() );

            MingleIdentifier id = MingleIdentifier.create( sb );
    
            return idMap.idFor( id ); // ensures our field add
        }
    }

    private
    ServiceGeneratorParameters
    genParams()
    {
        return getGeneratorParameters( ServiceGeneratorParameters.class );
    }

    private
    final
    class ServiceBuilder
    extends AbstractBuilder
    {
        private
        JvType
        getSuperClass()
        {
            ServiceGeneratorParameters gp = genParams();

            JvType res = JV_QNAME_BOUND_SERVICE;

            if ( gp != null )
            {
                JvType bsCls = gp.getServiceBaseClass();
                if ( bsCls != null ) res = bsCls;
            }

            return state.notNull( res );
        }

        private
        void
        addPublicConstFields()
        {
            Set< MingleIdentifier > s = Lang.newSet();

            for ( JvOpContext opCtx : opCtxs )
            {
                for ( JvField f : opCtx.jvFields )
                {
                    MingleIdentifier id = f.mgField.getName();

                    if ( s.add( id ) )
                    {
                        JvField idFld = JvField.createConstField();
                        idFld.mods.vis = JvVisibility.PUBLIC;
                        idFld.name = new JvId( "ID_" + upcaseIdent( id ) );
                        idFld.type = JV_QNAME_MG_IDENTIFIER;
                        idFld.assign = idMap.idFor( id );

                        cls.fields.add( idFld );
                    }
                }
            }
        }

        private
        void
        addOptValidateAuthExecutor( JvConstructor c )
        {
            if ( authTypCtx != null )
            {
                c.body.add(
                    new JvStatement(
                        JvFuncCall.create( 
                            JV_ID_IMPL_VALIDATE_AUTH_EXECUTOR, JV_ID_B ) ) );
            }
        }

        private
        void
        addImplConstructor()
        {
            JvConstructor c = new JvConstructor();
            c.vis = JvVisibility.PROTECTED;
    
            c.params.add(
                JvParam.create(
                    JV_ID_B,
                    JvTypeExpression.withParams(
                        JV_TYPE_NAME_BUILDER, JV_TYPE_WILDCARD )
                )
            );
 
            c.body.add( 
                new JvStatement( JvFuncCall.create( JvId.SUPER, JV_ID_B ) ) );

            addOptValidateAuthExecutor( c );
 
            cls.constructors.add( c );
        }
 
        private
        void
        initClass()
        {
            cls.name = new JvTypeName( "Abstract" + jvTypeNameOf( typeDef() ) );
            cls.mods.vis = JvVisibility.PUBLIC;
            cls.mods.isFinal = false;
            cls.mods.isAbstract = true;
            cls.sprTyp = getSuperClass();
    
            addImplConstructor();
            addPublicConstFields();
        }

        private
        void
        addBuilderTypeInfo( JvClass bldrCls )
        {
            bldrCls.typeParams.add(
                new JvTypeExtendParameter(
                    JV_TYPE_NAME_B,
                    JvTypeExpression.withParams(
                        JV_TYPE_NAME_BUILDER, JV_TYPE_NAME_B )
                )
            );

            bldrCls.sprTyp =
                JvTypeExpression.withParams(
                    JvTypeExpression.dotTypeOf(
                        getSuperClass(), JV_TYPE_NAME_BUILDER ),
                    JV_TYPE_NAME_B
                );
        }

        private
        void
        addBuilderConstructor( JvClass bldrCls )
        {
            JvConstructor cons = new JvConstructor();
            cons.vis = JvVisibility.PUBLIC;

            cons.body.add(
                new JvStatement(
                    JvFuncCall.create( 
                        JV_ID_IMPL_SET_TYPE_NAME,
                        idMap.idFor( typeDef().getName() ) ) ) );

            cons.body.add(
                new JvStatement(
                    JvFuncCall.create( JV_ID_SET_NAMESPACE, nsIdFieldId() ) ) );
            
            cons.body.add(
                new JvStatement(
                    JvFuncCall.create( JV_ID_SET_SVC_ID, svcIdFieldId() ) ) );
            
            bldrCls.constructors.add( cons );
        }

        private
        JvType
        authProtoType( JvTypeName subType )
        {
            return
                JvTypeExpression.dotTypeOf(
                    jvTypeOf( authTypCtx.protoDef.getName() ), subType );
        }

        private
        JvType
        authExecType()
        {
            return 
                JvTypeExpression.withParams(
                    JV_QNAME_PROCESS_OPERATION,
                    authProtoType( JV_TYPE_NAME_INPUT ),
                    authTypCtx.authOutType
                );
        }

        private
        void
        addBuilderAuthExecSetter( JvClass bldrCls )
        {
            JvMethod m = new JvMethod();
            m.mods.vis = JvVisibility.PUBLIC;
            m.retType = JV_TYPE_NAME_B;
            m.name = JV_ID_SET_AUTH_EXECUTOR;

            JvType execTyp = authExecType();
            JvParam p = JvParam.create( JV_ID_AUTH_EXECUTOR, execTyp );
            m.params.add( p );

            m.body.add(
                new JvReturn(
                    JvFuncCall.create( 
                        JV_ID_IMPL_SET_AUTH_EXECUTOR, 
                        idMap.jvNotNull( p.id ) 
                    )
                )
            );
 
            bldrCls.methods.add( m );
        }

        private
        void
        addBuilder()
        {
            JvClass bldrCls = new JvClass();
            bldrCls.mods.vis = JvVisibility.PUBLIC;
            bldrCls.mods.isFinal = false;
            bldrCls.mods.isStatic = true;
            bldrCls.name = JV_TYPE_NAME_BUILDER;
            addBuilderTypeInfo( bldrCls );
            addBuilderConstructor( bldrCls );
            if ( authTypCtx != null ) addBuilderAuthExecSetter( bldrCls );

            cls.nestedTypes.add( bldrCls );
        }

        private
        JvId
        addLocalAuthValAssign( JvId opId,
                               JvMethod m )
        {
            JvLocalVar v = new JvLocalVar();

            v.name = nextJvId();

            v.type = jvTypeOf( authTypCtx.authInType );

            v.assign = 
                JvFuncCall.create(
                    JV_ID_IMPL_GET_JAVA_AUTH_VALUE, 
                    idMap.idFor( authTypCtx.authInType ),
                    opId
                );

            m.body.add( v );

            return v.name;
        }

        private
        JvExpression
        addCreateAuthInputReturn( JvId authValId )
        {
            return
                new JvReturn(
                    JvFuncCall.create(
                        new JvAccess(
                            authProtoType( JV_TYPE_NAME_INPUT ),
                            JV_ID_CREATE
                        ),
                        authValId
                    )
                );
        }

        private
        void
        addImplCreateAuthInput()
        {
            JvMethod m = new JvMethod();
            m.anns.add( JvAnnotation.OVERRIDE );
            m.mods.vis = JvVisibility.PROTECTED;
            m.mods.isFinal = true;
            m.retType = authProtoType( JV_TYPE_NAME_INPUT ); 
            m.name = JV_ID_IMPL_CREATE_AUTH_INPUT;

            JvType opType = 
                JvTypeExpression.withParams(
                    JV_TYPE_NAME_ABSTRACT_OPERATION, JV_TYPE_WILDCARD );

            JvParam p = JvParam.create( nextJvId(), opType );
            m.params.add( p );

            JvId authValId = addLocalAuthValAssign( p.id, m );
            m.body.add( addCreateAuthInputReturn( authValId ) );

            cls.methods.add( m );
        }

        private
        void
        addOptAuthProcessing()
        {
            if ( authTypCtx != null ) addImplCreateAuthInput();
        }

        // this adds and returns the abstract start( Op ) method that user code
        // will implement; the return value is used by addOpImplStartMethod to
        // actually make the call defined here
        private
        JvMethod
        addOpStartMethod( JvOpContext opCtx )
        {
            JvMethod m = new JvMethod();
            m.mods.isAbstract = true;
            m.mods.isFinal = false;
            m.mods.vis = JvVisibility.PROTECTED;
            m.name = opCtx.getOpImplStartMethodName();
            m.retType = JvPrimitiveType.VOID;
            m.thrown.add( JV_QNAME_JLANG_EXCEPTION );
    
            m.params.add( JvParam.create( JV_ID_OP, opCtx.getOpClassName() ) );
    
            cls.methods.add( m );
            return m;
        }

        private
        void
        addOperationBaseClassConstructor( JvClass c )
        {
            JvConstructor cons = new JvConstructor();
            cons.vis = JvVisibility.PRIVATE;

            JvFuncCall call = new JvFuncCall();
            call.target = JvId.SUPER;

            call.params.add( addParam( cons, JV_QNAME_MG_CALL_CTX ) );
            call.params.add( addParam( cons, JV_TYPE_MG_RESP_RPC_RESP_CTX ) );
            call.params.add( addParam( cons, JV_QNAME_MG_TYPE_REF ) );
            call.params.add( addParam( cons, JvPrimitiveType.BOOLEAN ) );
            call.params.add( addParam( cons, JV_TYPE_PROC_ACTIVITY_CTX ) );
            
            cons.body.add( new JvStatement( call ) );

            c.constructors.add( cons );
        }

        private
        JvClass
        addOperationBaseClass()
        {
            JvClass bsCls = new JvClass();
            bsCls.mods.vis = JvVisibility.PROTECTED;
            bsCls.mods.setAbstract();
            bsCls.name = JV_TYPE_NAME_OPERATION;

            JvTypeName typeArg = new JvTypeName( "V" );

            bsCls.sprTyp = 
                JvTypeExpression.withParams(
                    JV_TYPE_NAME_ABSTRACT_OPERATION, typeArg );

            bsCls.typeParams.add( typeArg );

            addOperationBaseClassConstructor( bsCls );

            cls.nestedTypes.add( bsCls );

            return bsCls;
        }

        private
        void
        addOpImplStartMethod( JvClass opCls,
                              JvMethod opMeth )
        {
            JvMethod m = new JvMethod();
            m.mods.vis = JvVisibility.PROTECTED;
            m.mods.isFinal = true;
            m.retType = JvPrimitiveType.VOID;
            m.name = JV_ID_IMPL_START;

            m.body.add(
                new JvStatement( JvFuncCall.create( opMeth.name, JvId.THIS ) )
            );

            m.thrown.add( JV_QNAME_JLANG_EXCEPTION );

            opCls.methods.add( m );
        }
    
        private
        void
        addOpImplFields( JvClass opCls, 
                         JvOpContext opCtx )
        {
            for ( JvField f : opCtx.jvFields )
            {
                f = f.copyOf();
                f.mods.vis = JvVisibility.PUBLIC;
                f.mods.isFinal = true;
    
                opCls.fields.add( f );
            }
        }
    
        private
        void
        addOpImplConstructorFieldAssigns( JvConstructor cons,
                                          JvId callCtxId,
                                          JvOpContext opCtx )
        {
            for ( JvField f : opCtx.jvFields )
            {
                cons.body.add(
                    new JvStatement(
                        new JvAssign(
                            new JvAccess( JvId.THIS, f.name ),
                            optUnbox(
                                f.type,
                                JvFuncCall.create(
                                    JV_ID_AS_JAVA_VALUE,
                                    idMap.idFor( f.mgField ),
                                    callCtxId,
                                    expBoolLiteral( useOpaqueJavaType( f ) )
                                )
                            )
                        )
                    )
                );
            }
        }

        private
        JvQname
        expectExceptionQname( MingleTypeReference ref )
        {
            // If we were doing type params, we'd also check here that there are
            // none unless we supporting mingle exceptions with type
            state.isTrue( ref instanceof AtomicTypeReference );

            return (JvQname) jvTypeOf( ref );
        }

        private
        void
        addOpImplConstructorSuperCall( JvConstructor cons,
                                       JvId callCtxId,
                                       JvId respCtxId,
                                       JvId procCtxId,
                                       JvOpContext opCtx )
        {
            cons.body.add(
                new JvStatement( 
                    JvFuncCall.create( 
                        JvId.SUPER, 
                        callCtxId, 
                        respCtxId,
                        idMap.idFor( opCtx.sig.getReturnType() ),
                        expBoolLiteral( opCtx.useOpaqueReturnType() ),
                        procCtxId
                    )
                )
            );
        }
    
        private
        void
        addOpImplConstructor( JvClass opCls,
                              JvOpContext opCtx )
        {
            JvConstructor cons = new JvConstructor();
            cons.vis = JvVisibility.PRIVATE;
            
            JvId callCtxId = addParam( cons, JV_QNAME_MG_CALL_CTX );
            JvId respCtxId = addParam( cons, JV_TYPE_MG_RESP_RPC_RESP_CTX );
            JvId procCtxId = addParam( cons, JV_TYPE_PROC_ACTIVITY_CTX );
    
            addOpImplConstructorSuperCall( 
                cons, callCtxId, respCtxId, procCtxId, opCtx );

            addOpImplConstructorFieldAssigns( cons, callCtxId, opCtx );
    
            opCls.constructors.add( cons );
        }

        private
        void
        addAuthInputGetter( JvClass targCls )
        {
            JvMethod m = new JvMethod();
            m.mods.vis = JvVisibility.PUBLIC;
            m.retType = jvTypeOf( authTypCtx.authInType );
            m.name = JV_ID_AUTH_INPUT;

            m.body.add(
                new JvReturn( JvFuncCall.create( JV_ID_IMPL_AUTH_INPUT ) ) );
            
            targCls.methods.add( m );
        }

        private
        void
        addAuthResultGetter( JvClass targCls )
        {
            JvMethod m = new JvMethod();
            m.mods.vis = JvVisibility.PUBLIC;
            m.retType = authTypCtx.authOutType;
            m.name = JV_ID_AUTH_RESULT;

            m.body.add(
                new JvReturn( JvFuncCall.create( JV_ID_IMPL_AUTH_RESULT ) ) );
            
            targCls.methods.add( m );
        }

        private
        void
        addOptAuthAccessors( JvClass targCls )
        {
            if ( authTypCtx != null )
            {
                addAuthInputGetter( targCls );
                addAuthResultGetter( targCls );
            }
        }

        private
        JvType
        getOpClassSuperType( JvOpContext opCtx )
        {
            return
                JvTypeExpression.withParams(
                    JV_TYPE_NAME_OPERATION, opCtx.retType
                );
        }
 
        private
        void
        addJvOpImpl( JvOpContext opCtx )
        {
            JvClass opCls = new JvClass();
    
            opCls.name = new JvTypeName( opCtx.getOpClassName() );
            opCls.mods.isStatic = false;
            opCls.mods.isFinal = true;
            opCls.mods.vis = JvVisibility.PROTECTED;
            opCls.sprTyp = getOpClassSuperType( opCtx );
    
            addOpImplFields( opCls, opCtx );
            addOpImplConstructor( opCls, opCtx );
 
            cls.nestedTypes.add( opCls );
            JvMethod opMeth = addOpStartMethod( opCtx );
            addOpImplStartMethod( opCls, opMeth );
        }

        private 
        void 
        processImplOpCtx( JvOpContext opCtx ) 
        { 
            addJvOpImpl( opCtx ); 
        }
    
        private
        final
        class ImplStartRequestBuilder
        {
            private final JvMethod m = new JvMethod();
            
            private final JvId callCtxId = nextJvId();
            private final JvId respCtxId = nextJvId();
            private final JvId procCtxId = nextJvId();
    
            private ImplStartRequestBuilder() {}
    
            private
            JvExpression
            createOpIdTest( JvOpContext opCtx )
            {
                return
                    JvFuncCall.create(
                        JV_ID_IS_ROUTE_MATCH,
                        callCtxId,
                        idMap.idFor( opCtx.opDef.getName() )
                    );
            }
    
            private
            void
            addOpDispatch( List< JvExpression > block,
                           JvOpContext opCtx )
            {
                JvInstantiate inst = new JvInstantiate();
                inst.target = opCtx.getOpClassName();
                inst.params.add( callCtxId );
                inst.params.add( respCtxId );
                inst.params.add( procCtxId );
    
                block.add( new JvReturn( inst ) );
            }
    
            private
            void
            addHandleNoSuchOp( List< JvExpression > block )
            {
                block.add( new JvReturn( JV_ID_NULL ) );
            }

            private
            void
            addNontrivialBody( Iterator< JvOpContext > it )
            {
                JvBranch root = new JvBranch();
                JvBranch cur = root;
                do
                {
                    JvOpContext opCtx = it.next();
                    cur.test = createOpIdTest( opCtx );
                    addOpDispatch( cur.onTrue, opCtx );
    
                    if ( it.hasNext() )
                    {
                        JvBranch next = new JvBranch();
                        cur.onFalse.add( next );
                        cur = next;
                    }
                    else addHandleNoSuchOp( cur.onFalse );
                }
                while ( it.hasNext() );
    
                m.body.add( root );
            }

            private
            void
            addTrivialBody()
            {
                m.body.add( new JvReturn( JV_ID_NULL ) );
            }

            private
            void
            addBody()
            {
                Iterator< JvOpContext > it = opCtxs.iterator();

                if ( it.hasNext() ) addNontrivialBody( it ); 
                else addTrivialBody();
            }
    
            private 
            void
            build()
            {
                m.mods.vis = JvVisibility.PROTECTED;
                m.mods.isFinal = true;
                m.retType = 
                    JvTypeExpression.withParams(
                        JV_TYPE_NAME_ABSTRACT_OPERATION, JV_TYPE_WILDCARD );

                m.name = JV_ID_IMPL_GET_OPERATION;

                m.params.add( 
                    JvParam.create( callCtxId, JV_QNAME_MG_CALL_CTX ) );

                m.params.add(
                    JvParam.create( respCtxId, JV_TYPE_MG_RESP_RPC_RESP_CTX ) );

                m.params.add(
                    JvParam.create( procCtxId, JV_TYPE_PROC_ACTIVITY_CTX ) );
        
                addBody();
        
                cls.methods.add( m );
            }
        }
    
        private
        JvCompilationUnit
        build()
        {
            initClass();
            addBuilder();
            new ImplStartRequestBuilder().build();
                
            JvClass bsCls = addOperationBaseClass();
            addOptAuthProcessing();
            addOptAuthAccessors( bsCls );
    
            if ( ! opCtxs.isEmpty() )
            {
                for ( JvOpContext opCtx : opCtxs ) processImplOpCtx( opCtx );
            }
    
            JvCompilationUnit res = new JvCompilationUnit();
            res.pkg = jvPackageOf( typeDef() );
            res.decl = cls;
    
            return res;
        }
    }

    private
    final
    class ClientBuilder
    extends AbstractBuilder
    {
        private
        JvTypeName
        getCallClassName( JvOpContext opCtx )
        {
            return new JvTypeName( opCtx.upcaseName() + "Call" );
        }

        private
        void
        initCls()
        {
            cls.mods.vis = JvVisibility.PUBLIC;
            cls.mods.isFinal = true;
            cls.name = new JvTypeName( jvTypeNameOf( typeDef() ) + "Client" );
            cls.sprTyp = JV_QNAME_BOUND_CLIENT;
        }

        private
        void
        addConstructor()
        {
            JvConstructor cons = new JvConstructor();
            cons.vis = JvVisibility.PRIVATE;

            cons.params.add( JvParam.create( JV_ID_B, JV_TYPE_NAME_BUILDER ) );

            cons.body.add(
                new JvStatement( JvFuncCall.create( JvId.SUPER, JV_ID_B ) ) );
            
            cls.constructors.add( cons );
        }

        private
        void
        addBuilderTypeInfo( JvClass bldrCls )
        {
            bldrCls.sprTyp = 
                JvTypeExpression.withParams(
                    JvTypeExpression.dotTypeOf(
                        JV_QNAME_BOUND_CLIENT, JV_TYPE_NAME_BUILDER ),
                    cls.name,
                    JV_TYPE_NAME_BUILDER
                );
        } 

        private
        void
        addImplSetAuthInputTypeCall( JvConstructor cons )
        {
            cons.body.add(
                new JvStatement(
                    JvFuncCall.create(
                        JV_ID_IMPL_SET_AUTH_INPUT_TYPE,
                        idMap.idFor( authTypCtx.authInType )
                    )
                )
            );
        }

        private
        void
        addBuilderConstructor( JvClass bldrCls )
        {
            JvConstructor cons = new JvConstructor();
            cons.vis = JvVisibility.PUBLIC;

            cons.body.add(
                new JvStatement(
                    JvFuncCall.create( JV_ID_SET_NAMESPACE, nsIdFieldId() ) ) );
            
            cons.body.add(
                new JvStatement(
                    JvFuncCall.create( JV_ID_SET_SVC_ID, svcIdFieldId() ) ) );
            
            if ( authTypCtx != null ) addImplSetAuthInputTypeCall( cons );

            bldrCls.constructors.add( cons );
        }

        private
        JvMethod
        createSetAuthMethod( JvType authType,
                             JvType methRetType )
        {
            JvMethod m = new JvMethod();
            m.mods.vis = JvVisibility.PUBLIC;
            m.retType = methRetType;
            m.name = JV_ID_SET_AUTHENTICATION;

            JvParam p = JvParam.create( new JvId( "auth" ), authType );
            m.params.add( p );

            m.body.add(
                new JvReturn(
                    JvFuncCall.create(
                        JV_ID_IMPL_SET_AUTHENTICATION, 
                        p.id,
                        new JvString( p.id ) 
                    )
                )
            );
            
            return m;
        }

        private
        void
        addOptAuthSetter( JvClass targCls )
        {
            if ( authTypCtx != null )
            {
                JvType authTyp = jvTypeOf( authTypCtx.authInType );
                JvMethod m = createSetAuthMethod( authTyp, targCls.name );

                targCls.methods.add( m );
            }
        }

        private
        void
        addBuilderBuildMethod( JvClass bldrCls )
        {
            JvMethod build = new JvMethod();
            build.mods.vis = JvVisibility.PUBLIC;
            build.retType = cls.name;
            build.name = JV_ID_BUILD;

            JvInstantiate instantiate = new JvInstantiate();
            instantiate.target = cls.name;
            instantiate.params.add( JvId.THIS );
            build.body.add( new JvReturn( instantiate ) );

            bldrCls.methods.add( build );
        }

        private
        void
        addBuilder()
        {
            JvClass bldrCls = new JvClass();
            bldrCls.mods.vis = JvVisibility.PUBLIC;
            bldrCls.mods.isFinal = true;
            bldrCls.mods.isStatic = true;
            bldrCls.name = JV_TYPE_NAME_BUILDER;
            addBuilderTypeInfo( bldrCls );
            addBuilderConstructor( bldrCls );
            addOptAuthSetter( bldrCls );
            addBuilderBuildMethod( bldrCls );

            cls.nestedTypes.add( bldrCls );
        }

        private
        void
        addOpCallClassTypeInfo( JvClass opCls,
                                JvOpContext opCtx )
        {
            opCls.sprTyp = 
                JvTypeExpression.withParams(
                    JvTypeExpression.dotTypeOf(
                        JV_QNAME_BOUND_CLIENT, JV_TYPE_NAME_ABSTRACT_CALL ),
                    opCtx.retType,
                    opCls.name
                );
        }

        private
        void
        addOpCallClassConstructorSuperCall( JvConstructor cons,
                                            JvOpContext opCtx )
        {
            cons.body.add(
                new JvStatement(
                    JvFuncCall.create(
                        JvId.SUPER,
                        idMap.idFor( opCtx.sig.getReturnType() ),
                        expBoolLiteral( opCtx.useOpaqueReturnType() )
                    )
                )
            );
        }

        private
        void
        addOpCallClassConstructor( JvClass opCls,
                                   JvOpContext opCtx )
        {
            JvConstructor cons = new JvConstructor();
            cons.vis = JvVisibility.PRIVATE;

            addOpCallClassConstructorSuperCall( cons, opCtx );

            cons.body.add(
                new JvStatement(
                    JvFuncCall.create(
                        JV_ID_SET_OPERATION,
                        idMap.idFor( opCtx.opDef.getName() )
                    )
                )
            );

            opCls.constructors.add( cons );
        }

        private
        void
        addOpCallFieldSetter( JvField f,
                              JvClass opCls )
        {
            JvMethod m = new JvMethod();
            m.mods.vis = JvVisibility.PUBLIC;
            m.retType = opCls.name;
            m.name = createMethodName( "set", f.name );
            m.params.add( JvParam.forField( f ) );

            m.body.add(
                new JvStatement(
                    new JvAssign( 
                        new JvAccess( JvId.THIS, f.name ), 
                        JvFuncCall.create(
                            new JvAccess( 
                                JV_QNAME_MINGLE_BINDERS, 
                                JV_ID_VALIDATE_FIELD_VALUE 
                            ),
                            f.name,
                            idMap.idFor( f.mgField.getType() ),
                            idMap.jvFldPathFor( f.name )
                        )
                    )
                )
            );

            m.body.add( new JvReturn( JvId.THIS ) );
            
            opCls.methods.add( m );
        }

        private
        void
        addOpCallFields( JvClass opCls,
                         JvOpContext opCtx )
        {
            for ( JvField f : opCtx.jvFields )
            {
                f = f.copyOf();
                f.mods.vis = JvVisibility.PRIVATE;
                f.mods.isFinal = false;

                if ( f.type instanceof JvPrimitiveType ) 
                {
                    f.type = ( (JvPrimitiveType) f.type ).boxed;
                }

                opCls.fields.add( f );
                addOpCallFieldSetter( f, opCls );
            }
        }

        private
        void
        addOpCallImplSetFieldsBody( JvMethod m,
                                    JvOpContext opCtx,
                                    JvId mpBldrId,
                                    JvId mbId,
                                    JvId pathId )
        {
            for ( JvField f : opCtx.jvFields )
            {
                m.body.add(
                    new JvStatement(
                        JvFuncCall.create(
                            JV_ID_SET_PARAMETER,
                            mpBldrId,
                            idMap.idFor( f.mgField ),
                            new JvString( f.name ),
                            f.name,
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
        addOpCallImplSetFields( JvClass opCls,
                                JvOpContext opCtx )
        {
            JvMethod m = new JvMethod();
            m.mods.vis = JvVisibility.PROTECTED;
            m.name = JV_ID_IMPL_SET_PARAMETERS;
            m.retType = JvPrimitiveType.VOID;

            JvId mpBldrId = nextJvId();
            JvId mbId = nextJvId();
            JvId pathId = nextJvId();

            m.params.add( 
                JvParam.create( mpBldrId, JV_QNAME_SYM_MAP_BUILDER ) );
            
            m.params.add( JvParam.create( mbId, JV_QNAME_MINGLE_BINDER ) );
            m.params.add( JvParam.create( pathId, JV_TYPE_OBJ_PATH_STRING ) );

            addOpCallImplSetFieldsBody( m, opCtx, mpBldrId, mbId, pathId );

            opCls.methods.add( m );
        }

        private
        JvClass
        addOpCallClass( JvOpContext opCtx )
        {
            JvClass opCls = new JvClass();
            opCls.mods.vis = JvVisibility.PUBLIC;
            opCls.mods.isFinal = true;
            opCls.name = getCallClassName( opCtx );
            addOpCallClassTypeInfo( opCls, opCtx );
            addOpCallClassConstructor( opCls, opCtx );
            addOpCallFields( opCls, opCtx );
            addOpCallImplSetFields( opCls, opCtx );
            addOptAuthSetter( opCls );

            cls.nestedTypes.add( opCls );
            return opCls;
        }

        private
        void
        addOpCallMethod( JvClass opCls,
                         JvOpContext opCtx )
        {
            JvMethod m = new JvMethod();
            m.mods.vis = JvVisibility.PUBLIC;
            m.retType = opCls.name;
            m.name = opCtx.name;

            JvInstantiate inst = new JvInstantiate();
            inst.target = opCls.name;
            m.body.add( new JvReturn( inst ) );

            cls.methods.add( m );
        }

        private
        void
        addCalls()
        {
            for ( JvOpContext opCtx : opCtxs )
            {
                JvClass opCls = addOpCallClass( opCtx );
                addOpCallMethod( opCls, opCtx );
            }
        }

        private
        JvCompilationUnit
        build()
        {
            initCls();
            addConstructor();
            addBuilder();
            addCalls();

            JvCompilationUnit res = new JvCompilationUnit();
            res.pkg = jvPackageOf( typeDef() );
            res.decl = cls;
    
            return res;
        }
    }

    BuildResult
    buildImpl()
    {
        init();
        
        BuildResult res = new BuildResult();

        res.units.add( new ServiceBuilder().build() );
        res.units.add( new ClientBuilder().build() );

        return res;
    }
}
