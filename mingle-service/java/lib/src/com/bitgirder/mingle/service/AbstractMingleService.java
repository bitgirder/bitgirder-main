package com.bitgirder.mingle.service;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;
import com.bitgirder.lang.Completion;

import com.bitgirder.lang.reflect.ReflectUtils;
import com.bitgirder.lang.reflect.MethodInvocation;
import com.bitgirder.lang.reflect.ReflectedInvocation;

import com.bitgirder.mingle.parser.MingleParsers;

import com.bitgirder.mingle.model.MingleException;
import com.bitgirder.mingle.model.MingleServiceRequest;
import com.bitgirder.mingle.model.MingleServiceResponse;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleNull;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleValidationException;

import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.ProcessRpcServer;
import com.bitgirder.process.ProcessRpcClient;
import com.bitgirder.process.ProcessExit;
import com.bitgirder.process.ProcessBehavior;

import java.util.List;
import java.util.Iterator;
import java.util.Map;
import java.util.Set;
import java.util.IdentityHashMap;
import java.util.Collection;

import java.lang.reflect.Method;
import java.lang.reflect.Member;
import java.lang.reflect.Constructor;
import java.lang.reflect.AnnotatedElement;

import java.lang.annotation.Annotation;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Retention;
import java.lang.annotation.Target;
import java.lang.annotation.ElementType;

import java.util.concurrent.CopyOnWriteArrayList;
import java.util.concurrent.Callable;

public
abstract
class AbstractMingleService
extends AbstractVoidProcess
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Map< MingleIdentifier, MingleResponder > responders =
        Lang.newMap();

    private final Map< AbstractProcess< ? >, ResponderContextImpl >
        responderSpawns = 
            new IdentityHashMap< AbstractProcess< ? >, ResponderContextImpl >();

    // Because we don't place any restrictions on when a call to addFilter() may
    // be made, other than that such calls must come from within the process
    // thread, we use a copy on write list here so that filter chains can be
    // built on top of the result of calling filters.iterator() at the time the
    // request chain is begun without concern for the safety of that iterator if
    // a call to addFilter() arrives while the chain is still live. Indeed, by
    // using CopyOnWriteArrayList we could even remove the requirement that adds
    // come from the process thread, but keep that restriction in place
    // nonetheless to be consistent with the general principle that processes
    // should never attempt to access their own state from outside of the
    // process thread.
    private final List< RequestFilter > filters = 
        new CopyOnWriteArrayList< RequestFilter >();

    private boolean stopped;

    protected
    AbstractMingleService( Builder< ? > b )
    {
        super( 
            b.mixinAll( 
                ProcessRpcClient.create(),
                ProcessRpcServer.createStandard() ) );
    }

    protected
    AbstractMingleService( ProcessBehavior... behaviors )
    {
        this( new Builder< Builder >().mixinAll( behaviors ) );
    }

    protected AbstractMingleService() { this( new Builder< Builder >() ); }

    @Override
    protected
    final
    void
    childExited( AbstractProcess< ? > proc,
                 ProcessExit< ? > exit )
    {
        ResponderContextImpl ctx = responderSpawns.remove( proc );

        if ( ctx != null )
        {
            if ( exit.isOk() ) ctx.respond( exit.getResult() );
            else ctx.fail( exit.getThrowable() );
        }
    }

    protected
    final
    void
    holdRequests()
    {
        behavior( ProcessRpcServer.class ).holdRequests();
    }

    protected
    final
    void
    resumeRequests()
    {
        behavior( ProcessRpcServer.class ).resumeRequests();
    }

    protected
    static
    MingleValue
    asMingleValue( Object obj )
    {
        return MingleModels.asMingleValue( obj );
    }

    // Could use some caching or id pool at some point (will likely build some
    // sort of identifier factory in model to abstract this)
    protected
    static
    MingleIdentifier
    getIdentifier( CharSequence str )
    {
        inputs.notNull( str, "str" );
        return MingleParsers.createIdentifier( str );
    }

    // ctx is currently not used, but in all likelihood will be (for setting
    // auditing, timing, responseTo ids, etc
    protected
    final
    MingleServiceResponse
    createSuccessResponse( MingleServiceCallContext ctx,
                           MingleValue val )
    {
        inputs.notNull( ctx, "ctx" );
        inputs.notNull( val, "val" );

        return MingleServiceResponse.createSuccess( val );
    }

    protected
    final
    MingleServiceResponse
    createSuccessResponse( MingleServiceCallContext ctx )
    {
        return createSuccessResponse( ctx, MingleNull.getInstance() );
    }

    protected
    final
    MingleServiceResponse
    createFailureResponse( MingleServiceCallContext ctx,
                           MingleException me )
    {
        inputs.notNull( ctx, "ctx" );
        inputs.notNull( me, "me" );

        return MingleServiceResponse.createFailure( me );
    }

    static
    interface MingleResponder
    {
        public
        void
        respond( MingleServiceCallContext ctx,
                 ProcessRpcServer.ResponderContext< Object > respCtx )
            throws Exception;
        
        public
        Completion< MingleServiceResponse >
        getResponse( MingleServiceCallContext ctx,
                     Completion< ? > comp )
            throws Exception;
        
        public
        Object
        getInvocationTarget();
    }

    static
    interface AnnotatedMingleResponder
    extends MingleResponder
    {
        // behaves exactly as AnnotatedElement.getAnnotation()
        public
        < T extends Annotation >
        T
        getAnnotation( Class< T > cls );
    }

    private
    MingleResponder
    getResponder( MingleServiceRequest req )
    {
        return responders.get( req.getOperation() );
    }
 
    // ResponderContext wrapper implementation for rewriting responses as mingle
    // service responses. In many cases the response will already be a mingle
    // service response, such as for the MingleResponder implementations in this
    // class.  In those cases the code path through this filter will not create
    // any new objects and will ultimately return the response it is given.
    private
    final
    class ResponderContextImpl
    implements ProcessRpcServer.ResponderContext< Object >
    {
        private final MingleServiceCallContext ctx;
        private final MingleResponder resp; // could be null
        private final ProcessRpcServer.ResponderContext< Object > respCtx;

        private
        ResponderContextImpl( 
            MingleServiceCallContext ctx,
            MingleResponder resp,
            ProcessRpcServer.ResponderContext< Object > respCtx )
        {
            this.ctx = ctx;
            this.resp = resp;
            this.respCtx = respCtx;
        }

        private
        Completion< MingleServiceResponse >
        failedFilterResponse( Exception ex )
        {
            warn(
                ex, "getResponse() failed for responder", resp, "-- returning",
                "internal service error to caller" );
 
            return
                Lang.successCompletion(
                    createFailureResponse(
                        ctx, MingleServices.getInternalServiceException() ) );
        }

        private
        Completion< MingleServiceResponse >
        asServiceCompletion( Completion< MingleServiceResponse > comp )
        {
            if ( comp.isOk() ) return comp;
            else
            {
                String failTarget = 
                    "Response for operation " +
                    ctx.getRequest().getOperation().getExternalForm();
 
                MingleServiceResponse mgResp =
                    asExceptionResponse( ctx, comp.getThrowable(), failTarget );
 
                return Lang.successCompletion( mgResp );
            }
        }

        public
        Completion< ? >
        filterCompletion( Completion< ? > comp )
        {
            if ( resp == null ) return comp;
            else
            {
                try
                {
                    return asServiceCompletion( resp.getResponse( ctx, comp ) );
                }
                catch ( Exception ex ) { return failedFilterResponse(  ex ); }
            }
        }

        private
        void
        respondCompletion( Completion< ? > comp )
        {
            comp = filterCompletion( comp );

            if ( comp.isOk() ) respCtx.respond( comp.getResult() );
            else respCtx.fail( comp.getThrowable() );
        }

        public
        void
        spawn( AbstractProcess< ? > proc )
        {
            responderSpawns.put( proc, this );
            AbstractMingleService.this.spawn( proc );
        }

        public 
        void 
        respond( Object resp )
        {
            respondCompletion( Lang.successCompletion( resp ) );
        }

        public
        void
        fail( Throwable th )
        {
            respondCompletion( Lang.failureCompletion( th ) );
        }
    }

    final
    CharSequence
    makeErrorLocation( Member m,
                       int indx )
    {
        return "parameter at index " + indx + " in " + m;
    }

    private boolean hasControlAnnotation( Annotation[] arr ) { return false; }

    private
    boolean
    isMethodResponderParameter( Class< ? > cls,
                                Method m )
    {
        return getControlKeyFromParameterType( cls, m ) != null;
    }

    // more correctly: is this method a responder that this class can deal with?
    private
    boolean
    isMethodResponder( Method m )
    {
        Class< ? >[] paramTypes = m.getParameterTypes();
        Annotation[][] anns = ReflectUtils.getParameterAnnotations( m );

        int ctxIndx = -1;

        for ( int i = 0, e = paramTypes.length; i < e; ++i )
        {
            Annotation[] arr = anns[ i ];

            if ( arr == null || ! hasControlAnnotation( arr ) )
            {
                if ( paramTypes[ i ].equals( MingleServiceCallContext.class ) )
                {
                    if ( ctxIndx == -1 ) ctxIndx = i;
                    else return false;
                }
                else return isMethodResponderParameter( paramTypes[ i ], m );
            }
        }

        return ctxIndx >= 0;
    }

    static
    boolean
    isInternalValidationException( MingleServiceCallContext ctx,
                                   Throwable th )
    {
        if ( th instanceof MingleValidationException )
        {
            MingleValidationException mve = (MingleValidationException) th;

            return ! ctx.isInboundValidationException( mve );
        }
        else return false;
    }

    // Special logic for handling exceptions thrown from an operation.
    // Specifically, we don't throw validation exceptions to callers unless they
    // are tied to the input itself (validation exceptions encountered
    // internally during operation processing are considered internal service
    // exceptions instead). Other logic or cases may be added eventually
    private
    static
    MingleException
    asServiceException( MingleServiceCallContext ctx,
                        Throwable th )
    {
        if ( isInternalValidationException( ctx, th ) ) th = null;
 
        return th == null ? null : MingleServices.asServiceException( th );
    }

    private
    static
    MingleServiceResponse
    asExceptionResponse( MingleServiceCallContext ctx,
                         Throwable th,
                         String failTarget )
    {
        MingleException me = asServiceException( ctx, th );

        if ( me == null )
        {
            CodeLoggers.warn( 
                th, 
                failTarget, "failed with non-service throwable (attached); " +
                "returning internal service exception to caller" );
 
            me = MingleServices.getInternalServiceException();
        }

        return MingleServiceResponse.createFailure( me );
    }

    static
    abstract
    class AbstractAnnotatedMingleResponder
    implements AnnotatedMingleResponder
    {
        private final AnnotatedElement elt;

        AbstractAnnotatedMingleResponder( AnnotatedElement elt ) 
        { 
            this.elt = state.notNull( elt, "elt" ); 
        }

        public
        final
        boolean
        isAnnotationPresent( Class< ? extends Annotation > cls )
        {
            return elt.isAnnotationPresent( inputs.notNull( cls, "cls" ) );
        }

        public
        final
        < T extends Annotation >
        T
        getAnnotation( Class< T > cls )
        {
            return elt.getAnnotation( inputs.notNull( cls, "cls" ) );
        }

        public final Object getInvocationTarget() { return elt; }
    }

    private
    static
    abstract
    class AbstractResponderImpl
    extends AbstractAnnotatedMingleResponder
    {
        private AbstractResponderImpl( AnnotatedElement elt ) { super( elt ); }

        // if comp is ok we do a sanity check that the child process did in fact
        // exit with a MingleServiceResponse; if comp is a failure just return
        // it and let the ResponderContextImpl handle it there
        public
        final
        Completion< MingleServiceResponse >
        getResponse( MingleServiceCallContext ctx,
                     Completion< ? > comp )
        {
            if ( comp.isOk() )
            {
                Object result = comp.getResult();

                state.isTrue( 
                    result instanceof MingleServiceResponse,
                    this, "exited with instance of something other than a " +
                    "MingleServiceResponse:", result );
            }
            
            @SuppressWarnings( "unchecked" )
            Completion< MingleServiceResponse > res =
                (Completion< MingleServiceResponse >) comp;
            
            return res;
        }
    }

    private
    final
    static
    class MethodCall
    implements Callable< Object >
    {
        private final MethodInvocation mi;
        private final Map< Object, Object > params;

        private
        MethodCall( MethodInvocation mi,
                    Map< Object, Object > params )
        {
            this.mi = mi;
            this.params = params;
        }

        public Object call() throws Exception { return mi.invoke( params ); }
    }

    private
    void
    completeCall( MingleServiceCallContext ctx,
                  ProcessRpcServer.ResponderContext< Object > respCtx,
                  MethodCall call )
    {
        try 
        { 
            Object res = call.call();

            if ( ! call.mi.hasKey( ProcessRpcServer.ResponderContext.class ) )
            {
                respCtx.respond( res );
            }
        }
        catch ( Throwable th ) { respCtx.fail( th ); }
    }

    final
    void
    invokeCall( AnnotatedMingleResponder amr,
                final MingleServiceCallContext ctx,
                final ProcessRpcServer.ResponderContext< Object > respCtx,
                MethodInvocation mi,
                Map< Object, Object > params )
    {
        MethodCall call = new MethodCall( mi, params );

        completeCall( ctx, respCtx, call );
    }

    private
    final
    class ImmediateResponder
    extends AbstractResponderImpl
    {
        private final MethodInvocation mi;

        private 
        ImmediateResponder( MethodInvocation mi ) 
        { 
            super( (AnnotatedElement) mi.getTarget() );
            this.mi = mi;
        }

        private
        Map< Object, Object >
        makeParams( MingleServiceCallContext ctx,
                    ProcessRpcServer.ResponderContext< Object > respCtx )
        {
            Map< Object, Object > res = Lang.newMap();

            res.put( MingleServiceCallContext.class, ctx );
            res.put( ProcessRpcServer.ResponderContext.class, respCtx );
            
            return res;
        }

        public
        void
        respond( final MingleServiceCallContext ctx,
                 final ProcessRpcServer.ResponderContext< Object > respCtx )
        {
            invokeCall( this, ctx, respCtx, mi, makeParams( ctx, respCtx ) );
        }

        @Override
        public 
        String 
        toString() 
        { 
            return "Immediate responder [ " + mi.getTarget() + " ]"; 
        }
    }

    private
    final
    class ChildProcessResponder
    extends AbstractResponderImpl
    {
        private final Constructor< ? extends AbstractProcess > cons;

        private
        ChildProcessResponder( Constructor< ? extends AbstractProcess > cons )
        {
            super( cons.getDeclaringClass() );

            this.cons = cons;
        }

        public boolean requiresProcessThread() { return true; }

        public
        void
        respond( MingleServiceCallContext ctx,
                 ProcessRpcServer.ResponderContext< Object > respCtx )
            throws Exception
        {
            Object[] args = new Object[] { AbstractMingleService.this, ctx };

            AbstractProcess< ? > child = ReflectUtils.invoke( cons, args );
            respCtx.spawn( child );
        }

        @Override
        public
        String
        toString()
        {
            return "Child process responder [ " + cons + " ]";
        }
    }

    final
    void
    putResponder( MingleIdentifier op,
                  MingleResponder responder )
    {
        if ( responders.containsKey( op ) )
        {
            throw new MingleServices.OverloadedOperationException( op );
        }
        else responders.put( op, responder );
    }

    // overrideable so subclasses can reuse getControlKeyFromAnnotations but
    // still define which annotations are relevant to it or how to process them
    Object
    asControlKey( Annotation ann,
                  Object target,
                  CharSequence errLoc )
    {
        return null;
    }

    // could return null; will fail if more than one valid result could be
    // inferred. See note at asControlKey()
    final
    Object
    getControlKeyFromAnnotations( Annotation[] paramAnns,
                                  Object target,
                                  CharSequence errLoc )
    {
        Object res = null;

        for ( Annotation ann : paramAnns )
        {
            Object key = asControlKey( ann, target, errLoc );

            if ( key != null )
            {
                if ( res == null ) res = key;
                else 
                {
                    throw state.createFail( 
                        "Duplicate control keys at", errLoc ); 
                }
            }
        }

        return res;
    }

    // overrideable to parameterize behavior of getControlKey()
    Object
    getControlKeyFromParameterType( Class< ? > typ,
                                    Object target )
    {
        if ( typ.equals( MingleServiceCallContext.class ) ||
             typ.equals( ProcessRpcServer.ResponderContext.class ) )
        {
            return typ;
        }
        else return null;
    }

    // behavior is parameterized by getControlKeyFromAnnotations() and
    // getControlKeyFromParameterType()
    final
    Object
    getControlKey( Class< ? > paramTyp,
                   Annotation[] paramAnns,
                   Object target,
                   CharSequence errLoc )
    {
        Object res = getControlKeyFromAnnotations( paramAnns, target, errLoc );

        if ( res == null ) 
        {
            res = getControlKeyFromParameterType( paramTyp, target );
        }

        if ( res == null )
        {
            throw state.createFail( "Don't know how to handle", errLoc );
        }
        else return res;
    }

    private
    void
    completeInvocationBuild( MethodInvocation.Builder b,
                             Method m )
    {
        b.setIgnoreUnmatchedKeys( true );

        Class< ? >[] paramTypes = m.getParameterTypes();
        Annotation[][] anns = ReflectUtils.getParameterAnnotations( m );

        for ( int i = 0, e = anns.length; i < e; ++i )
        {
            CharSequence errLoc = makeErrorLocation( m, i );

            Object ctlKey = 
                getControlKey( paramTypes[ i ], anns[ i ], m, errLoc );

            b.setKey( ctlKey, i );
        }
    }

    private
    MethodInvocation
    createMethodInvocation( Method m )
    {
        MethodInvocation.Builder b = 
            new MethodInvocation.Builder().
                setTarget( m ).
                setInstance( this );
        
        completeInvocationBuild( b, m );

        return b.build();
    }
    
    private
    void
    initNativeResponder( Method m )
    {
        m.setAccessible( true );
        MethodInvocation mi = createMethodInvocation( m );

        MingleIdentifier op = MingleServices.getOperationName( m );
        putResponder( op, new ImmediateResponder( mi ) );
    }

    private
    void
    initMethodBasedResponders()
        throws Exception
    {
        Collection< Method > methods = 
            ReflectUtils.getDeclaredAncestorMethods(
                getClass(), MingleServices.Operation.class );
 
        for ( Method m : methods )
        {
            if ( ! initMingleResponder( m ) ) initNativeResponder( m );
        }
    }

    private
    Class< ? extends AbstractProcess >
    asAbstractProcessClass( Class< ? > cls )
    {
        try { return cls.asSubclass( AbstractProcess.class ); }
        catch ( ClassCastException cce )
        {
            throw state.createFail( "Not an abstract process:", cls );
        }
    }

    // since the test for whether a class is a child responder is to try and get
    // its single-arg constructor, we conflate the check and the getting of that
    // constructor together in this method, returning null if there is no such
    // constructor.
    //
    // The prohibition against static child responders is arbitrary and there
    // really is no reason we couldn't have them. For now all responders in the
    // system are non-static, so rather than put in checks and coverage of the
    // static case (for now), we just assert that no static responders are
    // encountered.
    private
    Constructor< ? extends AbstractProcess >
    getChildProcessResponderConstructor( 
        Class< ? > enclosing,
        Class< ? extends AbstractProcess > cls )
    {
        try 
        { 
            Constructor< ? extends AbstractProcess > res =
                cls.getDeclaredConstructor( 
                    enclosing, MingleServiceCallContext.class );
        
            res.setAccessible( true );

            state.isFalse( 
                ReflectUtils.isStatic( cls ),
                "static child responders not supported" );
            
            return res;
        }
        catch ( NoSuchMethodException nsme ) { return null; }
    }

    // subclasses can override the initMingleResponder() methods to prevent this
    // class from attempting to interpret the target as a native responder.
    // return values of true indicate that the target has been handled; false
    // indicates that this class should expect to be able to handle it
    boolean
    initMingleResponder( Class< ? > cls )
        throws Exception
    { 
        return false;
    }

    boolean initMingleResponder( Method m ) throws Exception { return false; }

    private
    void
    initNativeResponder( Class< ? > cls )
        throws Exception
    {
        Class< ? extends AbstractProcess > procCls = 
            asAbstractProcessClass( cls );
        
        Constructor< ? extends AbstractProcess > cons =
            getChildProcessResponderConstructor( getClass(), procCls );

        MingleIdentifier op = MingleServices.getOperationName( procCls );
        putResponder( op, new ChildProcessResponder( cons ) );
    }

    private
    void
    initInstanceBasedResponders()
        throws Exception
    {
        Collection< Class< ? > > classes =
            ReflectUtils.getDeclaredAncestorClasses(
                getClass(), MingleServices.Operation.class );
        
        for ( Class< ? > cls : classes )
        {
            if ( ! initMingleResponder( cls ) ) initNativeResponder( cls );
        }
    }

    private
    void
    initResponders()
        throws Exception
    {
        initMethodBasedResponders();
        initInstanceBasedResponders();
    }

    // for subclasses in place of startImpl(), which we make final. This will be
    // called as the final part of startImpl(). Subclasses can use this to
    // continue any immediate or asynchronous initialization of the service, but
    // regardless startImpl() will return once this method does. Services which
    // need to do some sort of asynchronous or other long-running initialization
    // can call holdRequests() as part of initService(), eventually calling
    // resumeRequests() once initialization is complete.
    //
    // Overridden versions of this method should call super.initService() to
    // properly chain invocations
    protected void initService() throws Exception {}

    protected 
    final
    void 
    startImpl()
        throws Exception
    {
        initResponders();
        initService();
    }

    private
    MingleServiceResponse
    getNoSuchOperationFailure( MingleServiceCallContext ctx )
    {
        MingleIdentifier op = ctx.getRequest().getOperation();

        // we could cache the mingle exceptions created below keyed by op if
        // needed for performance
        return
            createFailureResponse(
                ctx,
                MingleServices.asServiceException(
                    new NoSuchOperationException( op ) ) );
    }

    public
    static
    interface RequestFilter
    {
        public
        void
        process( MingleServiceCallContext ctx,
                 RequestFilterContext filterCtx )
            throws Exception;
    }

    public
    static
    interface RequestFilterContext
    {
        // pass ctx to the next element in the filter chain
        public
        void
        pass( MingleServiceCallContext ctx );
        
        // abort the filter chain and respond with resp
        public
        void
        respond( Object resp );

        // abort the filter chain with the given failure
        public
        void
        fail( Throwable th );

        // could return null either because the target of the invocation is not
        // an AnnotatedMingleResponder, doesn't have the given annotation, or
        // because there is no target for the invocation at all (ie, this
        // request would default to throwing a NoSuchOperationException)
        public
        < T extends Annotation >
        T
        getTargetAnnotation( Class< T > cls );

        public
        Object
        getInvocationTarget();
    }

    // Should not be called from outside the process thread or during
    // construction; expectation is that this will not be called after
    // completion of initService() for normal usage
    //
    // We may eventually need to add an overloaded version that takes some
    // insertion position parameter, perhaps something like BEFORE_HEAD,
    // BEFORE_TAIL, AFTER_HEAD, AFTER_TAIL, indicating where in the existing
    // chain to insert the filter. The current version would just call into that
    // one with the default behavior (BEFORE_TAIL).
    protected
    final
    void
    addFilter( RequestFilter f )
    {
        filters.add( inputs.notNull( f, "f" ) );
    }
    
    private
    final
    class FilterChain
    implements RequestFilterContext
    {
        private final MingleResponder resp; // could be null
        private final ProcessRpcServer.ResponderContext< Object > respCtx;
        private final Iterator< RequestFilter > it = filters.iterator();

        private
        FilterChain( MingleResponder resp,
                     ProcessRpcServer.ResponderContext< Object > respCtx )
        {
            this.resp = resp;
            this.respCtx = respCtx;
        }

        private
        void
        advance( MingleServiceCallContext ctx )
        {
            try
            {
                if ( it.hasNext() ) it.next().process( ctx, this );
                else
                {
                    MingleServiceRequest req = ctx.getRequest();
            
                    if ( resp == null ) 
                    {
                        respCtx.respond( getNoSuchOperationFailure( ctx ) );
                    }
                    else resp.respond( ctx, respCtx );
                }
            }
            catch ( Throwable th ) { respCtx.fail( th ); }
        }

        public void pass( MingleServiceCallContext ctx ) { advance( ctx ); }

        public void respond( Object resp ) { respCtx.respond( resp ); }
        public void fail( Throwable th ) { respCtx.fail( th ); }

        public
        < T extends Annotation >
        T
        getTargetAnnotation( Class< T > cls )
        {
            inputs.notNull( cls, "cls" );

            return 
                resp instanceof AnnotatedMingleResponder 
                    ? ( (AnnotatedMingleResponder) resp ).getAnnotation( cls )
                    : null;
        }

        public
        Object
        getInvocationTarget()
        {
            return resp == null ? null : resp.getInvocationTarget(); 
        }
    }

    @ProcessRpcServer.Responder
    private
    void
    handle( MingleServiceCallContext ctx,
            ProcessRpcServer.ResponderContext< Object > respCtx )
        throws Exception
    {
        MingleResponder resp = getResponder( ctx.getRequest() );

        respCtx = new ResponderContextImpl( ctx, resp, respCtx );

        new FilterChain( resp, respCtx ).advance( ctx );
    }

    @ProcessRpcServer.Responder
    private
    void
    handle( MingleServiceRequest req,
            ProcessRpcServer.ResponderContext< Object > respCtx )
        throws Exception
    {
        handle( MingleServiceCallContext.create( req ), respCtx );
    }

    public
    static
    class Builder< B extends Builder >
    extends AbstractProcess.Builder< B >
    {}
}
