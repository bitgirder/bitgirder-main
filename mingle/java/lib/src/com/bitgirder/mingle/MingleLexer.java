package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.io.Rfc4627Reader;
import com.bitgirder.io.IoUtils;
import com.bitgirder.io.CharReader;
import com.bitgirder.io.CountingCharReader;

import java.io.IOException;

import java.util.Arrays;
import java.util.List;

final
class MingleLexer
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final CountingCharReader cr;

    private 
    MingleLexer( CharReader cr ) 
    { 
        this.cr = new CountingCharReader( cr ); 
    }
    
    enum SpecialLiteral
    {
        COLON( ":" ),
        TILDE( "~" ),
        OPEN_PAREN( "(" ),
        CLOSE_PAREN( ")" ),
        OPEN_BRACKET( "[" ),
        CLOSE_BRACKET( "]" ),
        COMMA( "," ),
        QUESTION_MARK( "?" ),
        MINUS( "-" ),
        FORWARD_SLASH( "/" ),
        PERIOD( "." ),
        ASTERISK( "*" ),
        PLUS( "+" ),
        ASPERAND( "@" );
    
        final static String ALPHABET = ":~()[],?-/.*+@";
    
        private final String lit;
    
        private SpecialLiteral( String lit ) { this.lit = lit; }
    }

    private
    MingleSyntaxException
    fail( int adj,
          String msg )
    {
        return new MingleSyntaxException( msg, (int) ( cr.position() + adj ) );
    }

    private
    MingleSyntaxException
    fail( String msg )
    {
        return fail( 0, msg );
    }

    private
    MingleSyntaxException
    failf( int adj,
           String tmpl,
           Object... args )
    {
        return fail( adj, String.format( tmpl, args ) );
    }

    private
    MingleSyntaxException
    failf( String tmpl,
           Object... args )
    {
        return failf( 0, tmpl, args );
    }

    private
    MingleSyntaxException
    asSyntaxFailure( Rfc4627Reader.ReadResult rr,
                     int startCol )
    {
        return new MingleSyntaxException( 
            rr.errorMessage(), startCol + rr.errorCol() );
    }

    void
    checkUnexpectedEnd( String msg )
        throws MingleSyntaxException,
               IOException
    {
        int v = cr.peek();

        if ( v < 0 ) throw fail( 1, msg );
    }

    void
    checkUnexpectedEnd()
        throws MingleSyntaxException,
               IOException
    {
        checkUnexpectedEnd( "Unexpected end of input" );
    }

    void
    checkNoTrailing()
        throws MingleSyntaxException,
               IOException
    {
        int v = cr.peek();

        if ( v < 0 ) return;

        throw failf( 1, 
            "Unexpected trailing data \"%c\" (U+%04X)", (char) v, v );
    }

    private
    MingleString
    parseStringToken()
        throws MingleSyntaxException,
               IOException
    {
        int startCol = (int) cr.position();

        Rfc4627Reader.StringRead rd = Rfc4627Reader.readString( cr );

        if ( rd.isOk() ) return new MingleString( rd.string() );

        throw asSyntaxFailure( rd, startCol );
    }

    private
    boolean
    isDigit( int v )
    {
        return v >= (int) '0' && v <= (int) '9';
    }

    private
    boolean
    isUpperCase( int v )
    {
        return v >= (int) 'A' && v <= (int) 'Z';
    }

    private
    boolean
    isLowerCase( int v )
    {
        return v >= (int) 'a' && v <= (int) 'z';
    }

    private
    boolean
    isIdTailChar( int v )
    {
        return isDigit( v ) || isLowerCase( v );
    }

    private boolean isIdStart( int v ) { return isLowerCase( v ); }

    // Implements equals()/hashCode() for testing purposes only at the moment
    // (and therefore does so somewhat inefficiently)
    final
    static
    class Number
    {
        final boolean neg;
        final CharSequence i;
        final CharSequence f;
        final CharSequence e;

        Number( boolean neg,
                CharSequence i,
                CharSequence f,
                CharSequence e )
        {
            this.neg = neg;
            this.i = i;
            this.f = f;
            this.e = e;
        }

        private
        Object[]
        makeEqArr()
        {
            return new Object[] {
                neg,
                i.toString(),
                f == null ? null : f.toString(),
                e == null ? null : e.toString()
            };
        }

        public int hashCode() { return Arrays.hashCode( makeEqArr() ); }

        public
        boolean
        equals( Object o )
        {
            if ( o == this ) return true;
            if ( ! ( o instanceof Number ) ) return false;

            return Arrays.equals( makeEqArr(), ( (Number) o ).makeEqArr() );
        }

        @Override
        public
        String
        toString()
        {
            StringBuilder sb = new StringBuilder();

            if ( neg ) sb.append( '-' );
            sb.append( i );
            if ( f != null ) sb.append( '.' ).append( f );
            if ( e != null ) sb.append( 'e' ).append( e );

            return sb.toString();
        }
    }

    private
    Number
    parseNumber()
        throws MingleSyntaxException,
               IOException
    {
        int startCol = (int) cr.position();

        Rfc4627Reader.NumberRead rd = 
            Rfc4627Reader.readNumber( cr, Rfc4627Reader.ALLOW_LEADING_ZEROES );

        if ( rd.isOk() ) 
        {
            return new Number( 
                rd.negative(), rd.integer(), rd.fraction(), rd.exponent() );
        }

        throw asSyntaxFailure( rd, startCol );
    }

    private
    MingleIdentifierFormat
    detectIdFormat()
        throws IOException
    {
        int v = cr.peek();

        if ( v == (int) '-' ) return MingleIdentifierFormat.LC_HYPHENATED;
        if ( v == (int) '_' ) return MingleIdentifierFormat.LC_UNDERSCORE;
        if ( isUpperCase( v ) ) return MingleIdentifierFormat.LC_CAMEL_CAPPED;

        return null;
    }

    // Side effect: reads past a '-' or '_'
    private
    boolean
    nextIsIdSep( MingleIdentifierFormat fmt )
        throws IOException
    {
        int v = cr.peek();
        boolean res = false;
        boolean pass = false;

        switch ( fmt )
        {
            case LC_HYPHENATED: res = pass = v == (int) '-'; break;
            case LC_UNDERSCORE: res = pass = v == (int) '_'; break;
            case LC_CAMEL_CAPPED: res = isUpperCase( v ); break;
            default: throw state.createFail( "Unhandled id format:", fmt );
        }

        if ( pass ) cr.read();

        return res;
    }

    private
    MingleSyntaxException
    failEmptyIdPart()
        throws IOException
    {
        int v = cr.peek();

        if ( v < 0 ) return fail( 1, "Empty identifier part" );

        String tmpl = "Illegal start of identifier part: \"%c\" (U+%04X)";
        return failf( 1, tmpl, (char) v, v );
    }

    private
    void
    readIdPartStart( MingleIdentifierFormat fmt,
                     StringBuilder sb )
        throws MingleSyntaxException,
               IOException
    {
        if ( fmt == MingleIdentifierFormat.LC_CAMEL_CAPPED )
        {
            char ch = (char) cr.read();
            state.isTrue( isUpperCase( ch ) );
            sb.append( Character.toLowerCase( ch ) );
        }
        else
        {
            int v = cr.read();
            
            if ( v < 0 ) return; // will handle error in calling method

            if ( isLowerCase( v ) ) sb.append( (char) v );
            else 
            {
                throw failf( 
                    "Illegal start of identifier part: \"%c\" (U+%04X)",
                    (char) v, v
                );
            }
        }
    }

    private
    String
    expectIdPart( MingleIdentifierFormat fmt )
        throws MingleSyntaxException,
               IOException
    {
        StringBuilder sb = new StringBuilder( 8 );
        readIdPartStart( fmt, sb );

        for ( int ch = cr.peek(); isIdTailChar( ch ); ch = cr.peek() )
        {
            sb.append( (char) cr.read() );
        }

        if ( sb.length() == 0 ) throw failEmptyIdPart();
        else return sb.toString();
    }

    private
    void
    checkIdEnd()
        throws MingleSyntaxException,
               IOException
    {
        int v = cr.peek();

        if ( v < 0 ) return;

        if ( SpecialLiteral.ALPHABET.indexOf( (char) v ) < 0 )
        {
            throw failf( 1, 
                "Unexpected identifier character: \"%c\" (U+%04X)", (char) v, v
            );
        }
    }

    // fmt may be null, in which case it will be inferred
    MingleIdentifier
    parseIdentifier( MingleIdentifierFormat fmt )
        throws MingleSyntaxException,
               IOException
    {
        checkUnexpectedEnd( "Empty identifier" );

        List< String > parts = Lang.newList();
        parts.add( expectIdPart( fmt ) );

        if ( fmt == null ) fmt = detectIdFormat();
        code( "Id fmt:", fmt );

        if ( fmt != null ) 
        {
            while ( nextIsIdSep( fmt ) ) parts.add( expectIdPart( fmt ) );
        }

        checkIdEnd();
        if ( parts.isEmpty() ) throw fail( "Empty identifier" );
 
        String[] partsArr = new String[ parts.size() ];
        return new MingleIdentifier( parts.toArray( partsArr ) );
    }

    Object
    nextToken()
        throws MingleSyntaxException,
               IOException
    {
        int v = cr.peek();

        if ( v < 0 ) return null;

        if ( v == (int) '"' ) return parseStringToken();
        if ( v == (int) '-' || isDigit( v ) ) return parseNumber();
        if ( isIdStart( v ) ) return parseIdentifier( null );

        throw failf( 
            1, "Unrecognized token start: \"%c\" (U+%04X)", (char) v, v );
    }

    static
    MingleLexer
    forString( CharSequence cs )
    {
        inputs.notNull( cs, "cs" );
        return new MingleLexer( IoUtils.charReaderFor( cs ) );
    }
}
