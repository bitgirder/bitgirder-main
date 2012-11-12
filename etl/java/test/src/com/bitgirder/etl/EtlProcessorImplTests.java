package com.bitgirder.etl;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;
import com.bitgirder.lang.TestSums;
import com.bitgirder.lang.ObjectReceiver;

import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.ProcessExit;
import com.bitgirder.process.ProcessRpcClient;
import com.bitgirder.process.ProcessRpcServer;
import com.bitgirder.process.Processes;
import com.bitgirder.process.ProcessActivity;

import com.bitgirder.concurrent.Duration;

import com.bitgirder.test.Test;
import com.bitgirder.test.InvocationFactory;
import com.bitgirder.test.LabeledTestObject;

import java.util.List;

@Test
final
class EtlProcessorImplTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private
    static
    class TestProcessor
    extends AbstractEtlProcessor
    {
        boolean initDone;
        int count;
        private int addend;

        boolean isAsync() { return false; }

        boolean shouldProcess( RecordProcessor rp ) { return true; }

        @Override protected void init() { initDone = true; }

        @Override
        protected
        void
        setProcessorState( Object st )
        {
            this.addend = (Integer) st;
        }

        final
        class RecordProcessor
        extends AbstractRecordSetProcessor
        {
            void
            execProcess()
            {
                for ( Object rec : recordSet() )
                {
                    int i = state.cast( Integer.class, rec );
                    state.equalInt( count++, i );
                }
    
                respond( -count - 1 + addend );
            }

            protected
            void
            startProcess()
            {
                if ( shouldProcess( this ) )
                {
                    if ( isAsync() )
                    {
                        submit(
                            new AbstractTask() {
                                protected void runImpl() { execProcess(); }
                            },
                            Duration.fromMillis( 200 )
                        );
                    }
                    else execProcess();
                }
            }
        }
    }

    private
    abstract
    class TestImpl< P extends AbstractProcess< ? > >
    extends AbstractEtlProcessorTest< P, MemoryEtlTestReactor >
    {
        protected 
        MemoryEtlTestReactor 
        createReactor() 
        { 
            return new MemoryEtlTestReactor(); 
        }

        protected
        EtlTestRecordGenerator
        createGenerator()
        {
            return EtlTests.intGenerator();
        }
    }

    private
    abstract
    class DefaultTest
    extends TestImpl< TestProcessor >
    {
        boolean testProcIsAsync() { return false; }

        Integer
        getExpectedFinalState()
        {
            return Integer.valueOf( -( getFeedLength() + 1 ) );
        }

        protected
        void
        beginAssert( TestProcessor tp,
                     Object lastState,
                     Runnable onComp )
        {
            state.isTrue( tp.initDone );
            state.equalInt( getFeedLength(), tp.count );

            state.equal( getExpectedFinalState(), (Integer) lastState );
            onComp.run();
        }

        protected
        TestProcessor
        createTestProcessor()
        {
            return new TestProcessor() {
                boolean isAsync() { return testProcIsAsync(); }
            };
        }
    }

    @Test private final class BasicTest extends DefaultTest {}

    @Test
    private
    final
    class InitialStateSetTest
    extends DefaultTest
    {
        private final int addend = 5000;

        @Override protected Object getInitialState() { return addend; }

        @Override
        Integer
        getExpectedFinalState()
        {
            return super.getExpectedFinalState() + addend;
        }
    }

    // Make sure that a processor which never processes any records still
    // responds normally to its lifecycle
    @Test
    private
    final
    class EmptyProcessorRunTest
    extends DefaultTest
    {
        @Override protected int getFeedLength() { return 0; }
        @Override Integer getExpectedFinalState() { return null; }
    }

    @Test
    private
    final
    class AsyncProcessCompletionTest
    extends DefaultTest
    {
        @Override boolean testProcIsAsync() { return true; }
    }

    private
    final
    static
    class ShutdownTester
    extends TestProcessor
    {
        private final boolean completeRecords;

        private RecordProcessor rp;

        private
        ShutdownTester( boolean completeRecords )
        {
            this.completeRecords = completeRecords;
        }

        @Override
        protected 
        boolean
        shouldProcess( RecordProcessor rp )
        {
            this.rp = rp;
            return false;
        }

        @Override
        void
        handledShutdownRequest()
        {
            state.notNull( rp );
            if ( completeRecords ) rp.execProcess();
        }
    }

    private
    final
    class ShutdownBehaviorTest
    extends TestImpl< ShutdownTester >
    implements LabeledTestObject
    {
        private final boolean isUrgent;

        private
        ShutdownBehaviorTest( boolean isUrgent )
        {
            this.isUrgent = isUrgent;
        }

        @Override protected int getFeedLength() { return 100; }
        @Override protected int getBatchSize() { return 100; }
        @Override protected boolean sendShutdownOnComplete() { return false; }
        @Override protected boolean expectRecordSetAbort() { return isUrgent; }

        public
        CharSequence
        getLabel()
        {
            return isUrgent ? "abortOnUrgent" : "completesInProgressRecords";
        }

        public Object getInvocationTarget() { return this; }

        protected
        void
        beginAssert( ShutdownTester st,
                     Object lastState,
                     Runnable onComp )
        {
            state.equalInt( isUrgent ? 0 : getFeedLength(), st.count );
            onComp.run();
        }

        protected 
        ShutdownTester 
        createTestProcessor() 
        { 
            return new ShutdownTester( ! isUrgent );
        }

        @Override
        protected
        EventHandlerImpl
        createEventHandler()
        {
            return new EventHandlerImpl()
            {
                public
                void
                recordsSent( EtlRecordSet rs,
                             EtlTestProcessorFeed f )
                {
                    f.sendShutdown( isUrgent );
                }
            };
        }
    }

    @InvocationFactory
    private
    List< LabeledTestObject >
    testShutdownRequestWithRecsInProgress()
    {
        return
            Lang.< LabeledTestObject >asList(
                new ShutdownBehaviorTest( true ),
                new ShutdownBehaviorTest( false )
            );
    }

    private
    final
    static
    class AccTestProcessor
    extends AccumulatingEtlProcessor
    {
        private final AccProcessorTest t;

        private int batches;
        private int sum;

        private Integer procState;
        
        private 
        AccTestProcessor( AccProcessorTest t ) 
        { 
            super(
                new AccumulatingEtlProcessor.Builder() {}.
                    setMinBatchSize( t.minBatchSize )
            );

            this.t = t; 
        }

        @Override
        protected
        boolean
        acceptRecord( Object rec )
        {
            int i = (Integer) rec;
            
            procState = -i - 2;

            return t.filterAt == 0 || ( (Integer) rec ) % t.filterAt > 0;
        }

        private
        final
        class RecordProcessor
        extends AbstractBatchProcessor
        {
            private
            void
            completeBatch()
            {
                if ( t.useAsync )
                {
                    submit(
                        new AbstractTask() { 
                            protected void runImpl() { 
                                batchDone( procState ); 
                            } 
                        },
                        Duration.fromMillis( 100 )
                    );
                }
                else batchDone( procState );
            }
 
            protected
            void
            startBatch()
            {
                ++batches;
                state.isFalse( batch().isEmpty() );
 
                int i = Integer.MIN_VALUE;
                for ( Object o : batch() ) sum += ( i = (Integer) o );
 
                // either this is a fully loaded batch or the end of the feed
                state.isTrue(
                    batch().size() >= t.minBatchSize || recordSet().isFinal() );
 
                completeBatch();
            }
        }
    }

    private
    final
    class AccProcessorTest
    extends TestImpl< AccTestProcessor >
    implements LabeledTestObject
    {
        private final boolean useAsync;
        private final int filterAt;
        private final int minBatchSize;
        private final int feedLen;
        private final int batchesExpct;
        private final Integer lastStateExpct;

        private
        AccProcessorTest( boolean useAsync,
                          int filterAt,
                          int minBatchSize,
                          int feedLen,
                          int batchesExpct,
                          Integer lastStateExpct )
        {
            this.useAsync = useAsync;
            this.filterAt = filterAt;
            this.minBatchSize = minBatchSize;
            this.feedLen = feedLen;
            this.batchesExpct = batchesExpct;
            this.lastStateExpct = lastStateExpct;
        }

        public
        CharSequence
        getLabel()
        {
            return Strings.crossJoin( "=", ",",
                "useAsync", useAsync,
                "filterAt", filterAt,
                "minBatchSize", minBatchSize,
                "feedLen", feedLen
            );
        }

        public Object getInvocationTarget() { return this; }

        @Override protected int getFeedLength() { return feedLen; }

        // Some of the invocations we create assume that record sets have 100
        // elements in them when fed, so we ensure that 
        @Override protected int getBatchSize() { return 100; }

        private
        void
        assertSum( AccTestProcessor p )
        {
            int filtered =
                filterAt == 0 ? 0 : Lang.ceilI( getFeedLength(), filterAt );

            int fullSum = TestSums.ofSequence( 0, getFeedLength() );

            int sumExpct = 
                TestSums.ofSequence( 0, getFeedLength() ) -
                ( TestSums.ofSequence( 0, filtered ) * filterAt );

            state.equalInt( sumExpct, p.sum );
        }

        private
        void
        assertBatchCount( AccTestProcessor p )
        {
            state.equalInt( batchesExpct, p.batches );
        }

        protected
        void
        beginAssert( AccTestProcessor p,
                     Object lastState,
                     Runnable onComp )
        {
            assertSum( p );
            assertBatchCount( p );
            state.equal( lastStateExpct, lastState );
                        
            onComp.run();
        }

        protected
        AccTestProcessor
        createTestProcessor()
        {
            return new AccTestProcessor( this );
        }
    }

    private
    void
    addAccProcTest( List< LabeledTestObject > l,
                    int filterAt,
                    int minBatchSize,
                    int feedLen,
                    int batchesExpct,
                    Integer lastStateExpct )
    {
        for ( int i = 0; i < 2; ++i )
        {
            l.add( 
                new AccProcessorTest( 
                    i == 0, 
                    filterAt, 
                    minBatchSize, 
                    feedLen, 
                    batchesExpct,
                    lastStateExpct ) );
        }
    }

    @InvocationFactory
    private
    List< LabeledTestObject >
    testAccProcessor()
    {
        List< LabeledTestObject > res = Lang.newList();

        // aligned batch/feed, no filter
        addAccProcTest( res, 0, 100, 1000, 10, -1001 );

        // non-aligned batch/feed with nontrivial remainder; no filter
        addAccProcTest( res, 0, 300, 1000, 4, -1001 );

        // batch with batch size bigger than entire feed (single batch)
        addAccProcTest( res, 0, 1001, 1000, 1, -1001 );
 
        // single batch with filter
        addAccProcTest( res, 5, 1000, 1000, 1, -1001 );

        // multiple batches with filter; will accumulate into 5 batches, the
        // first of which coming after the second feed batch and containing 160
        // elements -- all non-multiples of 5 in [0,200).
        addAccProcTest( res, 5, 100, 1000, 5, -1001 );
        
        // filter that filters last feed item
        addAccProcTest( res, 5, 100, 1001, 5, -1001 );
        
        // empty feed with filter
        addAccProcTest( res, 5, 100, 0, 0, null );
        
        // empty feed with no filter
        addAccProcTest( res, 0, 100, 0, 0, null );

        return res;
    }

    // Next bunch of classes and tests are for various detections/allowances in
    // rpc responder registration in AbstractEtlProcessor

    private
    static
    class RecSetProcessor
    extends AbstractRecordSetProcessor
    {
        protected void startProcess() { state.fail(); }
    }
    
    private
    static
    class MultipleRecSetDetectionTest
    extends AbstractEtlProcessor
    {
        @Override protected final void init() { state.fail(); }
    }

    @Test( expected = IllegalStateException.class,
           expectedPattern = 
            "More than one descendant of class com\\.bitgirder\\.etl\\.AbstractRecordSetProcessor exists in class com\\.bitgirder\\.etl\\.EtlProcessorImplTests\\$MultipleLocalRecSetProcessorsDetectedTest or its superclasses: class com\\.bitgirder\\.etl\\.EtlProcessorImplTests\\$MultipleLocalRecSetProcessorsDetectedTest\\$Processor2 and class com\\.bitgirder\\.etl\\.EtlProcessorImplTests\\$MultipleLocalRecSetProcessorsDetectedTest\\$Processor1" )
    private
    final
    class MultipleLocalRecSetProcessorsDetectedTest
    extends MultipleRecSetDetectionTest
    {
        private final class Processor1 extends RecSetProcessor {}
        private final class Processor2 extends RecSetProcessor {}
    }

    private
    static
    class MultipleRecSetProcessorsInheritedDetectedTestBase
    extends MultipleRecSetDetectionTest
    {
        private final class Processor extends RecSetProcessor {}
    }

    @Test( expected = IllegalStateException.class,
           expectedPattern = "More than one descendant of class com\\.bitgirder\\.etl\\.AbstractRecordSetProcessor exists in class com\\.bitgirder\\.etl\\.EtlProcessorImplTests\\$MultipleRecSetProcessorsInheritedDetectedTest or its superclasses: class com\\.bitgirder\\.etl\\.EtlProcessorImplTests\\$MultipleRecSetProcessorsInheritedDetectedTestBase\\$Processor and class com\\.bitgirder\\.etl\\.EtlProcessorImplTests\\$MultipleRecSetProcessorsInheritedDetectedTest\\$Processor" )
    private
    final
    class MultipleRecSetProcessorsInheritedDetectedTest
    extends MultipleRecSetProcessorsInheritedDetectedTestBase
    {
        private final class Processor extends RecSetProcessor {}
    }

    private
    abstract
    static
    class AbstractRecSetProcessorAcceptableTestBase
    extends AbstractEtlProcessor
    {
        abstract class ProcBase extends AbstractRecordSetProcessor {}
    }  

    @Test
    private
    final
    class AbstractRecSetProcessorsAcceptableTest
    extends AbstractRecSetProcessorAcceptableTestBase
    {
        private 
        final 
        class Proc 
        extends ProcBase
        {
            protected void startProcess() { state.fail(); }
        }

        // If init runs then we're good so just exit
        @Override protected void init() { exit(); }
    }
}
