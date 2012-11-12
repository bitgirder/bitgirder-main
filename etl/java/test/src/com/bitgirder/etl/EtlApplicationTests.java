package com.bitgirder.etl;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.concurrent.Duration;

import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.Stoppable;

import com.bitgirder.process.management.AbstractProcessControl;
import com.bitgirder.process.management.ProcessThrashTracker;
import com.bitgirder.process.management.ProcessControl;
import com.bitgirder.process.management.ProcessFactory;
import com.bitgirder.process.management.ProcessManagement;

import com.bitgirder.test.Test;

//@Test
public
final
class EtlApplicationTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private
    final
    static
    class BinaryProcessor
    extends AbstractEtlProcessor
    {
        private
        final
        class RecordProcessor
        extends AbstractRecordSetProcessor
        {
            protected
            void
            startProcess()
            {
                throw new UnsupportedOperationException( "Unimplemented" );
            }
        }
    }

    private
    final
    static
    class LineProcessor
    extends AbstractEtlProcessor
    {
        private
        final
        class RecordProcessor
        extends AbstractRecordSetProcessor
        {
            protected
            void
            startProcess()
            {
                throw new UnsupportedOperationException( "Unimplemented" );
            }
        }
    }

    private
    static
    abstract
    class SourceListerImpl
    extends AbstractVoidProcess
    implements Stoppable
    {
        protected
        void
        startImpl()
        {
            throw new UnsupportedOperationException( "Unimplemented" );
        }

        public final void stop() { exit(); }
    }

    private
    final
    static
    class BinarySourceLister
    extends SourceListerImpl
    {}

    private
    final
    static
    class LineSourceLister
    extends SourceListerImpl
    {}

    private
    static
    abstract
    class RecordReaderImpl
    extends AbstractVoidProcess
    implements Stoppable
    {
        protected 
        void 
        startImpl() 
        {
            throw new UnsupportedOperationException( "Unimplemented" );
        }

        public final void stop() { exit(); }
    }

    private
    final
    static
    class BinaryRecordReader
    extends RecordReaderImpl
    {}

    private
    final
    static
    class LineRecordReader
    extends RecordReaderImpl
    {}

    @Test
    private
    final
    class BasicBehaviorTest
    extends EtlApplicationTest
    {
        private
        ProcessControl< ? >
        createStateManagerControl()
        {
            return
                ProcessManagement.createRpcStoppableControl(
                    new ProcessFactory< MemoryStateManager >() {
                        public MemoryStateManager newProcess() {
                            return new MemoryStateManager();
                        }
                    },
                    getActivityContext(),
                    1,
                    Duration.fromMinutes( 1 )
                );
        }
    
        private
        EtlProcessorGroup
        createBinaryGroup()
        {
            return
                new EtlProcessorGroup.Builder().
                    addProcessor( 
                        nextId(), factoryFor( BinaryProcessor.class ) ).
                    addProcessor(
                        nextId(), factoryFor( BinaryProcessor.class ) ).
                    setStateManager( createStateManagerControl() ).
                    setSourceLister( controlFor( BinarySourceLister.class ) ).
                    setRecordReader( controlFor( BinaryRecordReader.class ) ).
                    build();
        }

        private
        EtlProcessorGroup
        createLineGroup()
        {
            return
                new EtlProcessorGroup.Builder().
                    addProcessor( nextId(), factoryFor( LineProcessor.class ) ).
                    addProcessor( nextId(), factoryFor( LineProcessor.class ) ).
                    setStateManager( createStateManagerControl() ).
                    setSourceLister( controlFor( LineSourceLister.class ) ).
                    setRecordReader( controlFor( LineRecordReader.class ) ).
                    build();
        }

        protected
        void
        startTest()
        {
            EtlApplication.Builder b = nextAppBuilder();

            b.addGroup( "binGrp", createBinaryGroup() );
            b.addGroup( "lineGrp", createLineGroup() );

            spawnApplication( b.build() );
        }
    }

    // To test:
    //
    //  - basic operation in which all procs start from head, process through to
    //  tail, app shuts down, and final states are correct
    //
    //  - basic operation as above but in which one of the processors fails
    //  beyond restart; others should complete and we should receive the failure
    //  notification somehow and assert that the partial progress of the crashed
    //  processor was recorded
    //
    //  - resumption from failure with a processor that does not fail as it did
    //  before and confirmation that it ultimately completes its processing
    //
    //  - operation in which all processors are correct but lister or record
    //  sources periodically crash and restart
    //
    //  - basic stop/resume of an otherwise normally functioning app should
    //  be equivalent of letting it run to completion normally (nothing should
    //  be missed, okay if records are seen more than once)
    //
    //  - configuration failures (groups with no processors; groups with no
    //  lister or rec set source, groups with same proc ids within the group,
    //  distinct groups each containing the same proc id)
    //
    //  - make sure EtlApplication with no groups fails build()
    //
    //  - failure in getting or setting initial state for a processor should
    //  remove that processor from service but continue with others
    //
    //  - test of procs with and without initial state
}
