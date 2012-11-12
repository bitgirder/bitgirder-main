package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;
import com.bitgirder.lang.TypedString;
import com.bitgirder.lang.Inspectable;
import com.bitgirder.lang.Inspector;

import com.bitgirder.lang.reflect.ReflectUtils;

import com.bitgirder.test.Test;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.io.Charsets;
import com.bitgirder.io.IoUtils;

import java.lang.reflect.Constructor;

import java.nio.ByteBuffer;
import java.nio.CharBuffer;

import java.util.List;
import java.util.Map;
import java.util.Iterator;

@Test
final
class GrammarTests
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final static String BUILTIN_FILE_NAME = "(test)";

    private final static Grammar< Class, Character > LEXICAL_GRAMMAR_1;

    private final static RecursiveDescentParserFactory< Class, Character >
        LEXICAL_GRAMMAR_1_PF;

    private final static Grammar< Class, TestToken > SYNTACTIC_GRAMMAR_1;

    private static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static boolean doLog = false;
    private final static boolean logTokens = doLog;
    private final static boolean logSyntaxTokens = doLog;
    private final static boolean logParser = doLog;
    private final static boolean logSyntaxParser = doLog;
 
    private final static boolean doTokenLog = false;
    private final static boolean logTokenizerChars = doTokenLog;
    private final static boolean logTokenizerParser = doTokenLog;
    private final static boolean logLexerFedChars = doTokenLog;
    private final static boolean logLexerParser = doTokenLog;

    private
    static
    interface TestToken
    {}

    private
    final
    static
    class TestIdentifierPiece
    extends TypedString< TestIdentifierPiece >
    {
        private TestIdentifierPiece( CharSequence cs ) { super( cs ); }
    }

    private
    final
    static
    class TestIdentifier
    extends TypedString< TestIdentifier >
    implements TestToken
    {
        private TestIdentifier( CharSequence s ) { super( s ); }
    }

    private
    final
    static
    class TestIntegral
    extends TypedString< TestIntegral >
    implements TestToken
    {
        private TestIntegral( CharSequence s ) { super( s ); }
    }

    private
    final
    static
    class WhitespaceToken
    extends TypedString< TestIdentifier >
    implements TestToken
    {
        private WhitespaceToken( CharSequence s ) { super( s ); }
    }

    private
    final
    static
    class TestScriptComment
    extends TypedString< TestScriptComment >
    implements TestToken
    {
        private TestScriptComment( CharSequence s ) { super( s ); }
    }

    private
    abstract
    static
    class LiteralToken< T extends LiteralToken >
    extends TypedString< LiteralToken< T > >
    implements TestToken
    {
        private LiteralToken( CharSequence cs ) { super( cs ); }
    }

    private 
    final 
    static 
    class OpenBrace 
    extends LiteralToken< OpenBrace >
    {
        private OpenBrace() { super( "{" ); }
    }

    private
    final
    static
    class CloseBrace
    extends LiteralToken< CloseBrace >
    {
        private CloseBrace() { super( "}" ); }
    }

    private
    final
    static
    class OpenBracket
    extends LiteralToken< OpenBracket >
    {
        private OpenBracket() { super( "[" ); }
    }

    private
    final
    static
    class CloseBracket
    extends LiteralToken< CloseBracket >
    {
        private CloseBracket() { super( "]" ); }
    }

    private
    final
    static
    class Comma
    extends LiteralToken< Comma >
    {
        private Comma() { super( "," ); }
    }

    private
    final
    static
    class Semicolon
    extends LiteralToken< Semicolon >
    {
        private Semicolon() { super( ";" ); }
    }

    private
    final
    static
    class OpenParen
    extends LiteralToken< OpenParen >
    {
        private OpenParen() { super( "(" ); }
    }

    private
    final
    static
    class CloseParen
    extends LiteralToken< CloseParen >
    {
        private CloseParen() { super( ")" ); }
    }

    private static interface Expression1Op {}

    private
    final
    static
    class Plus
    extends LiteralToken< Plus >
    implements Expression1Op
    {
        private Plus() { super( "+" ); }
    }

    private
    final
    static
    class Minus
    extends LiteralToken< Minus >
    implements Expression1Op
    {
        private Minus() { super( "-" ); }
    }

    private static interface Expression2Op {}

    private
    final
    static
    class Asterisk
    extends LiteralToken< Asterisk >
    implements Expression2Op
    {
        private Asterisk() { super( "*" ); }
    }

    private
    final
    static
    class ForwardSlash
    extends LiteralToken< ForwardSlash >
    implements Expression2Op
    {
        private ForwardSlash() { super( "/" ); }
    }

    private
    final
    static
    class AssignmentOp
    extends LiteralToken< AssignmentOp >
    {
        private AssignmentOp() { super( "=" ); }
    }

    private
    static
    abstract
    class KeywordToken< K extends KeywordToken >
    extends LiteralToken< K >
    {
        private KeywordToken( CharSequence cs ) { super( cs ); }
    }

    private
    final
    static
    class KwdNull
    extends KeywordToken< KwdNull >
    {
        private KwdNull() { super( "null" ); }
    }

    private
    final
    static
    class StringLiteral
    implements TestToken,
               TestValue,
               Factor1
    {
        private final String jstring;

        private StringLiteral( String jstring ) { this.jstring = jstring; }

        public
        Inspector
        accept( Inspector i )
        {
            return i.add( "jstring", jstring );
        }
    }

    private
    static
    interface TestChar
    {}

    private
    static
    final
    class MnemonicCharEscape
    implements TestChar
    {}

    private
    static
    final
    class Utf16CodeUnitEscape
    implements TestChar
    {}

    private
    static
    final
    class UnescapedChar
    implements TestChar
    {}

    private
    final
    static
    class Expression
    {
        private final Expression1 lhs;
        private final Expression1 rhs;

        private
        Expression( Expression1 lhs,
                    Expression1 rhs )
        {
            this.lhs = lhs;
            this.rhs = rhs;
        }
    }

    private
    final
    static
    class Expression1
    {
        private final Expression2 expr2;
        private final Expression1Rest rest;

        private
        Expression1( Expression2 expr2,
                     Expression1Rest rest )
        {
            this.expr2 = expr2;
            this.rest = rest;
        }
    }

    private
    final
    static
    class Expression1Rest
    {
        private final Expression1Op op;
        private final Expression2 expr2;
        private final Expression1Rest rest;

        private
        Expression1Rest( Expression1Op op,
                         Expression2 expr2,
                         Expression1Rest rest )
        {   
            this.op = op;
            this.expr2 = expr2;
            this.rest = rest;
        }
    }

    private
    final
    static
    class Expression2
    {
        private final Factor1 fact;
        private final Expression2Rest rest;

        private 
        Expression2( Factor1 fact,
                     Expression2Rest rest )
        {
            this.fact = fact;
            this.rest = rest;
        }
    }

    private
    final
    static
    class Expression2Rest
    {
        private final Expression2Op op;
        private final Factor1 fact;
        private final Expression2Rest rest;

        private
        Expression2Rest( Expression2Op op,
                         Factor1 fact,
                         Expression2Rest rest )
        {
            this.op = op;
            this.fact = fact;
            this.rest = rest;
        }
    }

    private
    static
    interface Factor1
    {}
 
    private
    final
    static
    class IdentifierFactor1
    implements Factor1
    {
        private final TestIdentifier identifier;

        private 
        IdentifierFactor1( TestIdentifier identifier )
        {
            this.identifier = identifier;
        }
    }

    private
    final
    static
    class IntegralFactor1
    implements Factor1
    {
        private final TestIntegral integral;

        private
        IntegralFactor1( TestIntegral integral )
        {
            this.integral = integral;
        }
    }

    private
    final
    static
    class Expression1Factor1
    implements Factor1
    {
        private final Expression1 expr;

        private Expression1Factor1( Expression1 expr ) { this.expr = expr; }
    }

    private
    final
    static
    class ListExpression
    implements Factor1
    {
        private final List< Expression1 > exprs;

        private
        ListExpression( List< Expression1 > exprs )
        {
            this.exprs = Lang.unmodifiableCopy( exprs );
        }
    }

    private
    static
    interface Statement
    {}

    private
    static
    final
    class ExpressionStatement
    implements Statement
    {
        private final Expression expr;

        private ExpressionStatement( Expression expr ) { this.expr = expr; }
    }

    private
    final
    static
    class ScriptBody
    {
        private final List< Statement > body;

        private 
        ScriptBody( List< Statement > body )
        {
            this.body = Lang.unmodifiableList( body );
        }
    }

    private
    final
    static
    class TestGoal
    {}

    private
    final
    static
    class Env
    {
        private final Map< TestIdentifier, TestValue > vals;

        private
        Env( Map< TestIdentifier, TestValue > vals )
        {
            this.vals = vals;
        }

        void
        bind( TestIdentifier ident,
              TestValue val )
        {
            vals.put( ident, val );
        }

        static
        Env
        create()
        {
            return new Env( Lang.< TestIdentifier, TestValue >newMap() );
        }
    }

    private 
    static 
    interface TestValue 
    extends Inspectable 
    {}

    private
    static
    final
    class NumValue
    implements TestValue
    {
        private final int num;

        private NumValue( int num ) { this.num = num; }

        public Inspector accept( Inspector i ) { return i.add( "num", num ); }
    }

    private
    final
    static
    class ListValue
    implements TestValue
    {
        private final List< TestValue > vals;

        private
        ListValue( List< TestValue > vals )
        {
            this.vals = Lang.unmodifiableCopy( vals );
        }

        public Inspector accept( Inspector i ) { return i.add( "vals", vals ); }
    }

    private
    static
    < T extends TestToken >
    T
    createTestTokenInstance( Class< T > cls,
                             CharSequence str )
    {
        try
        {
            Constructor< T > cons = 
                ReflectUtils.getDeclaredConstructor(
                    cls, new Class< ? >[] { CharSequence.class } );

            return ReflectUtils.invoke( cons, new Object[] { str } );
        }
        catch ( Exception ex ) { throw new RuntimeException( ex ); }
    }

    private
    static
    Class< ? extends TestToken >
    getTokenClass( DerivationMatch< Class, Character > dm )
    {
        return ( (Class< ? >) dm.getHead() ).asSubclass( TestToken.class );
    }

    private
    static
    TestToken
    buildStringToken( DerivationMatch< Class, Character > dm )
    {
        Class< ? extends TestToken > cls = getTokenClass( dm );

        CharSequence str = Parsers.asString( dm.getMatch() );

        return createTestTokenInstance( cls, str );
    }

    private
    static
    TestToken
    buildLiteralToken( DerivationMatch< Class, Character > dm )
    {
        UnionMatch< Character > um = Parsers.castUnionMatch( dm.getMatch() );

        DerivationMatch< Class, Character > dm2 =
            Parsers.castDerivationMatch( um.getMatch() );

        Class< ? extends TestToken > cls = getTokenClass( dm2 );

        try { return ReflectUtils.newInstance( cls ); }
        catch ( Exception ex ) { throw new RuntimeException( ex ); }
    }

    private
    static
    char
    getMnemonicCharEscape( ProductionMatch< Character > pm )
    {
        char ch = Parsers.asString( pm ).charAt( 1 );

        switch ( ch )
        {
            case 't': return '\t';
            case 'n': return '\n';
            case 'r': return '\r';
            case 'f': return '\f';
            case '"': return '"';
            case '\\': return '\\';

            default: throw state.createFail( "Unexpected escape:", ch );
        }
    }

    private
    static
    char
    getUtf16CodeUnitEscape( ProductionMatch< Character > pm )
    {
        String escStr = Parsers.asString( pm ).toString();
        return (char) Integer.parseInt( escStr.substring( 2 ), 16 );
    }

    private
    static
    char
    getUnescapedChar( ProductionMatch< Character > pm )
    {
        return Parsers.asString( pm ).charAt( 0 );
    }

    private
    static
    char
    getStringChar( ProductionMatch< Character > pm )
    {
        // pm is TestChar, we just want to work with it's union match:
        DerivationMatch< Class, Character > tcDm =
            Parsers.castDerivationMatch( pm );

        UnionMatch< Character > um = Parsers.castUnionMatch( tcDm.getMatch() );

        ProductionMatch< Character > pm2 = um.getMatch();
        int alt = um.getAlternative();
        
        switch ( alt )
        {
            case 0: return getMnemonicCharEscape( pm2 );
            case 1: return getUtf16CodeUnitEscape( pm2 );
            case 2: return getUnescapedChar( pm2 );

            default: throw state.createFail( "Unexpected alt:", alt );
        }
    }

    private
    static
    StringLiteral
    buildStringLiteral( DerivationMatch< Class, Character > dm )
    {
        SequenceMatch< Character > sm = 
            Parsers.castSequenceMatch( dm.getMatch() );

        QuantifierMatch< Character > qm =
            Parsers.castQuantifierMatch( sm.get( 1 ) );

        char[] str = new char[ qm.size() ];

        int i = 0;
        for ( ProductionMatch< Character > pm : qm )
        {
            str[ i++ ] = getStringChar( pm );
        }

        return new StringLiteral( new String( str ) );
    }

    private
    static
    TestToken
    buildTestToken( DerivationMatch< Class, Character > dm )
    {
        UnionMatch< Character > um = Parsers.castUnionMatch( dm.getMatch() );

        int alt = um.getAlternative();
        
        DerivationMatch< Class, Character > dm2 = 
            Parsers.castDerivationMatch( um.getMatch() );

        switch ( alt )
        {
            case 0: return buildLiteralToken( dm2 );

            case 1:
            case 2: 
            case 3:
            case 5:
                return buildStringToken( dm2 );

            case 4: return buildStringLiteral( dm2 );

            default: throw state.createFail( "Unrecognized alt:", alt );
        }
    }

    private
    final
    static
    class LocatedToken< T >
    {
        private final T token;
        private final SourceTextLocation loc;

        private
        LocatedToken( T token,
                      SourceTextLocation loc )
        {
            this.token = token;
            this.loc = loc;
        }

        @Override
        public
        String
        toString()
        {
            return Strings.inspect( 
                this, true, "token", token, "loc", loc ).toString();
        }
    }

    // Implements both interfaces, but is not meant to be used as both
    // simultaneously
    private
    final
    static
    class LocationAccumulator< T >
    implements Lexer.Listener< T >,
               DocumentParser.TokenLocationListener< T >
    {
        private final List< LocatedToken< T > > acc = Lang.newList();

        private
        void
        acc( T token,
             SourceTextLocation loc )
        {
            acc.add( new LocatedToken< T >( token, loc ) );
        }

        public
        void
        lexerTokenized( T token,
                        SourceTextLocation loc )
        {
            acc( token, loc );
        }

        public
        void
        markToken( T token,
                   SourceTextLocation loc )
        {
            acc( token, loc );
        }

        public void lexerComplete() {}
    }

    private
    static
    Lexer< Class, TestToken >
    createTestTokenLexer( Lexer.Listener< TestToken > listener )
    {
        RecursiveDescentParserFactory< Class, Character > lexerPf =
            RecursiveDescentParserFactory.forGrammar( LEXICAL_GRAMMAR_1 );
 
        return
            new Lexer.Builder< Class, TestToken >().
                setFileName( BUILTIN_FILE_NAME ).
                setCharset( Charsets.UTF_8.charset() ).
                setParserFactory( lexerPf ).
                setListener( listener ).
                setGoal( TestToken.class ).
                setTokenBuilder( new TestTokenBuilder() ).
                setLogFedChars( logLexerFedChars ).
                setLogParser( logLexerParser ).
                build();
    }

    private
    static
    void
    feedLexer( Lexer< ?, ? > lexer,
               ByteBuffer src,
               int chunkSize )
        throws Exception
    {
        for ( int e = src.limit(); src.hasRemaining(); )
        {
            int len = Math.min( src.remaining(), chunkSize );
            src.limit( src.position() + len );
 
            boolean isEnd = src.limit() == e;
            lexer.update( src, src.limit() == e );
            src.limit( e );
        }
    }

    private
    static
    List< TestToken >
    tokenize( CharSequence text,
              int chunkSize )
        throws Exception
    {
        LocationAccumulator< TestToken > acc =
            new LocationAccumulator< TestToken >();

        Lexer< Class, TestToken > lexer = createTestTokenLexer( acc );

        ByteBuffer src = Charsets.UTF_8.asByteBuffer( text );
        feedLexer( lexer, src, chunkSize );

        List< TestToken > res = Lang.newList();

        for ( LocatedToken< TestToken > lt : acc.acc )
        {
            if ( ! ( lt.token instanceof WhitespaceToken ||
                     lt.token instanceof TestScriptComment ) ) 
            {
                res.add( lt.token );
            }
        }

        return res;
    }

    private
    static
    List< TestToken >
    tokenize( CharSequence text )
        throws Exception
    {
        return tokenize( text, Integer.MAX_VALUE );
    }

    private
    DerivationMatch< Class, TestToken >
    parseScript( CharSequence script )
        throws Exception
    {
        List< TestToken > toks = tokenize( script );

        ParseResult< Class, TestToken > pr = 
            getParseResult( SYNTACTIC_GRAMMAR_1, ScriptBody.class, toks );
        
        state.isFalse( pr.match == null, "Parser didn't match script" );

        return pr.match;
    }

    private
    static
    Expression2Rest
    buildExpression2Rest( ProductionMatch< TestToken > m )
    {
        SequenceMatch< TestToken > sm = Parsers.castSequenceMatch( m );

        UnionMatch< TestToken > um = Parsers.castUnionMatch( sm.get( 0 ) );
        TerminalMatch< TestToken > tm = 
            Parsers.castTerminalMatch( um.getMatch() );
        Expression2Op op = (Expression2Op) tm.getTerminal();

        Factor1 fact = buildDerivation( Factor1.class, sm.get( 1 ) );

        Expression2Rest rest = 
            buildUnaryDerivation( Expression2Rest.class, sm.get( 2 ) );
        
        return new Expression2Rest( op, fact, rest );
    }

    private
    static
    Expression2
    buildExpression2( ProductionMatch< TestToken > pm )
    {
        SequenceMatch< TestToken > sm = Parsers.castSequenceMatch( pm );

        Factor1 fact = buildDerivation( Factor1.class, sm.get( 0 ) );

        QuantifierMatch< TestToken > qm =
            Parsers.castQuantifierMatch( sm.get( 1 ) );
        Expression2Rest rest = qm.size() > 0
            ? buildDerivation( Expression2Rest.class, qm.get( 0 ) ) : null;

        return new Expression2( fact, rest );
    }

    private
    static
    Factor1
    buildFactor1Alt0( ProductionMatch< TestToken > pm )
    {
        TerminalMatch< TestToken > tm = Parsers.castTerminalMatch( pm );
        return new IdentifierFactor1( (TestIdentifier) tm.getTerminal() );
    }

    private
    static
    Factor1
    buildFactor1Alt1( ProductionMatch< TestToken > pm )
    {
        TerminalMatch< TestToken > tm = Parsers.castTerminalMatch( pm );
        return new IntegralFactor1( (TestIntegral) tm.getTerminal() );
    }

    private
    static
    Factor1
    buildFactor1Alt2( ProductionMatch< TestToken > pm )
    {
        TerminalMatch< TestToken > tm = Parsers.castTerminalMatch( pm );
        return (StringLiteral) tm.getTerminal();
    }

    private
    static
    Factor1
    buildFactor1Alt3( ProductionMatch< TestToken > pm )
    {
        return buildDerivation( ListExpression.class, pm );
    }

    private
    static
    Factor1
    buildFactor1Alt4( ProductionMatch< TestToken > pm )
    {
        SequenceMatch< TestToken > sm = Parsers.castSequenceMatch( pm );
        
        Expression1 expr = buildDerivation( Expression1.class, sm.get( 1 ) );
        return new Expression1Factor1( expr );
    }

    private
    static
    Factor1
    buildFactor1( ProductionMatch< TestToken > pm )
    {
        UnionMatch< TestToken > um = Parsers.castUnionMatch( pm );

        int alt = um.getAlternative();
        ProductionMatch< TestToken > pm2 = um.getMatch();

        switch ( alt )
        {
            case 0: return buildFactor1Alt0( pm2 );
            case 1: return buildFactor1Alt1( pm2 );
            case 2: return buildFactor1Alt2( pm2 );
            case 3: return buildFactor1Alt3( pm2 );
            case 4: return buildFactor1Alt4( pm2 );
            default: throw state.createFail( "Unhandled alt:", alt );
        }
    }

    private
    static
    Expression1Rest
    buildExpression1Rest( ProductionMatch< TestToken > pm )
    {
        SequenceMatch< TestToken > sm = Parsers.castSequenceMatch( pm );

        Expression1Op op = buildDerivation( Expression1Op.class, sm.get( 0 ) );

        Expression2 expr2 = buildDerivation( Expression2.class, sm.get( 1 ) );

        Expression1Rest rest = 
            buildUnaryDerivation( Expression1Rest.class, sm.get( 2 ) );

        return new Expression1Rest( op, expr2, rest );
    }

    private
    static
    Expression1Op
    buildExpression1Op( ProductionMatch< TestToken > pm )
    {
        UnionMatch< TestToken > um = Parsers.castUnionMatch( pm );

        TerminalMatch< TestToken > tm = 
            Parsers.castTerminalMatch( um.getMatch() );

        return (Expression1Op) tm.getTerminal();
    }

    private
    static
    Expression1
    buildExpression1( ProductionMatch< TestToken > pm )
    {
        SequenceMatch< TestToken > sm = Parsers.castSequenceMatch( pm );
        
        Expression2 expr2 = buildDerivation( Expression2.class, sm.get( 0 ) );

        QuantifierMatch< TestToken > qm = 
            Parsers.castQuantifierMatch( sm.get( 1 ) );
        Expression1Rest rest = qm.size() > 0 
            ? buildDerivation( Expression1Rest.class, qm.get( 0 ) ): null;

        return new Expression1( expr2, rest );
    }

    private
    static
    ListExpression
    buildListExpression( ProductionMatch< TestToken > pm )
    {
        List< Expression1 > exprs = Lang.newList();

        SequenceMatch< TestToken > sm = Parsers.castSequenceMatch( pm );

        QuantifierMatch< TestToken > qm = 
            Parsers.castQuantifierMatch( sm.get( 1 ) );

        if ( qm.size() > 0 )
        {
            SequenceMatch< TestToken > sm2 = 
                Parsers.castSequenceMatch( qm.get( 0 ) );

            exprs.add( buildDerivation( Expression1.class, sm2.get( 0 ) ) );

            QuantifierMatch< TestToken > qm2 = 
                Parsers.castQuantifierMatch( sm2.get( 1 ) );

            for ( int i = 0, e = qm2.size(); i < e; ++i )
            {
                SequenceMatch< TestToken > sm3 =
                    Parsers.castSequenceMatch( qm2.get( i ) );
                
                exprs.add( buildDerivation( Expression1.class, sm3.get( 1 ) ) );
            }
        }

        return new ListExpression( exprs );
    }

    private
    static
    ScriptBody
    buildScriptBody( ProductionMatch< TestToken > pm )
    {
        QuantifierMatch< TestToken > qm = Parsers.castQuantifierMatch( pm );

        List< Statement > body = Lang.newList( qm.size() );

        for ( int i = 0, e = qm.size(); i < e; ++i )
        {
            body.add( buildDerivation( Statement.class, qm.get( i ) ) );
        }

        return new ScriptBody( body );
    }


    private
    static
    ExpressionStatement
    buildExpressionStatement( ProductionMatch< TestToken > pm )
    {
        SequenceMatch< TestToken > sm = Parsers.castSequenceMatch( pm );

        Expression expr = buildDerivation( Expression.class, sm.get( 0 ) );
        return new ExpressionStatement( expr );
    }

    private
    static
    Statement
    buildStatement( ProductionMatch< TestToken > pm )
    {
        return buildDerivation( ExpressionStatement.class, pm );
    }

    private
    static
    Expression
    buildExpression( ProductionMatch< TestToken > pm )
    {
        SequenceMatch< TestToken > sm = Parsers.castSequenceMatch( pm );

        Expression1 lhs = buildDerivation( Expression1.class, sm.get( 0 ) );

        QuantifierMatch< TestToken > qm = 
            Parsers.castQuantifierMatch( sm.get( 1 ) );

        Expression1 rhs = null;

        if ( qm.size() > 0 )
        {
            SequenceMatch< TestToken > sm2 = 
                Parsers.castSequenceMatch( qm.get( 0 ) );
 
            rhs = buildDerivation( Expression1.class, sm2.get( 1 ) );
        }

        return new Expression( lhs, rhs );
    }

    private
    static
    Object
    buildDerivation( DerivationMatch< Class, TestToken > dm )
    {
        Class head = dm.getHead(); 
        ProductionMatch< TestToken > m = dm.getMatch();

        if ( head.equals( Expression1.class ) ) return buildExpression1( m );
        else if ( head.equals( Expression1Rest.class ) )
        {
            return buildExpression1Rest( m );
        }
        else if ( head.equals( Expression1Op.class ) )
        {
            return buildExpression1Op( m );
        }
        else if ( head.equals( Expression2.class ) ) 
        {
            return buildExpression2( m );
        }
        else if ( head.equals( Expression2Rest.class ) )
        {
            return buildExpression2Rest( m );
        }
        else if ( head.equals( Factor1.class ) ) return buildFactor1( m );
        else if ( head.equals( ListExpression.class ) )
        {
            return buildListExpression( m );
        }
        else if ( head.equals( ScriptBody.class ) ) return buildScriptBody( m );
        else if ( head.equals( Statement.class ) ) return buildStatement( m );
        else if ( head.equals( ExpressionStatement.class ) )
        {
            return buildExpressionStatement( m );
        }
        else if ( head.equals( Expression.class ) ) return buildExpression( m );
        throw state.createFail( "Unrecognized head:", head );
    }

    private
    static
    < T >
    T
    buildDerivation( Class< T > cls,
                     ProductionMatch< TestToken > pm )
    {
        DerivationMatch< Class, TestToken > dm = 
            Parsers.castDerivationMatch( pm );

        state.equal( cls, dm.getHead() );

        return cls.cast( buildDerivation( dm ) );
    }

    private
    static
    < T >
    T
    buildUnaryDerivation( Class< T > cls,
                          ProductionMatch< TestToken > pm )
    {
        QuantifierMatch< TestToken > qm = Parsers.castQuantifierMatch( pm );

        return qm.size() > 0 ? buildDerivation( cls, qm.get( 0 ) ) : null;
    }

    private
    static
    StringLiteral
    asString( TestValue val )
    {
        if ( val instanceof StringLiteral ) return (StringLiteral) val;
        else if ( val instanceof NumValue )
        {
            return new StringLiteral( 
                Integer.toString( ( (NumValue) val ).num ) );
        }
        else throw state.createFail( "Can't convert to string:", val );
    }

    private
    static
    StringLiteral
    concatenate( StringLiteral s1,
                 StringLiteral s2 )
    {
        return new StringLiteral( s1.jstring + s2.jstring );
    }

    private
    static
    TestValue
    eval( Expression1Op op,
          TestValue left,
          TestValue right,
          Env env )
    {
        if ( left instanceof StringLiteral || right instanceof StringLiteral )
        {
            state.isTrue( op instanceof Plus, "Got unexpected string op:", op );
            return concatenate( asString( left ), asString( right ) );
        }
        else
        {
            int iLeft = ( (NumValue) left ).num;
            int iRight = ( (NumValue) right ).num;
    
            int res;
    
            if ( op instanceof Plus ) res = iLeft + iRight;
            else if ( op instanceof Minus ) res = iLeft - iRight;
            else throw state.createFail( "Unrecognized expr1 op:", op );
    
            return new NumValue( res );
        }
    }

    private
    static
    TestValue
    eval( Expression2Op op,
          TestValue left,
          TestValue right,
          Env env )
    {
        int iLeft = ( (NumValue) left ).num;
        int iRight = ( (NumValue) right ).num;

        int res;

        if ( op instanceof Asterisk ) res = iLeft * iRight;
        else if ( op instanceof ForwardSlash ) res = iLeft / iRight;
        else throw state.createFail( "Unrecognized expr2 op:", op );

        return new NumValue( res );
    }

    private
    static
    TestValue
    eval( IdentifierFactor1 fact,
          Env env )
    {
        return state.get( env.vals, fact.identifier, "env.vals" );
    }

    private
    static
    NumValue
    eval( IntegralFactor1 fact,
          Env env )
    {
        int num = Integer.parseInt( fact.integral.toString() );
        return new NumValue( num );
    }

    private
    static
    TestValue
    eval( Expression1Factor1 fact,
          Env env )
    {
        return eval( fact.expr, env );
    }

    private
    static
    TestValue
    eval( Factor1 fact,
          Env env )
    {
        if ( fact instanceof IdentifierFactor1 )
        {
            return eval( (IdentifierFactor1) fact, env );
        }
        else if ( fact instanceof IntegralFactor1 )
        {
            return eval( (IntegralFactor1) fact, env );
        }
        else if ( fact instanceof Expression1Factor1 )
        {
            return eval( (Expression1Factor1) fact, env );
        }
        else if ( fact instanceof StringLiteral ) return (StringLiteral) fact;
        else if ( fact instanceof ListExpression )
        {
            return eval( (ListExpression) fact, env );
        }
        else throw state.createFail( "Unrecognized factor:", fact );
    }

    private
    static
    TestValue
    eval( Expression2 expr2,
          Env env )
    {
        TestValue res = eval( expr2.fact, env );

        for ( Expression2Rest rest = expr2.rest; 
                rest != null; rest = rest.rest )
        {
            TestValue right = eval( rest.fact, env );
            res = eval( rest.op, res, right, env );
        }

        return res;
    }

    private
    static
    TestValue
    eval( Expression1 expr,
          Env env )
    {
        TestValue res = eval( expr.expr2, env );

        for ( Expression1Rest rest = expr.rest; rest != null; rest = rest.rest )
        {
            TestValue right = eval( rest.expr2, env );
            res = eval( rest.op, res, right, env );
        }

        return res;
    }

    private
    static
    ListValue
    eval( ListExpression expr,
          Env env )
    {
        List< TestValue > res = Lang.newList( expr.exprs.size() ); 

        for ( Expression1 expr1 : expr.exprs )
        {
            res.add( eval( expr1, env ) );
        }

        return new ListValue( res );
    }

    private
    static
    TestValue
    eval( Expression1 expr )
    {
        return eval( expr, Env.create() );
    }

    private
    static
    TestIdentifier
    extractLhsIdentifier( Expression expr )
    {
        return ( (IdentifierFactor1) expr.lhs.expr2.fact ).identifier;
    }

    private
    static
    TestValue
    eval( Expression expr,
          Env env )
    {
        TestValue res;
        
        if ( expr.rhs == null ) res = eval( expr.lhs, env );
        else
        {
            TestIdentifier ident = extractLhsIdentifier( expr );
            TestValue rhsVal = eval( expr.rhs, env );

            env.bind( ident, rhsVal );
            res = rhsVal;
        }

        return res;
    }

    private
    static
    TestValue
    eval( ExpressionStatement st,
          Env env )
    {
        return eval( st.expr, env );
    }

    private
    static
    TestValue
    eval( Statement st,
          Env env )
    {
        if ( st instanceof ExpressionStatement )
        {
            return eval( (ExpressionStatement) st, env );
        }
        else throw state.createFail( "Unrecognized statement:", st );
    } 

    private
    static
    TestValue
    eval( ScriptBody script,
          Env env )
    {
        TestValue res = null;

        for ( Statement st : script.body ) res = eval( st, env );

        return state.notNull( res );
    }

    private
    static
    TestValue
    eval( ScriptBody script )
    {
        return eval( script, Env.create() );
    }

    private
    final
    static
    class ParseResult< N, T >
    {
        private final ProductionMatcherState ms;
        private final DerivationMatch< N, T > match;

        private
        ParseResult( ProductionMatcherState ms,
                     DerivationMatch< N, T > match )
        {
            this.ms = ms;
            this.match = match;
        }
    }

    private
    void
    logToken( Object tok )
    {
        if ( logTokens ) 
        {
            code( "Feeding token", tok, "(" + tok.getClass() + ")" );
        }
    }

    private
    void
    logParser( RecursiveDescentParser< ?, ? > p )
    {
        if ( logParser ) code( "Parser after feed:", Strings.inspect( p ) );
    }

    private
    < N, T >
    void
    execParse( RecursiveDescentParser< N, T > p,
               List< T > toks )
        throws Exception
    {
        Iterator< T > it = toks.iterator();

        while ( it.hasNext() && p.isMatching() )
        {
            T tok = it.next();

            logToken( tok );
            p.consumeTerminal( tok );
            logParser( p );
        }

        if ( p.isMatching() && ! it.hasNext() ) p.complete();

        if ( it.hasNext() && p.isMatched() )
        {
            throw new UnconsumedInputException();
        }
    }

    private
    < N, T >
    ParseResult< N, T >
    getParseResult( Grammar< N, T > g,
                    N goal,
                    List< T > toks )
        throws Exception
    {
        RecursiveDescentParserFactory< N, T > pf =
            RecursiveDescentParserFactory.forGrammar( g );
 
        RecursiveDescentParser< N, T > p = pf.createParser( goal );

        execParse( p, toks );

        return new ParseResult< N, T >( 
            p.getMatcherState(), p.isMatched() ? p.getMatch() : null );
    }

    private
    final
    static
    class ParserFeeder< N, T >
    implements Lexer.Listener< T >
    {
        private final RecursiveDescentParser< N, T > parser;
        private SourceTextLocation errLoc;

        private
        ParserFeeder( RecursiveDescentParser< N, T > parser )
        {
            this.parser = parser;
        }

        public
        void
        lexerTokenized( T token,
                        SourceTextLocation start )
        {
            if ( errLoc == null )
            {
                if ( ! ( token instanceof WhitespaceToken ||
                         token instanceof TestScriptComment ) ) 
                {
                    parser.consumeTerminal( token );
                    if ( parser.isUnmatched() ) errLoc = start;
                }
            }
        }

        public 
        void 
        lexerComplete() 
        { 
            if ( parser.isMatching() ) parser.complete();
        }
    }

    private
    final
    static
    class TestTokenBuilder
    implements Lexer.TokenBuilder< Class, TestToken >
    {
        public
        TestToken
        buildToken( DerivationMatch< Class, Character > dm )
        {
            return buildTestToken( dm );
        }
    }

    private
    final
    static
    class TestTokenFeedFilter
    implements DocumentParser.FeedFilter< TestToken >
    {
        public
        boolean
        shouldFeed( TestToken token )
        {
            return ! ( token instanceof WhitespaceToken ||
                       token instanceof TestScriptComment );
        }
    }

    private
    final
    static
    class ScriptBodyBuilder
    implements SyntaxBuilder< Class, TestToken, ScriptBody >
    {
        public
        ScriptBody
        buildSyntax( DerivationMatch< Class, TestToken > dm )
        {
            return buildDerivation( ScriptBody.class, dm );
        }
    }

    private
    InvalidSyntaxException
    createParserFailure( ParserFeeder< ?, ? > f )
    {
        return new InvalidSyntaxException( state.notNull( f.errLoc ) );
    }

    private
    Lexer< Class, TestToken >
    createLexer( Lexer.Listener< TestToken > listener,
                 CharSequence fileName )
    {
        RecursiveDescentParserFactory< Class, Character > lexerPf =
            RecursiveDescentParserFactory.forGrammar( LEXICAL_GRAMMAR_1 );
 
        return
            new Lexer.Builder< Class, TestToken >().
                setFileName( fileName ).
                setCharset( Charsets.UTF_8.charset() ).
                setParserFactory( lexerPf ).
                setListener( listener ).
                setGoal( TestToken.class ).
                setTokenBuilder( new TestTokenBuilder() ).
                setLogFedChars( logLexerFedChars ).
                setLogParser( logLexerParser ).
                build();
    }

    private
    List< LocatedToken< TestToken > >
    lexToList( CharSequence text )
        throws Exception
    {
        LocationAccumulator< TestToken > l = 
            new LocationAccumulator< TestToken >();

        Lexer< Class, TestToken > lexer = createLexer( l, BUILTIN_FILE_NAME );

        ByteBuffer body = Charsets.UTF_8.asByteBuffer( text );
        lexer.update( body, true );

        return l.acc;
    }

    private
    DocumentParser< ScriptBody >
    createScriptParser( 
        CharSequence scriptFileName,
        DocumentParser.TokenLocationListener< TestToken > locListener )
    {
        DocumentParserFactory< Class, Class, TestToken > dpf =
            new DocumentParserFactory.Builder< Class, Class, TestToken >().
                setLexicalGrammar( LEXICAL_GRAMMAR_1 ).
                setLexicalGoal( TestToken.class ).
                setTokenBuilder( new TestTokenBuilder() ).
                setSyntacticGrammar( SYNTACTIC_GRAMMAR_1 ).
                setSyntacticGoal( ScriptBody.class ).
                build();

        DocumentParser.Builder< Class, Class, TestToken, ScriptBody > dpb =
            dpf.< ScriptBody >createParserBuilder().
                setFileName( scriptFileName ).
                setCharset( Charsets.UTF_8.charset() ).
                setSyntaxBuilder( new ScriptBodyBuilder() ).
                setFeedFilter( new TestTokenFeedFilter() ).
                setLogChars( logLexerFedChars ).
                setLogLexerParser( logLexerParser ).
                setLogSyntaxTokens( logSyntaxTokens ).
                setLogSyntaxParser( logSyntaxParser );
 
        if ( locListener != null ) dpb.setTokenLocationListener( locListener );

        return dpb.build();
    }

    private
    ScriptBody
    parseScript( ByteBuffer body,
                 DocumentParser< ScriptBody > p )
        throws Exception
    {
        p.update( body, true );
        return p.buildSyntax();
    }

    private
    TestValue
    evalScript( ByteBuffer body,
                CharSequence scriptFileName,
                DocumentParser.TokenLocationListener< TestToken > locListener )
        throws Exception
    {
        DocumentParser< ScriptBody > dp = 
            createScriptParser( scriptFileName, locListener );

        ScriptBody script = parseScript( body, dp );

        return eval( script );
    }

    private
    TestValue
    evalScript( ByteBuffer body,
                CharSequence scriptFileName )
        throws Exception
    {
        return evalScript( body, scriptFileName, null );
    }

    private
    TestValue
    evalScriptText( CharSequence scriptText )
        throws Exception
    {
        ByteBuffer bb = Charsets.UTF_8.asByteBuffer( scriptText );

        return evalScript( bb, "STDIN" );
    }

    private
    TestValue
    evalScriptFile( CharSequence scriptFileName )
        throws Exception
    {
        ByteBuffer bb = 
            IoUtils.toByteBuffer(
                ReflectUtils.getResourceAsStream( 
                    getClass(), scriptFileName ) );

        return evalScript( bb, scriptFileName );
    }

    private
    void
    assertEvalInt( int expct,
                   CharSequence text )
        throws Exception
    {
        NumValue numVal = (NumValue) evalScriptText( text );
        state.equalInt( expct, numVal.num );
    }

    @Test
    private
    void
    test0()
        throws Exception
    {
        assertEvalInt( 6, "1 + 2 + 3;" );
    }

    @Test
    private
    void
    test1()
        throws Exception
    {
        assertEvalInt( 6, "1 + ( 2 + 3 );" );
    }

    @Test
    private
    void
    test2()
        throws Exception
    {
        assertEvalInt( 9, "((1+2)+((3)))+(2)+1;" );
    }

    @Test
    private
    void
    test3()
        throws Exception
    {
        assertEvalInt( 15, "3 * 5;" );
    }

    @Test
    private
    void
    test4()
        throws Exception
    {
        assertEvalInt( 30, "3 * 5 * 2;" );
    }

    @Test
    private
    void
    test5()
        throws Exception
    {
        assertEvalInt( 7, "3 * 5 * 7 / 15;" );
    }

    @Test
    private
    void
    test6()
        throws Exception
    {
        assertEvalInt( 720, "( ( 1 ) ) * 2 * ( 3 * ( ( 4 ) ) * ( 5 * 6 ) );" );
    }

    @Test
    private
    void
    test7()
        throws Exception
    {
        assertEvalInt( 7, "1 + 2 * 3;" );
    }

    @Test
    private
    void
    test8()
        throws Exception
    {
        assertEvalInt( 7, "2 * 3 + 1;" );
    }

    @Test
    private
    void
    test9()
        throws Exception
    {
        assertEvalInt( 10, "5 * ( 4 + 12 ) / ( ( 2 + 2 ) * 2 );" );
    }

    private
    void
    assertIntVal( int i,
                  TestValue tv )
    {
        state.equalInt( i, ( (NumValue) tv ).num );
    }

    @Test
    private
    void
    test10()
        throws Exception
    {
        ListValue lv = 
            (ListValue) evalScriptText( 
                "[ 1 , 2 , 3 + 4 , 5 * 6 + 7 , 8 * ( 9 + 10 ) ];" );

        Iterator< TestValue > it = lv.vals.iterator();

        assertIntVal( 1, it.next() );
        assertIntVal( 2, it.next() );
        assertIntVal( 7, it.next() );
        assertIntVal( 37, it.next() );
        assertIntVal( 152, it.next() );

        state.isFalse( it.hasNext() );
    }

    @Test
    private
    void
    test11()
        throws Exception
    {
        CharSequence script = 
            Strings.join( "\n",
                "an_ident = 1;",
                "b = 2;",
                "c = b * ( an_ident + b );",
                "c + 2;"
            );

        TestValue res = evalScriptText( script );

        state.equalInt( 8, ( (NumValue) res ).num );
    }

    // Runs a parse for a grammar using the exactly quantifier and returns the
    // parse result for further assertion
    private
    ParseResult< Class, TestToken >
    getIdentStringGrammarParseResult( int expctLen,
                                      CharSequence text )
        throws Exception
    {
        Grammar.Builder< Class, TestToken > b =
            Grammar.createBuilder( Class.class );
 
        b.addProduction(
            TestGoal.class,
            b.exactly( b.terminalInstance( TestIdentifier.class ), expctLen ) );
 
        Grammar< Class, TestToken > g = b.build();

        return getParseResult( g, TestGoal.class, tokenize( text ) );
    }

    // Test Grammar.Builder.exactly() which matches the entire input, also tests
    // that the lexer correctly tokenizes all the way to its endput
    @Test
    private
    void
    test12()
        throws Exception
    {
        ParseResult< Class, TestToken > pr = 
            getIdentStringGrammarParseResult( 4, "a b c d" );

        state.equal( ProductionMatcherState.MATCHED, pr.ms );

        QuantifierMatch< TestToken > qm = 
            Parsers.castQuantifierMatch( pr.match.getMatch() );
 
        List< TestIdentifier > matches = Lang.newList();
        for ( ProductionMatch< TestToken > pm : qm )
        {
            TerminalMatch< TestToken > tm = Parsers.castTerminalMatch( pm );
            matches.add( (TestIdentifier) tm.getTerminal() );
        }

        state.equalString( "a|b|c|d", Strings.join( "|", matches ) );
    }

    // Test exactly() with an overflow
    //
    // Not the most ideal way to test this, since we actually will match, but
    // the parser will begin the next match with 'e' and then remain in state
    // unmatched. We can add more direct/specific tests going forward, but this
    // is at least coarse coverage for this type of test.
    @Test( expected = UnconsumedInputException.class )
    private
    void
    test13()
        throws Exception
    {
        ParseResult< Class, TestToken > pr =
            getIdentStringGrammarParseResult( 4, "a b c d e" );

        state.equal( ProductionMatcherState.UNMATCHED, pr.ms );
    }

    // Test exactly() with an underflow
    @Test
    private
    void
    test14()
        throws Exception
    {
        ParseResult< Class, TestToken > pr =
            getIdentStringGrammarParseResult( 4, "a b c" );

        state.equal( ProductionMatcherState.UNMATCHED, pr.ms );
    }

    // Utility method for checking that an entire string is accepted by the
    // given grammar, and that the same string is inherent in the match result;
    // returns the parser used in the match for further processing if needed
    private
    < N >
    void
    assertStringMatch( Grammar< N, Character > g,
                       N goal,
                       CharSequence expctStr )
    {
        RecursiveDescentParserFactory< N, Character > pf =
            RecursiveDescentParserFactory.forGrammar( g );

        RecursiveDescentParser< N, Character > p = pf.createParser( goal );

        for ( int i = 0, e = expctStr.length(); p.isMatching() && i < e; ++i )
        {
            p.consumeTerminal( Character.valueOf( expctStr.charAt( i ) ) );
        }

        if ( p.isMatching() ) p.complete();

        state.isTrue( 
            p.isMatched(), "Parser was unmatched for input:", expctStr );

        CharSequence parsed = Parsers.asString( p.getMatch() );
        state.equalString( expctStr, parsed );
    }

    // The following sets of tests check that parser does not loop infinitely on
    // a nested quantifier match which matches empty

    @Test
    private
    void
    test15()
    {
        Grammar.LexBuilder< String > b = 
            Grammar.createLexBuilder( String.class );

        b.addProduction( 
            "p1",
            b.sequence(
                b.kleene( b.kleene( b.ch( 'a' ) ) ),
                b.ch( 'b' ) ) );
 
        assertStringMatch( b.build(), "p1", "b" );
    }

    @Test
    private
    void
    test16()
    {
        Grammar.LexBuilder< String > b =
            Grammar.createLexBuilder( String.class );

        b.addProduction( 
            "p1",
            b.sequence(
                b.atLeastOne( b.kleene( b.ch( 'a' ) ) ),
                b.ch( 'b' ) ) );
        
        assertStringMatch( b.build(), "p1", "aaab" );
    }

    @Test
    private
    void
    test17()
    {
        Grammar.LexBuilder< String > b =
            Grammar.createLexBuilder( String.class );

        b.addProduction(
            "p1",
            b.sequence(
                b.kleene(
                    b.sequence(
                        b.ch( 'a' ),
                        b.kleene( b.ch( 'b' ) ),
                        b.ch( 'c' ) ) ),
                b.string( "abbbbbd" ) ) );
        
        assertStringMatch( b.build(), "p1", "abbbbbd" );
    }

    // Put in to repro and fix the following bug in RecursiveDescentParser: p1
    // consumes all of the input up to the end, but upon complete() discovers
    // that it will not match (no trailing '@'); after p1 gives up p2 attempts
    // to match, but seeing the eof flag, set during complete(), only matches a
    // single character.
    @Test
    private
    void
    test18()
    {
        Grammar.LexBuilder< String > b =
            Grammar.createLexBuilder( String.class );
 
        b.addProduction( "p1", 
            b.sequence( b.atLeastOne( b.ch( 'a' ) ), b.ch( '@' ) ) );

        b.addProduction( "p2", b.atLeastOne( b.ch( 'a' ) ) );

        b.addProduction( "p3",
            b.union( b.derivation( "p1" ), b.derivation( "p2" ) ) );
 
        assertStringMatch( b.build(), "p3", "aa" );
    }

    // Test that DocumentParser fails with a syntax exception on an incomplete
    // document
    @Test( expected = SyntaxException.class,
           expectedPattern = "^Unmatched document$" )
    private
    void
    test19()
        throws Exception
    {
        DocumentParser< ScriptBody > p = 
            createScriptParser( BUILTIN_FILE_NAME, null );
        
        parseScript( Charsets.UTF_8.asByteBuffer( "a + b =" ), p );
    }

    private
    < T extends TestToken >
    T
    expectOneToken( CharSequence str,
                    Class< T > cls,
                    int chunkSize )
        throws Exception
    {
        List< TestToken > toks = tokenize( str, chunkSize );
        state.equalInt( 1, toks.size() );
        
        return state.cast( cls, toks.get( 0 ) );
    } 

    // The test20() and test21() tests are regression tests for bugs discovered
    // arising from the fact that Lexer was not equipped to deal with multibyte
    // chars that span calls to update. We add the character with codepoint
    // u271f (the original character in the field which led to discovery of the
    // bug -- it's a thick cross character) both in the middle and end of input
    // to the lexer and assert that things are okay by feeding the input bytes 1
    // at a time.  Note that we added just the charact u271f to the grammar for
    // TestIdentifier in order to achieve testability at the final character of
    // an input. 
    //
    // We exhaustively cover all chunk sizes from degenerate (1) to the full
    // string and various possible break configurations. (Note that this proved
    // especially useful and was not at first obvious: these tests passed during
    // the first dev iteration using chunk sizes of 1, but only after using
    // slightly larger chunk sizes was it discovered that there was a new bug
    // introduced in which valid chars produced in a chunk that ended in an
    // incomplete multibyte char were being dropped).
    private
    void
    assertSingleIdentifier( String s )
        throws Exception
    {
        for ( int i = 1; i < s.length(); ++i )
        {
            TestIdentifier lit = expectOneToken( s, TestIdentifier.class, i );
            state.equalString( s, lit );
        }
    }

    @Test
    private
    void
    test20()
        throws Exception
    {
        assertSingleIdentifier( "multibyte_char\u271fin_middle_of_input" );
    }

    @Test
    private
    void
    test21()
        throws Exception
    {
        assertSingleIdentifier( "ident_with_trailing_multibyte\u271f" );
    }

    @Test
    private
    void
    testTokenizerTakesLongestMatch()
        throws Exception
    {
        List< TestToken > toks = tokenize( "nully null" );

        // assert both the string value and the token type

        state.equalString( 
            "nully", state.cast( TestIdentifier.class, toks.get( 0 ) ) );

        state.cast( KwdNull.class, toks.get( 1 ) );
    }

    private
    void
    checkLocations( List< LocatedToken< TestToken > > locatedToks,
                    Object[] locAsserts )
    {
        state.equalInt( 0, locAsserts.length % 3 );
        state.equalInt( locAsserts.length / 3, locatedToks.size() );

        int i = 0;

        for ( LocatedToken< TestToken > lt : locatedToks )
        {
            int base = i * 3;

            state.isTrue(
                ( (Class< ? >) locAsserts[ base ] ).isInstance( lt.token ) );
            
            state.equalInt(
                (Integer) locAsserts[ base + 1 ], lt.loc.getLine() );

            state.equalInt(
                (Integer) locAsserts[ base + 2 ], lt.loc.getColumn() );

            state.equalString( BUILTIN_FILE_NAME, lt.loc.getFileName() );

            ++i;
        }
    }

    @Test
    private
    void
    testLexerTokenLocations()
        throws Exception
    {
        CharSequence text =
            Strings.join( "\n", "a = 1  + 2;", "", "a;", "big_ident = a;" );

        List< LocatedToken< TestToken > > locatedToks = lexToList( text );

        Object[] locAsserts = 
            new Object[]
            {
                TestIdentifier.class, 1, 1,
                WhitespaceToken.class, 1, 2,
                AssignmentOp.class, 1, 3,
                WhitespaceToken.class, 1, 4,
                TestIntegral.class, 1, 5,
                WhitespaceToken.class, 1, 6,
                Plus.class, 1, 8,
                WhitespaceToken.class, 1, 9,
                TestIntegral.class, 1, 10,
                Semicolon.class, 1, 11,
                WhitespaceToken.class, 1, 12,
                TestIdentifier.class, 3, 1,
                Semicolon.class, 3, 2,
                WhitespaceToken.class, 3, 3,
                TestIdentifier.class, 4, 1,
                WhitespaceToken.class, 4, 10,
                AssignmentOp.class, 4, 11,
                WhitespaceToken.class, 4, 12,
                TestIdentifier.class, 4, 13,
                Semicolon.class, 4, 14
            };

        checkLocations( locatedToks, locAsserts );
    }

    @Test
    private
    void
    testTokenLocationListener()
        throws Exception
    {
        ByteBuffer bb = Charsets.UTF_8.asByteBuffer( "7*9;" );

        LocationAccumulator< TestToken > acc = 
            new LocationAccumulator< TestToken >();
        
        TestValue res = evalScript( bb, BUILTIN_FILE_NAME, acc );
        state.equal( 63, ( (NumValue) res ).num );

        Object[] locAsserts =
        {
            TestIntegral.class, 1, 1,
            Asterisk.class, 1, 2,
            TestIntegral.class, 1, 3,
            Semicolon.class, 1, 4
        };

        checkLocations( acc.acc, locAsserts );
    }

    @Test
    private
    void
    testObjectInstance()
        throws Exception
    {
        Grammar.Builder< String, TestToken > b =
            Grammar.createBuilder( String.class );
        
        b.addProduction( "p1",
            b.sequence(
                b.objectInstance( new TestIdentifier( "reserved" ) ),
                b.terminalInstance( TestToken.class ),
                b.terminalInstance( Semicolon.class ) ) );
        
        Grammar< String, TestToken > g = b.build();

        List< TestToken > toks = tokenize( "reserved blah;" );
        ParseResult< String, TestToken > pr = getParseResult( g, "p1", toks );

        state.equal( ProductionMatcherState.MATCHED, pr.ms );
    }

    @Test
    private
    void
    testGetUnconsumedInput()
    {
        Grammar.LexBuilder< String > b =
            Grammar.createLexBuilder( String.class );

        b.addProduction( "p1", b.atLeastOne( b.ch( 'a' ) ) );
        
        RecursiveDescentParserFactory< String, Character > pf =
            RecursiveDescentParserFactory.forGrammar( b.build() );

        RecursiveDescentParser< String, Character > p = pf.createParser( "p1" );

        p.consumeTerminal( 'a' );
        state.isTrue( p.isMatching() );

        p.consumeTerminal( 'b' );
        state.isTrue( p.isMatched() );

        List< Character > unconsumed = p.getUnconsumedInput();
        state.equalInt( 1, unconsumed.size() );
        state.isTrue( 'b' == unconsumed.get( 0 ).charValue() );
    }

    @Test( expected = LeftRecursiveGrammarException.class )
    private
    void
    testImmediateLeftRecursionDetection()
    {
        Grammar.Builder< String, String > b =
            Grammar.createBuilder( String.class );

        b.addProduction( 
            "p1", b.sequence( b.derivation( "p1" ), b.derivation( "p2" ) ) );
        
        b.addProduction( "p2", b.terminalInstance( String.class ) );
        
        Grammar< String, String > g = b.build();

        RecursiveDescentParserFactory.forGrammar( g );
    }

    @Test( expected = LeftRecursiveGrammarException.class )
    private
    void
    testDerivedSequenceLeftRecursionDetection()
    {
        Grammar.Builder< String, String > b =
            Grammar.createBuilder( String.class );
        
        b.addProduction(
            "p1", b.sequence( b.derivation( "p2" ), b.derivation( "p3" ) ) );

        b.addProduction( 
            "p2", 
            b.sequence( b.derivation( "p3" ), b.derivation( "p4" ) ) );

        b.addProduction( 
            "p3", 
            b.sequence( b.derivation( "p1" ), b.derivation( "p4" ) ) );

        b.addProduction( "p4", b.terminalInstance( String.class ) );

        Grammar< String, String > g = b.build();

        RecursiveDescentParserFactory.forGrammar( g );
    }

    @Test( expected = LeftRecursiveGrammarException.class )
    private
    void
    testDerivedViaEpsilonProductionLeftRecursionDetection()
    {
        Grammar.Builder< String, String > b = 
            Grammar.createBuilder( String.class );

        b.addProduction( "p1", b.sequence( b.derivation( "p2" ) ) );

        b.addProduction( 
            "p2",
            b.sequence(
                b.unary( b.derivation( "p3" ) ),
                b.derivation( "p4" ) ) );
 
        b.addProduction( "p3", b.unary( b.derivation( "p4" ) ) );

        b.addProduction( "p4", b.derivation( "p1" ) );
        
        RecursiveDescentParserFactory.forGrammar( b.build() );
    }

    @Test( expected = LeftRecursiveGrammarException.class )
    private
    void
    testDerivedInLeadingUnaryProductionLeftRecursionDetected()
    {
        Grammar.Builder< String, String > b =
            Grammar.createBuilder( String.class );
        
        b.addProduction( "p1", b.derivation( "p2" ) );

        b.addProduction( 
            "p2", 
            b.sequence(
                b.unary( b.derivation( "p1" ) ),
                b.derivation( "p2" ) ) );
    
        RecursiveDescentParserFactory.forGrammar( b.build() );
    }

    @Test( expected = UnrecognizedNonTerminalException.class,
           expectedPattern = 
            ".*reference undefined non-terminal\\(s\\): p3, p2$" )
    private
    void
    testUnrecognizedNonTerminalException()
    {
        Grammar.Builder< String, String > b =
            Grammar.createBuilder( String.class );

        b.addProduction( "p1", b.sequence( b.unary( b.derivation( "p2" ) ) ) );
        b.addProduction( "p4", b.derivation( "p3" ) );

        b.build();
    }

    @Test
    private
    void
    testScript1()
        throws Exception
    {
        TestValue res = evalScriptFile( "test-script1" );

        CharSequence expct = 
            Strings.join( "|",
                "a string with\nnewlines",
                "a string with escaped quotes ('\"') and a backslash ('\\')",
                "a string with utf16 codepoint escapes: \u0008, \u01Df",
                "i1=7; i2=12; i3=19" );

        state.equalString( expct, ( (StringLiteral) res ).jstring );
    }

    private
    void
    testScriptFileParseError( 
        SourceTextLocation expctLoc,
        Class< ? extends AbstractLocatableSyntaxException > cls )
        throws Exception
    {
        try
        {
            evalScriptFile( expctLoc.getFileName() );
            state.fail();
        }
        catch ( Exception ex )
        {
            if ( cls.isInstance( ex ) )
            {
                AbstractLocatableSyntaxException se = cls.cast( ex );
                SourceTextLocation loc = se.getLocation();
                state.equalString( expctLoc.getFileName(), loc.getFileName() );
                state.equalInt( expctLoc.getLine(), loc.getLine() );
                state.equalInt( expctLoc.getColumn(), loc.getColumn() );
            }
            else throw ex;
        }
    }

    @Test
    private
    void
    testScript2()
        throws Exception
    {
        testScriptFileParseError(
            SourceTextLocation.create( "test-script2", 5, 31 ),
            InvalidTokenException.class );
    }

    @Test
    private
    void
    testScript3()
        throws Exception
    {
        testScriptFileParseError( 
            SourceTextLocation.create( "test-script3", 3, 12 ),
            InvalidSyntaxException.class );
    }

    @Test
    private
    void
    testExtractTerminals()
        throws Exception
    {
        List< TestToken > toks = tokenize( "a = b + 3 * 4" );
        ParseResult< Class, TestToken > pr = 
            getParseResult( SYNTACTIC_GRAMMAR_1, Expression.class, toks );

        List< TestIdentifier > idents = 
            Parsers.extractTerminals( pr.match, TestIdentifier.class );
        state.equalString( "a|b", Strings.join( "|", idents ) );

        List< TestIntegral > ints =
            Parsers.extractTerminals( pr.match, TestIntegral.class );
        state.equalString( "3|4", Strings.join( "|", ints ) );
    }

    @Test
    private
    void
    testExtractDerivations()
        throws Exception
    {
        DerivationMatch< Class, TestToken > script = 
            parseScript( "a = b + 2;" );
        
        List< DerivationMatch< Class, TestToken > > derivs =
            Parsers.extractDerivations( 
                script, (Class) Expression1Op.class, (Class) Factor1.class );
        
        state.equalInt( 4, derivs.size() );

        state.equal( Factor1.class, derivs.get( 0 ).getHead() );
        state.equal( Factor1.class, derivs.get( 1 ).getHead() );
        state.equal( Expression1Op.class, derivs.get( 2 ).getHead() );
        state.equal( Factor1.class, derivs.get( 3 ).getHead() );
    }
 
    private
    TestIdentifier
    runParsersMatchTest( CharSequence str,
                         boolean callParse )
        throws SyntaxException
    {
        TestIdentifier expct = new TestIdentifier( "test_identifier" );

        DerivationMatch< Class, Character > dm =
            callParse ?
                Parsers.parseStringMatch(
                    LEXICAL_GRAMMAR_1_PF, TestIdentifier.class, str ) :
                Parsers.createStringMatch(
                    LEXICAL_GRAMMAR_1_PF, TestIdentifier.class, str );
 
        return (TestIdentifier) buildStringToken( dm );
    }

    @Test
    private
    void
    testParsersCreateMatch()
        throws Exception
    {
        state.equalString( 
            "test_identifier", 
            runParsersMatchTest( "test_identifier", false ) );
    }

    @Test
    private
    void
    testParsersParseMatch()
        throws Exception
    {
        state.equalString( 
            "test_identifier", 
            runParsersMatchTest( "test_identifier", true ) );
    }

    @Test( expected = IllegalArgumentException.class,
           expectedPattern = 
            "^Trailing characters \\(match stopped at index 10\\): " +
            "test_ident\\+ifier$" )
    private
    void
    testParsersCreateMatchTrailingChars()
        throws Exception
    {
        runParsersMatchTest( "test_ident+ifier", false );
    }

    @Test( expected = SyntaxException.class,
           expectedPattern = 
            "^Trailing characters \\(match stopped at index 10\\): " +
            "test_ident\\+ifier$" )
    private
    void
    testParsersParseMatchTrailingChars()
        throws Exception
    {
        runParsersMatchTest( "test_ident+ifier", true );
    }

    @Test( expected = IllegalArgumentException.class,
           expectedPattern = "^Invalid syntax: \\+$" )
    private
    void
    testParsersCreateMatchUnmatched()
        throws Exception
    {
        runParsersMatchTest( "+", false );
    }

    @Test( expected = SyntaxException.class,
           expectedPattern = "^Invalid syntax: \\+$" )
    private
    void
    testParsersParseMatchUnmatched()
        throws Exception
    {
        runParsersMatchTest( "+", true );
    }

    // Build the lexical grammar
    static
    {
        Grammar.LexBuilder< Class > b = Grammar.createLexBuilder( Class.class );
 
        b.addProduction( 
            TestIdentifier.class,
            b.nonEmptyList(
                b.derivation( TestIdentifierPiece.class ), b.ch( '_' ) ) );

        b.addProduction(
            TestIdentifierPiece.class,
            b.sequence(
                b.charRange( 'a', 'z' ),
                b.kleene(
                    b.union(
                        b.charRange( 'a', 'z' ),
                        b.ch( '\u271f' ),
                        b.charRange( '0', '9' ) ) ) ) );

        b.addProduction(
            TestIntegral.class, b.atLeastOne( b.charRange( '0', '9' ) ) );

        b.addProduction(
            WhitespaceToken.class,
            b.atLeastOne( b.charSet( ' ', '\t', '\n', '\r' ) ) );

        b.addProduction(
            StringLiteral.class,
            b.sequence(
                b.ch( '"' ),
                b.kleene( b.derivation( TestChar.class ) ),
                b.ch( '"' ) ) );
 
        b.addProduction(
            TestChar.class,
            b.union(
                b.derivation( MnemonicCharEscape.class ),
                b.derivation( Utf16CodeUnitEscape.class ),
                b.derivation( UnescapedChar.class ) ) );
 
        b.addProduction(
            MnemonicCharEscape.class,
            b.sequence(
                b.ch( '\\' ),
                b.union( b.charSet( 't', 'n', 'r', 'f', '"', '\\' ) ) ) );
        
        b.addProduction(
            Utf16CodeUnitEscape.class,
                b.sequence(
                    b.ch( '\\' ),
                    b.ch( 'u' ),
                    b.exactly(
                        b.union(
                            b.charRange( 'a', 'f' ),
                            b.charRange( 'A', 'F' ),
                            b.charRange( '0', '9' ) ),
                        4 ) ) );
 
        b.addProduction(
            UnescapedChar.class, b.charSetComplement( '"', '\\' ) );
 
        b.addProduction(
            TestScriptComment.class,
                b.sequence(
                    b.ch( '#' ),
                    b.kleene( b.charSetComplement( '\n', '\r' ) ) ) );
 
        b.addProduction( OpenBrace.class, b.ch( '{' ) );
        b.addProduction( CloseBrace.class, b.ch( '}' ) );
        b.addProduction( OpenBracket.class, b.ch( '[' ) );
        b.addProduction( CloseBracket.class, b.ch( ']' ) );
        b.addProduction( OpenParen.class, b.ch( '(' ) );
        b.addProduction( CloseParen.class, b.ch( ')' ) );
        b.addProduction( Comma.class, b.ch( ',' ) );
        b.addProduction( Semicolon.class, b.ch( ';' ) );
        b.addProduction( Plus.class, b.ch( '+' ) );
        b.addProduction( Minus.class, b.ch( '-' ) );
        b.addProduction( Asterisk.class, b.ch( '*' ) );
        b.addProduction( ForwardSlash.class, b.ch( '/' ) );
        b.addProduction( AssignmentOp.class, b.ch( '=' ) );
        b.addProduction( KwdNull.class, b.string( "null" ) );

        b.addProduction( 
            LiteralToken.class,
            b.unionLongest(
                b.derivation( OpenBrace.class ),
                b.derivation( CloseBrace.class ),
                b.derivation( OpenBracket.class ),
                b.derivation( CloseBracket.class ),
                b.derivation( OpenParen.class ),
                b.derivation( CloseParen.class ),
                b.derivation( Comma.class ),
                b.derivation( Semicolon.class ),
                b.derivation( Plus.class ),
                b.derivation( Minus.class ),
                b.derivation( Asterisk.class ),
                b.derivation( ForwardSlash.class ),
                b.derivation( AssignmentOp.class ),
                b.derivation( KwdNull.class )
            )
        );
 
        // Note: literals need to precede anything else
        b.addProduction( 
            TestToken.class,
            b.unionLongest(
                b.derivation( LiteralToken.class ),
                b.derivation( TestIntegral.class ), 
                b.derivation( WhitespaceToken.class ),
                b.derivation( TestIdentifier.class ),
                b.derivation( StringLiteral.class ),
                b.derivation( TestScriptComment.class ) ) );

        LEXICAL_GRAMMAR_1 = b.build();
        
        LEXICAL_GRAMMAR_1_PF =
            RecursiveDescentParserFactory.forGrammar( LEXICAL_GRAMMAR_1 );

    }

    // Build the syntactic grammar
    static
    {
        Grammar.Builder< Class, TestToken > b = 
            Grammar.createBuilder( Class.class );

        b.addProduction(
            Expression.class,
            b.sequence(
                b.derivation( Expression1.class ),
                b.unary( 
                    b.sequence(
                        b.terminalInstance( AssignmentOp.class ),
                        b.derivation( Expression1.class ) ) ) ) );
 
        b.addProduction(
            Expression1.class, 
            b.sequence(
                b.derivation( Expression2.class ),
                b.unary( b.derivation( Expression1Rest.class ) ) ) );
 
        b.addProduction(
            Expression1Rest.class,
            b.sequence(
                b.derivation( Expression1Op.class ),
                b.derivation( Expression2.class ),
                b.unary( b.derivation( Expression1Rest.class ) ) ) );
 
        b.addProduction( 
            Expression1Op.class, 
            b.union(
                b.terminalInstance( Plus.class ),
                b.terminalInstance( Minus.class ) ) );

        b.addProduction(
            Expression2.class,
            b.sequence(
                b.derivation( Factor1.class ),
                b.unary( b.derivation( Expression2Rest.class ) ) ) );
 
        b.addProduction(
            Expression2Rest.class,
            b.sequence(
                b.union(
                    b.terminalInstance( Asterisk.class ),
                    b.terminalInstance( ForwardSlash.class ) ),
                b.derivation( Factor1.class ),
                b.unary( b.derivation( Expression2Rest.class ) ) ) );

        b.addProduction( 
            Factor1.class, 
            b.union(
                b.terminalInstance( TestIdentifier.class ),
                b.terminalInstance( TestIntegral.class ),
                b.terminalInstance( StringLiteral.class ),
                b.derivation( ListExpression.class ),
                b.sequence(
                    b.terminalInstance( OpenParen.class ),
                    b.derivation( Expression1.class ),
                    b.terminalInstance( CloseParen.class ) ) ) );
 
        b.addProduction(
            ListExpression.class,
            b.sequence(
                b.terminalInstance( OpenBracket.class ),
                b.list(
                    b.derivation( Expression1.class ),
                    b.terminalInstance( Comma.class ) ),
                b.terminalInstance( CloseBracket.class ) ) );
 
        b.addProduction( 
            ScriptBody.class, 
            b.atLeastOne( b.derivation( Statement.class ) ) );

        b.addProduction( 
            Statement.class, b.derivation( ExpressionStatement.class ) );

        b.addProduction(
            ExpressionStatement.class,
            b.sequence(
                b.derivation( Expression.class ),
                b.terminalInstance( Semicolon.class ) )
        );

        SYNTACTIC_GRAMMAR_1 = b.build();
    }

    private
    final
    static
    class UnconsumedInputException
    extends Exception
    {}
}
