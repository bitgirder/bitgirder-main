package com.bitgirder.json;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.parser.Grammar;
import com.bitgirder.parser.DocumentParserFactory;
import com.bitgirder.parser.DocumentParser;
import com.bitgirder.parser.Lexer;
import com.bitgirder.parser.Parsers;
import com.bitgirder.parser.DerivationMatch;
import com.bitgirder.parser.ProductionMatch;
import com.bitgirder.parser.UnionMatch;
import com.bitgirder.parser.QuantifierMatch;
import com.bitgirder.parser.SequenceMatch;

final
class JsonGrammars
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final static Grammar< LxNt, Character > RFC4627_LEXICAL_GRAMMAR;

    private final static Grammar< SxNt, JsonToken > RFC4627_SYNTACTIC_GRAMMAR;

    final static DocumentParser.FeedFilter< JsonToken > RFC4627_FEED_FILTER =
        new Rfc4627TokenFeedFilter();
    
    final static DocumentParserFactory< SxNt, LxNt, JsonToken >
        RFC4627_DOCUMENT_PARSER_FACTORY;

    static
    enum SxNt
    {
        JSON_TEXT,
        JSON_ARRAY,
        JSON_OBJECT,
        JSON_VALUE,
        JSON_OBJECT_MEMBER;
    }

    static
    enum LxNt
    {
        JSON_TOKEN,
        JSON_TOKEN_NULL,
        JSON_TOKEN_TRUE,
        JSON_TOKEN_FALSE,
        JSON_TOKEN_NUMBER,
        JSON_TOKEN_NUMBER_INT_PART,
        JSON_TOKEN_NUMBER_FRAC_PART,
        JSON_TOKEN_NUMBER_EXP_PART,
        JSON_TOKEN_STRING,
        JSON_TOKEN_UNESCAPED_CHAR,
        JSON_TOKEN_MNEMONIC_CHAR_ESCAPE,
        JSON_TOKEN_UTF16_CODE_UNIT_ESCAPE,
        JSON_TOKEN_WHITESPACE,
        JSON_TOKEN_BEGIN_ARRAY,
        JSON_TOKEN_END_ARRAY,
        JSON_TOKEN_BEGIN_OBJECT,
        JSON_TOKEN_END_OBJECT,
        JSON_TOKEN_NAME_SEPARATOR,
        JSON_TOKEN_VALUE_SEPARATOR;
    }

    private
    static
    DerivationMatch< LxNt, Character >
    extractTokenMatch( DerivationMatch< LxNt, Character > dm )
    {
        state.equal( LxNt.JSON_TOKEN, dm.getHead() );
        
        UnionMatch< Character > um = Parsers.castUnionMatch( dm.getMatch() );

        DerivationMatch< LxNt, Character > res = 
            Parsers.castDerivationMatch( um.getMatch() );

        return res;
    }

    private
    static
    char
    fromUnescapedChar( ProductionMatch< Character > pm )
    {
        return Parsers.asString( pm ).charAt( 0 );
    }

    private
    static
    char
    fromMnemonicCharEscape( ProductionMatch< Character > pm )
    {
        char esc = Parsers.asString( pm ).charAt( 1 );

        switch ( esc )
        {
            case 't': return '\t';
            case 'n': return '\n';
            case 'r': return '\r';
            case 'f': return '\f';
            case 'b': return '\b';
            case '/': return '/';
            case '\\': return '\\';
            case '"': return '"';
            
            default: 
                throw state.createFail( "Invalid escape sequence: \\" + esc );
        }
    }

    private
    static
    char
    fromUtf16CodeUnitEscape( ProductionMatch< Character > pm )
    {
        CharSequence escStr = Parsers.asString( pm );
        String hexStr = escStr.subSequence( 2, 6 ).toString();

        return (char) Integer.parseInt( hexStr, 16 );
    }

    private
    static
    char
    asJavaChar( ProductionMatch< Character > pm )
    {
        UnionMatch< Character > um = Parsers.castUnionMatch( pm );

        DerivationMatch< LxNt, Character > dm = 
            Parsers.castDerivationMatch( um.getMatch() );

        ProductionMatch< Character > pm2 = dm.getMatch();
        LxNt charType = dm.getHead();

        switch ( charType )
        {
            case JSON_TOKEN_UNESCAPED_CHAR: return fromUnescapedChar( pm2 );

            case JSON_TOKEN_MNEMONIC_CHAR_ESCAPE: 
                return fromMnemonicCharEscape( pm2 );

            case JSON_TOKEN_UTF16_CODE_UNIT_ESCAPE:
                return fromUtf16CodeUnitEscape( pm2 );
            
            default: 
                throw state.createFail( "Unrecognized char type:", charType );
        }
    }

    private
    static
    JsonTokenString
    createJsonTokenString( ProductionMatch< Character > pm )
    {
        SequenceMatch< Character > sm = Parsers.castSequenceMatch( pm );

        QuantifierMatch< Character > qm =
            Parsers.castQuantifierMatch( sm.get( 1 ) );

        char[] str = new char[ qm.size() ];

        for ( int i = 0, e = str.length; i < e; ++i )
        {
            str[ i ] = asJavaChar( qm.get( i ) );
        }

        return new JsonTokenString( str );
    }

    private
    static
    boolean
    isNumNegation( ProductionMatch< Character > pm )
    {
        CharSequence str = Parsers.asString( pm );

        return str.length() > 0 ? str.charAt( 0 ) == '-' : false;
    }

    private
    static
    CharSequence
    asNumFracPart( ProductionMatch< Character > pm )
    {
        DerivationMatch< LxNt, Character > dm =
            Parsers.extractUnaryDerivationMatch( pm );

        if ( dm == null ) return null;
        else 
        {
            SequenceMatch< Character > sm =
                Parsers.castSequenceMatch( dm.getMatch() );

            return Parsers.asString( sm.get( 1 ) );
        }
    }

    private
    static
    JsonTokenNumber.Exponent
    getNumExponent( ProductionMatch< Character > pm )
    {
        DerivationMatch< LxNt, Character > dm =
            Parsers.extractUnaryDerivationMatch( pm );

        if ( dm == null ) return null;
        else
        {
            SequenceMatch< Character > sm = 
                Parsers.castSequenceMatch( dm.getMatch() );

            CharSequence signStr = Parsers.asString( sm.get( 1 ) );

            boolean negated = 
                signStr.length() > 0 ? signStr.charAt( 0 ) == '-' : false;

            CharSequence num = Parsers.asString( sm.get( 2 ) );

            return new JsonTokenNumber.Exponent( negated, num );
        }
    }

    private
    static
    JsonTokenNumber
    createJsonTokenNumber( ProductionMatch< Character > pm )
    {
        SequenceMatch< Character > sm = Parsers.castSequenceMatch( pm );
 
        boolean negated = isNumNegation( sm.get( 0 ) );
        CharSequence intPart = Parsers.asString( sm.get( 1 ) );
        CharSequence fracPart = asNumFracPart( sm.get( 2 ) );
        JsonTokenNumber.Exponent exp = getNumExponent( sm.get( 3 ) );

        return new JsonTokenNumber( negated, intPart, fracPart, exp );
    }

    private
    static
    JsonToken
    buildToken( DerivationMatch< LxNt, Character > dm )
    {
        DerivationMatch< LxNt, Character > tokMatch = extractTokenMatch( dm );

        ProductionMatch< Character > pm = tokMatch.getMatch();
        LxNt head = tokMatch.getHead();

        switch ( head )
        {
            case JSON_TOKEN_NULL: return JsonTokenNull.INSTANCE;
            case JSON_TOKEN_TRUE: return JsonTokenTrue.INSTANCE;
            case JSON_TOKEN_FALSE: return JsonTokenFalse.INSTANCE;
            case JSON_TOKEN_BEGIN_OBJECT: return JsonTokenBeginObject.INSTANCE;
            case JSON_TOKEN_END_OBJECT: return JsonTokenEndObject.INSTANCE;
            case JSON_TOKEN_BEGIN_ARRAY: return JsonTokenBeginArray.INSTANCE;
            case JSON_TOKEN_END_ARRAY: return JsonTokenEndArray.INSTANCE;
            case JSON_TOKEN_WHITESPACE: return JsonTokenWhitespace.INSTANCE;

            case JSON_TOKEN_NAME_SEPARATOR:
                return JsonTokenNameSeparator.INSTANCE;

            case JSON_TOKEN_VALUE_SEPARATOR:
                return JsonTokenValueSeparator.INSTANCE;

            case JSON_TOKEN_STRING: return createJsonTokenString( pm );
            case JSON_TOKEN_NUMBER: return createJsonTokenNumber( pm );

            default: throw state.createFail( "Unexpected head:", head );
        }
    }

    private
    final
    static
    class Rfc4627TokenFeedFilter
    implements DocumentParser.FeedFilter< JsonToken >
    {
        public
        boolean
        shouldFeed( JsonToken token )
        {
            return ! ( token instanceof JsonTokenWhitespace );
        }
    }

    static
    {
        Grammar.LexBuilder< LxNt > b = Grammar.createLexBuilder( LxNt.class );

        b.addProduction(
            LxNt.JSON_TOKEN,
            b.union(
                b.derivation( LxNt.JSON_TOKEN_NULL ),
                b.derivation( LxNt.JSON_TOKEN_TRUE ),
                b.derivation( LxNt.JSON_TOKEN_FALSE ),
                b.derivation( LxNt.JSON_TOKEN_NUMBER ),
                b.derivation( LxNt.JSON_TOKEN_STRING ),
                b.derivation( LxNt.JSON_TOKEN_WHITESPACE ),
                b.derivation( LxNt.JSON_TOKEN_BEGIN_ARRAY ),
                b.derivation( LxNt.JSON_TOKEN_END_ARRAY ),
                b.derivation( LxNt.JSON_TOKEN_BEGIN_OBJECT ),
                b.derivation( LxNt.JSON_TOKEN_END_OBJECT ),
                b.derivation( LxNt.JSON_TOKEN_NAME_SEPARATOR ),
                b.derivation( LxNt.JSON_TOKEN_VALUE_SEPARATOR ) ) );

        b.addProduction( LxNt.JSON_TOKEN_NULL, b.string( "null" ) );
        b.addProduction( LxNt.JSON_TOKEN_TRUE, b.string( "true" ) );
        b.addProduction( LxNt.JSON_TOKEN_FALSE, b.string( "false" ) );

        b.addProduction(
            LxNt.JSON_TOKEN_NUMBER,
            b.sequence(
                b.unary( b.ch( '-' ) ),
                b.derivation( LxNt.JSON_TOKEN_NUMBER_INT_PART ),
                b.unary( b.derivation( LxNt.JSON_TOKEN_NUMBER_FRAC_PART ) ),
                b.unary( b.derivation( LxNt.JSON_TOKEN_NUMBER_EXP_PART ) ) ) );

        b.addProduction(
            LxNt.JSON_TOKEN_NUMBER_INT_PART,
            b.union(
                b.ch( '0' ),
                b.sequence(
                    b.charRange( '1', '9' ),
                    b.kleene( b.charRange( '0', '9' ) ) ) ) );

        b.addProduction(
            LxNt.JSON_TOKEN_NUMBER_FRAC_PART,
            b.sequence(
                b.ch( '.' ), b.atLeastOne( b.charRange( '0', '9' ) ) ) );

        b.addProduction(
            LxNt.JSON_TOKEN_NUMBER_EXP_PART,
            b.sequence(
                b.charSet( 'e', 'E' ),
                b.unary( b.charSet( '-', '+' ) ),
                b.atLeastOne( b.charRange( '0', '9' ) ) ) );

        b.addProduction(
            LxNt.JSON_TOKEN_STRING,
            b.sequence(
                b.ch( '"' ),
                b.kleene(
                    b.union(
                        b.derivation( LxNt.JSON_TOKEN_UNESCAPED_CHAR ),
                        b.derivation( LxNt.JSON_TOKEN_MNEMONIC_CHAR_ESCAPE ),
                        b.derivation( 
                            LxNt.JSON_TOKEN_UTF16_CODE_UNIT_ESCAPE ) ) ),
                b.ch( '"' ) ) );
 
        b.addProduction(
            LxNt.JSON_TOKEN_UNESCAPED_CHAR, b.charSetComplement( '"', '\\' ) );
 
        b.addProduction(
            LxNt.JSON_TOKEN_MNEMONIC_CHAR_ESCAPE,
            b.sequence(
                b.ch( '\\' ),
                b.charSet( '"', '/', '\\', 'b', 'f', 'n', 'r', 't' ) ) );
        
        b.addProduction(
            LxNt.JSON_TOKEN_UTF16_CODE_UNIT_ESCAPE,
            b.sequence(
                b.string( "\\u" ),
                b.exactly(
                    b.union(
                        b.charRange( '0', '9' ),
                        b.charRange( 'a', 'f' ),
                        b.charRange( 'A', 'F' ) ),
                    4 ) ) );

        b.addProduction(
            LxNt.JSON_TOKEN_WHITESPACE, 
            b.atLeastOne( b.charSet( ' ', '\t', '\r', '\n' ) ) );

        b.addProduction( LxNt.JSON_TOKEN_BEGIN_ARRAY, b.ch( '[' ) );
        b.addProduction( LxNt.JSON_TOKEN_END_ARRAY, b.ch( ']' ) );
        b.addProduction( LxNt.JSON_TOKEN_BEGIN_OBJECT, b.ch( '{' ) );
        b.addProduction( LxNt.JSON_TOKEN_END_OBJECT, b.ch( '}' ) );
        b.addProduction( LxNt.JSON_TOKEN_NAME_SEPARATOR, b.ch( ':' ) );
        b.addProduction( LxNt.JSON_TOKEN_VALUE_SEPARATOR, b.ch( ',' ) );
        
        RFC4627_LEXICAL_GRAMMAR = b.build();
    }

    static
    { 
        Grammar.Builder< SxNt, JsonToken > b =
            Grammar.createBuilder( SxNt.class );

        b.addProduction(
            SxNt.JSON_TEXT,
            b.union( 
                b.derivation( SxNt.JSON_ARRAY ), 
                b.derivation( SxNt.JSON_OBJECT ) ) );
        
        b.addProduction(
            SxNt.JSON_ARRAY,
            b.sequence(
                b.terminalInstance( JsonTokenBeginArray.class ),
                b.unary(
                    b.sequence(
                        b.derivation( SxNt.JSON_VALUE ),
                        b.kleene(
                            b.sequence(
                                b.terminalInstance(
                                    JsonTokenValueSeparator.class ),
                                b.derivation( SxNt.JSON_VALUE ) ) ) ) ),
                b.terminalInstance( JsonTokenEndArray.class  ) ) );

        b.addProduction(
            SxNt.JSON_OBJECT,
            b.sequence(
                b.terminalInstance( JsonTokenBeginObject.class ),
                b.unary(
                    b.sequence(
                        b.derivation( SxNt.JSON_OBJECT_MEMBER ),
                        b.kleene(
                            b.sequence(
                                b.terminalInstance(
                                    JsonTokenValueSeparator.class ),
                                b.derivation( SxNt.JSON_OBJECT_MEMBER ) ) ) ) ),
                b.terminalInstance( JsonTokenEndObject.class ) ) );
    
        b.addProduction(
            SxNt.JSON_OBJECT_MEMBER,
            b.sequence(
                b.terminalInstance( JsonTokenString.class ),
                b.terminalInstance( JsonTokenNameSeparator.class ),
                b.derivation( SxNt.JSON_VALUE ) ) );

        b.addProduction(
            SxNt.JSON_VALUE,
            b.union(
                b.terminalInstance( JsonTokenNull.class ),
                b.terminalInstance( JsonTokenFalse.class ),
                b.terminalInstance( JsonTokenTrue.class ),
                b.derivation( SxNt.JSON_OBJECT ),
                b.derivation( SxNt.JSON_ARRAY ),
                b.terminalInstance( JsonTokenNumber.class ),
                b.terminalInstance( JsonTokenString.class ) ) );

        RFC4627_SYNTACTIC_GRAMMAR = b.build();
    }

    private
    final
    static
    class JsonTokenBuilder
    implements Lexer.TokenBuilder< LxNt, JsonToken >
    {
        public
        JsonToken
        buildToken( DerivationMatch< LxNt, Character > dm )
        {
            return JsonGrammars.buildToken( dm );
        }
    }

    static
    {
        RFC4627_DOCUMENT_PARSER_FACTORY =
            new DocumentParserFactory.Builder< SxNt, LxNt, JsonToken >().
                setLexicalGrammar( RFC4627_LEXICAL_GRAMMAR ).
                setLexicalGoal( LxNt.JSON_TOKEN ).
                setTokenBuilder( new JsonTokenBuilder() ).
                setSyntacticGrammar( RFC4627_SYNTACTIC_GRAMMAR ).
                setSyntacticGoal( SxNt.JSON_TEXT ).
                build();
    }
}
