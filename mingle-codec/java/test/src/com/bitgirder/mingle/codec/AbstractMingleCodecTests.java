package com.bitgirder.mingle.codec;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.ObjectReceiver;
import com.bitgirder.lang.Strings;
import com.bitgirder.lang.Lang;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.io.IoProcessor;
import com.bitgirder.io.IoUtils;
import com.bitgirder.io.IoTestFactory;
import com.bitgirder.io.IoTestSupport;
import com.bitgirder.io.FileWrapper;
import com.bitgirder.io.ProtocolProcessor;
import com.bitgirder.io.ProtocolProcessors;
import com.bitgirder.io.ProtocolProcessorTests;
import com.bitgirder.io.ProtocolCopies;
import com.bitgirder.io.ByteBufferAccumulator;

import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.ProcessExit;

import com.bitgirder.mingle.model.ModelTestInstances;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleNull;
import com.bitgirder.mingle.model.MingleList;
import com.bitgirder.mingle.model.MingleStruct;

import com.bitgirder.test.Test;
import com.bitgirder.test.TestFailureExpector;
import com.bitgirder.test.InvocationFactory;
import com.bitgirder.test.LabeledTestObject;
import com.bitgirder.test.TestRuntime;

import java.util.List;
import java.util.Map;
import java.util.Iterator;

import java.nio.ByteBuffer;
import java.nio.ByteOrder;

public
abstract
class AbstractMingleCodecTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final TestRuntime rt;

    protected
    AbstractMingleCodecTests( TestRuntime rt )
    {
        this.rt = inputs.notNull( rt, "rt" );
    }

    protected final TestRuntime testRuntime() { return rt; }

    protected
    abstract
    MingleCodec
    getCodec()
        throws Exception;

    protected
    final
    MingleDecoder< MingleStruct >
    structDecoder()
        throws Exception
    {
        return getCodec().createDecoder( MingleStruct.class );
    }
 
    protected void debugEncoded( ByteBuffer ser ) throws Exception {}

    protected
    final
    ByteBuffer
    toByteBuffer( MingleStruct ms )
        throws Exception
    {
        return MingleCodecs.toByteBuffer( getCodec(), ms );
    }

    protected
    final
    MingleStruct
    fromByteBuffer( ByteBuffer bb )
        throws Exception
    {
        return 
            MingleCodecs.fromByteBuffer( getCodec(), bb, MingleStruct.class );
    }

    protected
    abstract
    class CodecRoundtripTest< T extends CodecRoundtripTest< T > >
    extends AbstractVoidProcess
    implements TestFailureExpector
    {
        private MingleStruct expct;
        private int encBufSz = 10;
        private int decBufSz = 10;
        private ByteOrder order = ByteOrder.LITTLE_ENDIAN;

        private Class< ? extends Throwable > errCls;
        private CharSequence errPat;

        public Object getInvocationTarget() { return this; }

        protected void addLabelToks( List< Object > toks ) {}

        public
        final
        CharSequence
        getLabel()
        {
            List< Object > toks = Lang.newList();
            addLabelToks( toks );

            toks.addAll(
                Lang.< Object >asList(
                    "encBufSz", encBufSz,
                    "decBufSz", decBufSz,
                    "order", order
                )
            );

            return Strings.crossJoin( "=", ",", toks );
        }

        public 
        Class< ? extends Throwable >
        expectedFailureClass()
        {
            return errCls;
        }

        public CharSequence expectedFailurePattern() { return errPat; }

        private T castThis() { return Lang.castUnchecked( this ); }

        public
        final
        T
        setStruct( MingleStruct expct )
        {
            this.expct = inputs.notNull( expct, "expct" );
            return castThis();
        }

        public
        final
        T
        setEncodeBufferSize( int encBufSz )
        {
            this.encBufSz = inputs.positiveI( encBufSz, "encBufSz" );
            return castThis();
        }

        public
        final
        T
        setByteOrder( ByteOrder order )
        {
            this.order = inputs.notNull( order, "order" );
            return castThis();
        }

        public
        final
        T
        setDecodeBufferSize( int decBufSz )
        {
            this.decBufSz = inputs.positiveI( decBufSz, "decBufSz" );
            return castThis();
        }

        public
        final
        T
        expectError( Class< ? extends Throwable > errCls,
                     CharSequence errPat )
        {
            this.errCls = inputs.notNull( errCls, "errCls" );
            this.errPat = inputs.notNull( errPat, "errPat" );

            return castThis();
        }

        private
        ByteBuffer
        createXferBuf( int sz )
        {
            return ByteBuffer.allocate( sz ).order( order );
        }

        private
        void
        beginAssert( MingleStruct actual )
            throws Exception
        {
            ModelTestInstances.assertEqual( expct, actual );
            exit();
        }

        private
        ProtocolProcessor< ByteBuffer >
        createSender( MingleCodec c )
        {
            return MingleCodecs.createSendProcessor( c, expct );
        }

        private
        ProtocolProcessor< ByteBuffer >
        getDecodeProcessor( MingleCodec codec )
            throws Exception
        {
            return
                MingleCodecs.createReceiveProcessor( 
                    codec, 
                    MingleStruct.class,
                    new ObjectReceiver< MingleStruct >() {
                        public void receive( MingleStruct actual )
                            throws Exception
                        {
                            beginAssert( actual );
                        }
                    }
                );
        }

        private
        void
        encoded( ByteBuffer enc,
                 MingleCodec codec )
            throws Exception
        {
            debugEncoded( enc );

            ProtocolCopies.copyByteStream(
                ProtocolProcessors.createBufferSend( enc ),
                getDecodeProcessor( codec ),
                createXferBuf( decBufSz ),
                getActivityContext(),
                new AbstractTask() { protected void runImpl() {} }
            );
        }

        protected
        final
        void
        startImpl()
            throws Exception
        {
            state.notNull( expct, "expct" );
            final MingleCodec codec = getCodec();

            final ByteBufferAccumulator acc = 
                ByteBufferAccumulator.create( 512 );

            ProtocolCopies.copyByteStream(
                MingleCodecs.createSendProcessor( codec, expct ),
                acc,
                createXferBuf( encBufSz ),
                getActivityContext(),
                new AbstractTask() {
                    protected void runImpl() throws Exception 
                    {
                        ByteBuffer enc = 
                            ProtocolProcessorTests.toByteBuffer( acc );

                        encoded( enc, codec );
                    }
                }
            );
        }
    }

    protected
    final
    class BasicRoundtripTest
    extends CodecRoundtripTest< BasicRoundtripTest >
    implements LabeledTestObject
    {
        private final CharSequence objName;

        public
        BasicRoundtripTest( CharSequence objName )
        {
            this.objName = inputs.notNull( objName, "objName" );
        }

        @Override
        protected
        void
        addLabelToks( List< Object > toks )
        {
            toks.add( "objName" );
            toks.add( objName );
        }
    }

    private
    void
    addStandardBasicRoundtripTests( List< BasicRoundtripTest > l )
    {
        for ( Map.Entry< MingleIdentifier, MingleStruct > e :
                MingleCodecTests.getStandardTestStructs().entrySet() )
        {
            l.add(
                new BasicRoundtripTest( e.getKey().getExternalForm() ).
                    setStruct( e.getValue() )
            );
        }
    }

    protected
    void
    implAddBasicRoundtripTests( List< BasicRoundtripTest > l )
    {}

    @InvocationFactory
    private
    List< BasicRoundtripTest >
    testBasicRoundtrip()
    {
        List< BasicRoundtripTest > res = Lang.newList();

        addStandardBasicRoundtripTests( res );
        implAddBasicRoundtripTests( res );

        return res;
    }

    @Test
    private
    final
    class CodecReceiverSquashesEmptyBufferTest
    extends AbstractVoidProcess
    {
        private final MingleStruct ms = 
            MingleModels.structBuilder().
                setType( "ns1@v1/S1" ).
                build();
        
        private
        List< ByteBuffer >
        getInputBufs( MingleCodec codec )
            throws Exception
        {
            ByteBuffer bb = MingleCodecs.toByteBuffer( codec, ms );

            List< ByteBuffer > res = Lang.newList();

            res.add( (ByteBuffer) bb.slice().limit( 10 ) );
            res.add( IoUtils.emptyByteBuffer() );
            res.add( (ByteBuffer) bb.slice().position( 10 ) );

            return res;
        }

        private
        void
        feed( final Iterator< ByteBuffer > it,
              final ProtocolProcessor< ByteBuffer > proc )
            throws Exception
        {
            ProtocolProcessors.process(
                proc,
                it.next(),
                ! it.hasNext(), // important that this follows it.next()
                getActivityContext(),
                new ObjectReceiver< Boolean >() {
                    public void receive( Boolean b ) throws Exception {
                        if ( ! b ) feed( it, proc );
                    }
                }
            );
        }

        private
        void
        checkDecode( MingleStruct ms2 )
            throws Exception
        {
            ModelTestInstances.assertEqual( ms, ms2 );
            exit();
        }

        protected
        void
        startImpl()
            throws Exception
        {
            MingleCodec codec = getCodec();

            feed( 
                getInputBufs( codec ).iterator(),
                MingleCodecs.createReceiveProcessor(
                    codec.createDecoder( MingleStruct.class ),
                    new ObjectReceiver< MingleStruct >() {
                        public void receive( MingleStruct ms2 ) 
                            throws Exception 
                        {
                            checkDecode( ms2 );
                        }
                    }
                )
            );
        }
    }

    @Test
    private
    void
    testFinalEmptyBuffer()
        throws Exception
    {
        MingleCodec codec = getCodec();

        MingleStruct ms = ModelTestInstances.TEST_STRUCT1_INST1;

        ByteBuffer bb = MingleCodecs.toByteBuffer( codec, ms );

        MingleDecoder< MingleStruct > dec =
            codec.createDecoder( MingleStruct.class );

        if ( ! dec.readFrom( bb, false ) )
        {
            bb.position( bb.limit() ); // make it empty
            dec.readFrom( bb, true );
        }

        ModelTestInstances.assertEqual( ms, dec.getResult() );
    }

    @Test
    private
    final
    class MingleStructFileRoundtripTest
    extends AbstractVoidProcess
    {
        private final MingleStruct mgVal =
            ModelTestInstances.TEST_STRUCT1_INST1;

        private
        MingleStructFileRoundtripTest()
        {
            super( IoTestSupport.create( rt ) );
        }

        private
        void
        readObject( MingleCodec codec,
                    FileWrapper fw )
        {
            MingleCodecs.fromFile(
                codec,
                fw,
                MingleStruct.class,
                getActivityContext(),
                behavior( IoTestSupport.class ).ioProcessor(),
                new ObjectReceiver< MingleStruct >() {
                    public void receive( MingleStruct ms ) 
                        throws Exception
                    {
                        ModelTestInstances.assertEqual( mgVal, ms );
                        exit();
                    }
                }
            );
        }

        protected
        void
        startImpl()
            throws Exception
        {
            final MingleCodec codec = getCodec();
            final FileWrapper fw = IoTestFactory.createTempFile();

            MingleCodecs.toFile( 
                codec, 
                mgVal, 
                fw,
                getActivityContext(),
                behavior( IoTestSupport.class ).ioProcessor(),
                new AbstractTask() {
                    protected void runImpl() { readObject( codec, fw ); }
                }
            );
        }
    }
}
