package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.io.Rfc4627Reader;
import com.bitgirder.io.IoUtils;
import com.bitgirder.io.CharReader;
import com.bitgirder.io.CountingCharReader;

import java.io.IOException;

import java.util.Arrays;

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
        return new MingleSyntaxException( msg, adj );
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

    Object
    nextToken()
        throws MingleSyntaxException,
               IOException
    {
        int v = cr.peek();

        if ( v < 0 ) return null;

        if ( v == (int) '"' ) return parseStringToken();
        if ( v == (int) '-' || isDigit( v ) ) return parseNumber();

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
