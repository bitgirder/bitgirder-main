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
class Rfc4627RecognizerTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private
    abstract
    static
    class AbstractRecognizerTest< R extends Rfc4627Recognizer >
    extends LabeledTestCall
    {
        final CharSequence[] parts;
        final CharSequence errExpct;
        final int tailLenExpct;
        final R rec;

        private
        AbstractRecognizerTest( CharSequence lbl,
                                CharSequence errExpct,
                                int tailLenExpct,
                                R rec,
                                CharSequence... parts )
        {
            super( lbl );

            this.errExpct = errExpct;
            this.tailLenExpct = tailLenExpct;
            this.rec = rec;
            this.parts = parts;
        }

        private
        int
        recognize( CharSequence str,
                   boolean isEnd )
        {
            int res = rec.recognize( str, 0, isEnd );

            if ( isEnd )
            {
                state.isFalse(
                    rec.getStatus() == Rfc4627Recognizer.Status.RECOGNIZING );
            }

            return res;
        }

        private
        void
        assertFailure( int consumed )
        {
            String msg = rec.getErrorMessage();
            state.equal( Rfc4627Recognizer.Status.FAILED, rec.getStatus() );

            if ( tailLenExpct < 0 )
            {
                state.equalInt( -consumed - 1, tailLenExpct );
                state.equalString( errExpct, msg );
            }
            else throw state.createFail( msg );
        }

        private
        int
        getTailLenActual( int consumed )
        {
            int tot = 0;
            for ( CharSequence part : parts ) tot += part.length();

            return tot - consumed;
        }

        abstract
        void
        assertSuccess();

        public
        final
        void
        call()
            throws Exception
        {
            int len = 0;

            for ( int i = 0, e = parts.length; i < e && ! rec.failed(); ++i )
            {
                len += recognize( parts[ i ], i + 1 == e );
                if ( rec.failed() ) assertFailure( len );
            }

            if ( rec.failed() ) assertFailure( len ); 
            else 
            {
                state.equalInt( tailLenExpct, getTailLenActual( len ) );
                assertSuccess();
            }
        }
    }
    
    private
    final
    class StringTestCall
    extends AbstractRecognizerTest< Rfc4627StringRecognizer >
    {
        private final CharSequence expct;

        private
        StringTestCall( CharSequence lbl,
                        CharSequence expct,
                        int tailLenExpct,
                        CharSequence... parts )
        {
            super( 
                lbl, 
                expct, // Is also errExpct when we expect failures
                tailLenExpct, 
                Rfc4627StringRecognizer.create(), 
                parts 
            );

            this.expct = expct;
        }

        void
        assertSuccess()
        {
            code( "Asserting success" );
            state.isFalse( tailLenExpct < 0, "Expected a failure" );
            state.equalString( expct, rec.getString() );
        }
    }

    @InvocationFactory
    private
    List< StringTestCall >
    testRfc4627StringRecognizer()
    {
        return Lang.asList(
            
            new StringTestCall( 
                "single-part-vanilla", 
                "hello there", 
                0,
                "\"hello there\"" 
            ),

            new StringTestCall(
                "single-part-with-unescaped-forward-solidus",
                "/",
                0,
                "\"/\""
            ),

            new StringTestCall( "single-part-empty-string", "", 0, "\"\"" ),
            
            new StringTestCall( "multi-part-empty-string", "", 0, "\"", "\"" ),

            new StringTestCall(
                "single-part-vanilla-with-tail",
                "hello there",
                5,
                "\"hello there\" abcd"
            ),

            // test case-insensitivity of unicode hex-digits as well
            new StringTestCall( 
                "single-part-with-escapes",
                "\"/\\\b\f\n\r\t some normal stuff too \u0123\u01ff\u01ff",
                0,
                "\"\\\"\\/\\\\\\b\\f\\n\\r\\t " +
                    "some normal stuff too " +
                    "\\u0123\\u01ff\\u01Ff\""
            ),

            new StringTestCall( 
                "multi-part-clean-breaks",
                "this\nhas\t\u0123\fno split escapes.",
                0,
                "\"this", "\\nhas\\t", "\\u0123\\fno split", " escapes.\""
            ),

            new StringTestCall(
                "multi-part-with-tail",
                "this has a tail",
                5,
                "\"this has a tail\" abcd"
            ),

            new StringTestCall(
                "multi-part-split-escapes",
                "\"\n\u0123\u0123\u0123\u0123",
                0,
                "\"\\", "\"", "\\", "n\\u012", "3\\u01", "23\\u", "0123",
                    "\\", "u01", "23\""
            ),

            new StringTestCall( 
                "fail-unterminated-string", 
                "Unterminated string", 
                -7, 
                "\"hello" 
            ),

            new StringTestCall( 
                "invalid-single-char-escape", 
                "Unrecognized escape: \\c", 
                -8, 
                "\"hi: \\c\"" 
            ),
        
            new StringTestCall( 
                "incomplete-unicode-escape", 
                "Invalid hex char in escape: s", 
                -7, 
                "\"\\u23stuff\"" 
            ),

            new StringTestCall(
                "unterminated-unicode-escape", 
                "Invalid hex char in escape: \"", 
                -8,
                "\"\\u012\"" 
            ),

            new StringTestCall( 
                "unterminated-any-escape", 
                "Unterminated string", 
                -4, 
                "\"\\\"" 
            ),

            new StringTestCall(
                "unescaped-control-char",
                "Invalid unescaped character literal",
                -6,
                "\"Bad\n\""
            )
        );
    }

    @Test
    private
    void
    testStringRecognizerRespectsBeginIndex()
        throws Exception
    {
        Rfc4627StringRecognizer rec = Rfc4627StringRecognizer.create();

        state.equalInt( 9, rec.recognize( "  \"hello\"  ", 2, true ) );

        state.equal( 
            Rfc4627StringRecognizer.Status.COMPLETED, rec.getStatus() );

        state.equalString( "hello", rec.getString() );
    }

    private
    final
    class NumTestCall
    extends AbstractRecognizerTest< Rfc4627NumberRecognizer >
    {
        private final CharSequence intExpct;
        private final CharSequence fracExpct;
        private final CharSequence expExpct;

        private
        NumTestCall( CharSequence lbl,
                     CharSequence intExpct,
                     CharSequence fracExpct,
                     CharSequence expExpct,
                     CharSequence errExpct,
                     int tailLenExpct,
                     CharSequence... parts )
        {
            super( 
                lbl,
                errExpct,
                tailLenExpct,
                Rfc4627NumberRecognizer.create(),
                parts
            );

            this.intExpct = intExpct;
            this.fracExpct = fracExpct;
            this.expExpct = expExpct;
        }

        private
        NumTestCall( CharSequence lbl,
                     CharSequence errExpct,
                     int tailLenExpct,
                     CharSequence... parts )
        {
            this( lbl, null, null, null, errExpct, tailLenExpct, parts );
        }

        void
        assertSuccess()
        {
            state.equalString( rec.getIntPart(), intExpct );
            state.equalString( rec.getFracPart(), fracExpct );
            state.equalString( rec.getExponent(), expExpct );
            code( "Num done" );
        }
    }

    @InvocationFactory
    private
    List< NumTestCall >
    testNumberRecognizer()
    {
        return Lang.< NumTestCall >asList(
            
            new NumTestCall( "posint", "1", null, null, null, 0, "1" ),
            new NumTestCall( "negint", "-1", null, null, null, 0, "-1" ),
            new NumTestCall( "zero", "0", null, null, null, 0, "0" ),

            new NumTestCall( 
                "zero-point-zero", "0", "0", null, null, 0, "0.0" ),

            new NumTestCall(
                "zero-point-multizeroes", "0", "000", null, null, 0, "0.000" ),

            new NumTestCall(
                "elaborate-zero", "0", "000", "00", null, 0, "0.000e00" ),

            new NumTestCall( "pi", "3", "14", null, null, 0, "3.14" ),

            new NumTestCall(
                "neg-decimal", "-3", "14", null, null, 0, "-3.14" ),

            new NumTestCall( 
                "neg-int-with-exp", "-314", null, "-2", null, 0, "-314e-2" ),
            
            new NumTestCall(
                "dec-with-exp", "55", "23", "5", null, 0, "55.23e5" ),
                
            new NumTestCall(
                "neg-dec-with-plus-exp", 
                "-55", "23", "52", null, 0, "-55.23e+52" ),

            new NumTestCall(
                "degenerative-splits",
                "-55", 
                "23", 
                "52", 
                null, 
                0,
                "-", "5", "5", ".", "2", "3", "e", "+", "5", "2" 
            ),

            new NumTestCall(
                "implicit-token-end", "123", null, null, null, 2, "123x4" ),

            new NumTestCall(
                "illegal-leading-zeroes-posint", 
                "Illegal leading zero(es) in int part",
                -4,
                "000"
            ),

            new NumTestCall(
                "illegal-leading-zeroes-negint",
                "Illegal leading zero(es) in int part",
                -5,
                "-000"
            ),
            
            new NumTestCall(
                "multiple-leading-signs", "Empty int", -2, "--234" ),

            new NumTestCall(
                "unterminated-decimal", "Unterminated decimal", -3, "1." ),
            
            new NumTestCall(
                "unterminated-exp", "Unterminated exponent", -5, "1.1e" )
        );
    }
    
    @Test
    private
    void
    testNumberStringBoundsRespected()
    {
        Rfc4627NumberRecognizer rec = Rfc4627NumberRecognizer.create();

        state.equalInt( 5, rec.recognize( "  123x s", 2, false ) );

        state.equalString( "123", rec.getIntPart() );
        state.isTrue( rec.getFracPart() == null );
        state.isTrue( rec.getExponent() == null );
    }

    // Regression test against a bug that first appeared ("Unhandled mode
    // EXP_DONE")
    @Test
    private
    void
    testNumberTrailingCommaRegression()
    {
        Rfc4627NumberRecognizer rec = Rfc4627NumberRecognizer.create();
        state.equalInt( 5, rec.recognize( "1.1e5,stuff", 0, false ) );
        
        state.equalString( "1", rec.getIntPart() );
        state.equalString( "1", rec.getFracPart() );
        state.equalString( "5", rec.getExponent() );
    }

    @Test( expected = IllegalArgumentException.class,
           expectedPattern = "indx >= input length" )
    private
    void
    testEmptyNumberFails()
    {
        Rfc4627NumberRecognizer.create().recognize( "", 0, true );
    }
}
