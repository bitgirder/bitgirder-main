package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Strings;

import java.io.IOException;

public
final
class Rfc4627Reader
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private Rfc4627Reader() {}

    public
    abstract
    static
    class ReadResult
    {
        private int col;

        private String errMsg;
        private int errCol;

        private ReadResult() {}

        private
        void
        checkErr( String methName )
        {
            state.isFalsef( errMsg == null, 
                "Attempt to call %s() when isOk() is true", methName );
        }

        public
        final
        String
        errorMessage()
        {
            checkErr( "errorMessage" );
            return errMsg;
        }

        public
        final
        int
        errorCol()
        {
            checkErr( "errorCol" );
            return errCol;
        }

        public final boolean isOk() { return errMsg == null; }

        final
        void
        checkOk( String methName )
        {
            if ( ! isOk() )
            {
                state.failf( 
                    "Attempt to call %s() when isOk() is false", methName );
            }
        }
    }

    private
    static
    abstract
    class AbstractReader< R extends ReadResult >
    {
        private final CharReader cr;
        final R readRes;

        private int col;

        private
        AbstractReader( CharReader cr,
                        R readRes )
        {
            this.cr = cr;
            this.readRes = readRes;
        }

        final
        boolean
        fail( int errColAdj,
              String msg )
        {
            readRes.errMsg = msg;
            readRes.errCol = col + errColAdj;

            return false;
        }

        final boolean fail( String msg ) { return fail( 0, msg ); }

        final
        boolean
        failf( int errColAdj,
               String tmpl,
               Object... args )
        {
            return fail( errColAdj, String.format( tmpl, args ) );
        }

        final
        boolean
        failf( String tmpl,
               Object... args )
        {
            return failf( 0, tmpl, args );
        }

        final
        int
        read()
            throws IOException
        {
            int res = cr.read();
            ++col;

            return res;
        }

        final
        int
        peek()
            throws IOException
        {
            return cr.peek();
        }

        final
        boolean
        expectChar( char ch,
                    String errTmpl,
                    Object... errArgs )
            throws IOException
        {
            int res = read();

            if ( res < 0 ) return fail( "Unexpected end of input" );

            char chRes = (char) res;

            if ( chRes == ch ) return true;

            Object[] args = new Object[ errArgs.length + 2 ];
            System.arraycopy( errArgs, 0, args, 0, errArgs.length );
            args[ args.length - 2 ] = chRes;
            args[ args.length - 1 ] = (int) chRes;

            return failf( errTmpl, args );
        }

        final
        boolean
        expectChar( char ch )
            throws IOException
        {
            return expectChar(
                ch,
                "Expected '%c' (U+%04X) but got '%c' (U+%04X)", 
                ch, (int) ch 
            );
        }

        abstract
        void
        implExec()
            throws IOException;
        
        final
        R
        exec()
            throws IOException
        {
            implExec();
            return readRes;
        }
    }

    public
    final
    static
    class StringRead
    extends ReadResult
    {
        private StringBuilder acc = new StringBuilder();

        private StringRead() {}

        public
        CharSequence
        string()
        {
            checkOk( "string" );
            return acc;
        }
    }

    private
    final
    static
    class StringReader
    extends AbstractReader< StringRead >
    {
        private 
        StringReader( CharReader cr ) 
        { 
            super( cr, new StringRead() ); 
        }

        private 
        boolean 
        append( char ch ) 
        { 
            readRes.acc.append( ch ); 
            return true;
        }

        private
        boolean
        unterminatedString()
        {
            return fail( "Unterminated string literal" );
        }
    
        private
        boolean
        isHexDigit( char ch )
        {
            return 
                ( ch >= '0' && ch <= '9' ) ||
                ( ch >= 'a' && ch <= 'f' ) ||
                ( ch >= 'A' && ch <= 'F' );
        }

        private
        char
        asHexChar( int v )
        {
            char ch = (char) v;

            if ( isHexDigit( ch ) ) return ch;
            else 
            {
                String tmpl = "Invalid hex char in escape: \"%c\" (U+%04X)";
                failf( tmpl, ch, (int) ch );

                return (char) 0;
            }
        }

        private
        int
        readHexCharVal()
            throws IOException
        {
            char[] arr = new char[ 4 ];

            for ( int i = 0; i < 4; ++i )
            {
                int v = read();

                if ( v < 0 )
                {
                    unterminatedString();
                    return -1;
                }
                else 
                {
                    char ch = asHexChar( v );
                    if ( ch == 0 ) return -1; else arr[ i ] = ch;
                }
            }

            return Integer.parseInt( new String( arr ), 16 );
        }

        private
        boolean
        accumulateTrailingSurrogate()
            throws IOException
        {
            String errTmpl = 
                "Expected trailing surrogate, found: \"%c\" (U+%04X)";

            if ( ! expectChar( '\\', errTmpl ) ) return false;

            int v = read();
            if ( v < 0 ) return unterminatedString();
            if ( v != (int) 'u' )
            {
                return failf( -1,
                    "Expected trailing surrogate, found: \\%c", (char) v );
            }

            v = readHexCharVal();
            if ( v < 0 ) return false;
            char ch = (char) v;
            if ( Character.isLowSurrogate( ch ) ) return append( ch );

            return failf( -5, errTmpl, ch, (int) ch );
        }

        private
        boolean
        accumulateHexChar()
            throws IOException
        {
            int val = readHexCharVal();

            if ( val < 0 ) return false;

            char ch = (char) val;
            append( ch );

            if ( Character.isLowSurrogate( ch ) )
            {
                String tmpl =
                    "Trailing surrogate U+%04X is not preceded by a leading " +
                    "surrogate";

                return failf( -5, tmpl, (int) ch );
            }

            if ( Character.isHighSurrogate( ch ) ) 
            {
                return accumulateTrailingSurrogate();
            }

            return true;
        }

        private
        boolean
        accumulateEscape()
            throws IOException
        {
            int rd = read();
            if ( rd < 0 ) return unterminatedString();

            char ch = (char) rd;
            switch ( ch )
            {
                case 't': return append( '\t' );
                case 'n': return append( '\n' );
                case 'r': return append( '\r' );
                case 'f': return append( '\f' );
                case 'b': return append( '\b' );
                case '\\': return append( '\\' );
                case '"': return append( '"' );
                case '/': return append( '/' );
                case 'u': return accumulateHexChar();

                default:
                    return failf( -1, 
                        "Unrecognized escape: \\%c (U+%04X)", ch, (int) ch );
            }
        }

        private
        boolean
        accumulate( char ch )
            throws IOException
        {
            if ( ch == '\\' ) return accumulateEscape();

            if ( ch < ' ' ) 
            {
                String tmpl = 
                    "Invalid control character U+%04X in string literal";

                return failf( tmpl, (int) ch );
            }

            return append( ch );
        }

        void
        implExec()
            throws IOException
        {
            if ( ! expectChar( '"' ) ) return;

            while ( true )
            {
                int rd = read();

                if ( rd == (int) '"' ) break;

                if ( rd < 0 ) 
                {
                    unterminatedString();
                    break;
                }

                if ( ! accumulate( (char) rd ) ) break;
            }
        }
    }

    public
    static
    StringRead
    readString( CharReader cr )
        throws IOException
    {
        inputs.notNull( cr, "cr" );
        return new StringReader( cr ).exec();
    }

    public
    final
    static
    class NumberOptions
    {
        private final boolean allowLeadingZeroes;
        private final int[] delims;
        
        private
        NumberOptions( NumberOptionsBuilder b )
        {
            this.allowLeadingZeroes = b.allowLeadingZeroes;
            this.delims = b.delims;
        }
    }

    public
    final
    static
    class NumberOptionsBuilder
    {
        private final int[] DEFAULT_DELIMS = new int[] {
            (int) ',',
            (int) ']',
            (int) '}'
        };

        private int[] delims = DEFAULT_DELIMS;
        private boolean allowLeadingZeroes = false;

        public
        NumberOptionsBuilder
        setDelimiters( int[] delims )
        {
            this.delims = inputs.notNull( delims, "delims" );
            return this;
        }

        public
        NumberOptionsBuilder
        setAllowLeadingZeroes( boolean flag )
        {
            this.allowLeadingZeroes = flag;
            return this;
        }

        public NumberOptions build() { return new NumberOptions( this ); }
    }

    private final static NumberOptions DEFAULT =
        new NumberOptionsBuilder().build();

    public
    final
    static
    class NumberRead
    extends ReadResult
    {
        private boolean neg;
        private CharSequence i;
        private CharSequence f;
        private CharSequence e;

        public 
        boolean
        negative() 
        {
            checkOk( "negative" );
            return neg;
        }

        public 
        CharSequence 
        integer() 
        {
            checkOk( "integer" );
            return i;
        }

        public 
        CharSequence 
        fraction() 
        {
            checkOk( "fraction" );
            return f;
        }

        public 
        CharSequence 
        exponent() 
        {
            checkOk( "exponent" );
            return e;
        }

        @Override
        public
        String
        toString()
        {
            return Strings.inspect( this, true,
                "neg", neg,
                "i", i,
                "f", f,
                "e", e
            ).
            toString();
        }
    }

    private
    final
    static
    class NumberReader
    extends AbstractReader< NumberRead >
    {
        private final NumberOptions opts;

        private 
        NumberReader( CharReader cr,
                      NumberOptions opts ) 
        { 
            super( cr, new NumberRead() ); 
            
            this.opts = opts;
        }

        // Don't error if what we see is EOF or invalid; let later calls handle
        // that
        private
        void
        setNegative()
            throws IOException
        {
            if ( peek() == (int) '-' ) 
            {
                readRes.neg = true;
                read();
            }
        }

        private
        boolean
        isDigit( int v )
        {
            char c = (char) v;
            return c >= '0' && c <= '9';
        }

        private
        StringBuilder
        readIntString( StringBuilder sb )
            throws IOException
        {
            while ( true )
            {
                int v = peek();
                if ( ! isDigit( v ) ) break;
                sb.append( (char) read() );
            }

            return sb.length() == 0 ? null : sb;
        }

        private
        StringBuilder
        readIntString()
            throws IOException
        {
            return readIntString( new StringBuilder() );
        }

        private
        boolean
        isNaturalNumDelim( int v )
        {
            for ( int i = 0, e = opts.delims.length; i < e; ++i )
            {
                if ( v == opts.delims[ i ] ) return true;
            }

            return false;
        }

        private
        boolean
        checkBadTrail( String errTyp,
                       char... valids )
            throws IOException
        {
            int i = peek();

            if ( i < 0 ) return true;

            if ( isNaturalNumDelim( i ) ) return true;
            for ( char valid : valids ) if ( i == (int) valid ) return true;

            return failf( 1, 
                "Unexpected char in %s: \"%c\" (U+%04X)", errTyp, (char) i, i );
        }

        private
        boolean
        setIntPart()
            throws IOException
        {
            CharSequence i = readIntString();

            if ( i == null ) 
            {
                return fail( 1, "Number has invalid or empty integer part" );
            }

            if ( ( ! opts.allowLeadingZeroes ) && i.length() > 1 && 
                 i.charAt( 0 ) == '0' && i.charAt( 1 ) == '0' )
            {
                return fail( 1 - i.length(),
                    "Illegal leading zero(es) in integer part" );
            }

            readRes.i = i;

            return checkBadTrail( "integer part", '.', 'e', 'E' );
        }

        private
        boolean
        setFrac()
            throws IOException
        {
            if ( peek() == (int) '.' ) read(); else return true;

            CharSequence f = readIntString();

            if ( f == null ) 
            {
                return fail( 
                    peek() < 0 ? 0 : 1,
                    "Number has empty or invalid fractional part" );
            }

            readRes.f = f;
            return checkBadTrail( "fractional part", 'e', 'E' );
        }

        private
        void
        setExp()
            throws IOException
        {
            int v = peek();
            if ( v == (int) 'e' || v == (int) 'E' ) read(); else return;

            StringBuilder e = new StringBuilder();

            v = peek();
            if ( v == (int) '-' || v == (int) '+' )
            {
                read();
                if ( v == (int) '-' ) e.append( (char) v );
            }

            // readRes.e could end up null here, such as from the input "1e" or
            // "1eBadness". 
            readRes.e = readIntString( e ); 

            // Order here matters, since we want to fail with bad input if there
            // is any
            if ( ! checkBadTrail( "exponent" ) ) return;
            if ( readRes.e == null )
            {
                fail( "Number has empty or invalid exponent" );
            }
        }

        void
        implExec()
            throws IOException
        {
            setNegative();
            if ( ! setIntPart() ) return;
            if ( ! setFrac() ) return;
            setExp();
        }
    }

    public
    static
    NumberRead
    readNumber( CharReader cr,
                NumberOptions opts )
        throws IOException
    {
        inputs.notNull( cr, "cr" );
        inputs.notNull( opts, "opts" );

        return new NumberReader( cr, opts ).exec();
    }

    public
    static
    NumberRead
    readNumber( CharReader cr )
        throws IOException
    {
        return readNumber( cr, DEFAULT );
    }
}
