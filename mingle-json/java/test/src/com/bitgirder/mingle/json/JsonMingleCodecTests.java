package com.bitgirder.mingle.json;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;
import com.bitgirder.lang.ObjectReceiver;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.mingle.codec.AbstractMingleCodecTests;
import com.bitgirder.mingle.codec.MingleCodec;
import com.bitgirder.mingle.codec.MingleCodecFactory;
import com.bitgirder.mingle.codec.MingleCodecException;
import com.bitgirder.mingle.codec.MingleCodecs;
import com.bitgirder.mingle.codec.MingleCodecTests;
import com.bitgirder.mingle.codec.AbstractCodecDetectionTest;
import com.bitgirder.mingle.codec.NoSuchMingleCodecException;

import com.bitgirder.mingle.model.ModelTestInstances;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleIdentifierFormat;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleList;
import com.bitgirder.mingle.model.MingleStruct;
import com.bitgirder.mingle.model.MingleStructBuilder;

import com.bitgirder.mingle.parser.MingleParsers;

import com.bitgirder.io.Charsets;

import com.bitgirder.json.JsonObject;
import com.bitgirder.json.JsonSerialization;
import com.bitgirder.json.JsonParserFactory;

import com.bitgirder.test.Test;
import com.bitgirder.test.TestRuntime;
import com.bitgirder.test.LabeledTestCall;
import com.bitgirder.test.TestFailureExpector;
import com.bitgirder.test.InvocationFactory;

import java.nio.ByteBuffer;

import java.util.List;

// In addition to the standard json codec tests from the superclass, this class
// also has some extra tests of just the json piece that are independent of any
// particular codec
@Test
final
class JsonMingleCodecTests
extends AbstractMingleCodecTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static JsonMingleCodec codec = JsonMingleCodec.create();

    private final static JsonParserFactory JPF = JsonParserFactory.create();

    private JsonMingleCodecTests( TestRuntime rt ) { super( rt ); }

    @Override
    protected
    void
    debugEncoded( ByteBuffer ser )
        throws Exception
    {
//        code( "Encoded:", Charsets.UTF_8.asString( ser ) );
    }

    protected MingleCodec getCodec() { return codec; }

    private
    void
    assertFormattedId( MingleStruct ms,
                       MingleIdentifierFormat fmt,
                       String expctStr )
        throws Exception
    {
        JsonMingleCodec codec =
            new JsonMingleCodec.Builder().
                setIdentifierFormat( fmt ).
                build();
 
        JsonObject jsonObj = codec.asCodecObject( ms );
        ByteBuffer bb = JsonSerialization.toByteBuffer( jsonObj );
        String jsonStr = Charsets.UTF_8.asString( bb ).toString();

        state.isTrue( jsonStr.indexOf( expctStr ) >= 0 );
    }

    // Coverage that json serialization formats ids according to the given style
    @Test
    private
    void
    testIdentifierFormats()
        throws Exception
    {
        MingleStructBuilder b = MingleModels.structBuilder();

        b.setType( "mingle:json@v1/TestType" );
        b.fields().setString( "the-great-identifier", "howdy" );

        MingleStruct ms = b.build();

        assertFormattedId( 
            ms, MingleIdentifierFormat.LC_UNDERSCORE, "the_great_identifier" );
        
        assertFormattedId(
            ms, MingleIdentifierFormat.LC_HYPHENATED, "the-great-identifier" );

        assertFormattedId(
            ms, MingleIdentifierFormat.LC_CAMEL_CAPPED, "theGreatIdentifier" );
    }

    @Test
    private
    void
    testCodecOmitsTypeFields()
        throws Exception
    {
        MingleStructBuilder b = MingleModels.structBuilder();
        b.setType( "test@v1/Type1" );
        
        // build a struct with some nested structs and structs in lists
        MingleStruct ms = ModelTestInstances.TEST_STRUCT1_INST1;
        b.fields().set( "field1", ms );
        b.fields().set( "field2", MingleList.create( ms, ms ) );

        JsonMingleCodec codec =
            new JsonMingleCodec.Builder().
                setOmitTypeFields().
                build();
        
        ByteBuffer buf = MingleCodecs.toByteBuffer( codec, b.build() );
        CharSequence str = Charsets.UTF_8.asString( buf );

        state.isFalse( str.toString().indexOf( "$type" ) >= 0 );
    }

    @Test
    private
    void
    testCodecExplicitEnumExpansion()
        throws Exception
    {
        JsonMingleCodec codec =
            new JsonMingleCodec.Builder().
                setExpandEnums().
                build();
        
        MingleStruct expct = ModelTestInstances.TEST_STRUCT1_INST1;

        ByteBuffer buf = MingleCodecs.toByteBuffer( codec, expct );
        String str = Charsets.UTF_8.asString( buf ).toString();

        state.isTrue(
            str.indexOf( 
                "\"enum1\":{\"$type\":\"mingle:test@v1/TestEnum1\"," +
                "\"$constant\":\"constant1\"}" ) > 0
        );
        
        MingleStruct actual = 
            MingleCodecs.fromByteBuffer( codec, buf, MingleStruct.class );

        ModelTestInstances.assertEqual( expct, actual );
    }

    @Test( expected = IllegalStateException.class,
           expectedPattern = 
            "Illegal combination of expandEnums and omitTypeFields" )
    private
    void
    testExpandEnumOmitFieldsComboFails()
    {
        new JsonMingleCodec.Builder().
            setOmitTypeFields().
            setExpandEnums().
            build();
    }

    @Test( expected = MingleCodecException.class,
           expectedPattern = "\\Q$type: Missing value\\E" )
    private
    void
    testEmptyDocFailsNoType()
        throws Exception
    {
        MingleCodecs.fromByteBuffer(
            JsonMingleCodecs.getJsonCodec(),
            Charsets.UTF_8.asByteBuffer( "{}" ),
            MingleStruct.class
        );
    }

    private
    final
    class CodecDetectionTest
    extends AbstractCodecDetectionTest< CodecDetectionTest >
    {
        private 
        CodecDetectionTest( CharSequence lbl ) 
        { 
            super( lbl ); 

            expectCodec( JsonMingleCodec.class );
        }

        protected
        MingleCodecFactory
        codecFactory()
        {
            return
                new MingleCodecFactory.Builder().
                    addCodec( 
                        MingleIdentifier.create( "json" ),
                        codec,
                        JsonMingleCodecs.getCodecDetectorFactory()
                    ).
                    build();
        }
    }

    private
    void
    addCodecDetectionStringTests( List< CodecDetectionTest > l )
        throws Exception
    {
        String[] csArr = 
            new String[] { 
                "utf-8", "utf-16le", "utf-16be", "utf-32le", "utf-32be"
            };

        for ( String cs : csArr )
        {
            for ( String s : new String[] { "{   ", " \t\r\n  {" } )
            for ( int xferSz : new int[] { 1, 3, 20 } )
            {
                CharSequence lbl =
                    Strings.crossJoin( "=", ",",
                        "cs", cs, "str", s, "xferSz", xferSz );

                l.add(
                    new CodecDetectionTest( lbl ).
                        setCopyBufferSize( xferSz ).
                        setInput( s.getBytes( cs ) )
                );
            }
        }
    }

    @InvocationFactory
    private
    List< CodecDetectionTest >
    testCodecDetection()
        throws Exception
    {
        List< CodecDetectionTest > res = Lang.newList();

        addCodecDetectionStringTests( res );
        
        res.add(
            new CodecDetectionTest( "non-json-charset" ).
                setInput( (byte) 0, (byte) 0, (byte) 'x', (byte) 'y' ).
                expectFailure( NoSuchMingleCodecException.class, "" )
        );

//        res.add(
//            new CodecDetectionTest( "insufficient-input" ).
//                setInput( (byte) '{', (byte) '}' )
//        );

        res.add(
            new CodecDetectionTest( "binary-data" ).
                setInput( (byte) 1, (byte) 2, (byte) 3, (byte) 4 ).
                expectFailure( NoSuchMingleCodecException.class, "" )
        );

        return res;
    }

    private
    void
    assertFactoryRoundtrip( CharSequence id,
                            final char lastCharExpct )
        throws Exception
    {
        MingleCodecTests.assertFactoryRoundtrip(
            id, testRuntime(), new ObjectReceiver< ByteBuffer >() {
                public void receive( ByteBuffer bb ) 
                {
                    state.equal( 
                        (byte) lastCharExpct, bb.get( bb.limit() - 1 ) );
                }
            });
    }

    @Test
    private
    void
    testFactoryDefaultCodec()
        throws Exception
    {
        assertFactoryRoundtrip( "json", '}' );
    }

    @Test
    private
    void
    testFactoryNewlineCodec()
        throws Exception
    {
        assertFactoryRoundtrip( "json-newline", '\n' );
    }
}
