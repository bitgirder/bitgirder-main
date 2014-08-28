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

    private final CountingCharReader cr;

    private 
    MingleLexer( CharReader cr ) 
    { 
        this.cr = new CountingCharReader( cr ); 
    }
    
    enum SpecialLiteral
    {
        COLON( ":" ),
        OPEN_BRACKET( "[" ),
        CLOSE_BRACKET( "]" ),
        FORWARD_SLASH( "/" ),
        PERIOD( "." ),
        ASPERAND( "@" );
    
        private final String lit;
    
        private SpecialLiteral( String lit ) { this.lit = lit; }

        String inspect() { return "'" + lit + "'"; }

        static
        boolean
        couldStartWith( char ch ) 
        {
            for ( SpecialLiteral sl : SpecialLiteral.class.getEnumConstants() )
            {
                if ( sl.lit.charAt( 0 ) == ch ) return true;
            }

            return false;
        }
    }

    long position() { return cr.position(); }

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

    private
    void
    implCheckUnexpectedEnd( String msg,
                            Object... args )
        throws MingleSyntaxException,
               IOException
    {
        int v = cr.peek();

        if ( v < 0 ) throw failf( 1, msg, args );
    }

    void
    checkUnexpectedEnd( String errExpct )
        throws MingleSyntaxException,
               IOException
    {
        implCheckUnexpectedEnd( "Expected %s but found END", errExpct );
    }

    void
    checkUnexpectedEnd()
        throws MingleSyntaxException,
               IOException
    {
        implCheckUnexpectedEnd( "Unexpected end of input" );
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

        if ( ! SpecialLiteral.couldStartWith( (char) v ) )
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
        implCheckUnexpectedEnd( "Empty identifier" );

        List< String > parts = Lang.newList();
        parts.add( expectIdPart( fmt ) );

        if ( fmt == null ) fmt = detectIdFormat();

        if ( fmt != null ) 
        {
            while ( nextIsIdSep( fmt ) ) parts.add( expectIdPart( fmt ) );
        }

        checkIdEnd();
        if ( parts.isEmpty() ) throw fail( "Empty identifier" );
 
        String[] partsArr = new String[ parts.size() ];
        return new MingleIdentifier( parts.toArray( partsArr ) );
    }

    private
    boolean
    isDeclNmStart( int v )
    {
        return isUpperCase( v );
    }

    private
    boolean
    isSpecStart( int v )
    {
        return SpecialLiteral.couldStartWith( (char) v );
    }

    private
    void
    readDeclNameTail( StringBuilder sb )
        throws MingleSyntaxException,
               IOException
    {
        for ( int v = cr.peek(); 
              ! ( v < 0 || isSpecStart( v ) ); 
              v = cr.peek() )
        {
            if ( isUpperCase( v ) || isLowerCase( v ) || isDigit( v ) )
            {
                sb.append( (char) v );
                cr.read();
            }
            else
            {
                throw failf( 1, "Illegal type name rune: \"%c\" (U+%04X)",
                    (char) v, v );
            }
        }
    }

    DeclaredTypeName
    parseDeclaredTypeName()
        throws MingleSyntaxException,
               IOException
    {
        implCheckUnexpectedEnd( "Empty type name" );

        StringBuilder sb = new StringBuilder();

        int v = cr.read();

        if ( isDeclNmStart( v ) ) sb.append( (char) v );
        else
        {
            throw failf( "Illegal type name start: \"%c\" (U+%04X)", 
                (char) v, v );
        }

        readDeclNameTail( sb );

        return new DeclaredTypeName( sb.toString() );
    }

    private
    SpecialLiteral
    parseSpecial()
        throws MingleSyntaxException,
               IOException
    {
        int v = cr.read();
        
        switch ( v ) {
        case (int) ':': return SpecialLiteral.COLON;
        case (int) '[': return SpecialLiteral.OPEN_BRACKET;
        case (int) ']': return SpecialLiteral.CLOSE_BRACKET;
        case (int) '/': return SpecialLiteral.FORWARD_SLASH;
        case (int) '.': return SpecialLiteral.PERIOD;
        case (int) '@': return SpecialLiteral.ASPERAND;
        }
        throw state.failf( "Unhandled spec start: %c", (char) v );
    }

    private
    MingleSyntaxException
    unrecognizedTokStart( int v )
    {
        return failf( 
            1, "Unrecognized token start: \"%c\" (U+%04X)", (char) v, v );
    }

    Object
    nextToken()
        throws MingleSyntaxException,
               IOException
    {
        int v = cr.peek();

        if ( v < 0 ) return null;

        if ( isIdStart( v ) ) return parseIdentifier( null );
        if ( isDeclNmStart( v ) ) return parseDeclaredTypeName();
        if ( isSpecStart( v ) ) return parseSpecial();

        throw unrecognizedTokStart( v );
    }

    static
    MingleLexer
    forString( CharSequence cs )
    {
        inputs.notNull( cs, "cs" );
        return new MingleLexer( IoUtils.charReaderFor( cs ) );
    }
}
