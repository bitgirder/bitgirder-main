package com.bitgirder.mingle.json;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.json.JsonSerializer;
import com.bitgirder.json.JsonParsers;

import com.bitgirder.mingle.codec.MingleCodecDetectorFactory;
import com.bitgirder.mingle.codec.MingleCodecDetector;
import com.bitgirder.mingle.codec.MingleCodecFactory;
import com.bitgirder.mingle.codec.MingleCodecFactoryInitializer;

import com.bitgirder.mingle.model.MingleIdentifier;

import com.bitgirder.parser.SyntaxException;

import com.bitgirder.io.IoUtils;

import java.nio.ByteBuffer;
import java.nio.ByteOrder;

import java.nio.charset.Charset;

public
final
class JsonMingleCodecs
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static JsonMingleCodec JSON_CODEC = JsonMingleCodec.create();

    private JsonMingleCodecs() {}

    public static JsonMingleCodec getJsonCodec() { return JSON_CODEC; }

    // Main expected use case will be to call with something like "\n" to get a
    // codec which serializes objects followed by a line terminator
    public
    static
    JsonMingleCodec
    getJsonCodec( String serialSuffix )
    {
        inputs.notNull( serialSuffix, "serialSuffix" );

        JsonSerializer.Options opts =
            new JsonSerializer.Options.Builder().
                setSerialSuffix( serialSuffix ).
                build();

        return
            new JsonMingleCodec.Builder().
                setSerializerOptions( opts ).
                build();
    }

    // codec detector impl; alg:
    //
    //  - expect at least 4 bytes to allow for json char detection. 
    //
    //  - use the charset to create a ByteBuffer charAcc into which we'll
    //  accumulate one char at a time while we are matching
    //
    //  - skip any whitespace chars until we see a '{', in which case we
    //  register detection TRUE, or any other non-ws char, in which case
    //  we register a detection FALSE
    //
    // Note that we're able to use this simplisitic decode, particularly
    // ignoring any utf-8 chars that are not single byte, since the only ones we
    // could need to interpret are the single-byte ones (' ', '\r', '\n', '\t',
    // and '{').
    private
    final
    static
    class DetectorImpl
    implements MingleCodecDetector
    {
        private final ByteBuffer acc = ByteBuffer.allocate( 4 );
        private ByteBuffer charAcc;

        private
        void
        initCharAcc( int sz,
                     ByteOrder bo )
        {
            charAcc = ByteBuffer.allocate( sz );
            if ( bo != null ) charAcc.order( bo );
        }

        private
        void
        initCharAcc( Charset cs )
        {
            if ( cs.name().equalsIgnoreCase( "utf-8" ) ) initCharAcc( 1, null );
            else if ( cs.name().equalsIgnoreCase( "utf-16le" ) )
            {
                initCharAcc( 2, ByteOrder.LITTLE_ENDIAN );
            }
            else if ( cs.name().equalsIgnoreCase( "utf-16be" ) )
            {
                initCharAcc( 2, ByteOrder.BIG_ENDIAN );
            }
            else if ( cs.name().equalsIgnoreCase( "utf-32le" ) )
            {
                initCharAcc( 4, ByteOrder.LITTLE_ENDIAN );
            }
            else if ( cs.name().equalsIgnoreCase( "utf-32be" ) )
            {
                initCharAcc( 4, ByteOrder.BIG_ENDIAN );
            }
            else state.fail( "Unrecognized charset:", cs.name() );
        }

        private
        Character
        getChar()
        {
            switch ( charAcc.remaining() )
            {
                case 1: return (char) charAcc.get();
                case 2: return charAcc.getChar();

                case 4: 
                    char[] arr = Character.toChars( charAcc.getInt() );
                    return arr.length == 1 ? arr[ 0 ] : null;

                default: throw state.createFail();
            }
        }

        // charAcc will have no remaining upon entry; will be cleared on return
        private
        Boolean
        scanChar()
        {
            charAcc.flip();

            Character ch = getChar();
            charAcc.clear();

            if ( ch == null ) return Boolean.FALSE;
            else if ( ch == '{' ) return Boolean.TRUE;
            else if ( Character.isWhitespace( ch ) ) return null;
            else return Boolean.FALSE;
        }

        private
        Boolean
        scan( ByteBuffer bb )
        {
            Boolean res = null;

            while ( res == null && bb.hasRemaining() )
            {
                IoUtils.copy( bb, charAcc );
    
                if ( ! charAcc.hasRemaining() ) res = scanChar();
            }

            return res;
        }

        // leaves acc ready for scan()
        private
        Boolean
        detectCharset()
        {
            state.isFalse( acc.hasRemaining() );
            acc.flip();

            try 
            { 
                Charset cs = JsonParsers.detectCharset( acc.slice() ); 
                initCharAcc( cs );
                
                return null;
            }
            catch ( SyntaxException se ) { return Boolean.FALSE; }
        }

        public
        Boolean
        update( ByteBuffer bb )
        {
            Boolean res = null;

            if ( acc.hasRemaining() ) 
            {
                IoUtils.copy( bb, acc );

                if ( ! acc.hasRemaining() ) 
                {
                    res = detectCharset();
                    if ( res == null ) res = scan( acc );
                }
            }

            if ( res == null && ( ! acc.hasRemaining() ) ) return scan( bb );
            else return res;
        }
    }

    static
    MingleCodecDetectorFactory
    getCodecDetectorFactory()
    {
        return new MingleCodecDetectorFactory() {
            public MingleCodecDetector createCodecDetector() {
                return new DetectorImpl();
            }
        };
    }

    private
    final
    static
    class FactoryInitializer
    implements MingleCodecFactoryInitializer
    {
        public
        void
        initialize( MingleCodecFactory.Builder b )
        {
            b.addCodec(
                MingleIdentifier.create( "json" ),
                getJsonCodec(),
                getCodecDetectorFactory()
            );

            b.addCodec(
                MingleIdentifier.create( "json-newline" ),
                getJsonCodec( "\n" )
            );
        }
    }
}
