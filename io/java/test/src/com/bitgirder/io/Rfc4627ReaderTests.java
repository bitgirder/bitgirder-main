package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

import com.bitgirder.test.Test;
import com.bitgirder.test.InvocationFactory;
import com.bitgirder.test.LabeledTestCall;

import java.util.List;

@Test
final
class Rfc4627ReaderTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private
    abstract
    static
    class AbstractReaderTest< T extends AbstractReaderTest< T, ? >,
                              R extends Rfc4627Reader.ReadResult >
    extends LabeledTestCall
    {
        private CharSequence in;
        private CharSequence errMsg;
        private int errCol;
        private CharSequence tail;

        private AbstractReaderTest( CharSequence lbl ) { super( lbl ); }
        
        private T castThis() { return Lang.< T >castUnchecked( this ); }

        final
        T
        setIn( CharSequence in )
        {
            this.in = in;
            return castThis();
        }

        final
        T
        setError( CharSequence msg,
                  int col )
        {
            this.errMsg = msg;
            this.errCol = col;

            return castThis();
        }

        final
        T
        setTail( CharSequence tail )
        {
            this.tail = tail;
            return castThis();
        }

        abstract
        R
        read( CharReader cr )
            throws Exception;

        private
        void
        assertFailure( R res )
        {
            if ( errMsg == null )
            {
                state.fail( "Unexpected parse failure:", res.errorMessage() );
            }

            state.equalString( errMsg, res.errorMessage() );
            state.equalInt( errCol, res.errorCol() );
        }

        abstract
        void
        assertResult( R res );

        private
        CharSequence
        tailOf( CharReader cr )
            throws Exception
        {
            StringBuilder sb = new StringBuilder();

            for ( int i = cr.read(); i >= 0; i = cr.read() )
            {
                sb.append( (char) i );
            }

            return sb;
        }

        public
        final
        void
        call()
            throws Exception
        {
            CharReader cr = IoUtils.charReaderFor( in );
            R res = read( cr );

            if ( res.isOk() )
            {
                state.isTruef( errMsg == null,
                    "Got successful read but expected error: %s", errMsg );

                assertResult( res );
                state.equalString( tail == null ? "" : tail, tailOf( cr ) );
            }
            else assertFailure( res );
        }
    }
    
    private
    final
    class StringTestCall
    extends AbstractReaderTest< StringTestCall, Rfc4627Reader.StringRead >
    {
        private CharSequence expct;

        private StringTestCall( CharSequence lbl ) { super( lbl ); }

        private
        StringTestCall
        expect( CharSequence expct )
        {
            this.expct = expct;
            return this;
        }

        Rfc4627Reader.StringRead
        read( CharReader cr )
            throws Exception
        {
            return Rfc4627Reader.readString( cr );
        }

        void
        assertResult( Rfc4627Reader.StringRead res )
        {
            state.equalString( expct, res.string() );
        }
    }

    @InvocationFactory
    private
    List< StringTestCall >
    testRfc4627StringReader()
    {
        return Lang.asList(
 
            new StringTestCall( "basic-ascii" ).
                setIn( "\"hello there\"" ).
                expect( "hello there" ),

            new StringTestCall( "unescaped-forward-solidus" ).
                expect( "/" ).
                setIn( "\"/\"" ),

            new StringTestCall( "empty-string" ).
                expect( "" ).
                setIn( "\"\"" ),
            
            new StringTestCall( "single-part-vanilla-with-tail" ).
                expect( "hello there" ).
                setTail( " abcd" ).
                setIn( "\"hello there\" abcd" ),

            // test case-insensitivity of unicode hex-digits as well
            new StringTestCall( "misc-escapes" ).
                expect(
                    "\"/\\\b\f\n\r\t some normal stuff too " +
                    "\u0123\u01ff\u01ff" ).
                setIn(
                    "\"\\\"\\/\\\\\\b\\f\\n\\r\\t " +
                    "some normal stuff too " +
                    "\\u0123\\u01ff\\u01Ff\""
                ),

            new StringTestCall( "gclef" ).
                setIn( "\"a\\ud834\\udd1eb\"" ).
                expect( "a\ud834\udd1eb" ),

            new StringTestCall( "fail-unterminated-string" ).
                setError( "Unterminated string literal", 7 ).
                setIn( "\"hello" ),

            new StringTestCall( "invalid-single-char-escape" ).
                setError( "Unrecognized escape: \\c (U+0063)", 6 ).
                setIn( "\"hi: \\c\"" ),
        
            new StringTestCall( "incomplete-unicode-escape" ).
                setError( "Invalid hex char in escape: \"s\" (U+0073)", 6 ).
                setIn( "\"\\u23stuff\"" ),

            new StringTestCall( "unterminated-unicode-escape" ).
                setError( "Invalid hex char in escape: \"\"\" (U+0022)", 7 ).
                setIn( "\"\\u012\"" ),

            new StringTestCall( "unterminated-any-escape" ).
                setError( "Unterminated string literal", 4 ).
                setIn( "\"\\\""  ),

            new StringTestCall( "unescaped-control-char" ).
                setError(
                    "Invalid control character U+000A in string literal", 5 ).
                setIn( "\"Bad\nstuff\"" ),
            
            new StringTestCall( "last-char-is-high-surrogate" ).
                setError( 
                    "Expected trailing surrogate, found: \"\"\" (U+0022)", 8 ).
                setIn( "\"\\ud834\"" ),

            new StringTestCall( "low-surrogate-missing-unescaped-trail" ).
                setError(
                    "Expected trailing surrogate, found: \"|\" (U+007C)", 9 ).
                setIn( "\"a\\ud834|\\udd1e\"" ),

            new StringTestCall( "low-surrogate-missing-escaped-trail" ).
                setError( "Expected trailing surrogate, found: \\t", 9 ).
                setIn( "\"a\\ud834\\t\\udd1e\"" ),

            new StringTestCall( "invalid-surrogate-pair" ).
                setError( 
                    "Expected trailing surrogate, found: \"a\" (U+0061)", 9 ).
                setIn( "\"a\\ud834\\u0061\"" ),

            new StringTestCall( "unexpected-low-surrogate" ).
                setError(
                    "Trailing surrogate U+DD1E is not preceded by a leading " +
                    "surrogate",
                    3
                ).
                setIn( "\"a\\udd1e\\ud834\"" ),
            
            new StringTestCall( "not-a-string" ).
                setIn( "123" ).
                setError( "Expected '\"' (U+0022) but got '1' (U+0031)", 1 )
        );
    }

    private
    final
    class NumTestCall
    extends AbstractReaderTest< NumTestCall, Rfc4627Reader.NumberRead >
    {
        private boolean negExpct;
        private CharSequence intExpct;
        private CharSequence fracExpct;
        private CharSequence expExpct;
        private Rfc4627Reader.NumberOptions opts;

        private NumTestCall( CharSequence lbl ) { super( lbl ); }

        private
        NumTestCall
        setNumber( boolean negExpct,
                   CharSequence intExpct,
                   CharSequence fracExpct,
                   CharSequence expExpct )
        {
            this.negExpct = negExpct;
            this.intExpct = intExpct;
            this.fracExpct = fracExpct;
            this.expExpct = expExpct;

            return this;
        }

        private
        NumTestCall
        setOptions( Rfc4627Reader.NumberOptions opts )
        {
            this.opts = opts;
            return this;
        }

        Rfc4627Reader.NumberRead
        read( CharReader cr )
            throws Exception
        {
            if ( opts == null ) return Rfc4627Reader.readNumber( cr );
            else return Rfc4627Reader.readNumber( cr, opts );
        }

        void
        assertResult( Rfc4627Reader.NumberRead res )
        {
            state.equal( negExpct, res.negative() );
            state.equalString( intExpct, res.integer() );
            state.equalString( fracExpct, res.fraction() );
            state.equalString( expExpct, res.exponent() );
        }
    }

    @InvocationFactory
    private
    List< NumTestCall >
    testNumberReader()
    {
        return Lang.< NumTestCall >asList(
            
            new NumTestCall( "posint" ).
                setNumber( false, "1", null, null ).
                setIn( "1" ),

            new NumTestCall( "negint" ).
                setNumber( true, "1", null, null ).
                setIn( "-1" ),
            
            new NumTestCall( "zero" ).
                setNumber( false, "0", null, null ).
                setIn( "0" ),

            new NumTestCall( "neg-zero" ).
                setNumber( true, "0", null, null ).
                setIn( "-0" ),

            new NumTestCall( "zero-point-zero" ).
                setNumber( false, "0", "0", null ).
                setIn( "0.0" ),

            new NumTestCall( "zero-point-multizeroes" ).
                setNumber( false, "0", "000", null ).
                setIn( "0.000" ),

            new NumTestCall( "elaborate-zero" ).
                setNumber( false, "0", "000", "00" ).
                setIn( "0.000e00" ),

            new NumTestCall( "pi" ).
                setNumber( false, "3", "14", null ).
                setIn( "3.14" ),

            new NumTestCall( "neg-decimal" ).
                setNumber( true, "3", "14", null ).
                setIn( "-3.14" ),

            new NumTestCall( "neg-int-with-exp" ).
                setNumber( true, "314", null, "-2" ).
                setIn( "-314e-2" ),
            
            new NumTestCall( "dec-with-exp" ).
                setNumber( false, "55", "23", "5" ).
                setIn( "55.23e5" ),
                
            new NumTestCall( "neg-dec-with-plus-exp" ).
                setNumber( true, "55", "23", "52" ).
                setIn( "-55.23e+52" ),
            
            new NumTestCall( "comma-delim" ).
                setNumber( true, "1", null, null ).
                setIn( "-1, true" ).
                setTail( ", true" ),
            
            new NumTestCall( "multi-zero-int" ).
                setIn( "000000" ).
                setOptions( 
                    new Rfc4627Reader.NumberOptionsBuilder().
                        setAllowLeadingZeroes( true ).
                        build()
                ).
                setNumber( false, "000000", null, null ),
            
            new NumTestCall( "multi-zero-int-with-dec" ).
                setIn( "-000000.001" ).
                setOptions( 
                    new Rfc4627Reader.NumberOptionsBuilder().
                        setAllowLeadingZeroes( true ).
                        build()
                ).
                setNumber( true, "000000", "001", null ),
            
            new NumTestCall( "custom-delim" ).
                setIn( "1234," ).
                setOptions(
                    new Rfc4627Reader.NumberOptionsBuilder().
                        setDelimiters( new int[] { (int) ',' } ).
                        build()
                ).
                setNumber( false, "1234", null, null ).
                setTail( "," ),

            new NumTestCall( "illegal-leading-zeroes-posint" ).
                setError( "Illegal leading zero(es) in integer part", 1 ).
                setIn( "000" ),

            new NumTestCall( "illegal-leading-zeroes-negint" ).
                setError( "Illegal leading zero(es) in integer part", 2 ).
                setIn( "-000" ),
            
            new NumTestCall( "multiple-leading-signs" ).
                setError( "Number has invalid or empty integer part", 2 ).
                setIn( "--234" ),

            new NumTestCall( "unterminated-decimal" ).
                setError( "Number has empty or invalid fractional part", 2 ).
                setIn( "1." ),
            
            new NumTestCall( "unterminated-exp" ).
                setError( "Number has empty or invalid exponent", 4 ).
                setIn( "1.1e" ),

            new NumTestCall( "bad-frac-part-leading-char" ).
                setError( "Number has empty or invalid fractional part", 3 ).
                setIn( "0.x3" ),
            
            new NumTestCall( "multi-dot" ).
                setError( 
                    "Unexpected char in fractional part: \".\" (U+002E)", 4 ).
                setIn( "1.2.3" ),
            
            new NumTestCall( "trailing-bad-text-int" ).
                setError( 
                    "Unexpected char in integer part: \"x\" (U+0078)", 2 ).
                setIn( "1x" ),
            
            new NumTestCall( "trailing-bad-text-frac" ).
                setError( 
                    "Unexpected char in fractional part: \"x\" (U+0078)", 4 ).
                setIn( "1.1x" ),
            
            new NumTestCall( "trailing-bad-text-exp" ).
                setError( "Unexpected char in exponent: \"x\" (U+0078)", 6 ).
                setIn( "1.1e1x" ),
            
            new NumTestCall( "empty-number" ).
                setError( "Number has invalid or empty integer part", 1 ).
                setIn( "" ),
            
            new NumTestCall( "not-a-number" ).
                setIn( "\"abc\"" ).
                setError( "Number has invalid or empty integer part", 1 )
        );
    }
}
