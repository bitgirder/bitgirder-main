package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Strings;

import com.bitgirder.log.CodeLoggers;

import java.nio.ByteBuffer;

import java.nio.charset.Charset;
import java.nio.charset.CharacterCodingException;

public
final
class DocumentParser< D >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final ParseChain< ?, ?, ?, D > pc;

    private
    DocumentParser( ParseChain< ?, ?, ?, D > pc )
    {
        this.pc = pc;
    }

    public
    void
    update( ByteBuffer bb,
            boolean endOfInput )
        throws CharacterCodingException,
               SyntaxException
    {
        inputs.notNull( bb, "bb" );
        pc.update( bb, endOfInput );
    }

    public D buildSyntax() { return pc.buildSyntax(); }

    private
    final
    static
    class ParseChain< N, L, T, D >
    {
        private final Lexer< L, T > lexer;
        private final SyntaxParserFeeder< N, T, D > spf;

        private
        ParseChain( Lexer< L, T > lexer,
                    SyntaxParserFeeder< N, T, D > spf )
        {
            this.lexer = lexer;
            this.spf = spf;
        }

        private
        void
        update( ByteBuffer bb,
                boolean endOfInput )
            throws CharacterCodingException,
                   SyntaxException
        {
            lexer.update( bb, endOfInput );

            if ( spf.errLoc != null )
            {
                throw new InvalidSyntaxException( spf.errLoc );
            }
        }

        private
        D
        buildSyntax()
        {
            return spf.syntaxBuilder.buildSyntax( spf.parser.getMatch() );
        }
    }

    private
    final
    static
    class SyntaxParserFeeder< N, T, D >
    implements Lexer.Listener< T >
    {
        private final RecursiveDescentParser< N, T > parser;
        private final SyntaxBuilder< N, T, D > syntaxBuilder;
        private final TokenLocationListener< T > locListener;
        private final FeedFilter< T > feedFilter;
        private final boolean logSyntaxTokens;
        private final boolean logSyntaxParser;
        
        private int fedTokens;
        private SourceTextLocation lastLoc;

        private SourceTextLocation errLoc;

        private
        SyntaxParserFeeder( Builder< N, ?, T, D > b )
        {
            this.parser =
                b.dpFact.getSyntacticParserFactory().createParserBuilder().
                    setGoal( b.dpFact.getSyntacticGoal() ).
                    setLogMatchEvents( b.logSyntaxMatchEvents ).
                    setLogTerminalConsume( b.logSyntaxTerminalConsume ).
                    build();

            this.syntaxBuilder =
                inputs.notNull( b.syntaxBuilder, "syntaxBuilder" );

            this.locListener = b.locListener;
            this.feedFilter = b.feedFilter;

            this.logSyntaxTokens = b.logSyntaxTokens;
            this.logSyntaxParser = b.logSyntaxParser;
        }

        private
        void
        logPreFeed( T token )
        {
            if ( logSyntaxTokens )
            {
                code( "Feeding token to syntax parser:", token );
            }

            if ( logSyntaxParser )
            {
                code( "Syntax parser:", Strings.inspect( parser ) );
            }
        }

        // If the parser matched but not by consuming all of the input it was
        // given, we register it as a failure occurring at the last token
        // location. This is a common way of failing, since many grammars will
        // consist of a goal which is some type of Kleene match, and an invalid
        // document may contain a valid number of matches preceding the parse
        // error, allowing the underlying parser to match, but still rendering
        // the document invalid. 
        //
        // An example would be a scripting language grammar which allows a
        // script to consist of any number of valid statements. A script might
        // have 2 valid statements followed by an invalid one. The parser will
        // correctly register a match stopping after the first 2 statements, but
        // the overall document does not parse, due to the offending token in
        // the invalid 3rd statement. That will be the token that causes the
        // parser to give up trying to match the 3rd statement but successfully
        // match the first 2, and so we register that last token as the
        // error-reporting location
        //
        private
        void
        parserMatched()
        {
            if ( parser.getConsumedTerminals() != fedTokens ) errLoc = lastLoc;
        }

        private
        void
        feedToken( T token,
                   SourceTextLocation loc )
        {
            logPreFeed( token );

            parser.consumeTerminal( token );
            ++fedTokens;
            lastLoc = loc;

            switch ( parser.getMatcherState() )
            {
                case MATCHING: break;
                case MATCHED: parserMatched(); break;
                case UNMATCHED: errLoc = loc; break;
            }
        }

        public
        void
        lexerTokenized( T token,
                        SourceTextLocation start )
        {
            if ( locListener != null ) locListener.markToken( token, start );

            if ( errLoc == null )
            {
                if ( feedFilter == null || feedFilter.shouldFeed( token ) )
                {
                    feedToken( token, start );
                }
            }
        }

        public
        void
        lexerComplete()
            throws SyntaxException
        {
            if ( parser.isMatching() ) parser.complete();

            switch ( parser.getMatcherState() )
            {
                case MATCHED: break;

                case UNMATCHED: 
                    throw new SyntaxException( "Unmatched document" );
                
                case MATCHING: throw new PrematureEndOfInputException();
            }
        }
    }

    private
    static
    < L, T >
    Lexer< L, T >
    createLexer( Lexer.Listener< T > listener,
                 Builder< ?, L, T, ? > b )
    {
        return
            new Lexer.Builder< L, T >().
                setFileName( b.fileName ).
                setCharset( b.charset ).
                setParserFactory( b.dpFact.getLexicalParserFactory() ).
                setListener( listener ).
                setGoal( b.dpFact.getLexicalGoal() ).
                setTokenBuilder( b.dpFact.getTokenBuilder() ).
                setLogFedChars( b.logChars ).
                setLogParser( b.logLexerParser ).
                build();
    }

    private
    static
    < N, L, T, D >
    DocumentParser< D >
    build( Builder< N, L, T, D > b )
    {
        inputs.notNull( b.charset, "charset" );
        inputs.notNull( b.fileName, "fileName" );
        inputs.notNull( b.syntaxBuilder, "syntaxBuilder" );

        SyntaxParserFeeder< N, T, D > spf = 
            new SyntaxParserFeeder< N, T, D >( b );

        Lexer< L, T > lexer = createLexer( spf, b );

        ParseChain< N, L, T, D > pc = 
            new ParseChain< N, L, T, D >( lexer, spf );
 
        return new DocumentParser< D >( pc );
    }

    public
    static
    interface TokenLocationListener< T >
    {
        public
        void
        markToken( T token,
                   SourceTextLocation start );
    }

    public
    static
    interface FeedFilter< T >
    {
        public
        boolean
        shouldFeed( T token );
    }

    public
    final
    static
    class DebugOptions
    {
        public final static DebugOptions VERBOSE_SYNTAX =
            new Builder().
                setLogSyntaxTokens( true ).
                setLogSyntaxParser( true ).
                setLogSyntaxMatchEvents( true ).
                setLogSyntaxTerminalConsume( true ).
                build();

        private final boolean logChars;
        private final boolean logLexerParser;
        private final boolean logSyntaxTokens;
        private final boolean logSyntaxParser;
        private final boolean logSyntaxMatchEvents;
        private final boolean logSyntaxTerminalConsume;

        private 
        DebugOptions( DebugOptions.Builder b )
        {
            this.logChars = b.logChars;
            this.logLexerParser = b.logLexerParser;
            this.logSyntaxTokens = b.logSyntaxTokens;
            this.logSyntaxParser = b.logSyntaxParser;
            this.logSyntaxMatchEvents = b.logSyntaxMatchEvents;
            this.logSyntaxTerminalConsume = b.logSyntaxTerminalConsume;
        }

        public
        final
        static
        class Builder
        {
            private boolean logChars;
            private boolean logLexerParser;
            private boolean logSyntaxTokens;
            private boolean logSyntaxParser;
            private boolean logSyntaxMatchEvents;
            private boolean logSyntaxTerminalConsume;
    
            public
            Builder
            setLogChars( boolean logChars )
            {
                this.logChars = logChars;
                return this;
            }
    
            public
            Builder
            setLogLexerParser( boolean logLexerParser )
            {
                this.logLexerParser = logLexerParser;
                return this;
            }
    
            public
            Builder
            setLogSyntaxTokens( boolean logSyntaxTokens )
            {
                this.logSyntaxTokens = logSyntaxTokens;
                return this;
            }
    
            public
            Builder
            setLogSyntaxParser( boolean logSyntaxParser )
            {
                this.logSyntaxParser = logSyntaxParser;
                return this;
            }
    
            public
            Builder
            setLogSyntaxMatchEvents( boolean logSyntaxMatchEvents )
            {
                this.logSyntaxMatchEvents = logSyntaxMatchEvents;
                return this;
            }

            public
            Builder
            setLogSyntaxTerminalConsume( boolean logSyntaxTerminalConsume )
            {
                this.logSyntaxTerminalConsume = logSyntaxTerminalConsume;
                return this;
            }
            
            public
            DebugOptions
            build()
            {
                return new DebugOptions( this );
            }
        }
    }

    public
    static
    final
    class Builder< N, L, T, D >
    {
        private final DocumentParserFactory< N, L, T > dpFact;

        private CharSequence fileName;
        private Charset charset;
        private SyntaxBuilder< N, T, D > syntaxBuilder;
        private TokenLocationListener< T > locListener;
        private FeedFilter< T > feedFilter;
        private boolean logChars;
        private boolean logLexerParser;
        private boolean logSyntaxTokens;
        private boolean logSyntaxParser;
        private boolean logSyntaxMatchEvents;
        private boolean logSyntaxTerminalConsume;

        Builder( DocumentParserFactory< N, L, T > dpFact )
        {
            this.dpFact = dpFact;
        }

        public
        Builder< N, L, T, D >
        setFileName( CharSequence fileName )
        {
            this.fileName = inputs.notNull( fileName, "fileName" );
            return this;
        }

        public
        Builder< N, L, T, D >
        setCharset( Charset charset )
        {
            this.charset = inputs.notNull( charset, "charset" );
            return this;
        }

        public
        Builder< N, L, T, D >
        setSyntaxBuilder( SyntaxBuilder< N, T, D > syntaxBuilder )
        {
            this.syntaxBuilder = 
                inputs.notNull( syntaxBuilder, "syntaxBuilder" );

            return this;
        }

        public
        Builder< N, L, T, D >
        setTokenLocationListener( TokenLocationListener< T > locListener )
        {
            this.locListener = inputs.notNull( locListener, "locListener" );
            return this;
        }

        public
        Builder< N, L, T, D >
        setFeedFilter( FeedFilter< T > feedFilter )
        {
            this.feedFilter = inputs.notNull( feedFilter, "feedFilter" );
            return this;
        }

        public
        Builder< N, L, T, D >
        setLogChars( boolean logChars )
        {
            this.logChars = logChars;
            return this;
        }

        public
        Builder< N, L, T, D >
        setLogLexerParser( boolean logLexerParser )
        {
            this.logLexerParser = logLexerParser;
            return this;
        }

        public
        Builder< N, L, T, D >
        setLogSyntaxTokens( boolean logSyntaxTokens )
        {
            this.logSyntaxTokens = logSyntaxTokens;
            return this;
        }

        public
        Builder< N, L, T, D >
        setLogSyntaxParser( boolean logSyntaxParser )
        {
            this.logSyntaxParser = logSyntaxParser;
            return this;
        }

        public
        Builder< N, L, T, D >
        setLogSyntaxMatchEvents( boolean logSyntaxMatchEvents )
        {
            this.logSyntaxMatchEvents = logSyntaxMatchEvents;
            return this;
        }

        public
        Builder< N, L, T, D >
        setDebugOptions( DebugOptions dbgOpts )
        {
            inputs.notNull( dbgOpts, "dbgOpts" );

            this.logChars = dbgOpts.logChars;
            this.logLexerParser = dbgOpts.logLexerParser;
            this.logSyntaxTokens = dbgOpts.logSyntaxTokens;
            this.logSyntaxParser = dbgOpts.logSyntaxParser;
            this.logSyntaxMatchEvents = dbgOpts.logSyntaxMatchEvents;
            this.logSyntaxTerminalConsume = dbgOpts.logSyntaxTerminalConsume;

            return this;
        }

        public
        DocumentParser< D >
        build()
        {
            return DocumentParser.< N, L, T, D >build( this );
        }
    }
}
