package com.bitgirder.mingle.bind;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.Stoppable;
import com.bitgirder.process.ProcessExit;
import com.bitgirder.process.ProcessActivity;
import com.bitgirder.process.ProcessBehavior;
import com.bitgirder.process.ProcessRpcClient;
import com.bitgirder.process.ProcessFailureTarget;

import com.bitgirder.mingle.service.MingleServiceEndpoint;
import com.bitgirder.mingle.service.MingleRpcClient;
import com.bitgirder.mingle.service.MingleServices;

import com.bitgirder.test.TestRuntime;

import com.bitgirder.testing.Testing;

import java.util.List;

public
abstract
class AbstractBoundServiceTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final TestRuntime rt;
    private final MingleBinder mb;

    protected
    AbstractBoundServiceTests( TestRuntime rt )
    {
        this.rt = inputs.notNull( rt, "rt" );
        
        this.mb = 
            Testing.expectObject(
                rt, MingleBindTests.KEY_BINDER, MingleBinder.class );
    }

    protected final TestRuntime testRuntime() { return rt; }
    protected final MingleBinder mingleBinder() { return mb; }

    public
    abstract
    class ServiceTest
    extends AbstractVoidProcess
    {
        private final List< Stoppable > stoppables = Lang.newList();
        private MingleServiceEndpoint ep;
        
        protected 
        ServiceTest( ProcessBehavior... behavs )
        {
            super(
                new Builder().
                    mixinAll( behavs ).
                    mixin( ProcessRpcClient.create() )
            );
        }

        private
        < V, P extends AbstractProcess< V > & Stoppable >
        P
        spawnStoppable( P proc )
        {
            spawn( proc );
            stoppables.add( proc );

            return proc;
        }

        // for LabeledTestObject impls
        public Object getInvocationTarget() { return this; }

        @Override
        protected
        final
        void
        childExited( AbstractProcess< ? > proc,
                     ProcessExit< ? > exit )
        {
            if ( ! exit.isOk() ) fail( exit.getThrowable() );
            if ( ! hasChildren() ) exit();
        }

        protected 
        final 
        void 
        testDone() 
        { 
            for ( Stoppable s : stoppables ) s.stop(); 
        }

        protected
        final
        MingleRpcClient
        mingleRpcClient()
        {
            return
                MingleServices.createRpcClient(
                    ep, behavior( ProcessRpcClient.class )
                );
        }

        // A common pattern during testing is for client C to call service S and
        // for service S to crash due to some issue that is being debugged,
        // developed, etc. In this case, C might receive and fail with a
        // ProcessRpcClient.RemoteExitException, masking the more interesting
        // exeption that crashed S. So, we intercept remote exit exceptions,
        // issuing a small warning, but otherwise expect the remote failure to
        // crash the test. Currently hardcoded, though we could let subclasses
        // opt in to how this is handled.
        private
        ProcessActivity.Context
        getClientActivityContext()
        {
            return
                getActivityContext( new ProcessFailureTarget() {
                    public void fail( Throwable th ) 
                    {
                        if ( th instanceof 
                                ProcessRpcClient.RemoteExitException )
                        {
                            warn( 
                                "Remote service exited; ignoring client side" );
                        }
                        else self().fail( th );
                    }
                });
        }
    
        protected
        final
        < C extends BoundServiceClient, 
          B extends BoundServiceClient.Builder< C, B > >
        C
        createClient( B b )
        {
            return
                b.setBinder( mb ).
                  setRpcClient( mingleRpcClient() ).
                  setActivityContext( getClientActivityContext() ).
                  build();
        }

        protected
        abstract
        void
        startTest()
            throws Exception;

        protected
        final
        void
        initDone( MingleServiceEndpoint.Builder b )
        {
            ep = spawnStoppable( b.build() );
            try { startTest(); } catch ( Throwable th ) { fail( th ); }
        }

        protected
        final
        < B extends BoundService.Builder< B > >
        B
        initBuilder( B b )
        {
            b.setBinder( mb );

            return b;
        }

        protected
        final
        void
        spawnService( BoundService svc,
                      MingleServiceEndpoint.Builder b )
        {
            inputs.notNull( svc, "svc" );

            spawnStoppable( svc );
            BoundServices.addRoute( svc, b );
        }

        protected
        abstract
        void
        initServices( MingleServiceEndpoint.Builder b )
            throws Exception;

        protected
        final
        void
        startImpl()
            throws Exception
        {
            initServices( new MingleServiceEndpoint.Builder() );
        }
    }
}
