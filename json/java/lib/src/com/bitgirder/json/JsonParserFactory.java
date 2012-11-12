package com.bitgirder.json;

import static com.bitgirder.json.JsonGrammars.SxNt;
import static com.bitgirder.json.JsonGrammars.LxNt;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.Charsets;

import com.bitgirder.parser.DocumentParserFactory;
import com.bitgirder.parser.DocumentParser;
import com.bitgirder.parser.SyntaxException;

import java.nio.ByteBuffer;

import java.nio.charset.Charset;
import java.nio.charset.CharacterCodingException;

public
final
class JsonParserFactory
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final static String ANONYMOUS_FILE_NAME = "<>";

    private DocumentParserFactory< SxNt, LxNt, JsonToken > dpf;

    private 
    JsonParserFactory( DocumentParserFactory< SxNt, LxNt, JsonToken > dpf )
    {
        this.dpf = dpf;
    }

    DocumentParserFactory< SxNt, LxNt, JsonToken >
    getDocumentParserFactory()
    {
        return dpf;
    }

    public
    static
    JsonParserFactory
    create()
    {
        return new JsonParserFactory( 
            JsonGrammars.RFC4627_DOCUMENT_PARSER_FACTORY );
    }

    < T >
    JsonParser.Builder< T >
    createParserBuilder()
    {
        return new JsonParser.Builder< T >( this );
    }

    // null checks fileName on behalf of public frontends; creates a parser that
    // will return (via cast) the expected type and which will detect the
    // charset in the document it is fed
    private
    < T extends JsonText >
    JsonParser< T >
    createParser( String fileName,
                  Class< T > cls )
    {
        return
            this.< T >createParserBuilder().
                setFileName( inputs.notNull( fileName, "fileName" ) ).
                setSyntaxBuilder( new JsonModelBuilder< T >( cls ) ).
                build();
    }

    public
    JsonParser< JsonText >
    createTextParser( String fileName )
    {
        return createParser( fileName, JsonText.class );
    }

    public
    JsonParser< JsonObject >
    createObjectParser( String fileName )
    {
        return createParser( fileName, JsonObject.class );
    }

    public
    JsonParser< JsonArray >
    createArrayParser( String fileName )
    {
        return createParser( fileName, JsonArray.class );
    }

    public
    JsonText
    parseJsonText( ByteBuffer data,
                   Charset charset )
        throws SyntaxException,
               CharacterCodingException
    {
        inputs.notNull( data, "data" );
        inputs.notNull( charset, "charset" );

        JsonModelBuilder< JsonText > bld =
            new JsonModelBuilder< JsonText >( JsonText.class );

        JsonParser< JsonText > p = 
            this.< JsonText >createParserBuilder().
                setFileName( ANONYMOUS_FILE_NAME ).
                setCharset( charset ).
                setSyntaxBuilder( bld ).
                build();
        
        p.update( data, true );

        return p.buildResult();
    }

    public
    JsonText
    parseJsonText( byte[] data,
                   Charset charset )
        throws SyntaxException,
               CharacterCodingException
    {
        inputs.notNull( data, "data" );

        return parseJsonText( ByteBuffer.wrap( data ), charset );
    }

    public
    JsonText
    parseJsonText( ByteBuffer data )
        throws SyntaxException,
               CharacterCodingException
    {
        inputs.notNull( data, "data" );
        return parseJsonText( data, JsonParsers.detectCharset( data ) );
    }

    public
    JsonText
    parseJsonText( byte[] data )
        throws SyntaxException,
               CharacterCodingException
    {
        inputs.notNull( data, "data" );
        return parseJsonText( ByteBuffer.wrap( data ) );
    }

    public
    JsonObject
    parseJsonObject( ByteBuffer data )
        throws SyntaxException,
               CharacterCodingException
    {
        return (JsonObject) parseJsonText( data );
    }
}
