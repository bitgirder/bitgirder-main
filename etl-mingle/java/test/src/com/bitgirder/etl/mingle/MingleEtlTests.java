package com.bitgirder.etl.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.TestSums;

import com.bitgirder.etl.AbstractEtlProcessorTest;
import com.bitgirder.etl.AccumulatingEtlProcessor;
import com.bitgirder.etl.EtlTests;
import com.bitgirder.etl.EtlTestReactor;
import com.bitgirder.etl.EtlTestRecordGenerator;

import com.bitgirder.mingle.model.MingleStruct;
import com.bitgirder.mingle.model.MingleStructBuilder;
import com.bitgirder.mingle.model.MingleSymbolMapAccessor;
import com.bitgirder.mingle.model.MingleInt32;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.AtomicTypeReference;

import com.bitgirder.process.AbstractProcess;

import com.bitgirder.test.Test;
import com.bitgirder.test.InvocationFactory;
import com.bitgirder.test.LabeledTestObject;

import java.util.List;

@Test
public
final
class MingleEtlTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public final static AtomicTypeReference TYPE0 = (AtomicTypeReference)
        MingleTypeReference.
            create( "com:bitgirder:etl:mingle@v1/MingleEtlTests/Struct0" );

    public final static AtomicTypeReference TYPE1 = (AtomicTypeReference)
        MingleTypeReference.
            create( "com:bitgirder:etl:mingle@v1/MingleEtlTests/Struct1" );

    private final static MingleIdentifier ID_VAL = 
        MingleIdentifier.create( "val" );

    private MingleEtlTests() {}

    private
    static
    int
    indexOf( MingleStruct ms )
    {
        if ( TYPE0.equals( ms.getType() ) ) return 0;
        else if ( TYPE1.equals( ms.getType() ) ) return 1;
        else throw state.createFail();
    }

    private
    static
    MingleStruct
    getStruct( long l )
    {
        return 
            MingleModels.structBuilder().
                setType( l % 2 == 0 ? TYPE0 : TYPE1 ).
                f().setString( "uid", Lang.randomUuid() ).
                f().setInt32( "val", ( (int) l ) / 2 ).
                build();
    }

    private
    final
    static
    class GeneratorImpl
    implements EtlTestRecordGenerator< MingleStruct >
    {
        public MingleStruct next( long l ) { return getStruct( l ); }
    }

    public
    static
    EtlTestRecordGenerator< MingleStruct >
    createStructGenerator()
    {
        return new GeneratorImpl();
    }

    private
    static
    abstract
    class AbstractTest< P extends AbstractProcess< ? > >
    extends AbstractEtlProcessorTest< P, EtlTestReactor >
    {
        @Override protected int getFeedLength() { return 2000; }

        protected
        EtlTestReactor
        createReactor()
        {
            return EtlTests.createMemoryTestReactor();
        }

        protected
        EtlTestRecordGenerator< MingleStruct >
        createGenerator()
        {
            return createStructGenerator();
        }

        abstract
        void
        beginAssert( P p );

        protected
        void
        beginAssert( P p,
                     Object lastState,
                     Runnable onComp )
        {
            beginAssert( p );
            onComp.run();
        }
    }

    private
    static
    class DefaultProcessor
    extends MingleRecordProcessor
    {
        final int filterMode;

        boolean completeInitCalled;

        final int[] sums = new int[ 2 ];

        private 
        DefaultProcessor( int filterMode )
        {
            this.filterMode = filterMode;
        }

        @Override protected void completeInit() { completeInitCalled = true; }

        @Override
        protected
        void
        initFilter()
        {
            if ( filterMode > 0 )
            {
                acceptType( TYPE0 );
                if ( filterMode == 2 ) acceptType( TYPE1 );
            }
        }

        private
        final
        class Processor
        extends AbstractBatchProcessor
        {
            protected
            void
            startBatch()
            {
                for ( Object o : batch() )
                {
                    MingleStruct ms = (MingleStruct) o;
                    
                    int indx = indexOf( ms );

                    int val =
                        ( (MingleInt32) ms.getFields().get( ID_VAL ) ).
                        intValue();

                    sums[ indx ] += val;
                }

                batchDone( null );
            }
        }
    }

    private
    class DefaultTest
    extends AbstractEtlProcessorTest< DefaultProcessor, EtlTestReactor >
    implements LabeledTestObject
    {
        final int filterMode;

        private DefaultTest( int filterMode ) { this.filterMode = filterMode; }
        private DefaultTest() { this( 0 ); }

        public 
        CharSequence 
        getLabel() 
        { 
            switch ( filterMode )
            {
                case 0 : return "accept-all-implicit";
                case 1 : return "accept-struct0";
                case 2 : return "accept-all-explicit";

                default: throw state.createFail();
            }
        }

        public Object getInvocationTarget() { return this; }

        @Override protected int getFeedLength() { return 2000; }

        protected
        EtlTestReactor
        createReactor()
        {
            return EtlTests.createMemoryTestReactor();
        }

        protected
        EtlTestRecordGenerator< MingleStruct >
        createGenerator()
        {
            return createStructGenerator();
        }

        protected
        void
        beginAssert( DefaultProcessor dp,
                     Object lastState,
                     Runnable onComp )
        {
            state.isTrue( dp.completeInitCalled );

            int fullSum = TestSums.ofSequence( 0, getFeedLength() / 2 );

            state.equalInt( fullSum, dp.sums[ 0 ] );
            state.equalInt( filterMode == 1 ? 0 : fullSum, dp.sums[ 1 ] );

            onComp.run();
        }

        protected
        DefaultProcessor
        createTestProcessor()
        {
            return new DefaultProcessor( filterMode ); 
        }
    }

    @InvocationFactory
    private
    List< LabeledTestObject >
    testMingleRecordProcessor()
    {
        List< LabeledTestObject > res = Lang.newList();

        for ( int i = 0; i < 3; ++i ) res.add( new DefaultTest( i ) );

        return res;
    }

    private
    final
    static
    class BatchAsTypedObjectsProcessor
    extends MingleRecordProcessor
    {
        private int structSum;
        private int symMapSum;

        private
        final
        class Proc
        extends MingleBatchProcessor
        {
            protected
            void
            startBatch()
            {
                for ( MingleSymbolMapAccessor m : symbolMaps() )
                {
                    symMapSum += m.expectInt( "val" );
                }

                for ( MingleStruct ms : structs() )
                {
                    structSum += 
                        ( (MingleInt32) ms.getFields().get( ID_VAL ) ).
                        intValue();
                }

                batchDone( null );
            }
        }
    }

    @Test
    private
    final
    class BatchAsTypedObjectsTest
    extends AbstractTest< BatchAsTypedObjectsProcessor >
    {
        void
        beginAssert( BatchAsTypedObjectsProcessor p )
        {
            int sumExpct = 2 * TestSums.ofSequence( 0, getFeedLength() / 2 );

            state.equalInt( sumExpct, p.symMapSum );
            state.equalInt( sumExpct, p.structSum );
        }

        @Override
        protected
        BatchAsTypedObjectsProcessor
        createTestProcessor()
        {
            return new BatchAsTypedObjectsProcessor();
        }
    }
}
