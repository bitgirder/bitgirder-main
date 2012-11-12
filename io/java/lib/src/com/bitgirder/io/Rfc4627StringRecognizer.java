package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

public
final
class Rfc4627StringRecognizer
extends Rfc4627Recognizer
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private StringBuilder sb; // initialized after opening '"'

    private int escapeState = -1;
    private final char[] hexEscape = new char[ 4 ];

    private Rfc4627StringRecognizer() {}

    public
    CharSequence
    getString()
    {
        state.isTrue( 
            getStatus() == Status.COMPLETED,
            "Attempt to access result string but recognizer has status",
                getStatus()
        );

        return sb;
    }

    private
    void
    expectOpeningQuote( char ch )
    {
        if ( ch == '"' ) sb = new StringBuilder();
        else setFailure( "Expected opening '\"'" );
    }

    private
    void
    addRegularResult( CharSequence input,
                      int start,
                      int indx )
    {
        int end = Integer.MIN_VALUE;

        switch ( getStatus() )
        {
            case COMPLETED: end = indx - 1; break;
            case RECOGNIZING: end = indx - ( escapeState + 1 ); break;
        }

        if ( end > start ) sb.append( input.subSequence( start, end ) );
    }

    private
    void
    assertUnescaped( char ch )
    {
        if ( ch < ' ' || ch == '"' || ch == '\\' )
        {
            setFailure( "Invalid unescaped character literal" );
        }
    }

    private
    int
    accumulateRegular( CharSequence input,
                       int indx )
    {
        int start = indx;

        for ( int e = input.length(); 
              getStatus() == Status.RECOGNIZING && 
                indx < e && 
                escapeState == -1; 
              ++indx )
        {
            char ch = input.charAt( indx );

            switch ( ch )
            {
                case '"': setStatus( Status.COMPLETED ); break;
                case '\\': escapeState = 0; break;
                default: assertUnescaped( ch ); 
            }
        }

        addRegularResult( input, start, indx );

        return indx;
    }

    private
    void
    setMnemonicEscape( char ch )
    {
        char res = 0;

        switch ( ch )
        {
            case '"'    : res ='"'; break;
            case '/'    : res ='/'; break;
            case '\\'   : res ='\\'; break;
            case 'b'    : res ='\b'; break;
            case 'f'    : res ='\f'; break;
            case 'n'    : res ='\n'; break;
            case 'r'    : res ='\r'; break;
            case 't'    : res ='\t'; break;

            default: setFailure( "Unrecognized escape: \\" + ch );
        }

        if ( res > 0 )
        {
            sb.append( res );
            escapeState = -1;
        }
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

    // escapeState is 1 more than the index into the hex string
    private
    void
    accumulateUnicodeEscape( char ch )
    {
        if ( isHexDigit( ch ) ) 
        {
            hexEscape[ escapeState - 1 ] = ch;
        
            if ( ++escapeState == 5 )
            {
                String hexStr = new String( hexEscape );
                sb.append( (char) Integer.parseInt( hexStr, 16 ) );
                escapeState = -1;
            }
        }
        else setFailure( "Invalid hex char in escape: " + ch );
    }

    // Currently only processes a single char of the escape at a time, even if
    // it turns out that we are reading a unicode escape and have the 'uXXXX'
    // available. We can optimize this later if needed, but for now code is
    // simpler one char per call.
    private
    int
    accumulateEscape( CharSequence input,
                      int indx )
    {
        state.isTrue( escapeState >= 0 ); // sanity check on our impl

        char ch = input.charAt( indx );

        if ( escapeState == 0 )
        {
            if ( ch == 'u' ) ++escapeState; else setMnemonicEscape( ch );
        }
        else accumulateUnicodeEscape( ch );
        
        return indx + 1;
    }

    // alg is currently optimized for the common case in which a string contains
    // few if any escape sequences
    private
    int
    accumulate( CharSequence input,
                int indx )
    {
        if ( escapeState == -1 ) return accumulateRegular( input, indx );
        else return accumulateEscape( input, indx );
    }

    int
    recognizeImpl( CharSequence input,
                   int indx,
                   boolean isEnd )
    {
        if ( sb == null ) 
        {
            expectOpeningQuote( input.charAt( indx ) );
            ++indx;
        }
            
        for ( int e = input.length(); 
              indx < e && getStatus() == Status.RECOGNIZING; )
        {
            indx = accumulate( input, indx );
        }

        if ( isEnd && getStatus() == Status.RECOGNIZING )
        {
            setFailure( "Unterminated string" );
        }

        return indx;
    }

    public
    static
    Rfc4627StringRecognizer
    create()
    {
        return new Rfc4627StringRecognizer();
    }
}
