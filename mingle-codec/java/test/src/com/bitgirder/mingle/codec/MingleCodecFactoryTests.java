package com.bitgirder.mingle.codec;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

import com.bitgirder.io.Charsets;
import com.bitgirder.io.IoUtils;

import com.bitgirder.test.Test;
import com.bitgirder.test.InvocationFactory;

import java.util.List;

import java.nio.ByteBuffer;

@Test
final
class MingleCodecFactoryTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private
    static
    abstract
    class NoOpCodec
    implements MingleCodec
    {
        public
        MingleEncoder
        createEncoder( Object obj )
        {
            throw new UnsupportedOperationException( "Unimplemented" );
        }

        public
        < E >
        MingleDecoder< E >
        createDecoder( Class< E > cls )
        {
            throw new UnsupportedOperationException( "Unimplemented" );
        }
    }

    private final static class Codec1 extends NoOpCodec {}
    private final static class Codec2 extends NoOpCodec {}
    private final static class Codec3 extends NoOpCodec {}
    private final static class Codec4 extends NoOpCodec {}

    private
    final
    static
    class NoOpCodecDetector
    implements MingleCodecDetector
    {
        private final ByteBuffer match;

        private
        NoOpCodecDetector( ByteBuffer match )
        {
            this.match = match;
        }

        public
        Boolean
        update( ByteBuffer bb )
        {
            while ( bb.hasRemaining() && match.hasRemaining() )
            {
                if ( match.get() != bb.get() ) return Boolean.FALSE;
            }

            return match.hasRemaining() ? null : Boolean.TRUE;
        }
    }

    private
    final
    static
    class NoOpCodecDetectorFactory
    implements MingleCodecDetectorFactory
    {
        private final ByteBuffer match;

        private
        NoOpCodecDetectorFactory( ByteBuffer match )
        {
            this.match = match;
        }

        public
        MingleCodecDetector
        createCodecDetector()
        {
            return new NoOpCodecDetector( match.duplicate() );
        }
    }

    private
    void
    addCodec( MingleCodecFactory.Builder b,
              MingleCodec codec,
              CharSequence pref )
        throws Exception
    {
        NoOpCodecDetectorFactory detFact =
            new NoOpCodecDetectorFactory( 
                Charsets.UTF_8.asByteBuffer( pref ) );
        
        b.addCodec( 
            codec.getClass().getSimpleName().toLowerCase(),
            codec,
            detFact
        );
    }

    private
    MingleCodecFactory
    noOpCodecFactory()
        throws Exception
    {
        MingleCodecFactory.Builder b = new MingleCodecFactory.Builder();

        addCodec( b, new Codec1(), "aaa" );
        addCodec( b, new Codec2(), "aab" );
        addCodec( b, new Codec3(), "ddd" );
        addCodec( b, new Codec4(), "ddd" );

        return b.build();
    }

    private
    final
    class NoOpCodecDetectionTest
    extends AbstractCodecDetectionTest< NoOpCodecDetectionTest >
    {
        private NoOpCodecDetectionTest( CharSequence lbl ) { super( lbl ); }

        private
        NoOpCodecDetectionTest
        setInput( CharSequence str )
            throws Exception
        {
            return setInput( Charsets.UTF_8.asByteBuffer( str ) );
        }

        protected
        MingleCodecFactory
        codecFactory()
            throws Exception
        {
            return noOpCodecFactory();
        }
    }

    @InvocationFactory
    private
    List< NoOpCodecDetectionTest >
    testCodecDetection()
        throws Exception
    {
        return Lang.asList(
 
            new NoOpCodecDetectionTest( "expct-codec1" ).
                setInput( "aaaaa" ).
                expectCodec( Codec1.class ),
 
            new NoOpCodecDetectionTest( "expct-codec2" ).
                setInput( "aabaa" ).
                expectCodec( Codec2.class ),
 
            new NoOpCodecDetectionTest( "expct-no-codec" ).
                setInput( "aaccccc" ).
                expectFailure( NoSuchMingleCodecException.class, "" ),
 
            new NoOpCodecDetectionTest( "ambiguous-codec-fails" ).
                setInput( "dddbbb" ).
                expectFailure(
                    MingleCodecDetectionException.class,
                    "\\QMultiple matching codecs detected: codec4, codec3\\E"
                ),
            
            new NoOpCodecDetectionTest( "insufficient-input-fails" ).
                setInput( "aa" ).
                expectFailure(
                    MingleCodecDetectionException.class,
                    "\\QDetection has not completed\\E"
                ),

            new NoOpCodecDetectionTest( "expct-codec1-single-feed" ).
                setInput( "aaaaa" ).
                expectCodec( Codec1.class ).
                setCopyBufferSize( 10 )
        );
    }

    @Test( expected = NoSuchMingleCodecException.class )
    private
    void
    testEmptyFactoryDetectionWellDefined()
        throws Exception
    {
        MingleCodecFactory f = new MingleCodecFactory.Builder().build();

        MingleCodecDetection det = f.createCodecDetection();
        state.isTrue( det.update( ByteBuffer.allocate( 1 ) ) );
        det.getResult();
    }

    private
    MingleCodec
    detectCodec( CharSequence input )
        throws Exception
    {
        ByteBuffer bb = Charsets.UTF_8.asByteBuffer( input );
        ByteBuffer expctPost = bb.duplicate();

        MingleCodec res = MingleCodecs.detectCodec( noOpCodecFactory(), bb );

        // if successfuly, assert that MingleCodecs.detectCodec() does not
        // affect original buf
        state.equal( expctPost, bb ); 
        
        return res;
    }

    // Just get some basic coverage of the ByteBuffer-based detect methods
    @Test
    private
    void
    testDetectFromByteBuffer()
        throws Exception
    {
        state.cast( Codec2.class, detectCodec( "aabaa" ) );
    }

    @Test( expected = MingleCodecDetectionException.class,
           expectedPattern = "Detection has not completed" )
    private
    void
    testDetectFromByteBufferInsufficientInput()
        throws Exception
    {
        detectCodec( "aa" );
    }

    @Test( expected = NoSuchMingleCodecException.class )
    private
    void
    testDetectFromByteBufferNoMatchDefinitive()
        throws Exception
    {
        detectCodec( "ff" );
    }
}
