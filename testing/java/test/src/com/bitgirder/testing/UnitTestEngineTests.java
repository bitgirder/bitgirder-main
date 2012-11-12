package com.bitgirder.testing;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.PatternHelper;

import com.bitgirder.test.Test;
import com.bitgirder.test.Before;
import com.bitgirder.test.After;
import com.bitgirder.test.TestPhase;

import java.util.SortedSet;
import java.util.Map;

import java.util.regex.Pattern;

@Test
final
class UnitTestEngineTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static String[] FILTERABLE_NAMES = 
    {
        "com.bitgirder.testing.Filterable/testCall/call1",
        "com.bitgirder.testing.Filterable/testCall/call2",
        "com.bitgirder.testing.Filterable/inst1/test1",
        "com.bitgirder.testing.Filterable/inst1/test2",
        "com.bitgirder.testing.Filterable/inst2/test1",
        "com.bitgirder.testing.Filterable/inst2/test2",
        "com.bitgirder.testing.Filterable/test1",
        "com.bitgirder.testing.Filterable/test2",
        "com.bitgirder.testing.Filterable/static1/testCall/call1",
        "com.bitgirder.testing.Filterable/static1/testCall/call2",
        "com.bitgirder.testing.Filterable/static1/inst1/test1",
        "com.bitgirder.testing.Filterable/static1/inst1/test2",
        "com.bitgirder.testing.Filterable/static1/inst2/test1",
        "com.bitgirder.testing.Filterable/static1/inst2/test2",
        "com.bitgirder.testing.Filterable/static1/test1",
        "com.bitgirder.testing.Filterable/static1/test2",
        "com.bitgirder.testing.Filterable/static2/testCall/call1",
        "com.bitgirder.testing.Filterable/static2/testCall/call2",
        "com.bitgirder.testing.Filterable/static2/inst1/test1",
        "com.bitgirder.testing.Filterable/static2/inst1/test2",
        "com.bitgirder.testing.Filterable/static2/inst2/test1",
        "com.bitgirder.testing.Filterable/static2/inst2/test2",
        "com.bitgirder.testing.Filterable/static2/test1",
        "com.bitgirder.testing.Filterable/static2/test2"
    };

    private
    final
    static
    class NameAccumulator
    extends AbstractInvocationEventHandler
    {
        // sorted so we can traverse and, if necessary, fail in a deterministic
        // order
        private final SortedSet< String > names = Lang.newSortedSet();

        @Override
        public
        void
        invocationCompleted( InvocationDescriptor id,
                             Throwable th )
        {
            if ( th != null ) throw new RuntimeException( th );
            if ( id.getPhase() == TestPhase.TEST ) names.add( id.getName() );
        }

        private
        void
        assertMatches( String patStr )
        {
            Pattern pat = PatternHelper.compile( patStr );

            SortedSet< String > expct = Lang.newSortedSet();

            for ( String nm : FILTERABLE_NAMES )
            {
                if ( pat.matcher( nm ).matches() ) expct.add( nm );
            }

            state.equal( expct, names );
        }
    }

    private
    void
    assertFilterPattern( String pat,
                         String... expcts )
        throws Exception
    {
        NameAccumulator acc = new NameAccumulator();

        UnitTestEngine.Builder b =
            new UnitTestEngine.Builder().
                setPoolSize( 1 ).
                setEventHandler( acc ).
                setClassNames( 
                    Lang.asList( "com.bitgirder.testing.Filterable" ) );
        
        if ( pat != null ) b.setFilterPattern( pat );

        b.build().execute();
        acc.assertMatches( pat == null ? ".*" : pat );
    }

    @Test
    private
    void
    testFilterPattern()
        throws Exception
    {
        String[] pats = new String[] {
            null, 
            ".*", 
            "test1", 
            ".*test1",
            "XXXXXX",
            ".*Filterable.*",
            ".*/static1/.*", 
            ".*/inst2/.*", 
            ".*/call[12]$"
        };

        for ( String pat : pats ) assertFilterPattern( pat );
    }

    private
    final
    static
    class ResultAccumulator
    extends AbstractInvocationEventHandler
    {
        // Default identity equals() of keys is okay
        private final Map< InvocationDescriptor, Throwable > results =
            Lang.newMap();

        @Override
        public
        void
        invocationCompleted( InvocationDescriptor desc,
                             Throwable th )
        {
            results.put( desc, th );
        }
    }

    private
    void
    assertBeforeAbortedRun( ResultAccumulator acc )
    {
        state.equalInt( 1, acc.results.size() );

        InvocationDescriptor id = acc.results.keySet().iterator().next();
        state.equal( TestPhase.BEFORE, id.getPhase() );

        state.isTrue( 
            acc.results.get( id ) instanceof BeforeFailer.MarkerException );
    }

    @Test
    private
    void
    testBeforeFailureAbortsTestInstance()
        throws Exception
    {
        ResultAccumulator acc = new ResultAccumulator();

        new UnitTestEngine.Builder().
            setEventHandler( acc ).
            setClassNames( Lang.asList( BeforeFailer.class.getName() ) ).
            setPoolSize( 1 ).
            setInstantiationHandler(
                new UnitTestEngine.InstantiationHandler() {
                    public void instantiated( Object obj ) {
                        if ( obj instanceof BeforeFailer ) {
                            ( (BeforeFailer) obj ).doFail = true;
                        }
                    }
                }
            ).
            build().
            execute();
 
        assertBeforeAbortedRun( acc );
    }
}
