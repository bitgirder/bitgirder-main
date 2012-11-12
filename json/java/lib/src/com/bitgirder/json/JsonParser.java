package com.bitgirder.json;

import static com.bitgirder.json.JsonGrammars.SxNt;
import static com.bitgirder.json.JsonGrammars.LxNt;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.io.IoUtils;

import com.bitgirder.parser.SyntaxException;
import com.bitgirder.parser.SyntaxBuilder;
import com.bitgirder.parser.DocumentParserFactory;
import com.bitgirder.parser.DocumentParser;
import com.bitgirder.parser.DerivationMatch;

import java.nio.ByteBuffer;

import java.nio.charset.CharacterCodingException;
import java.nio.charset.Charset;

public
final
class JsonParser< D >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final DocumentParser.Builder< SxNt, LxNt, JsonToken, D > dpb;

    // dp will initially be null unless a charset was explicitly provided. If
    // one is not, we first accumulate up to 4 bytes in csDetectBuf during
    // initial calls to update to detect the charset. Once detected, dp is built
    // and the parse continues normally
    private DocumentParser< D > dp;
    private ByteBuffer csDetectBuf;

    private 
    JsonParser( Builder< D > b )
    {
        dpb = b.jpf.getDocumentParserFactory().createParserBuilder();
        
        dpb.setFileName( inputs.notNull( b.fileName, "fileName" ) );

        dpb.setSyntaxBuilder( 
            inputs.notNull( b.syntaxBuilder, "syntaxBuilder" ) );

        dpb.setFeedFilter( JsonGrammars.RFC4627_FEED_FILTER );
        
        if ( b.charset != null ) dp = dpb.setCharset( b.charset ).build();
    }

    private
    void
    readParserInit( ByteBuffer bb,
                    boolean endOfInput )
        throws CharacterCodingException,
               SyntaxException
    {
        if ( csDetectBuf == null ) csDetectBuf = ByteBuffer.allocate( 4 );

        IoUtils.copy( bb, csDetectBuf );

        if ( endOfInput || ! csDetectBuf.hasRemaining() )
        {
            csDetectBuf.flip();
            Charset cs = JsonParsers.detectCharset( csDetectBuf );
            dpb.setCharset( cs );

            dp = dpb.build();
            dp.update( csDetectBuf, endOfInput && ! bb.hasRemaining() );
        }
    }

    public
    void
    update( ByteBuffer bb,
            boolean endOfInput )
        throws CharacterCodingException,
               SyntaxException
    {
        if ( dp == null ) readParserInit( bb, endOfInput );

        if ( bb.hasRemaining() && dp != null ) dp.update( bb, endOfInput );
    }

    public 
    D 
    buildResult() 
    { 
        state.isFalse( dp == null, "Parse is not complete" );
        return dp.buildSyntax(); 
    }

    static
    final
    class Builder< D >
    {
        private final JsonParserFactory jpf;

        private CharSequence fileName;
        private Charset charset;
        private SyntaxBuilder< SxNt, JsonToken, D > syntaxBuilder;

        Builder( JsonParserFactory jpf )
        {
            this.jpf = state.notNull( jpf, "jpf" );
        }

        public
        Builder< D >
        setFileName( CharSequence fileName )
        {
            this.fileName = inputs.notNull( fileName, "fileName" );
            return this;
        }

        public
        Builder< D >
        setCharset( Charset charset )
        {
            this.charset = inputs.notNull( charset, "charset" );
            return this;
        }

        public
        Builder< D >
        setSyntaxBuilder( SyntaxBuilder< SxNt, JsonToken, D > syntaxBuilder )
        {
            this.syntaxBuilder = 
                inputs.notNull( syntaxBuilder, "syntaxBuilder" );

            return this;
        }

        public JsonParser< D > build() { return new JsonParser< D >( this ); }
    }
}
