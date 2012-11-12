package com.bitgirder.mingle.json;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.io.ProtocolRoundtripTest;
import com.bitgirder.io.ProtocolProcessors;
import com.bitgirder.io.ProtocolProcessor;
import com.bitgirder.io.AbstractProtocolProcessor;
import com.bitgirder.io.IoUtils;

import com.bitgirder.mingle.model.MingleStruct;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleInt64;

import com.bitgirder.mingle.codec.MingleCodecs;

import com.bitgirder.test.Test;

import java.nio.ByteBuffer;

import java.util.List;

@Test
final
class JsonMingleStructObjectLineProcessorTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static MingleIdentifier ID_ID = 
        MingleIdentifier.create( "id" );

    private
    static
    MingleStruct
    createTestEntry( int i )
    {
        return
            MingleModels.structBuilder().
                setType( "test@v1/Entry" ).
                f().setInt64( ID_ID, i ).
                build();
    }
 
    private
    abstract
    class AbstractRoundtripTest
    extends ProtocolRoundtripTest
    {
        private final JsonMingleCodec codec = 
            JsonMingleCodecs.getJsonCodec( "\n" );

        private int nextIdExpct;

        int objCount() { return 100; }

        protected
        void
        beginAssert()
        {
            state.equalInt( objCount(), nextIdExpct );
            assertDone();
        }

        private
        ByteBuffer
        createStructLine( int id )
            throws Exception
        {
            MingleStruct ms = createTestEntry( id );
            return MingleCodecs.toByteBuffer( codec, ms );
        }

        private
        List< ByteBuffer >
        createStructLines( int begin,
                           int end )
            throws Exception
        {
            List< ByteBuffer > res = Lang.newList( end - begin );

            for ( int i = begin; i < end; ++i )
            {
                res.add( createStructLine( i ) );
            }

            return res;
        }

        private
        List< ByteBuffer >
        createStructLines( int count )
            throws Exception
        {
            return createStructLines( 0, count );
        }

        ProtocolProcessor< ByteBuffer >
        createSender()
            throws Exception
        {
            return 
                ProtocolProcessors.
                    createBufferSend( createStructLines( objCount() ) );
        }

        final
        void
        consume( MingleStruct ms )
        {
            int id = 
                ( (MingleInt64) ms.getFields().get( ID_ID ) ).intValue();

            state.equal( nextIdExpct, id );
            ++nextIdExpct;
        }

        private
        final
        class StructProcessor
        extends AbstractProtocolProcessor< List< MingleStruct > >
        {
            protected
            void
            processImpl( ProcessContext< List< MingleStruct > > ctx )
            {
                for ( MingleStruct ms : ctx.object() ) consume( ms );
                doneOrData( ctx );
            }
        }

        ProtocolProcessor< List< MingleStruct > >
        createInputProcessor()
        {
            return new StructProcessor();
        }

        private
        JsonMingleStructLineProcessor.Reactor
        getReactor()
        {
            return new JsonMingleStructLineProcessor.AbstractReactor() {};
        }

        private
        ProtocolProcessor< ByteBuffer >
        createReceiver()
        {
            return
                new JsonMingleStructLineProcessor.Builder().
                    setInputProcessor( createInputProcessor() ).
                    setReactor( getReactor() ).
                    build();
        }

        protected
        void
        startTest()
            throws Exception
        {
            setSender( createSender() );
            setReceiver( createReceiver() );

            testReady();
        }
    }

    @Test 
    private final class ImmediateRoundtripTest extends AbstractRoundtripTest {}

    @Test
    private
    final
    class EmptyInputRoundtripTest
    extends AbstractRoundtripTest
    {
        @Override int objCount() { return 0; }

        @Override
        ProtocolProcessor< ByteBuffer >
        createSender()
        {
            return
                ProtocolProcessors.
                    createBufferSend( IoUtils.emptyByteBuffer() );
        }
    }
}
