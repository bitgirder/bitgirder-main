package com.bitgirder.mingle.bind;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.process.Stoppable;
import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.ProcessRpcServer;
import com.bitgirder.process.ProcessRpcClient;
import com.bitgirder.process.ProcessActivity;
import com.bitgirder.process.ProcessFailureTarget;
import com.bitgirder.process.ProcessOperation;
import com.bitgirder.process.AbstractProcessOperationContext;

import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleNamespace;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleException;
import com.bitgirder.mingle.model.MingleNull;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleServiceResponse;
import com.bitgirder.mingle.model.MingleServiceRequest;
import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.MingleValidation;
import com.bitgirder.mingle.model.QualifiedTypeName;
import com.bitgirder.mingle.model.AtomicTypeReference;
import com.bitgirder.mingle.model.FieldDefinition;
import com.bitgirder.mingle.model.TypeDefinition;
import com.bitgirder.mingle.model.ServiceDefinition;
import com.bitgirder.mingle.model.OperationDefinition;
import com.bitgirder.mingle.model.PrototypeDefinition;

import com.bitgirder.mingle.service.MingleServiceCallContext;
import com.bitgirder.mingle.service.InternalServiceException;
import com.bitgirder.mingle.service.NoSuchNamespaceException;
import com.bitgirder.mingle.service.NoSuchOperationException;
import com.bitgirder.mingle.service.NoSuchServiceException;
import com.bitgirder.mingle.service.AuthenticationMissingException;

import java.util.Map;

public
abstract
class BoundService
extends AbstractVoidProcess
implements Stoppable
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static ObjectPath< String > JV_EXCEPTION_ROOT =
        ObjectPath.getRoot( "exception" );

    private final static ObjectPath< String > JV_RESULT_ROOT =
        ObjectPath.getRoot( "result" );

    private final MingleNamespace ns;
    private final MingleIdentifier svcId;

    // non-null when service is secured
    private final ProcessOperation< ?, ? > authExecutor; 

    private final Map< MingleIdentifier, Iterable< MingleTypeReference > >
        opExcptTypes;

    private final Iterable< MingleTypeReference > authExcptTypes;

    private final MingleBinder mb;
    private final EventHandler eh; // could be null

    private
    Map< MingleIdentifier, Iterable< MingleTypeReference > >
    initOpExcptTypes( ServiceDefinition sd )
    {
        Map< MingleIdentifier, Iterable< MingleTypeReference > > res =
            Lang.newMap();

        for ( OperationDefinition od : sd.getOperations() )
        {
            res.put( od.getName(), od.getSignature().getThrown() );
        }

        return Lang.unmodifiableMap( res );
    }

    private
    Iterable< MingleTypeReference >
    initAuthExcptTypes( ServiceDefinition sd )
    {
        state.notNull( this.mb ); // must be set before this is called

        QualifiedTypeName secType = sd.getSecurity();

        if ( secType == null ) return null;
        else
        {
            PrototypeDefinition secDef =
                state.cast(
                    PrototypeDefinition.class, 
                    mb.getTypes().expectType( secType ) 
                );
            
            return Lang.unmodifiableList( secDef.getSignature().getThrown() );
        }
    }

    protected
    BoundService( Builder< ? > b )
    {
        super( 
            inputs.notNull( b, "b" ).
                mixin( ProcessRpcServer.createStandard() ).
                mixin( ProcessRpcClient.create() )
        );

        this.ns = inputs.notNull( b.ns, "ns" );
        this.svcId = inputs.notNull( b.svcId, "svcId" );
        this.authExecutor = b.authExecutor;
        this.mb = inputs.notNull( b.mb, "mb" );
        this.eh = b.eh;

        ServiceDefinition sd = b.expectServiceDefinition();
        this.opExcptTypes = initOpExcptTypes( sd );
        this.authExcptTypes = initAuthExcptTypes( sd );
    }

    public final MingleNamespace getNamespace() { return ns; }
    public final MingleIdentifier getServiceId() { return svcId; }
    protected final MingleBinder mingleBinder() { return mb; }
    
    public final void stop() { behavior( ProcessRpcServer.class ).stop(); }

    protected
    final
    void
    initDone()
    {
        resumeRpcRequests();
    }

    protected void implInit() throws Exception { initDone(); }

    protected 
    final 
    void 
    startImpl() 
        throws Exception
    {
        holdRpcRequests();
        implInit();
    }

    protected
    abstract
    class OpRpcHandler
    extends ProcessRpcClient.DefaultResponseHandler
    {
        protected
        OpRpcHandler( AbstractOperation< ? > op )
        {
            super( activityContextFor( inputs.notNull( op, "op" ) ) );
        }
    }

    private
    void
    respond(
        MingleServiceCallContext callCtx,
        ProcessRpcServer.ResponderContext< MingleServiceResponse > respCtx,
        MingleServiceResponse mgResp )
    {
        respCtx.respond( mgResp );
    }

    private
    void
    respondSuccess( 
        MingleServiceCallContext callCtx,
        ProcessRpcServer.ResponderContext< MingleServiceResponse > respCtx,
        MingleValue respVal )
    {
        respond( 
            callCtx, respCtx, MingleServiceResponse.createSuccess( respVal ) );
    }

    private
    void
    respondFailure(
        MingleServiceCallContext callCtx,
        ProcessRpcServer.ResponderContext< MingleServiceResponse > respCtx,
        MingleException me )
    {
        respond( callCtx, respCtx, MingleServiceResponse.createFailure( me ) );
    }

    private
    boolean
    isThrowableInstance( Iterable< MingleTypeReference > typs,
                         MingleTypeReference inst )
    {
        for ( MingleTypeReference typ: typs )
        {
            if ( MingleModels.isAssignable( typ, inst, mb.getTypes() ) )
            {
                return true;
            }
        }

        return false;
    }

    private
    MingleException
    asMingleException( Throwable th,
                       Iterable< MingleTypeReference > excptTypes )
    {
        QualifiedTypeName qn = 
            MingleBinders.bindingNameForClass( th.getClass(), mb );
        
        if ( qn == null ) return null;
        else
        {
            AtomicTypeReference typ = AtomicTypeReference.create( qn );

            if ( isThrowableInstance( excptTypes, typ ) )
            {
                return (MingleException)
                    MingleBinders.
                        asMingleValue( mb, typ, th, JV_EXCEPTION_ROOT );
            }
            else return null;
        }
    }

    private
    MingleException
    asMingleException( Throwable th )
    {
        return asMingleException( th, MingleBinders.SVC_EXCEPTION_BINDINGS );
    }

    private
    void
    onInternalFailure( MingleServiceCallContext ctx,
                       Throwable th )
    {
        if ( eh == null ) 
        {
            warn( 
                th, 
                "Operation failed (see attached); " +
                "sending InternalServiceException to caller"
            );
        }
        else eh.onInternalFailure( ctx, th );
    }

    private
    void
    respondInternalFailure( 
        MingleServiceCallContext callCtx,
        ProcessRpcServer.ResponderContext< MingleServiceResponse > respCtx,
        Throwable th )
    {
        onInternalFailure( callCtx, th );

        respondFailure( 
            callCtx, 
            respCtx, 
            state.notNull( asMingleException( new InternalServiceException() ) )
        );
    }
 
    private
    void
    respondGeneralFailure( 
        MingleServiceCallContext callCtx,
        ProcessRpcServer.ResponderContext< MingleServiceResponse > respCtx,
        Throwable th )
    {
        MingleException me = asMingleException( th );
        
        if ( me == null ) respondInternalFailure( callCtx, respCtx, th );
        else respondFailure( callCtx, respCtx, me );
    }

    private
    boolean
    wasSerialized( Throwable th )
    {
        return MingleModels.wasSerialized( th ) ||
               MingleBinders.wasSerialized( th );
    }

    private
    void
    sendFailure( Throwable th,
                 AbstractOperation< ? > op,
                 Iterable< MingleTypeReference > excptTypes )
    {
        if ( wasSerialized( th ) ) 
        {
            respondInternalFailure( op.callCtx, op.respCtx, th );
        }
        else
        {
            MingleException me = asMingleException( th, excptTypes );
    
            if ( me == null ) 
            {
                respondGeneralFailure( op.callCtx, op.respCtx, th );
            }
            else respondFailure( op.callCtx, op.respCtx, me );
        }
    }

    protected
    abstract
    class AbstractOperation< V >
    extends ProcessActivity
    {
        private final MingleServiceCallContext callCtx;

        private final ProcessRpcServer.ResponderContext< MingleServiceResponse >
            respCtx;

        private final MingleTypeReference retTyp;
        private final boolean useOpaqueRetType;

        // Only set if a service takes authentication and after that service's
        // authentication phase is complete
        private Object authInput;
        private Object authRes;

        protected
        AbstractOperation(
            MingleServiceCallContext callCtx,
            ProcessRpcServer.ResponderContext< MingleServiceResponse > respCtx,
            MingleTypeReference retTyp,
            boolean useOpaqueRetType,
            ProcessActivity.Context pCtx )
        {
            super( pCtx );

            this.callCtx = callCtx;
            this.respCtx = respCtx;
            this.retTyp = retTyp;
            this.useOpaqueRetType = useOpaqueRetType;
        }

        // Our null test is on authInput since that can never be null (authRes
        // may be null for some service impls)
        private
        < V >
        V
        checkAuthAccess( Object obj,
                         String methName )
        {
            state.isFalse( 
                authInput == null,
                "Attempt to call", methName, "before authentication is complete"
            );

            return Lang.< V >castUnchecked( obj );
        }

        private
        Iterable< MingleTypeReference >
        exceptionTypes()
        {
            return 
                state.get( 
                    opExcptTypes, callCtx.getRequest().
                    getOperation(), 
                    "opExcptTypes" 
                );
        }

        // For implAuthInput() and implAuthResult() we pass the name of the
        // generated frontend method to checkAuthAccess as the methName to use
        // in errors

        protected
        final
        < V >
        V
        implAuthInput()
        {
            return this.< V >checkAuthAccess( authInput, "authInput()" );
        }

        protected
        final
        < V >
        V
        implAuthResult()
        {
            return this.< V >checkAuthAccess( authRes, "authResult()" );
        }

        private
        void
        doValidate( boolean passed,
                    MingleIdentifier optParam,
                    Object... msg )
        {
            ObjectPath< MingleIdentifier > path = optParam == null
                ? callCtx.getParametersPath()
                : callCtx.getParameterPath( optParam );

            MingleValidation.isTrue( passed, path, msg );
        }

        public
        final
        void
        validate( boolean passed,
                  Object... msg )
        {
            doValidate( passed, null, msg );
        }

        public
        final
        void
        validateParam( boolean passed,
                       MingleIdentifier param,
                       Object... msg )
        {
            doValidate( passed, inputs.notNull( param, "param" ), msg );
        }

        protected
        abstract
        void
        implStart()
            throws Exception;

        private
        MingleValue
        asMingleRespVal( V respVal )
        {
            if ( respVal == null && useOpaqueRetType ) 
            {
                return MingleNull.getInstance();
            }
            else
            {
                return
                    MingleBinders.asMingleValue( 
                        mb, retTyp, respVal, JV_RESULT_ROOT, useOpaqueRetType );
            }
        }

        public
        final
        void
        respond( V respVal )
        {
            try
            {
                MingleValue mv = asMingleRespVal( respVal );
                respondSuccess( callCtx, respCtx, mv );
            }
            catch ( Throwable th ) 
            { 
                respondInternalFailure( callCtx, respCtx, th ); 
            }
        }

        public final void respond() { respond( null ); }

        // core logic for failure targets and main entry point for failing an
        // operation due to a Throwable
        private
        void
        sendFailure( Throwable th )
        {
            inputs.notNull( th, "th" );
            BoundService.this.sendFailure( th, this, exceptionTypes() );
        }

        // Helper method for obtaining an activity context associated with the
        // operation but having a specified failure target
        private
        ProcessActivity.Context
        revealActivityContext( ProcessFailureTarget ft )
        {
            return getActivityContext( ft );
        }

        private
        ProcessActivity.Context
        revealActivityContext()
        {
            return revealActivityContext( this );
        }
    }

    protected
    abstract
    class OpTask
    implements Runnable
    {
        private final AbstractOperation< ? > op;

        protected
        OpTask( AbstractOperation< ? > op )
        {
            this.op = inputs.notNull( op, "op" );
        }

        protected
        abstract
        void
        runImpl()
            throws Exception;
        
        public
        final
        void
        run()
        {
            try { runImpl(); } catch ( Throwable th ) { op.fail( th ); }
        }
    }

    protected
    final
    ProcessActivity.Context
    activityContextFor( AbstractOperation< ? > op )
    {
        inputs.notNull( op, "op" );
        return op.revealActivityContext();
    }

    protected
    final
    boolean
    isRouteMatch( MingleServiceCallContext ctx,
                  MingleIdentifier op )
    {
        return ctx.getRequest().getOperation().equals( op );
    }

    protected
    final
    < V >
    V
    asJavaValue( FieldDefinition fd,
                 MingleServiceCallContext ctx,
                 boolean useOpaque )
    {
        return
            Lang.< V >castUnchecked(
                MingleBinders.asJavaValue(
                    fd,
                    ctx.getRawParameters(),
                    mb,
                    ctx.getParametersPath(),
                    useOpaque
                )
            );
    }

    private
    MingleServiceRequest
    accessRequest( 
        MingleServiceCallContext callCtx,
        ProcessRpcServer.ResponderContext< MingleServiceResponse > respCtx )
    {
        MingleServiceRequest req = callCtx.getRequest();
        Exception ex = null;

        MingleNamespace reqNs = req.getNamespace();
        MingleIdentifier reqSvcId = req.getService();

        if ( reqNs.equals( ns ) )
        {
            if ( ! reqSvcId.equals( svcId ) ) 
            {
                ex = new NoSuchServiceException( svcId );
            }
        }
        else ex = new NoSuchNamespaceException( reqNs );

        if ( ex == null ) return req;
        else
        {
            respondGeneralFailure( callCtx, respCtx, ex );
            return null;
        }
    } 

    protected
    abstract
    AbstractOperation< ? >
    implGetOperation( 
        MingleServiceCallContext callCtx,
        ProcessRpcServer.ResponderContext< MingleServiceResponse > respCtx,
        ProcessActivity.Context pCtx );
    
    private
    final
    class OpFailTarget
    implements ProcessFailureTarget
    {
        private final MingleServiceCallContext callCtx;

        private final
            ProcessRpcServer.ResponderContext< MingleServiceResponse > respCtx;
    
        private AbstractOperation< ? > op;

        private
        OpFailTarget(
            MingleServiceCallContext callCtx,
            ProcessRpcServer.ResponderContext< MingleServiceResponse > respCtx )
        {
            this.callCtx = callCtx;
            this.respCtx = respCtx;
        }

        public
        void
        fail( Throwable th )
        {
            if ( op == null ) respondGeneralFailure( callCtx, respCtx, th );
            else op.sendFailure( th );
        }
    }

    private
    AbstractOperation< ? >
    callGetImplOperation(
        MingleServiceCallContext callCtx,
        ProcessRpcServer.ResponderContext< MingleServiceResponse > respCtx )
    {
        OpFailTarget ft = new OpFailTarget( callCtx, respCtx );
        ProcessActivity.Context pCtx = getActivityContext( ft );
        AbstractOperation< ? > op = implGetOperation( callCtx, respCtx, pCtx );

        if ( op == null )
        {
            MingleIdentifier opId = callCtx.getRequest().getOperation();
            NoSuchOperationException ex = new NoSuchOperationException( opId );
            respondGeneralFailure( callCtx, respCtx, ex );

            return null;
        }
        else 
        {
            ft.op = op;
            return op;
        }
    }

    private
    AbstractOperation< ? >
    getOperation(
        MingleServiceCallContext callCtx,
        ProcessRpcServer.ResponderContext< MingleServiceResponse > respCtx )
    {
        try { return callGetImplOperation( callCtx, respCtx ); }
        catch ( Throwable th ) 
        {
            respondGeneralFailure( callCtx, respCtx, th );
            return null;
        }
    }

    protected
    final
    < A >
    A
    implGetJavaAuthValue( MingleTypeReference authType,
                          AbstractOperation< ? > op )
    {
        state.notNull( authType, "authType" );
        state.notNull( op, "op" );

        ObjectPath< MingleIdentifier > path = 
            op.callCtx.getAuthenticationPath();

        MingleValue mgAuth = op.callCtx.getAuthentication();

        if ( mgAuth == null ) 
        {
            throw new AuthenticationMissingException( 
                "Expected auth value of type: " + authType.getExternalForm() );
        }
        else
        {
            Object jvAuth = 
                MingleBinders.asJavaValue( mb, authType, mgAuth, path );
            
            return Lang.< A >castUnchecked( jvAuth );
        }
    }

    private
    void
    authenticationComplete( AbstractOperation< ? > op )
    {
        try { op.implStart(); } catch ( Throwable th ) { op.fail( th ); } 
    }

    // overridden by generated code for secured services
    protected
    Object
    implCreateAuthInput( AbstractOperation< ? > op )
    {
        throw state.createFail( "Service does not require authentication" );
    }

    private
    void
    failBeginAuthentication( Throwable th,
                             AbstractOperation< ? > op )
    {
        MingleException me = 
            asMingleException( th, MingleBinders.AUTH_EXCEPTION_BINDINGS );

        // if me == null just fall through to default handling
        if ( me == null ) respondGeneralFailure( op.callCtx, op.respCtx, th );
        else respondFailure( op.callCtx, op.respCtx, me );
    }

    private
    final
    class AuthFailTarget
    implements ProcessFailureTarget
    {
        private final AbstractOperation< ? > op;

        private AuthFailTarget( AbstractOperation< ? > op ) { this.op = op; }

        public 
        void 
        fail( Throwable th ) 
        {
            sendFailure( th, op, authExcptTypes );
        }
    }

    private
    final
    class AuthOpCtx
    extends AbstractProcessOperationContext< Object, Object >
    {
        private final AbstractOperation< ? > op;
        private final Object authInput;

        private
        AuthOpCtx( AbstractOperation< ? > op,
                   Object authInput )
        {
            super( op.revealActivityContext( new AuthFailTarget( op ) ) );

            this.op = op;
            this.authInput = authInput;
        }

        public Object input() { return authInput; }

        protected
        void
        implComplete( Object obj )
        {
            // set now so AbstractOperation.auth(Input|Result)() will return
            // normally
            op.authInput = authInput;
            op.authRes = obj; 

            authenticationComplete( op );
        }
    }

    private
    AuthOpCtx
    createAuthOpContext( final AbstractOperation< ? > op,
                         Object authInput )
    {
        return new AuthOpCtx( op, authInput );
    }

    private
    < I, O >
    void
    beginAuthExec( ProcessOperation< I, O > exec,
                   AuthOpCtx ctx )
    {
        // cast is okay since codegen ensures that in/out types are correct
        ProcessOperation< Object, Object > castOp = 
            Lang.< ProcessOperation< Object, Object > >castUnchecked( exec );

        ctx.begin( castOp );
    }

    private
    void
    beginAuthentication( AbstractOperation< ? > op )
    {
        if ( authExcptTypes == null ) authenticationComplete( op );
        else
        {
            Object authInput = implCreateAuthInput( op );
            AuthOpCtx ctx = createAuthOpContext( op, authInput );
            beginAuthExec( authExecutor, ctx );
        }
    }

    @ProcessRpcServer.Responder
    private
    void
    handleRequest( 
        MingleServiceCallContext callCtx,
        ProcessRpcServer.ResponderContext< MingleServiceResponse > respCtx )
    {
        MingleServiceRequest req = accessRequest( callCtx, respCtx );

        if ( req != null )
        {
            AbstractOperation< ? > op = getOperation( callCtx, respCtx );
    
            if ( op != null ) 
            {
                try { beginAuthentication( op ); }
                catch ( Throwable th ) { failBeginAuthentication( th, op ); }
            }
            // else: we responded with failure in getOperation()
        }
    }

    @ProcessRpcServer.Responder
    private
    void
    handleRequest(
        MingleServiceRequest req,
        ProcessRpcServer.ResponderContext< MingleServiceResponse > respCtx )
    {
        handleRequest( MingleServiceCallContext.create( req ), respCtx );
    }

    public
    static
    interface EventHandler
    {
        // response to caller will be sent separately from this method, possibly
        // before, possibly after, possibly at the same time. This is only for
        // informational purposes (logging, monitoring, etc)
        public
        void
        onInternalFailure( MingleServiceCallContext ctx,
                           Throwable th );
    }

    public
    static
    abstract
    class AbstractEventHandler
    implements EventHandler
    {
        public
        void
        onInternalFailure( MingleServiceCallContext ctx,
                           Throwable th )
        {}
    }

    protected
    final
    void
    implValidateAuthExecutor( Builder< ? > b )
    {
        state.notNull( b, "b" );
        state.isFalse( b.authExecutor == null, "No auth executor is set" );
    }

    public
    static
    abstract
    class Builder< B extends Builder< B > >
    extends AbstractVoidProcess.Builder< B >
    {
        private MingleNamespace ns;
        private MingleIdentifier svcId;
        private MingleBinder mb;
        private ProcessOperation< ?, ? > authExecutor;
        private EventHandler eh;
        private QualifiedTypeName svcQname;

        protected Builder() {}

        protected
        final
        B
        implSetTypeName( QualifiedTypeName svcQname )
        {
            this.svcQname = state.notNull( svcQname, "svcQname" );
            return castThis();
        }

        public
        final
        B
        setNamespace( MingleNamespace ns )
        {
            this.ns = inputs.notNull( ns, "ns" );
            return castThis();
        }

        public
        final
        B
        setNamespace( CharSequence ns )
        {
            inputs.notNull( ns, "ns" );
            return setNamespace( MingleNamespace.create( ns ) );
        }

        public
        final
        B
        setServiceId( MingleIdentifier svcId )
        {
            this.svcId = inputs.notNull( svcId, "svcId" );
            return castThis();
        }

        public
        final
        B
        setServiceId( CharSequence svcId )
        {
            inputs.notNull( svcId, "svcId" );
            return setServiceId( MingleIdentifier.create( svcId ) );
        }

        public
        final
        B
        setBinder( MingleBinder mb )
        {
            this.mb = inputs.notNull( mb, "mb" );
            return castThis();
        }
    
        protected
        final
        B
        implSetAuthExecutor( ProcessOperation< ?, ? > authExecutor )
        {
            this.authExecutor = state.notNull( authExecutor, "authExecutor" );
            return castThis();
        }

        public
        final
        B
        setEventHandler( EventHandler eh )
        {
            this.eh = inputs.notNull( eh, "eh" );
            return castThis();
        }

        private
        ServiceDefinition
        expectServiceDefinition()
        {
            state.notNull( svcQname, "svcQname" );
            inputs.notNull( mb, "mb" );

            return
                state.cast(
                    ServiceDefinition.class, 
                    mb.getTypes().expectType( svcQname ) );
        }
    }
}
