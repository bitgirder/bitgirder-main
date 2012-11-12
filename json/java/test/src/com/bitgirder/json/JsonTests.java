package com.bitgirder.json;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.lang.reflect.ReflectUtils;

import com.bitgirder.test.Test;

import com.bitgirder.io.Charsets;
import com.bitgirder.io.CharsetHelper;
import com.bitgirder.io.IoUtils;

import com.bitgirder.parser.SyntaxException;

import java.util.Arrays;
import java.util.List;
import java.util.Map;

import java.nio.ByteBuffer;

@Test
final
class JsonTests
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static String TEST0_FILE = "test0.json";

    private final static JsonParserFactory jpf = JsonParserFactory.create();

    private
    ByteBuffer
    getText( String fileName )
        throws Exception
    {
        return
            IoUtils.toByteBuffer(
                ReflectUtils.getResourceAsStream( getClass(), fileName ) );
    }

    private
    JsonText
    parseTestFile( String fileName )
        throws Exception
    {
        ByteBuffer text = getText( fileName );

        JsonModelBuilder< JsonText > bld = 
            new JsonModelBuilder< JsonText >( JsonText.class );

        JsonParser< JsonText > jp =
            jpf.< JsonText >createParserBuilder().
                setFileName( fileName ).
                setCharset( Charsets.UTF_8.charset() ).
                setSyntaxBuilder( bld ).
                build();

        jp.update( text, true );
        return jp.buildResult();
    }

    private
    void
    assertString( JsonString str,
                  Object expct )
    {
        state.equalString( (CharSequence) expct, str );
    }

    private
    void
    assertNumber( JsonNumber num,
                  Object expct )
    {
        if ( expct instanceof Double || expct instanceof Float )
        {
            state.equal( 
                Double.valueOf( ( (Number) expct ).doubleValue() ), 
                Double.valueOf( num.doubleValue() ) );
        }
        else if ( expct instanceof Byte ||
                  expct instanceof Short ||
                  expct instanceof Integer ||
                  expct instanceof Long )
        {
            state.equal( 
                Long.valueOf( ( (Number) expct ).longValue() ),
                Long.valueOf( num.longValue() ) );
        }
        else 
        {
            throw state.createFail( 
                "Unexpected number type:", expct.getClass() );
        }
    }

    private
    void
    assertBoolean( JsonBoolean bool,
                   Object expct )
    {
        state.equal( Boolean.valueOf( bool.booleanValue() ), (Boolean) expct );
    }

    private
    void
    assertNull( JsonNull jn,
                Object expct )
    {
        if ( expct != null ) state.cast( JsonNull.class, expct );
    }

    private
    void
    assertArray( JsonArray arr,
                 List< ? > expct )
    {
        state.equalInt( expct.size(), arr.size() );

        for ( int i = 0, e = arr.size(); i < e; ++i )
        {
            assertValue( arr.get( i ), expct.get( i ) );
        }
    }

    private
    void
    assertArray( JsonArray arr,
                 Object... expct )
    {
        assertArray( arr, Arrays.asList( expct ) );
    }

    private
    void
    assertArray( JsonArray arr,
                 Object expct )
    {
        if ( expct instanceof List< ? > )
        {
            assertArray( arr, (List< ? >) expct );
        }
        else if ( expct instanceof Object[] )
        {
            assertArray( arr, (Object[]) expct );
        }
        else
        {
            throw state.createFail(
                "Unexpected array representation:", expct );
        }
    }

    private
    void
    assertObject( JsonObject obj,
                  Map< ?, ? > expct )
    {
        state.equalInt( expct.size(), obj.size() );

        for ( Map.Entry< ?, ? > e : expct.entrySet() )
        {
            JsonString key = JsonString.create( ( (CharSequence) e.getKey() ) );

            JsonValue valActual = obj.getValue( key );
            assertValue( valActual, e.getValue() );
        }
    }

    private
    void
    assertObject( JsonObject obj,
                  Object... flatMap )
    {
        Map< Object, Object > map = 
            Lang.putAll( 
                Lang.< Object, Object >newMap(), Object.class, Object.class,
                flatMap );
 
        assertObject( obj, map );
    }

    private
    void
    assertObject( JsonObject obj,
                  Object expct )
    {
        if ( expct instanceof Object[] ) assertObject( obj, (Object[]) expct );
        else if ( expct instanceof Map< ?, ? > )
        {
            assertObject( obj, (Map< ?, ? >) expct );
        }
        else 
        {
            throw state.createFail( 
                "Unexpected object representation:", expct );
        }
    }

    private
    void
    assertValue( JsonValue val,
                 Object expct )
    {
        if ( val instanceof JsonString ) 
        {
            assertString( (JsonString) val, expct );
        }
        else if ( val instanceof JsonNumber )
        {
            assertNumber( (JsonNumber) val, expct );
        }
        else if ( val instanceof JsonNull ) assertNull( (JsonNull) val, expct );
        else if ( val instanceof JsonBoolean )
        {
            assertBoolean( (JsonBoolean) val, expct );
        }
        else if ( val instanceof JsonArray )
        {
            assertArray( (JsonArray) val, expct );
        }
        else if ( val instanceof JsonObject )
        {
            assertObject( (JsonObject) val, expct );
        }
        else throw state.createFail( "Unexpected json value:", val );
    }

    // Basic code coverage for the various array accessors
    @Test
    private
    void
    testArrayAccessors()
    {
        JsonArray arr =
            new JsonArray.Builder().
                add( JsonString.create( "hello" ) ).
                add( JsonNull.INSTANCE ).
                add( JsonBoolean.TRUE ).
                add( JsonNumber.forNumber( Integer.valueOf( 42 ) ) ).
                add( new JsonArray.Builder().build() ).
                add( new JsonObject.Builder().build() ).
                build();
        
        assertValue( arr.getString( 0 ), "hello" );
        assertValue( arr.getNull( 1 ), null );
        state.isTrue( arr.getBool( 2 ) );
        assertValue( arr.getNumber( 3 ), 42 );
        assertValue( arr.getArray( 4 ), Lang.emptyList() );
        assertValue( arr.getObject( 5 ), Lang.emptyMap() );
    }

    // We use the specific typed-accessors for top-level members, even though we
    // could just use jo.getValue() equally legally, in order to get some
    // coverage of the type-specific getters in JsonObject.
    private
    void
    assertTest0Object( JsonObject jo )
    {
        state.isTrue( jo.hasMember( "member-1" ) );
        state.isFalse( jo.hasMember( "nothing-to-see-here" ) );

        assertValue( jo.getNumber( "member-1" ), 1 );
        
        state.isTrue( jo.getBool( "member-2" ) );
        state.isFalse( jo.getBool( "member-3" ) );

        assertValue( jo.getNull( "member-4" ), null );
        assertValue( jo.getString( "member-5" ), "some string" );

        assertValue( 
            jo.getArray( "member-6" ), 
            Arrays.asList( 
                0, 1, 2, -3, -4.0, 0.50, 6.0123e-4, 7.0e4, -8e-9, -9.0e0 ) );
 
        assertValue(
            jo.getObject( "member-7" ),
            Lang.newMap(
                Object.class, Object.class,
                "member-1", 2,
                "member-2", 
                    Arrays.asList( 
                        "a", 123, "mixed list", true, false, null,
                        Arrays.asList( 1, 2, 3 ),
                        Lang.newMap(
                            Object.class, Object.class, "d", -4, "e", null ) ),
                "member-3",
                    Lang.newMap(
                        Object.class, Object.class, 
                            "a", true, 
                            "b", false,
                            "c", Arrays.asList( 7, 8, null, true ),
                            "d", Lang.newMap( Object.class, Object.class,
                                    "x", false, "y", "hello" ) ) ) );

        assertValue( jo.getObject( "member-8" ), Lang.newMap() );
        assertValue( jo.getArray( "member-9" ), Lang.emptyList() );

        assertValue(
            jo.getString( "member-10" ),
            "a\nstring\twith\fvarious\rescapes\b\\:\" (also /)" );

        assertValue(
            jo.getString( "member-11" ),
            "a string with an escaped gclef: \ud834\uDD1e" );
        
        assertValue(
            jo.getString( "member-12" ),
            "a string with an escaped ctl character (codepoint 3): \u0003" );
    }

    @Test
    private
    void
    test0()
        throws Exception
    {
        JsonObject jo = (JsonObject) parseTestFile( TEST0_FILE );
        assertTest0Object( jo );
    }

    @Test
    private
    void
    test0SmallMultiCallCharsetDetect()
        throws Exception
    {
        ByteBuffer bb = getText( TEST0_FILE );

        for ( int i = 1; i < 6; ++i )
        {
            JsonParser< JsonObject > p = jpf.createObjectParser( TEST0_FILE );

            ByteBuffer feed = bb.slice();
            feed.limit( i );

            p.update( feed, false );
            feed.limit( bb.limit() );

            p.update( feed, true );
            assertTest0Object( p.buildResult() );
        }
    }

    @Test
    private
    void
    testRoundtripTest0()
        throws Exception
    {
        JsonObject jo1 = (JsonObject) parseTestFile( TEST0_FILE );
        
        ByteBuffer bb = JsonSerialization.toByteBuffer( jo1 );

        JsonObject jo2 = (JsonObject) jpf.parseJsonText( bb );
        assertTest0Object( jo2 );
    }

    private
    void
    assertCharsetDetection( CharsetHelper hlp )
        throws Exception
    {
        String jsonStr = "{ \"a-key\": 123 }";
        ByteBuffer bb = hlp.asByteBuffer( jsonStr );

        JsonObject jsonObj = (JsonObject) jpf.parseJsonText( bb );
        state.equalInt( 123, jsonObj.getNumber( "a-key" ).intValue() );
    }

    @Test
    private
    void
    testCharsetDetectionUtf8()
        throws Exception
    {
        assertCharsetDetection( Charsets.UTF_8 );
    }

    @Test
    private
    void
    testCharsetDetectionUtf8EmptyObject()
        throws Exception
    {
        ByteBuffer bb = Charsets.UTF_8.asByteBuffer( "{}" );
        jpf.parseJsonText( bb );
    }

    @Test
    private
    void
    testCharsetDetectionUtf16Be()
        throws Exception
    {
        assertCharsetDetection( Charsets.UTF_16BE );
    }

    @Test
    private
    void
    testCharsetDetectionUtf16Le()
        throws Exception
    {
        assertCharsetDetection( Charsets.UTF_16LE );
    }

    @Test
    private
    void
    testCharsetDetectionUtf32Be()
        throws Exception
    {
        assertCharsetDetection( Charsets.UTF_32BE );
    }

    @Test
    private
    void
    testCharsetDetectionUtf32Le()
        throws Exception
    {
        assertCharsetDetection( Charsets.UTF_32LE );
    }

    @Test( expected = SyntaxException.class,
           expectedPattern = "Insufficient or malformed data in buffer" )
    private
    void
    testCharsetDetectionInsufficientData()
        throws Exception
    {
        jpf.parseJsonText( ByteBuffer.allocate( 3 ) );
    }

    @Test( expected = SyntaxException.class,
           expectedPattern =
            "Can't detect charset from first four octets of json text" )
    private
    void
    testCharsetDetectionUnidentifiedPattern()
        throws Exception
    {
        byte[] data = new byte[] { 0, 40, 0, 0, 40, 40, 40 };
        jpf.parseJsonText( data );
    }

    // Regression to make sure that we fail on a single-char invalid doc without
    // a declared charset with a SyntaxException (previously was an
    // IllegalArgumentException)
    @Test( expected = SyntaxException.class,
           expectedPattern = "Insufficient or malformed data in buffer" )
    private
    void
    testCharsetDetectionWithSingleCharInvalidDocument()
        throws Exception
    {
        jpf.parseJsonText( Charsets.UTF_8.asByteBuffer( "{" ) );
    }

    // This should pass the charset detection okay but still fail to parse as
    // normal
    @Test( expected = SyntaxException.class,
           expectedPattern = "Unmatched document" )
    private
    void
    testInvalidParseOfInvalidTwoCharDocument()
        throws Exception
    {
        jpf.parseJsonText( Charsets.UTF_8.asByteBuffer( "{ " ) );
    }

    private
    void
    assertSerializationFormatting( JsonSerializer.Options opts,
                                   String expct )
        throws Exception
    {
        JsonText text = 
            jpf.parseJsonText( Charsets.UTF_8.asByteBuffer( expct ) );

        JsonSerializer ser = JsonSerializer.create( text, opts );
        ByteBuffer json = JsonSerialization.toByteBuffer( ser );
        String actual = Charsets.UTF_8.asString( json ).toString();

        state.equalString( expct, actual );
    }

    @Test
    private
    void
    testSerializeWithTrailingNewline()
        throws Exception
    {
        JsonSerializer.Options opts =
            new JsonSerializer.Options.Builder().
                setCharset( Charsets.UTF_8.charset() ).
                setSerialSuffix( "\n" ).
                build();

        assertSerializationFormatting( opts, "{\"a-key\":\"a-value\"}\n" );
    }

    @Test
    private
    void
    testJsonAndJavaInterChange()
    {
        Map< String, Object > expct =
            Lang.newMap( String.class, Object.class,
                "a-null", null,
                "an-int", 12,
                "a-float", 12.2,
                "a-string", "hi",
                "a-bool", true,
                "an-array", 
                    Arrays.asList( 
                        1, null, true, "bye", Arrays.asList( 1, "hi" ),
                        Lang.newMap( String.class, Object.class, "key1", true )
                    ),
                "an-object",
                    Lang.newMap( String.class, Object.class,
                        "key1", "val1",
                        "key2", 12
                    )
            );
        
        JsonObject json = JsonValues.asJsonObject( expct );

        Map< String, Object > actual = JsonValues.asJavaMap( json );

        state.isTrue( actual.get( "a-null" ) == null );
        state.equalInt( 12, ( (Number) actual.get( "an-int" ) ).intValue() );

        state.isTrue( 
            12.2d == ( (Number) actual.get( "a-float" ) ).doubleValue() );
 
        state.equalString( "hi", (CharSequence) actual.get( "a-string" ) );
        state.isTrue( ( (Boolean) actual.get( "a-bool" ) ).booleanValue() );
        
        List< ? > anArray = 
            state.notNull( (List< ? >) actual.get( "an-array" ) );
        
        state.equalInt( 1, ( (Number) anArray.get( 0 ) ).intValue() );
        state.isTrue( anArray.get( 1 ) == null );
        state.isTrue( (Boolean) anArray.get( 2 ) );
        state.equalString( "bye", (CharSequence) anArray.get( 3 ) );

        // just use toString now to check nested structural types, since the
        // toString builtin to List and Map will make it clear whether they have
        // been transformed correctly
        state.equalString( 
            "1|hi", Strings.join( "|", (List< ? >) anArray.get( 4 ) ) );
        state.equalString( "{key1=true}", anArray.get( 5 ).toString() );
 
        Map< ?, ? > anObject = (Map< ?, ? >) actual.get( "an-object" );
        state.equalString( "val1", (CharSequence) anObject.get( "key1" ) );
        state.equalInt( 12, ( (Number) anObject.get( "key2" ) ).intValue() );
    }

    // Put in place to detect and fix a bug in JsonSerializer.writeTo which led
    // to an infinite loop. It turned out that if the call to writeTo() led to
    // CoderResult.OVERFLOW the loop in there was not keying off of this
    // condition and was continuing to loop, repeatedly overflowing, making no
    // progress, overflowing, etc. The fix involves both tracking whether
    // progress was made, so that writeTo() will now fail fast if an infinite
    // loop is detected, but more importantly it involves also detecting
    // overflow and exiting the loop normally in that case so that the caller
    // can drain and come back for more data.
    //
    // The object below is such that the original error condition was triggered
    // (infinite loop) but is now fixed. 
    @Test
    private
    void
    testJsonSerializerWriteToInfiniteLoopRegression()
        throws Exception
    {
        String str = "\ud834\uDD1e";

        JsonObject obj =
            new JsonObject.Builder().addMember( "key", str ).build();
 
        ByteBuffer buf = ByteBuffer.allocate( 50 );
        buf.limit( 9 );

        JsonSerializer ser = 
            JsonSerializer.create( obj,
                new JsonSerializer.Options.Builder().
                    setCharset( Charsets.UTF_8.charset() ).
                    build()
            );
 
        state.isFalse( ser.writeTo( buf ) );
        buf.limit( buf.capacity() );
        state.isTrue( ser.writeTo( buf ) );
        buf.flip();

        JsonObject obj2 = JsonParserFactory.create().parseJsonObject( buf );
        state.equalString( obj2.getString( "key" ), str );
    }

    private
    JsonObject
    feedMultiByteObj( ByteBuffer buf,
                      int chunkSize )
        throws Exception
    {
        JsonParser< JsonObject > p =
            JsonParserFactory.create().createObjectParser( "<>" );

        for ( int e = buf.limit(); buf.hasRemaining(); )
        {
            int lim = Math.min( buf.remaining(), chunkSize );
            buf.limit( buf.position() + lim );
            boolean isEnd = buf.limit() == e;

            p.update( buf, isEnd );
            buf.limit( e );
        }
        
        return p.buildResult();
    }

    // This came up in production first and is put in here as a regression test
    // and generally something that needs to be tested. In the case when it
    // first appeared, it turned out that the bug is actually in
    // com.bitgirder.parser.Lexer, and it is separately regressed and tested
    // there. But, since json parser will at some point move away from using
    // Lexer, and since this test is just a good one in general, it stays here
    // independently of its implicit coverage in the parser package.
    @Test
    private
    void
    testMultiByteCharSplitFeed()
        throws Exception
    {
        String expct = "\u271f";

        ByteBuffer buf =
            Charsets.UTF_8.asByteBuffer( "{ \"key\": \"" + expct + "\" }" );
 
        for ( int i = 1; i < buf.remaining(); ++i )
        {
            JsonObject obj = feedMultiByteObj( buf.slice(), i );

            state.equalString( expct, obj.getString( "key" ) );
        }
    }

    // Simple regression test for situations in which an errant attempt is made
    // to call buildResult() even before the internal document parser has been
    // built. Previously led to a NPE, now gives a more meaningful exception
    @Test( expected = IllegalStateException.class,
           expectedPattern = "Parse is not complete" )
    private
    void
    testJsonParserFailOnResultAccessBeforeDocParserBuilt()
    {
        JsonParserFactory.create().createObjectParser( "<>" ).buildResult();
    }

    // Regression test put in upon discovering a bug in serializer which was
    // causing serializer to return false when writing the last serialized data
    // when the end of data completely fills the writeTo destination buffer.
    @Test
    private
    void
    testSerializerPreciseOutBufAlignment()
        throws Exception
    {
        JsonObject jsonObj = new JsonObject.Builder().build();
        ByteBuffer bb = ByteBuffer.allocate( 2 );

        state.isTrue( JsonSerializer.create( jsonObj ).writeTo( bb ) );
    }

    @Test( expected = IllegalStateException.class,
           expectedPattern = "^Loop did not make progress$" )
    private
    final
    void testInsufficientSerializerBufCapacityFailure()
        throws Exception
    {
        JsonObject jsonObj =
            new JsonObject.Builder().
                addMember( "key", "\ud834\uDD1e" ).
                build();
        
        ByteBuffer bb = ByteBuffer.allocate( 1 );

        JsonSerializer ser = JsonSerializer.create( jsonObj );

        while ( ! ser.writeTo( bb ) ) bb.clear();
    }
}
