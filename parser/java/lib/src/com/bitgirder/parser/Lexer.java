package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Strings;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.io.IoUtils;

import java.util.ArrayDeque;
import java.util.Iterator;

import java.nio.ByteBuffer;
import java.nio.CharBuffer;

import java.nio.charset.Charset;
import java.nio.charset.CharsetDecoder;
import java.nio.charset.CharacterCodingException;
import java.nio.charset.CoderResult;

public
final
class Lexer< N, T >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static int DEFAULT_CHAR_BUF_SIZE = 1024;

    private final CharSequence fileName;
    private final TokenBuilder< N, T > tokenBuilder;
    private final Listener< T > listener;
    private final RecursiveDescentParser< N, Character > parser;
    private final CharsetDecoder dec;
    private final int charBufSize;

    private final boolean logFedChars;
    private final boolean logParser;

    private final ArrayDeque< LineMarker > lineMarkers;
    private boolean followsNewline;
    private int line;
    private SourceTextLocation errStartLoc;

    // accumulator for multibyte chars which are broken between calls to
    // update(); always positioned for read
    private ByteBuffer mbAcc;

    // Number of chars fed since the last completed match
    private int fedChars = 0;

    // Number of chars fed since this lexer began
    private long streamPos = 0;

    // Start position in stream of currently executing match
    private long matchStart = 0;

    // Start position in stream of the next char buffer to be decoded
    private int nextStartIndx = 0;

    private final ArrayDeque< CharBlock > toFeed = 
        new ArrayDeque< CharBlock >();

    private final ArrayDeque< CharBlock > rewind =
        new ArrayDeque< CharBlock >();

    private
    Lexer( Builder< N, T > b )
    {
        this.fileName = inputs.notNull( b.fileName, "fileName" );
        this.tokenBuilder = inputs.notNull( b.tokenBuilder, "tokenBuilder" );
        this.listener = inputs.notNull( b.listener, "listener" );

        this.parser = 
            inputs.notNull( b.parserFact, "parserFact" ).
                   createParser( inputs.notNull( b.goal, "goal" ) );

        this.dec = inputs.notNull( b.charset, "charset" ).newDecoder();
        this.charBufSize = b.charBufSize;

        this.logFedChars = b.logFedChars;
        this.logParser = b.logParser;

        line = 1;
        lineMarkers = new ArrayDeque< LineMarker >( 2 );
        lineMarkers.addLast( new LineMarker( line, 0 ) );
        errStartLoc = SourceTextLocation.create( fileName, line, 1 );
    }

    private
    CharBuffer
    getCharBuffer( int sz )
    {
        return CharBuffer.allocate( sz );
    }

    private void releaseCharBuffer( CharBuffer cb ) {}

    private
    final
    static
    class CharBlock
    {
        private final CharBuffer cb;
        private final long startIndx;

        private
        CharBlock( CharBuffer cb,
                   long startIndx )
        {
            this.cb = cb;
            this.startIndx = startIndx;
        }
    }

    private
    final
    static
    class LineMarker
    {
        private final int line;
        private final long streamIndx;

        private
        LineMarker( int line,
                    long streamIndx )
        {
            this.line = line;
            this.streamIndx = streamIndx;
        }

        @Override
        public
        String
        toString()
        {
            return Strings.inspect( this, true, 
                "line", line, "streamIndx", streamIndx ).toString();
        }
    }

    private
    boolean
    shouldProcessAdvance( CharBlock block )
    {
        long nextPos = block.startIndx + block.cb.position();

        if ( nextPos <= streamPos ) return false;
        else
        {
            state.isTrue( 
                nextPos == streamPos + 1,
                "Attempt to advance char at logical position", nextPos, 
                "beyond longest stream position", streamPos );
            
            return true;
        }
    }

    private
    char
    advanceChar( CharBlock block )
    {
        if ( shouldProcessAdvance( block ) )
        {
            ++streamPos;

            if ( followsNewline )
            {
                ++line;
    
                long streamIndx = block.startIndx + block.cb.position();
                LineMarker lm = new LineMarker( line, streamIndx );
                lineMarkers.addLast( lm );
    
                followsNewline = false;
            }
        }
 
        char res = block.cb.get();
        if ( res == '\n' ) followsNewline = true;

        return res;
    }

    private
    LineMarker
    getPredecessorFor( long streamIndx,
                       boolean canPurge )
    {
        Iterator< LineMarker > it = lineMarkers.iterator(); 

        LineMarker pred = it.next();
        state.isTrue( streamIndx >= pred.streamIndx );

        int purgeCount = 0; // only used if canPurge
        LineMarker succ = null; // not used except as a loop sentinel

        while ( it.hasNext() && succ == null )
        {
            LineMarker lm = it.next();
            if ( streamIndx >= lm.streamIndx ) 
            {
                pred = lm; 
                if ( canPurge ) ++purgeCount;
            }
            else succ = lm;
        }

        while ( canPurge && purgeCount-- > 0 ) lineMarkers.removeFirst();

        return pred;
    }

    private
    SourceTextLocation
    toLocation( long streamIndx,
                boolean canPurge )
    {
        LineMarker pred = getPredecessorFor( streamIndx, canPurge );

        long col = 1 + ( streamIndx - pred.streamIndx );

        state.isTrue(
            col <= Integer.MAX_VALUE, "Too-large column value:", col );

        return SourceTextLocation.create( fileName, pred.line, (int) col );
    }

    private
    SourceTextLocation
    getErrLoc()
    {
        return toLocation( matchStart, false );
    }

    private
    void
    feedMatch( DerivationMatch< N, Character > match )
    {
        T token = tokenBuilder.buildToken( match );

        SourceTextLocation tokenLoc = toLocation( matchStart, true );

        listener.lexerTokenized( token, tokenLoc );
    }

    private
    void
    logPreFeed( char ch )
    {
        if ( logFedChars ) code( "Feeding char:", ch );

        if ( logParser ) 
        {
            code( 
                "Lexer parser before consumeTerminal:",
                Strings.inspect( parser ) );
        }
    }

    // returns true if rewindTo was found and set in block
    private
    boolean
    setRewind( CharBlock block,
               long rewindTo )
    {
        state.isTrue( rewindTo >= block.startIndx );
        long newPos = rewindTo - block.startIndx;

        // In all imaginable cases, we will enter the if branch below, but for
        // completeness's sake, we still need to somehow account for the logical
        // case in which the parser matched but only after consuming and finally
        // backing off of more than Integer.MAX_VALUE characters.
        if ( newPos <= Integer.MAX_VALUE )
        {
            int newPosI = (int) newPos;

            if ( newPosI <= block.cb.limit() ) 
            {
                block.cb.position( newPosI );
                return true;
            }
            else return false;
        }
        else return false;
    }

    private
    void
    setRewind( long rewindTo )
    {
        boolean rewound = false;

        // first check whether rewind point is in head of toFeed
        if ( ! toFeed.isEmpty() ) 
        {
            rewound = setRewind( toFeed.getFirst(), rewindTo );
        }

        // if not, process rewinds until rewind point is found
        while ( ( ! rewind.isEmpty() ) && ( ! rewound ) )
        {
            CharBlock block = rewind.getLast();
            rewound = setRewind( block, rewindTo );

            if ( block.cb.hasRemaining() )
            {
                rewind.removeLast();
                toFeed.addFirst( block );
            }
        }

        state.isTrue(
            rewound,
            "Attempt to rewind to position", rewindTo, "beyond rewind limit" );
    }

    private
    void
    parserMatched()
    {
        // important that this comes before matchStart is updated below
        feedMatch( parser.getMatch() );

        int consumed = parser.getConsumedTerminals();
        long rewindTo = matchStart + consumed;

        setRewind( rewindTo );

        matchStart = rewindTo;
 
        // we won't rewind past where we are now so release char buffers
        for ( CharBlock block : rewind ) releaseCharBuffer( block.cb );

        parser.reset();
        fedChars = 0;
    }

    private
    void
    feedOneChar()
    {
        CharBlock block = toFeed.getFirst();
        char ch = advanceChar( block );

        logPreFeed( ch );
        parser.consumeTerminal( Character.valueOf( ch ) );
        ++fedChars;
        
        if ( ! block.cb.hasRemaining() ) rewind.addLast( toFeed.removeFirst() );
    }

    private
    void
    feedInput()
        throws SyntaxException
    {
        while ( ! toFeed.isEmpty() )
        {
            feedOneChar();
    
            switch ( parser.getMatcherState() )
            {
                case UNMATCHED: throw new InvalidTokenException( getErrLoc() );
                case MATCHING: break;
                case MATCHED: parserMatched();
            }
    
            // validate postcondition
            state.isTrue( parser.isMatching() );
        }
    }

    private
    void
    complete()
        throws SyntaxException
    {
        if ( ! toFeed.isEmpty() ) feedInput();
        state.isTrue( toFeed.isEmpty() );

        if ( fedChars == 0 ) listener.lexerComplete();
        else
        {
            if ( parser.isMatching() ) parser.complete();
    
            switch ( parser.getMatcherState() )
            {
                case MATCHED: 
                    parserMatched();
                    listener.lexerComplete(); 
                    break;

                case MATCHING: throw new PrematureEndOfInputException();
                case UNMATCHED: throw new InvalidTokenException( getErrLoc() );
            }
        }
    }

    private
    void
    initMbAcc( ByteBuffer bb )
    {
        state.isTrue( mbAcc == null );

        // heuristic: 6 bytes will be sufficient for most multibyte chars we'll
        // handle here
        mbAcc = ByteBuffer.allocate( Math.max( bb.remaining(), 6 ) );

        mbAcc.put( bb );
        mbAcc.flip(); // ensure read-positioning
    }

    private
    void
    addToMbAcc( ByteBuffer bb )
    {
        if ( mbAcc.limit() == mbAcc.capacity() ) 
        {
            mbAcc = IoUtils.expand( mbAcc );
            mbAcc.flip(); // set back to read
        }
        
        mbAcc.limit( mbAcc.limit() + 1 );
        mbAcc.put( mbAcc.limit() - 1, bb.get() );
    }

    private
    void
    processMbChar( boolean endOfInput )
        throws CharacterCodingException,
               SyntaxException
    {
        CharBuffer cb = CharBuffer.allocate( 1 );
        CoderResult cr = dec.decode( mbAcc, cb, endOfInput );

        if ( cr.isError() ) cr.throwException();
        else if ( cr.isOverflow() ) state.fail( "Unexpected overflow" );
        else
        {
            state.isTrue( cr.isUnderflow() );

            if ( cb.position() == 1 )
            {
                mbAcc = null;

                if ( endOfInput ) flushAndComplete( cb );
                else flipAddAndFeed( cb );
            }
            // else: still accumulating the char; calling loop will continue
        }
    }

    // on exit either mbAcc will be null or bb will be empty
    private
    void
    feedMultiByteAcc( ByteBuffer bb,
                      boolean endOfInput )
        throws CharacterCodingException,
               SyntaxException
    {
        while ( mbAcc != null && bb.hasRemaining() )
        {
            addToMbAcc( bb );
            processMbChar( endOfInput && ! bb.hasRemaining() );
        }
        
        // assert exit condition
        state.isTrue( mbAcc == null || ( ! bb.hasRemaining() ) );
    }

    private
    void
    flipAddAndFeed( CharBuffer cb )
        throws SyntaxException
    {
        cb.flip();

        toFeed.addLast( new CharBlock( cb, nextStartIndx ) );
        nextStartIndx += cb.remaining();

        feedInput();
    }

    private
    void
    flushAndComplete( CharBuffer cb )
        throws CharacterCodingException,
               SyntaxException
    {
        CoderResult cr;

        do
        {
            int nextSz = cb.capacity() * 2; // in case we have to flush more
            cr = dec.flush( cb );

            if ( cr.isError() ) cr.throwException();
            else
            {
                flipAddAndFeed( cb );

                if ( cr.isOverflow() ) cb = getCharBuffer( nextSz );
            }
        }
        while ( cr.isOverflow() );

        complete();
    }

    private
    void
    handleUpdateUnderflow( ByteBuffer bb,
                           CharBuffer cb,
                           boolean endOfInput )
        throws CharacterCodingException,
               SyntaxException
    {
        if ( bb.hasRemaining() )
        {
            // we may still have produced chars before stalling on the split
            // multi-byte char, so feed them if so
            if ( cb.position() > 0 ) flipAddAndFeed( cb );

            state.isFalse( endOfInput );
            initMbAcc( bb );
        }
        else
        {
            if ( endOfInput ) flushAndComplete( cb );
            else flipAddAndFeed( cb );
        }
    }

    public
    void
    update( ByteBuffer bb,
            boolean endOfInput )
        throws CharacterCodingException,
               SyntaxException
    {
        inputs.notNull( bb, "bb" );

        feedMultiByteAcc( bb, endOfInput );

        while ( bb.hasRemaining() )
        {
            CharBuffer cb = getCharBuffer( charBufSize );
            CoderResult cr = dec.decode( bb, cb, endOfInput );

            if ( cr.isError() ) cr.throwException();
            else if ( cr.isOverflow() ) flipAddAndFeed( cb );
            else
            {
                state.isTrue( cr.isUnderflow() );
                handleUpdateUnderflow( bb, cb, endOfInput );
            }
        }
    }

    public
    final
    static
    class Builder< N, T >
    {
        private CharSequence fileName;
        private RecursiveDescentParserFactory< N, Character > parserFact;
        private Charset charset;
        private TokenBuilder< N, T > tokenBuilder;
        private Listener< T > listener;
        private N goal;

        // hardcoded for now; later we'll likely add a setter
        private int charBufSize = DEFAULT_CHAR_BUF_SIZE;

        private boolean logFedChars;
        private boolean logParser;

        public
        Builder< N, T >
        setFileName( CharSequence fileName )
        {
            this.fileName = inputs.notNull( fileName, "fileName" );
            return this;
        }

        public
        Builder< N, T >
        setParserFactory( 
            RecursiveDescentParserFactory< N, Character > parserFact )
        {
            this.parserFact = inputs.notNull( parserFact, "parserFact" );
            return this;
        }

        public
        Builder< N, T >
        setCharset( Charset charset )
        {
            this.charset = inputs.notNull( charset, "charset" );
            return this;
        }

        public
        Builder< N, T >
        setTokenBuilder( TokenBuilder< N, T > tokenBuilder )
        {
            this.tokenBuilder = inputs.notNull( tokenBuilder, "tokenBuilder" );
            return this;
        }

        public
        Builder< N, T >
        setListener( Listener< T > listener )
        {
            this.listener = inputs.notNull( listener, "listener" );
            return this;
        }

        public
        Builder< N, T >
        setGoal( N goal )
        {
            this.goal = inputs.notNull( goal, "goal" );
            return this;
        }

        public
        Builder< N, T >
        setLogFedChars( boolean logFedChars )
        {
            this.logFedChars = logFedChars;
            return this;
        }

        public
        Builder< N, T >
        setLogParser( boolean logParser )
        {
            this.logParser = logParser;
            return this;
        }

        public
        Lexer< N, T >
        build()
        {
            return new Lexer< N, T >( this );
        }
    }

    public
    static
    interface Listener< T >
    {
        public
        void
        lexerTokenized( T token,
                        SourceTextLocation start );
 
        // Only called when lexer has consumed and matched all input
        public
        void
        lexerComplete()
            throws SyntaxException;
    }

    public
    static
    interface TokenBuilder< N, T >
    {
        public
        T
        buildToken( DerivationMatch< N, Character > match );
    }
}
