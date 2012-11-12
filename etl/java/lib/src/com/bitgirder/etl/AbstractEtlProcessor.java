package com.bitgirder.etl;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.reflect.ReflectUtils;

import com.bitgirder.process.ProcessRpcClient;
import com.bitgirder.process.ProcessRpcServer;
import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.ProcessBehavior;

import java.util.Collection;

public
abstract
class AbstractEtlProcessor
extends AbstractVoidProcess
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private static enum RpcStopState { INITIAL, WAIT, REQUESTED, STOPPED; }

    private AbstractRecordSetProcessor activeProc;
    private RpcStopState ss = RpcStopState.INITIAL;

    protected
    AbstractEtlProcessor( Builder< ? > b )
    {
        super(
            inputs.notNull( b, "b" ).
                mixin( new ProcessRpcServer.Builder().build() ).
                mixin( ProcessRpcClient.create() )
        );
    }

    protected AbstractEtlProcessor() { this( new Builder() {} ); }

    // overridable
    protected void init() throws Exception {}

    protected final void startImpl() throws Exception { init(); }

    // classes which use state objects should override one of the
    // setProcessorState() methods
    protected void setProcessorState( Object st ) throws Exception {}

    protected
    void
    setProcessorState( Object st,
                       Runnable onComp )
        throws Exception
    {
        setProcessorState( st );
        onComp.run();
    }

    @ProcessRpcServer.Responder
    private
    void
    handleSetProcessorState( 
        EtlProcessors.SetProcessorState req,
        final ProcessRpcServer.ResponderContext< Void > ctx )
            throws Exception
    {
        setProcessorState( 
            req.getObject(), 
            new AbstractTask() { 
                protected void runImpl() { ctx.respond( null ); }
            }
        );
    }

    private
    void
    exitConditional()
    {
        if ( ss == RpcStopState.STOPPED && activeProc == null ) exit();
    }

    @Override
    protected
    final
    void
    behaviorShutdown( ProcessBehavior b )
        throws Exception
    {
        super.behaviorShutdown( b );

        if ( b instanceof ProcessRpcServer ) 
        {
            // rpc server could be shutting down due to an error somewhere else
            // and not as the result of an explicity shutdown request, so we
            // only proceed if we are in the middle of a normal shutdown
            // initiated by handling a ShutdownRequest
            if ( ss == RpcStopState.WAIT )
            {
                ss = RpcStopState.STOPPED;
                exitConditional();
            }
        }
    }

    // (set|clear)ActiveProcessor() provided for access by
    // AbstractEtlRecordProcessor
    final
    void
    setActiveProcessor( AbstractRecordSetProcessor activeProc )
    {
        state.notNull( activeProc, "activeProc" );

        state.isTrue( 
            this.activeProc == null, 
            "A process operation is already in progress"
        );

        this.activeProc = activeProc;
    }

    final
    void
    clearActiveProcessor( AbstractRecordSetProcessor activeProc )
    {
        state.notNull( activeProc, "activeProc" );

        state.isTrue( 
            this.activeProc == activeProc, 
            "Unexpected activeProc passed to clearActiveProcessor()"
        );

        this.activeProc = null;

        if ( ss == RpcStopState.REQUESTED ) beginRpcStop();
    }

    private
    void
    beginRpcStop()
    {
        state.isTrue( 
            ss == RpcStopState.INITIAL || ss == RpcStopState.REQUESTED );

        if ( activeProc != null ) activeProc.abortProcess();
 
        ss = RpcStopState.WAIT;
        behavior( ProcessRpcServer.class ).stop();
    }

    private
    Class< ? >
    getSingleProcessorClass( Class< ? > clsDecl,
                             Class< ? > expctCls,
                             Class< ? > curRes )
    {
        // enter branch only for a proper non-abstract subclass of expctCls
        if ( expctCls.isAssignableFrom( clsDecl ) &&
             ( ! clsDecl.equals( expctCls ) ) &&
             ( ! ReflectUtils.isAbstract( clsDecl ) ) )
        {
            if ( curRes == null ) return clsDecl;
            else 
            {
                throw state.createFail( 
                    "More than one descendant of", expctCls, 
                    "exists in", getClass(), "or its superclasses:",
                    curRes, "and", clsDecl
                );
            }
        }
        else return curRes;
    }

    private
    Class< ? >
    getSingleProcessorClass( Collection< Class< ? > > classes,
                             Class< ? > expctCls )
    {
        Class< ? > res = null;

        for ( Class< ? > cls : classes )
        {
            Class< ? >[] clsDecls = cls.getDeclaredClasses();

            for ( Class< ? > clsDecl : clsDecls )
            {
                res = getSingleProcessorClass( clsDecl, expctCls, res );
            }
        }

        return res;
    }

    @ProcessRpcServer.Initializer
    private
    void
    initResultSetProcessor( ProcessRpcServer.InitializerContext ctx )
    {
        Collection< Class< ? > > classes =
            ReflectUtils.getAllAncestors( getClass() );
 
        Class< ? > expctCls = AbstractRecordSetProcessor.class;

        Class< ? > procCls = getSingleProcessorClass( classes, expctCls );
 
        if ( procCls == null )
        {
            state.fail( 
                "No subclass of", expctCls.getName(), "declared in", 
                getClass() );
        }
        else ctx.initHandler( procCls );
    }

    // package-only for the moment since the only use for this at the moment is
    // in coordinating some specific sequencing in test code; could be made
    // public though if a compelling reason presents itself
    void handledShutdownRequest() {}

    // Implemented as an immediate processor even if the shutdown ultimately
    // waits on other conditions to exit the process
    @ProcessRpcServer.Responder
    private
    Void
    handle( EtlProcessors.ShutdownRequest req )
    {
        // Only handle the first stop request
        if ( ss == RpcStopState.INITIAL )
        {
            if ( req.isUrgent() || activeProc == null ) beginRpcStop();
            else ss = RpcStopState.REQUESTED;

            handledShutdownRequest();
        }

        return null;
    }

    protected
    static
    class Builder< B extends Builder< ? > >
    extends AbstractVoidProcess.Builder< B >
    {
        protected Builder() {}
    }
}
