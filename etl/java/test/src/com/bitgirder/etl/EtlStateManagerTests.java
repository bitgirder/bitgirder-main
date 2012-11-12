package com.bitgirder.etl;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.ProcessExit;
import com.bitgirder.process.ProcessRpcClient;

import com.bitgirder.mingle.model.MingleIdentifiedName;
import com.bitgirder.mingle.model.MingleIdentifiedNameGenerator;

import com.bitgirder.test.Test;

public
abstract
class EtlStateManagerTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static MingleIdentifiedNameGenerator nmGen =
        MingleIdentifiedNameGenerator.forPrefix(
            "bitgirder:etl:sql@v1/etlStateManagerTests" );

    protected EtlStateManagerTests() {}

    protected
    abstract
    EtlTestReactor
    createTestReactor()
        throws Exception;
    
    private
    abstract
    class AbstractTest
    extends AbstractVoidProcess
    {
        final MingleIdentifiedName procId = nmGen.next();

        private EtlTestReactor reactor;

        private AbstractTest() { super( ProcessRpcClient.create() ); }

        @Override
        protected
        void
        childExited( AbstractProcess< ? > proc,
                     ProcessExit< ? > exit )
        {
            if ( ! exit.isOk() ) fail( exit.getThrowable() );
            if ( ! hasChildren() ) exit();
        }

        final
        void
        testDone()
            throws Exception
        {
            reactor.stopTestProcesses( getActivityContext() );            
        }

        final
        AbstractProcess< ? >
        stateManager() 
        { 
            return reactor.getStateManager(); 
        }

        abstract
        void
        startTest()
            throws Exception;
        
        protected
        final
        void
        startImpl()
            throws Exception
        {
            ( reactor = createTestReactor() ).startTestProcesses(
                getActivityContext(),
                new AbstractTask() {
                    protected void runImpl() throws Exception { startTest(); }
                }
            );
        }
    }

    // should return null if and only if i is null
    protected
    abstract
    Object
    createStateObject( Integer i )
        throws Exception;

    @Test
    private
    final
    class GetSetAndReplaceProcStateTest
    extends AbstractTest
    {
        private final Integer stopAt = Integer.valueOf( 2 );

        private
        void
        setState( final Integer nextNum )
            throws Exception
        {
            EtlProcessors.SetProcessorState req =
                EtlProcessors.createSetProcessorState( 
                    procId, createStateObject( nextNum ) 
                );

            beginRpc(
                stateManager(),
                req,
                new DefaultRpcHandler() {
                    @Override protected void rpcSucceeded() {
                        expectState( nextNum );
                    }
                }
            );
        }

        private
        void
        expectState( final Integer expctNum )
        {
            beginRpc(
                stateManager(),
                EtlProcessors.createGetProcessorState( procId ),
                new DefaultRpcHandler() 
                {
                    @Override 
                    protected void rpcSucceeded( Object resp ) throws Exception
                    {
                        state.equal( createStateObject( expctNum ), resp );

                        if ( stopAt.equals( expctNum ) ) testDone();
                        else setState( expctNum == null ? 1 : expctNum + 1 );
                    }
                }
            );
        }

        protected void startTest() { expectState( null ); }
    }
    
//    protected
//    abstract
//    Object
//    createFeedPosition()
//        throws Exception;
//
//    @Test
//    private
//    final
//    class GetSetAndReplaceProcFeedPositionTest
//    extends AbstractTest
//    {
//        private
//        void
//        setObject()
//            throws Exception
//        {
//            final Object pos = createFeedPosition();
//
//            beginRpc(
//                stateManager(),
//                EtlProcessors.createSetProcessorFeedPosition( procId, pos ),
//                new DefaultRpcHandler() {
//                    @Override protected void rpcSucceeded() {
//                        expectObject( pos );
//                    }
//                }
//            );
//        }
//
//        private
//        void
//        expectObject( final Object expct )
//        {
//            beginRpc(
//                stateManager(),
//                EtlProcessors.createGetProcessorFeedPosition( procId ),
//                new DefaultRpcHandler() 
//                {
//                    @Override 
//                    protected void rpcSucceeded( Object pos ) throws Exception
//                    {
//                        state.equal( expct, pos );
//                        if ( expct == null ) setObject(); else testDone();
//                    }
//                }
//            );
//        }
//
//        protected void startTest() { expectObject( null ); }
//    }
}
