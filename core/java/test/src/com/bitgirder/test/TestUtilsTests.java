package com.bitgirder.test;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

@Test
final
class TestUtilsTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private
    final
    static
    class MarkerException
    extends Exception
    {
        private MarkerException( String msg ) { super( msg ); }
    }

    private
    Throwable
    getFinalThrowable( final String pat,
                       final Class< ? extends Throwable > expctCls,
                       Throwable th )
    {
        TestFailureExpector tfe =
            new TestFailureExpector() {
                public Class< ? extends Throwable > expectedFailureClass() {
                    return expctCls;
                }
                public CharSequence expectedFailurePattern() { return pat; }
            };
        
        return 
            TestUtils.getFinalThrowable(
                TestUtils.failureExpectationFor( tfe ), th );
    }

    // strs[ 0 ] is the message (maybe null); strs[ 1 ] is the expect pattern
    // (maybe null);
    private
    Throwable
    getFinalThrowable( final String[] strs,
                       final Class< ? extends Throwable > expctCls )
    {
        return
            getFinalThrowable(
                strs[ 1 ],
                expctCls,
                expctCls == null ? null : new MarkerException( strs[ 0 ] )
            );
    }

    @Test
    private
    void
    testGetFinalThrowableFailureMatches()
    {
        for ( String[] strs : 
                new String[][] {
                    new String[] { null, null },
                    new String[] { "test-message", null },
                    new String[] { "test-message", ".*" },
                    new String[] { "test-message", "test-message" },
                    new String[] { "test-message", "^test-messa[g].*$" }
                } )
        {
            state.isTrue( 
                getFinalThrowable( strs, MarkerException.class ) == null );

            state.isTrue( getFinalThrowable( strs, Exception.class ) == null );
        }
    }

    @Test
    private
    void
    testGetFinalThrowableOnNormalSuccess()
    {
        for ( String[] strs :
                new String[][] {
                    new String[] { null, null },
                    new String[] { null, ".*ignored.*" }
                } )
        {
            state.isTrue( getFinalThrowable( strs, null ) == null );
        }
    }

    @Test
    private
    void
    testGetFinalThrowableNoFailureExpected()
    {
        MarkerException me = new MarkerException( null );

        state.equal( me, getFinalThrowable( null, null, me ) );
        state.equal( me, getFinalThrowable( ".*ignored", null, me ) );
    }

    @Test
    private
    void
    testGetFinalThrowableUnexpectedFailType()
    {
        MarkerException cause = new MarkerException( null );

        Throwable th =
            getFinalThrowable( null, NumberFormatException.class, cause );
        
        state.isTrue( th instanceof IllegalStateException );
        state.equal( cause, th.getCause() );

        state.equal(
            "Expected throwable of type java.lang.NumberFormatException " +
            "but got one of type " +
            "com.bitgirder.test.TestUtilsTests$MarkerException (see cause)",
            th.getMessage()
        );
    }

    @Test
    private
    void
    testGetFinalThrowableUnexpectedMessage()
    {
        MarkerException cause = new MarkerException( "unexpected" );

        Throwable th =  
            getFinalThrowable( "^expected$", MarkerException.class, cause );

        state.isTrue( th instanceof IllegalStateException );
        state.equal( cause, th.getCause() );

        state.equal(
            "Throwable message does not match pattern '^expected$': " +
            "unexpected (See cause for actual exception)",
            th.getMessage()
        );
    }
}
