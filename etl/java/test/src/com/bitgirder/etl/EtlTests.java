package com.bitgirder.etl;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.TestSums;
import com.bitgirder.lang.StandardThread;

import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.Processes;
import com.bitgirder.process.ProcessExecutor;
import com.bitgirder.process.ProcessWaiter;
import com.bitgirder.process.ProcessContext;
import com.bitgirder.process.ProcessExit;
import com.bitgirder.process.DefaultProcessExecutor;

import com.bitgirder.test.Test;

// This class furnishes public test methods and interfaces for use elsewhere in
// ETL testing. Because these methods are themselves part of a supported API, we
// have some tests of them in here too, though they're private and for our sake
// only
@Test
public
final
class EtlTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private EtlTests() {}

    public
    static
    AbstractProcess< ? >
    createMemoryStateManager()
    {
        return new MemoryStateManager();
    }

    public
    static
    EtlTestReactor
    createMemoryTestReactor()
    {
        return new MemoryEtlTestReactor();
    }

    public
    static
    void
    call( EtlTestProcessorFeed f )
        throws Exception
    {
        inputs.notNull( f, "f" );

        DefaultProcessExecutor exec = Processes.createExecutor( 1 );

        try
        {
            ProcessWaiter< Void > w = ProcessWaiter.create();
            ProcessContext< Void > ctx = Processes.createRootContext( w, exec );

            Processes.start( f, ctx );
            w.get();
        }
        finally { exec.shutdownNow(); }
    }

    public
    static
    void
    call( AbstractProcess< ? > proc,
          EtlTestRecordGenerator< ? > g,
          long feedLength )
        throws Exception
    {
        call(
            new EtlTestProcessorFeed.Builder().
                addProcessor( proc ).
                setGenerator( g ).
                setReactor( new MemoryEtlTestReactor() ).
                setFeedLength( feedLength ).
                setEventHandler( 
                    new EtlTestProcessorFeed.AbstractEventHandler() {} ).
                build()
            );
    }

    public
    static
    EtlTestRecordGenerator< Integer >
    intGenerator()
    {
        return
            new EtlTestRecordGenerator< Integer >() {
                public Integer next( long i ) { return (int) i; }
            };
    }

    private final static class MarkerException extends Exception {}

    private
    final
    static
    class IntProcessor
    extends AbstractEtlProcessor
    {
        private final int failAt;

        private int sum;

        private IntProcessor( int failAt ) { this.failAt = failAt; }
        private IntProcessor() { this( -1 ); } // don't fail

        private
        final
        class RecSetProcessor
        extends AbstractRecordSetProcessor
        {
            protected
            void
            startProcess()
                throws Exception
            {
                int last = -1;

                for ( Object o : recordSet() )
                {
                    last = (Integer) o;
                    if ( last == failAt ) throw new MarkerException();
                    sum += last;
                }

                respond( -last - 2 );
            }
        }
    }

    @Test
    private
    final
    class BlockingProcessorRunTest
    extends AbstractVoidProcess
    {
        private
        void
        runBlockingTest()
            throws Exception
        {
            int len = 10000;
            
            IntProcessor p = new IntProcessor();

            call( p, intGenerator(), len );
            state.equalInt( TestSums.ofSequence( 0, len ), p.sum );

            exit();
        }

        protected
        void
        startImpl()
        {
            new StandardThread( 
                new AbstractTask() {
                    protected void runImpl() throws Exception {
                        runBlockingTest(); 
                    }
                },
                "etl-blocking-test-%1$d"
            ).
            start();
        }
    }

    private
    abstract
    class FeedMultiProcessorsTest
    extends AbstractVoidProcess
    {
        private final int failAt;

        private FeedMultiProcessorsTest( int failAt ) { this.failAt = failAt; }

        @Override
        protected
        void
        childExited( AbstractProcess< ? > proc,
                     ProcessExit< ? > exit )
        {
            if ( ! exit.isOk() ) fail( exit.getThrowable() );
            else
            {
                EtlTestProcessorFeed f = (EtlTestProcessorFeed) proc;
                state.equalInt( 2, f.getProcessors().size() );

                for ( AbstractProcess< ? > p : f.getProcessors().keySet() )
                {
                    state.equalInt(
                        TestSums.ofSequence( 0, (int) f.getFeedLength() ),
                        ( (IntProcessor) p ).sum
                    );
                }

                exit();
            }
        }

        protected
        void
        startImpl()
        {
            spawn(
                new EtlTestProcessorFeed.Builder().
                    addProcessor( new IntProcessor( failAt ) ).
                    addProcessor( new IntProcessor() ).
                    setGenerator( intGenerator() ).
                    setReactor( new MemoryEtlTestReactor() ).
                    setFeedLength( 1000 ).
                    setBatchSize( 200 ).
                    setEventHandler( 
                        new EtlTestProcessorFeed.AbstractEventHandler() {} ).
                    build()
            );
        }
    }

    @Test
    private
    final
    class FeedMultiProcessorsSuccessTest
    extends FeedMultiProcessorsTest
    {
        private FeedMultiProcessorsSuccessTest() { super( -1 ); }
    }

    @Test( expected = MarkerException.class )
    private
    final
    class FeedMultiProcessorsFailureTest
    extends FeedMultiProcessorsTest
    {
        private FeedMultiProcessorsFailureTest() { super( 30 ); }
    }
}
