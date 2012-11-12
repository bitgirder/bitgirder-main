package com.bitgirder.mingle.bincodec;

import static com.bitgirder.mingle.bincodec.MingleBinaryCodecConstants.*;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

import com.bitgirder.io.IoUtils;
import com.bitgirder.io.IoTests;
import com.bitgirder.io.Charsets;

import com.bitgirder.mingle.codec.AbstractMingleCodecTests;
import com.bitgirder.mingle.codec.MingleCodec;
import com.bitgirder.mingle.codec.MingleCodecTests;
import com.bitgirder.mingle.codec.MingleDecoder;
import com.bitgirder.mingle.codec.MingleCodecException;

import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleStruct;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.ModelTestInstances;

import com.bitgirder.test.Test;
import com.bitgirder.test.Tests;
import com.bitgirder.test.TestFactory;
import com.bitgirder.test.TestRuntime;

import java.nio.ByteBuffer;
import java.nio.ByteOrder;

import java.util.List;

@Test
final
class BinCodecTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();
 
    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static MingleStruct STRUCT1_INST1 =
        MingleModels.structBuilder().
            setType( "ns@v1/S1" ).f().
            setString( "f1", "val1" ).
            build();
    
    private final static MingleStruct STRUCT2_INST1 =
        MingleModels.structBuilder().
            setType( "ns@v1/S2" ).f().
            setString( "test-id", "val1" ).
            build();
 
    // this will fail for the moment since we don't use type code 100
    private final static byte TYPE_CODE_FAIL = (byte) 100;

    private final TestRuntime rt;

    private BinCodecTests( TestRuntime rt ) { this.rt = rt; }

    private
    final
    static
    class TestImpl
    extends AbstractMingleCodecTests
    {
        private final MingleCodec codec = MingleBinaryCodecs.getCodec();

        private TestImpl( TestRuntime rt ) { super( rt ); }

        protected MingleCodec getCodec() { return codec; }

        @Override
        protected
        void
        debugEncoded( ByteBuffer ser )
        {
//            code( "Encoded:", IoUtils.asHexString( ser ) );
        }

        private
        void
        encodeByte( Byte b, 
                    List< ByteBuffer > bufs )
        {
            bufs.add( ByteBuffer.wrap( new byte[] { b.byteValue() } ) );
        }

        private
        void
        encodeInt32( Integer i,
                     List< ByteBuffer > bufs )
        {
            ByteBuffer res = MingleBinaryCodecs.allocateBuffer( 4 );
            res.putInt( i );
            res.flip();

            bufs.add( res );
        }

        private
        void
        encodeSizedBuffer( ByteBuffer bb,
                           List< ByteBuffer > bufs )
        {
            encodeInt32( bb.remaining(), bufs );
            bufs.add( bb );
        }

        private
        void
        encodeUtf8( CharSequence str,
                    List< ByteBuffer > bufs )
            throws Exception
        {
            encodeByte( TYPE_CODE_UTF8_STRING, bufs );
            encodeSizedBuffer( Charsets.UTF_8.asByteBuffer( str ), bufs );
        }

        private
        void
        encodeTypeReference( MingleTypeReference typ,
                             List< ByteBuffer > bufs )
            throws Exception
        {
            encodeUtf8( typ.getExternalForm(), bufs );
        }

        private
        void
        encodeIdentifier( MingleIdentifier id,
                          List< ByteBuffer > bufs )
            throws Exception
        {
            encodeUtf8( id.getExternalForm(), bufs );
        }

        private
        void
        encodeToken( Object obj,
                     List< ByteBuffer > bufs )
            throws Exception
        {
            if ( obj instanceof Byte ) encodeByte( (Byte) obj, bufs );
            else if ( obj instanceof Integer ) 
            {
                encodeInt32( (Integer) obj, bufs );
            }
            else if ( obj instanceof CharSequence )
            {
                encodeUtf8( (CharSequence) obj, bufs );
            }
            else if ( obj instanceof MingleTypeReference )
            {
                encodeTypeReference( (MingleTypeReference) obj, bufs );
            }
            else if ( obj instanceof MingleIdentifier )
            {
                encodeIdentifier( (MingleIdentifier) obj, bufs );
            }
            else throw state.createFail( "Unhandled tok:", obj );
        }

        private
        ByteBuffer
        encodeManual( Object... toks )
            throws Exception
        {
            List< ByteBuffer > bufs = Lang.newList();

            for ( int i = 0; i < toks.length; ++i ) 
            {
                encodeToken( toks[ i ], bufs );
            }

            return IoTests.toByteBuffer( bufs );
        }

        @Override
        protected
        void
        implAddBasicRoundtripTests( List< BasicRoundtripTest > l )
        {
            l.addAll( Lang.asList(
                
                new BasicRoundtripTest( "test-struct1-inst1" ).
                    setStruct( ModelTestInstances.TEST_STRUCT1_INST1 ).
                    setByteOrder( ByteOrder.BIG_ENDIAN ),
                
                new BasicRoundtripTest( "test-struct1-inst1" ).
                    setStruct( ModelTestInstances.TEST_STRUCT1_INST1 ).
                    setEncodeBufferSize( 8 ).
                    setDecodeBufferSize( 8 ),
                
                new BasicRoundtripTest( "encoder-progress-check-fails" ).
                    setStruct( ModelTestInstances.TEST_STRUCT1_INST1 ).
                    setEncodeBufferSize( 7 ).
                    setDecodeBufferSize( 8 ).
                    expectError(
                        IllegalStateException.class,
                        "Repeated calls to encoder with insufficient " +
                            "buffer capacity 7"
                    ),
 
                new BasicRoundtripTest( "decoder-progress-check-fails" ).
                    setStruct( ModelTestInstances.TEST_STRUCT1_INST1 ).
                    setEncodeBufferSize( 8 ).
                    setDecodeBufferSize( 7 ).
                    expectError(
                        IllegalStateException.class,
                        "Repeated calls to decoder with insufficient " +
                            "buffer capacity 7"
                    )
                
            ));
        }

        @Test
        private
        void
        testDecoderSignalsEndOnTrailingData()
            throws Exception
        {
            ByteBuffer enc = toByteBuffer( STRUCT1_INST1 );
            int expctPos = enc.limit();
            
            ByteBuffer bb = ByteBuffer.allocate( enc.remaining() + 10 );
            bb.put( enc );
            bb.flip();

            MingleDecoder< MingleStruct > dec = structDecoder();
            state.isTrue( dec.readFrom( bb, false ) );
            state.equalInt( expctPos, bb.position() );

            MingleStruct ms = dec.getResult();
            ModelTestInstances.assertEqual( STRUCT1_INST1, ms );
        }

        @Test( expected = MingleCodecException.class,
               expectedPattern = 
                "\\Q[offset 0] Saw type code 0x04 but expected 0x10\\E" )
        private
        void
        testUnexpectedTopLevelTypeCode()
            throws Exception
        {
            fromByteBuffer( ByteBuffer.wrap( new byte[] { TYPE_CODE_ENUM } ) );
        }

        @Test( expected = MingleCodecException.class,
               expectedPattern = 
                "\\Q[offset 28] Unrecognized type code: 0x64\\E" )
        private
        void
        testUnexpectedSymMapValueTypeCode()
            throws Exception
        {
            fromByteBuffer(
                encodeManual(
                    TYPE_CODE_STRUCT, -1, "ns1@v1/Blah",
                    MingleIdentifier.create( "f1" ), TYPE_CODE_FAIL
                )
            );
        }

        @Test( expected = MingleCodecException.class,
               expectedPattern = 
                "\\Q[offset 38] Unrecognized type code: 0x64\\E" )
        private
        void
        testUnexpectedListValTypeCode()
            throws Exception
        {
            fromByteBuffer(
                encodeManual(
                    TYPE_CODE_STRUCT, -1, "ns1@v1/Blah",
                    MingleIdentifier.create( "f1" ),
                    TYPE_CODE_LIST, -1,
                        TYPE_CODE_INT32, 10, // an okay list val
                        TYPE_CODE_FAIL
                )
            );
        }

        @Test( expected = MingleCodecException.class,
               expectedPattern =
                "\\Q[offset 10] Parsing type 'ns:ns2@v1/abc': " +
                "<> [1,11]: Type name segments must start with an " +
                "upper case char, got: a\\E" )
        private
        void
        testTypeReferenceSyntaxException()
            throws Exception
        {
            fromByteBuffer(
                encodeManual(
                    TYPE_CODE_STRUCT, -1, "ns:ns2@v1/abc" ) );
        }

        @Test( expected = MingleCodecException.class,
               expectedPattern = 
                "\\Q[offset 22] Parsing identifier 'bad_9id': " +
                "<> [1,5]: Invalid part beginning: 9\\E" )
        private
        void
        testIdentifierSyntaxException()
            throws Exception
        {
            fromByteBuffer(
                encodeManual( TYPE_CODE_STRUCT, -1, "ns@v1/S", "bad_9id" ) );
        }

        @Test( expected = MingleCodecException.class,
               expectedPattern = 
                "\\Q[offset 34] : Invalid timestamp: 2009-23-22222\\E" )
        private
        void
        testTimestampSyntaxException()
            throws Exception
        {
            fromByteBuffer(
                encodeManual(
                    TYPE_CODE_STRUCT, -1, "ns@v1/S1",
                    "time1", TYPE_CODE_RFC3339_STR, "2009-23-22222"
                )
            );
        }

        @Test
        private
        void
        testIdVariantsAccepted()
            throws Exception
        {
            for ( String id : new String[] { "test-id", "test_id", "testId" } )
            {
                ModelTestInstances.assertEqual(
                    STRUCT2_INST1,
                    fromByteBuffer(
                        encodeManual(
                            TYPE_CODE_STRUCT, -1, STRUCT2_INST1.getType(),
                            id, "val1",
                            TYPE_CODE_END
                        )
                    )
                );
            }
        }
    }

    @Test
    private
    void
    testFactoryLoaded()
        throws Exception
    {
        MingleCodecTests.assertFactoryRoundtrip( "binary", rt );
    }

    // May have other codec configurations that we want to test later, so set
    // tests up now to use TestFactory
    @TestFactory
    private
    static
    List< ? >
    testBinCodec( TestRuntime rt )
    {
        List< Object > res = Lang.newList();

        res.add( 
            Tests.createLabeledTestObject( new TestImpl( rt ), "default" ) );

        return res;
    }
}
