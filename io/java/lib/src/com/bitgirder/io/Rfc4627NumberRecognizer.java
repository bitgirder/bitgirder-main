package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

public
final
class Rfc4627NumberRecognizer
extends Rfc4627Recognizer
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private 
    static 
    enum Mode 
    { 
        INITIAL, 
        INT,
        FRAC_DECIMAL,
        FRAC_DIGITS,
        EXP_START,
        EXP_SIGN,
        EXP_DIGITS,
        EXP_DONE; 
    }

    private Mode mode = Mode.INITIAL;

    private StringBuilder intPart = new StringBuilder();
    private StringBuilder fracPart;
    private StringBuilder exp;

    private Rfc4627NumberRecognizer() {}

    private
    CharSequence
    ifComplete( String methNm,
                CharSequence val )
    {
        if ( getStatus() == Status.COMPLETED ) return val;
        else 
        {
            throw 
                state.createFail( 
                    "Attempt to call", methNm, "while recognizer has status",
                        getStatus()
                );
        }
    }

    public
    CharSequence
    getIntPart()
    {
        return ifComplete( "getIntPart()", intPart );
    }

    public
    CharSequence
    getFracPart()
    {
        return ifComplete( "getFracPart()", fracPart );
    }

    public
    CharSequence
    getExponent()
    {
        return ifComplete( "getExponent()", exp );
    }

    private boolean isDigit( char ch ) { return ch >= '0' && ch <= '9'; }

    private
    int
    accumulateDigits( CharSequence input,
                      int indx,
                      Mode nextMode,
                      StringBuilder sb )
    {
        Mode startMode = mode;
        int start = indx;

        for ( int e = input.length(); indx < e && mode == startMode; )
        {
            if ( isDigit( input.charAt( indx ) ) ) ++indx;
            else mode = nextMode;
        }

        if ( start < indx ) sb.append( input.subSequence( start, indx ) );

        return indx;
    }

    private
    boolean
    closeDigits( StringBuilder sb,
                 boolean signed,
                 String errMsg )
    {
        boolean isEmpty;

        if ( signed )
        {
            isEmpty = 
                ( sb.length() == 1 && sb.charAt( 0 ) == '-' ) ||
                sb.length() == 0;
        }
        else isEmpty = sb.length() == 0;

        if ( isEmpty )
        {
            setFailure( errMsg );
            return false;
        }
        else return true;
    }

    private 
    boolean 
    closeExp() 
    { 
        if ( closeDigits( exp, true, "Empty exponent" ) )
        {
            setStatus( Status.COMPLETED );
            return true;
        }
        else return false;
    }

    private
    int
    doExpDigits( CharSequence input,
                 int indx )
    {
        int res = accumulateDigits( input, indx, Mode.EXP_DONE, exp );

        if ( mode == Mode.EXP_DONE ) closeExp();
        
        return res;
    }

    private
    int
    doExpSign( CharSequence input,
               int indx )
    {
        char ch = input.charAt( indx );

        if ( ch == '-' || ch == '+' )
        {
            if ( ch == '-' ) exp.append( '-' );
            mode = Mode.EXP_DIGITS;

            return indx + 1;
        }
        else
        {
            mode = Mode.EXP_DIGITS;
            return indx;
        }
    }

    private
    int
    doExpStart( CharSequence input,
                int indx )
    {
        char ch = input.charAt( indx );

        if ( ch == 'e' || ch == 'E' )
        {
            exp = new StringBuilder();
            mode = Mode.EXP_SIGN;

            return indx + 1;
        }
        else
        {
            setStatus( Status.COMPLETED );
            return indx;
        }
    }

    private
    boolean
    closeFracDigits()
    {
        return closeDigits( fracPart, false, "Unterminated decimal" );
    }

    private
    int
    doFracDigits( CharSequence line,
                  int indx )
    {
        int res = accumulateDigits( line, indx, Mode.EXP_START, fracPart );

        if ( mode == Mode.EXP_START ) closeFracDigits();
        
        return res;
    }

    private
    int
    doFracDecimal( CharSequence line,
                   int indx )
    {
        if ( line.charAt( indx ) == '.' )
        {
            fracPart = new StringBuilder();
            mode = Mode.FRAC_DIGITS;
            return indx + 1;
        }
        else
        {
            mode = Mode.EXP_START;
            return indx;
        }
    }

    private
    boolean
    closeInt()
    {
        if ( closeDigits( intPart, true, "Empty int" ) )
        {
            int lead = intPart.charAt( 0 ) == '-' ? 1 : 0;
            
            if ( intPart.charAt( lead ) == '0' && intPart.length() > 1 + lead )
            {
                setFailure( "Illegal leading zero(es) in int part" );
                return false;
            }
            else return true;
        }
        else return false;
    }

    private
    int
    doInt( CharSequence input,
           int indx )
    {
        int res = accumulateDigits( input, indx, Mode.FRAC_DECIMAL, intPart );

        if ( mode == Mode.FRAC_DECIMAL ) closeInt();

        return res;
    }

    private
    int
    doInitial( CharSequence input,
               int indx )
    {
        // one way or the other we're going to move to INT mode now
        mode = Mode.INT;

        if ( input.charAt( indx ) == '-' )
        {
            intPart.append( '-' );
            return indx + 1;
        }
        else return indx;
    }

    private
    void
    forceEnd()
    {
        switch ( mode )
        {
            case INITIAL: 
                setFailure( "Empty number" ); 
                break;
            
            case INT: 
                if ( closeInt() ) setStatus( Status.COMPLETED );
                break;
            
            case FRAC_DECIMAL:
                setStatus( Status.COMPLETED );
                break;
 
            case FRAC_DIGITS:
                if ( closeFracDigits() ) setStatus( Status.COMPLETED );
                break;
            
            case EXP_START:
                setStatus( Status.COMPLETED );
                break;

            case EXP_SIGN:
                setFailure( "Unterminated exponent" );
                break;

            case EXP_DIGITS:
                if ( closeExp() ) setStatus( Status.COMPLETED );
                break;

            default: state.fail( "Unexpected mode in forceEnd():", mode );
        }
    }

    int
    recognizeImpl( CharSequence input,
                   int indx,
                   boolean isEnd )
    {
        for ( int e = input.length(); 
                indx < e && getStatus() == Status.RECOGNIZING; )
        {
            switch ( mode )
            {
                case INITIAL: indx = doInitial( input, indx ); break;
                case INT: indx = doInt( input, indx ); break;
                case FRAC_DECIMAL: indx = doFracDecimal( input, indx ); break;
                case FRAC_DIGITS: indx = doFracDigits( input, indx ); break;
                case EXP_START: indx = doExpStart( input, indx ); break;
                case EXP_SIGN: indx = doExpSign( input, indx ); break;
                case EXP_DIGITS: indx = doExpDigits( input, indx ); break;
                default: state.fail( "Unhandled mode:", mode );
            }
        }

        if ( isEnd && getStatus() == Status.RECOGNIZING ) forceEnd();

        return indx;
    }

    public
    static
    Rfc4627NumberRecognizer
    create()
    {
        return new Rfc4627NumberRecognizer();
    }
}
