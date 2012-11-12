package com.bitgirder.mingle.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

import com.bitgirder.mingle.model.MingleString;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleValue;

import com.bitgirder.io.Rfc4627Recognizer;
import com.bitgirder.io.Rfc4627StringRecognizer;
import com.bitgirder.io.Rfc4627NumberRecognizer;

import com.bitgirder.parser.SyntaxException;
import com.bitgirder.parser.SourceTextLocation;

import java.util.Queue;
import java.util.Collection;
import java.util.EnumSet;
import java.util.Iterator;

public
final
class MingleLexer
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static String LC_ALPHA = "abcdefghijklmnopqrstuvwxyz";
    private final static String UC_ALPHA = "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
    private final static String DIGIT = "0123456789";

    private final TokenHandler th;
    private final CharSequence fileName;

    private int lineNo = 1;
    private int colNo = 1;

    private Recognizer rec;

    private
    MingleLexer( Builder b )
    {
        this.th = inputs.notNull( b.th, "th" );
        this.fileName = inputs.notNull( b.fileName, "fileName" );
    }

    private
    static
    boolean
    contains( String alpha,
              char ch )
    {
        return alpha.indexOf( ch ) >= 0;
    }

    private
    static
    SyntaxException
    createSyntaxException( String msg,
                           SourceTextLocation loc )
    {
        return new SyntaxException( msg, loc );
    }

    private
    static
    abstract
    class Recognizer
    {
        private MingleToken tok;

        SourceTextLocation loc;

        final
        SyntaxException
        createSyntaxException( String msg )
        {
            return MingleLexer.createSyntaxException( msg, loc );
        }

        // returns null until done
        final MingleToken getToken() { return tok; }

        final 
        void 
        setTokenObject( Object obj )
        { 
            state.isTrue( this.tok == null );
            this.tok = new MingleToken( obj, loc ); 
        }

        abstract
        int
        recognize( CharSequence line,
                   int indx,
                   boolean hadEol )
            throws SyntaxException;
    }

    private
    final
    static
    class IdentifiableTextRecognizer
    extends Recognizer
    {
        private final StringBuilder sb = new StringBuilder();

        private
        boolean
        isIdentifiable( char ch )
        {
            return
                contains( DIGIT, ch ) ||
                contains( UC_ALPHA, ch ) ||
                contains( LC_ALPHA, ch );
        }

        int
        recognize( CharSequence line,
                   int indx,
                   boolean hadEol )
        {
            int start = indx;

            boolean loop = true;

            for ( int e = line.length(); indx < e && loop; )
            {
                if ( isIdentifiable( line.charAt( indx ) ) ) ++indx;
                else loop = false;
            }

            if ( indx > start ) sb.append( line, start, indx );

            if ( ( hadEol && loop ) || ( ! loop ) )
            {
                setTokenObject( new IdentifiableText( sb ) );
            }

            return indx;
        }
    }

    private
    final
    static
    class NumberRecognizer
    extends Recognizer
    {
        private final Rfc4627NumberRecognizer rec =
            Rfc4627NumberRecognizer.create();

        private
        MingleValue
        buildNumber()
        {
            StringBuilder sb = new StringBuilder( rec.getIntPart() );

            CharSequence frac = rec.getFracPart();
            if ( frac != null ) sb.append( '.' ).append( frac );

            CharSequence exp = rec.getExponent();
            if ( exp != null ) sb.append( 'e' ).append( exp );

            if ( frac == null && exp == null )
            {
                long l = Long.parseLong( sb.toString() );
                return MingleModels.asMingleInt64( l );
            }
            else 
            {
                double d = Double.parseDouble( sb.toString() );
                return MingleModels.asMingleDouble( d );
            }
        }

        int
        recognize( CharSequence line,
                   int indx,
                   boolean hadEol )
            throws SyntaxException
        {
            indx = MingleLexer.recognize( rec, line, indx, hadEol, loc );

            if ( rec.completed() ) setTokenObject( buildNumber() );

            return indx;
        }
    }

    private
    final
    static
    class WhitespaceRecognizer
    extends Recognizer
    {
        private final StringBuilder sb = new StringBuilder();

        private
        void
        setWhitespace()
        {
            setTokenObject( new WhitespaceText( sb ) );
        }

        private
        int
        consumeWhitespace( CharSequence line,
                           int indx )
        {
            boolean loop = true;

            for ( int e = line.length(); indx < e && loop; )
            {
                char ch = line.charAt( indx );

                if ( Character.isWhitespace( ch ) ) 
                {
                    sb.append( ch );
                    ++indx;
                }
                else 
                {
                    loop = false;
                    setWhitespace();
                }
            }

            return indx;
        }

        int
        recognize( CharSequence line,
                   int indx,
                   boolean hadEol )
        {
            indx = consumeWhitespace( line, indx );
            if ( indx == line.length() && hadEol ) setWhitespace();

            return indx;
        }
    }

    private
    final
    static
    class CommentRecognizer
    extends Recognizer
    {
        private StringBuilder comment;

        int
        recognize( CharSequence line,
                   int indx,
                   boolean hadEol )
        {
            if ( comment == null )
            {
                state.isTrue( line.charAt( indx ) == '#' );
                ++indx;

                comment = new StringBuilder();
            }

            comment.append( line, indx, line.length() );
            if ( hadEol ) setTokenObject( new CommentText( comment ) );

            return line.length();
        }
    }

    private
    final
    static
    class SpecialLiteralRecognizer
    extends Recognizer
    {
        private EnumSet< SpecialLiteral > possibles =
            EnumSet.allOf( SpecialLiteral.class );

        // how many characters in the (potential) result literal have we matched
        // so far
        private int pos; 

        private
        boolean
        matchChar( char ch )
        {
            EnumSet< SpecialLiteral > next = EnumSet.copyOf( possibles );

            for ( Iterator< SpecialLiteral > it = next.iterator(); 
                    it.hasNext(); )
            {
                SpecialLiteral sl = it.next();
                if ( pos >= sl.length() || sl.charAt( pos ) != ch ) it.remove();
            }

            if ( next.isEmpty() ) return false;
            else
            {
                ++pos;
                possibles = next;
                return true;
            }
        }

        private
        void
        retainFinalPossibles()
        {
            for ( Iterator< SpecialLiteral > it = possibles.iterator();
                    it.hasNext(); )
            {
                if ( it.next().length() != pos ) it.remove();
            }
        }

        private
        SpecialLiteral
        getSpecialLiteral()
            throws SyntaxException
        {
            retainFinalPossibles();

            switch ( possibles.size() )
            {
                case 0: throw createSyntaxException( "Unrecognized literal" );
                case 1: return possibles.iterator().next();
                default: throw createSyntaxException( "Ambiguous literal" );
            }
        }

        int
        recognize( CharSequence line,
                   int indx,
                   boolean hadEol )
            throws SyntaxException
        {
            boolean loop = true;

            for ( int e = line.length(); indx < e && loop; )
            {
                if ( matchChar( line.charAt( indx ) ) ) ++indx;
                else loop = false;
            }

            if ( ( loop && hadEol ) || ( ! loop ) ) 
            {
                setTokenObject( getSpecialLiteral() );
            }

            return indx;
        }
    }

    private
    static
    int
    recognize( Rfc4627Recognizer rec,
               CharSequence line,
               int indx,
               boolean isEnd,
               SourceTextLocation loc )
        throws SyntaxException
    {
        indx = rec.recognize( line, indx, isEnd );

        if ( rec.failed() )
        {
            throw createSyntaxException( rec.getErrorMessage(), loc );
        }
        else return indx;
    }

    private
    final
    static
    class StringLiteralRecognizer
    extends Recognizer
    {
        private final Rfc4627StringRecognizer rec = 
            Rfc4627StringRecognizer.create();

        private
        MingleString
        buildString()
        {
            return MingleModels.asMingleString( rec.getString() );
        }

        int
        recognize( CharSequence line,
                   int indx,
                   boolean hadEol )
            throws SyntaxException
        {
            indx = MingleLexer.recognize( rec, line, indx, hadEol, loc );

            if ( rec.completed() ) setTokenObject( buildString() );

            return indx;
        }
    }

    private
    Recognizer
    recognizerFor( char ch )
        throws SyntaxException
    {
        SourceTextLocation loc =
            SourceTextLocation.create( fileName, lineNo, colNo );

        Recognizer res;

//        if ( contains( LC_ALPHA, ch ) ) res = new IdentifierRecognizer();
//        else if ( contains( UC_ALPHA, ch ) ) res = new TypeNameRecognizer();
//        else if ( contains( DIGIT, ch ) ) res = new NumberRecognizer();
        if ( contains( DIGIT, ch ) ) res = new NumberRecognizer();
        else if ( contains( LC_ALPHA, ch ) || contains( UC_ALPHA, ch ) )
        {
            res = new IdentifiableTextRecognizer();
        }
        else if ( Character.isWhitespace( ch ) ) 
        {
            res = new WhitespaceRecognizer();
        }
        else if ( ch == '#' ) res = new CommentRecognizer();
        else if ( ch == '"' ) res = new StringLiteralRecognizer();
        else if ( contains( SpecialLiteral.ALPHABET, ch ) )
        {
            res = new SpecialLiteralRecognizer();
        }
        else 
        {
            throw createSyntaxException( "Unexpected token start: " + ch, loc );
        }

        res.loc = loc;

        return res;
    }

    private
    void
    handleToken( MingleToken tok,
                 int indx )
        throws Exception
    {
        rec = null;
        colNo = indx + 1;

//        code( "Calling handleToken with", tok );
        th.handleToken( tok );
    }

    private
    void
    advanceLine()
    {
        ++lineNo;
        colNo = 1;
    }

    public
    void
    update( CharSequence line,
            boolean hadEol )
        throws Exception
    {
        inputs.notNull( line, "line" );

        for ( int i = 0, e = line.length(); i < e; )
        {
            if ( rec == null ) rec = recognizerFor( line.charAt( i ) );
            i = rec.recognize( line, i, hadEol );

            MingleToken tok = rec.getToken();
            if ( tok != null ) handleToken( tok, i );
        }

        if ( hadEol ) advanceLine();
    }

    public
    static
    interface TokenHandler
    {
        public
        void
        handleToken( MingleToken tok )
            throws Exception;
    }

    public
    final
    static
    class TokenAccumulator
    implements TokenHandler
    {
        private final Queue< MingleToken > toks = Lang.newQueue();

        private TokenAccumulator() {}

        public void handleToken( MingleToken tok ) { toks.add( tok ); }
        public Queue< MingleToken > getTokens() { return toks; }
    }

    public
    static
    TokenAccumulator
    createTokenAccumulator()
    {
        return new TokenAccumulator();
    }

    public
    final
    static
    class Builder
    {
        private TokenHandler th;
        private CharSequence fileName;

        public
        Builder
        setTokenHandler( TokenHandler th )
        {
            this.th = inputs.notNull( th, "th" );
            return this;
        }

        public
        Builder
        setFileName( CharSequence fileName )
        {
            this.fileName = inputs.notNull( fileName, "fileName" );
            return this;
        }

        public MingleLexer build() { return new MingleLexer( this ); }
    }

    public
    static
    Queue< MingleToken >
    tokenizeString( CharSequence str )
        throws SyntaxException
    {
        inputs.notNull( str, "str" );

        TokenAccumulator acc = createTokenAccumulator();

        MingleLexer lex =
            new Builder().
                setTokenHandler( acc ).
                setFileName( "<>" ).
                build();
        
        try { lex.update( str, true ); }
        catch ( RuntimeException re ) { throw re; }
        catch ( SyntaxException se ) { throw se; }
        catch ( Exception ex )
        {
            throw new RuntimeException( 
                "Unexpected checked exception (attached as cause)", ex );
        }

        return acc.getTokens();
    }
}
