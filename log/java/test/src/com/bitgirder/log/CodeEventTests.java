package com.bitgirder.log;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Strings;
import com.bitgirder.lang.Lang;
import com.bitgirder.lang.PatternHelper;

import com.bitgirder.test.Test;

import java.util.Map;
import java.util.LinkedHashMap;

import java.util.regex.Pattern;

@Test
public
final
class CodeEventTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static Object[] MSG1 = new Object[] { "msg", 1 };

    // Various exeptions used to test formatting of stacktraces, nested and
    // otherwise. To test formatting, we manually unroll the trace and
    // reimplement the algorithm from DefaultCodeEventFormatter (see
    // throwableExpect() in this class). 
    //
    // We reimplement rather than just call since the whole point is to ensure
    // that the internals of DefaultCodeEventHandler are doing what we expect.
    // Original versions of these tests actually hardcoded known-good trace
    // formattings, but that proved too brittle in the face of changing code in
    // this class itself or in upstream sources, since those would lead to
    // line-number changes in the trace. 
    private final static Throwable EX1 = new Exception();
    private final static Throwable EX2 = new Exception( "test-fail" );
    private final static Throwable EX3 = new Exception( "test-cause", EX2 );
    private final static Throwable EX4 = new Exception( "deep", EX3 );

    // pattern part for the "[ 2011-09-23T09:17:48.113-07:00 ]" part
    private final static String TIME_PAT =
        "\\[ " +
        "\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}" +
        "\\.\\d{3}[\\-\\+]\\d{2}:\\d{2}" +
        " \\]";

    // We do this to stdout so it will still work even if CodeLoggers.code() is
    // not working (which could be the case if we're working with the library
    // itself)
    private
    static
    void
    code( Object... args )
    {
        System.out.println( Strings.join( " ", args ) );
    }

    public static long time() { return System.currentTimeMillis(); }

    private
    static
    String
    throwableExpect( String pref,
                     Throwable th )
    {
        return 
            Pattern.quote(
                Strings.join( "\n",
                    "-------",
                    pref + " " + th.getClass().getName() + 
                        ": " + th.getMessage(),
                    Strings.join( "\n", (Object[]) th.getStackTrace() )
                ).toString()
            );
    }

    public
    static
    void
    assertFormat( CodeEvent ev,
                  CodeEventFormatter fmtr,
                  CharSequence pat )
    {
        inputs.notNull( ev, "ev" );
        inputs.notNull( fmtr, "fmtr" );
        inputs.notNull( pat, "pat" );

        state.matches(
            CodeEvents.format( ev, fmtr ),
            "ev",
            PatternHelper.compile( pat )
        );
    }

    public
    static
    void
    assertStandardFormat( CodeEvent ev,
                          CodeEventFormatter fmtr,
                          CharSequence pat )
    {
        inputs.notNull( pat, "pat" );

        // The (?s) is the embedded DOTALL mode (see jdocs for
        // java.util.regex.Pattern), which we use since we may be matching
        // multiline stacktraces
        assertFormat( ev, fmtr, "^(?s)" + TIME_PAT + pat + "$" );
    }

    private
    static
    void
    assertStandardFormat( CodeEvent ev,
                          CharSequence pat )
    {
        assertStandardFormat( ev, new DefaultCodeEventFormatter(), pat );
    }

    @Test
    private
    void
    testFormatBasic()
    {
        for ( CodeEventType t : CodeEventType.class.getEnumConstants() )
        {
            assertStandardFormat( 
                CodeEvents.create( t, MSG1, null, null, time() ),
                "\\Q[][ " + t + " ]: msg 1\\E" 
            );
        }
    }

    @Test
    private
    void
    testFormatNoMessage()
    {
        assertStandardFormat(
            CodeEvents.create( CodeEventType.CODE, null, null, null, time() ),
            "\\Q[][ CODE ]\\E"
        );
    }

    // The checks below come from dumping the formatted output to the console,
    // ensuring manually that it is as desired, and then formatting it here as a
    // fixed regex for ongoing checks.

    @Test
    private
    void
    testFormatThrowableNoCause()
    {
        assertStandardFormat(
            CodeEvents.create( CodeEventType.CODE, MSG1, EX1, null, time() ),
            "\\Q[][ CODE ]: msg 1\\E\n" +
            throwableExpect( "Throwable", EX1 )
        );
    }

    @Test
    private
    void
    testFormatThrowableWithCause()
    {
        assertStandardFormat(
            CodeEvents.create( CodeEventType.WARN, MSG1, EX3, null, time() ),
            "\\Q[][ WARN ]: msg 1\\E\n" +
            throwableExpect( "Throwable", EX3 ) + "\n" +
            throwableExpect( "Caused by", EX2 )
        );
    }

    @Test
    private
    void
    testFormatThrowableWithManyCauses()
    {
        assertStandardFormat(
            CodeEvents.create( CodeEventType.WARN, MSG1, EX4, null, time() ),
            "\\Q[][ WARN ]: msg 1\\E\n" +
            throwableExpect( "Throwable", EX4 ) + "\n" +
            throwableExpect( "Caused by", EX3 ) + "\n" +
            throwableExpect( "Caused by", EX2 )
        );
    }

    // Uses a LinkedHashMap internally so that the order of add will be the
    // order of formatting
    public
    static
    Map< Object, Object >
    createAttachments( Object... pairs )
    {
        inputs.notNull( pairs, "pairs" );

        inputs.isTrue( 
            pairs.length % 2 == 0, "pairs array is not even length" );

        LinkedHashMap< Object, Object > res = 
            new LinkedHashMap< Object, Object >();
        
        for ( int i = 0, e = pairs.length; i < e; i += 2 )
        {
            res.put( pairs[ i ], pairs[ i + 1 ] );
        }

        return res;
    }

    @Test
    private
    void
    testDefaultFormatterWithAttachments()
    {
        Map< Object, Object > atts = 
            createAttachments( "key1", "val1", "key2", 2 );

        assertStandardFormat(
            CodeEvents.create( CodeEventType.CODE, MSG1, null, atts, time() ),
            "\\Q[ key1 = val1, key2 = 2 ][ CODE ]: msg 1\\E"
        );
    }

    @Test
    private
    void
    testFormatterWithUnprocessedButNonEmptyAttachments()
    {
        Map< Object, Object > atts = createAttachments( "k", "v" );

        assertStandardFormat(
            CodeEvents.create( CodeEventType.CODE, MSG1, null, atts, time() ),
            new DefaultCodeEventFormatter() 
            {
                @Override 
                protected 
                boolean 
                appendAttachment( StringBuilder sb,
                                  Object key,
                                  Object val,
                                  String sep )
                {
                    return false;
                }
            },
            "\\Q[][ CODE ]: msg 1\\E"
        );
    }

    // a little regression test action
    @Test
    private
    void
    testDefaultEventCreateAttsIsMutable()
    {
        CodeEvent ev = 
            CodeEvents.create( CodeEventType.CODE, MSG1, null, null, time() );
        
        ev.attachments().put( "a", "b" );
        state.equal( "b", ev.attachments().get( "a" ) );
    }
}
