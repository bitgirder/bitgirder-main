package com.bitgirder.mingle.http.json;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

import com.bitgirder.json.JsonSerializer;

import com.bitgirder.mingle.json.JsonMingleCodec;

import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleIdentifierFormat;

import com.bitgirder.mingle.codec.MingleCodec;
import com.bitgirder.mingle.codec.MingleCodecException;
import com.bitgirder.mingle.codec.NoSuchMingleCodecException;

import com.bitgirder.mingle.http.MingleHttpCodecFactory;
import com.bitgirder.mingle.http.MingleHttpCodecContext;

import com.bitgirder.parser.SyntaxException;

import com.bitgirder.http.HttpRequestMessage;
import com.bitgirder.http.HttpHeaderName;

import java.util.Map;

public
final
class JsonMingleHttpCodecFactory
implements MingleHttpCodecFactory
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static HttpHeaderName NAME_ID_STYLE =
        HttpHeaderName.forString( "x-service-id-style" );

    private final static MingleIdentifier DEFAULT_ID_STYLE;
    private final static JsonMingleHttpCodecFactory INSTANCE;

    private final Map< MingleIdentifier, MingleHttpCodecContext > contexts;

    private
    JsonMingleHttpCodecFactory( 
        Map< MingleIdentifier, MingleHttpCodecContext > contexts )
    {
        this.contexts = contexts;
    }

    private
    MingleIdentifier
    parseStyleId( CharSequence idStyleStr )
        throws MingleCodecException
    {
        try { return MingleIdentifier.parse( idStyleStr ); }
        catch ( SyntaxException se )
        {
            throw new MingleCodecException( "Invalid id style: " + idStyleStr );
        }
    }

    private
    MingleHttpCodecContext
    codecContextFor( CharSequence idStyleStr )
        throws MingleCodecException
    {
        return contexts.get( parseStyleId( idStyleStr ) );
    }

    public
    MingleHttpCodecContext
    getDefaultCodecContext()
    {
        return state.get( contexts, DEFAULT_ID_STYLE, "contexts" );
    }

    public
    MingleHttpCodecContext
    codecContextFor( HttpRequestMessage req )
        throws MingleCodecException
    {
        CharSequence idStyleStr = req.h().getFirst( NAME_ID_STYLE );

        if ( idStyleStr == null ) 
        {
            return state.get( contexts, DEFAULT_ID_STYLE, "contexts" );
        }
        else
        {
            MingleHttpCodecContext res = codecContextFor( idStyleStr );

            if ( res == null ) 
            {
                throw
                    new NoSuchMingleCodecException( 
                        "Invalid id style: " + idStyleStr );
            }
            else return res;
        }
    }

    public static JsonMingleHttpCodecFactory getInstance() { return INSTANCE; }

    private
    final
    static
    class ContextImpl
    implements MingleHttpCodecContext
    {
        private final JsonMingleCodec codec;

        private
        ContextImpl( MingleIdentifierFormat fmt )
        {
            codec =
                new JsonMingleCodec.Builder().
                    setSerializerOptions(
                        new JsonSerializer.Options.Builder().
                            setSerialSuffix( "\n" ).
                            build()
                    ).
                    setIdentifierFormat( fmt ).
                    build();
        }

        public MingleCodec codec() { return codec; }
        public CharSequence contentType() { return "application/json"; }
    }

    static
    {
        DEFAULT_ID_STYLE = MingleIdentifierFormat.LC_HYPHENATED.getIdentifier();

        Map< MingleIdentifier, MingleHttpCodecContext > m = Lang.newMap();

        for ( MingleIdentifierFormat fmt :
                MingleIdentifierFormat.class.getEnumConstants() )
        {
            m.put( fmt.getIdentifier(), new ContextImpl( fmt ) );
        }

        INSTANCE = new JsonMingleHttpCodecFactory( Lang.unmodifiableMap( m ) );
    }
}
