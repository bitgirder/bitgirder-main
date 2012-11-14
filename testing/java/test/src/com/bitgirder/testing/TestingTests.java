package com.bitgirder.testing;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.StandardThread;
import com.bitgirder.lang.Strings;
import com.bitgirder.lang.Completion;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.test.Test;
import com.bitgirder.test.TestCall;
import com.bitgirder.test.Tests;
import com.bitgirder.test.TestFactory;
import com.bitgirder.test.After;
import com.bitgirder.test.Before;
import com.bitgirder.test.LabeledTestObject;
import com.bitgirder.test.LabeledTestCall;
import com.bitgirder.test.InvocationFactory;
import com.bitgirder.test.TestFailureExpector;

import java.util.List;
import java.util.Set;

// Runs as a standard test as part of every test run and uses the Set expects to
// track that exactly the tests we expect to run do (and only run once). This
// class doesn't cover all testing mechanisms. Others not so easily covered here
// are covered in UnitTestEngineTests.
@Test
final
class TestingTests
{
    private static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static Set< Object > expects = Lang.newSynchronizedSet();
    private final static Set< String > stopNames = Lang.newSynchronizedSet();

    private boolean beforeRan;

    private final static List< FailContext > failContexts;

    private
    static
    void
    complete( Object... objs )
    {
        state.remove( expects, Lang.asList( objs ), "expects" );
    }

    private
    static
    void
    expect( Object... objs )
    {
        state.isTrue( expects.add( Lang.asList( objs ) ) );
    }

    static 
    {
        for ( String s : new String[] { "obj1", "obj2" } )
        {
            expect( s, "test1", ParameterizedTestObject.class );
        }

        expect( "obj2", "call-1" );
        expect( "test1" );
        expect( "testExpectedFailureNoMessage" );
        expect( "testExpectedFailureMessageMatch" );
        expect( "test-befores-precede-inv-fact" );
        expect( "inv1-call" );
        expect( "nested-succ-call" );
        expect( "nested-fail-call" );
        expect( "nested-descendant" );
    }

    static
    {
        Class< ? extends Throwable > cls = MarkerException.class;

        failContexts =
            Lang.< FailContext >asList(
                new FailContext( null, ".*", cls ),
                new FailContext( null, null, cls ),
                new FailContext( "test-msg", null, cls ),
                new FailContext( "test-msg", "test-m[st]g$", cls ),
                new FailContext( "test-msg", "test-msg", cls ),
                new FailContext( "test-msg", "test-msg", Exception.class ),
                new FailContext( null, null, null ) // should pass
            );

        for ( FailContext ctx : failContexts )
        {
            expect( makeExpectName( FailingCall.class, ctx ) );
        }
    }

    // We add this only in the @Before so as not to emit a warning at the
    // conclusion of a test run which detected and loaded this class's runtime
    // processes but had no intent of actually running tests in this class
    private
    void
    setStopNamesCheck()
    {
        Runtime.getRuntime().addShutdownHook(
            new StandardThread( "stoppable-chk-hook-%1$d" ) {
                public void run() 
                {
                    if ( ! stopNames.isEmpty() )
                    {
                        CodeLoggers.warn( 
                            "One or more stops did not run:", stopNames );
                        
                        Runtime.getRuntime().halt( 1 );
                    }
                }
            }
        );
    }

    @Before 
    private 
    void 
    before() 
    { 
        setStopNamesCheck();
        beforeRan = true; 
    }

    @Test
    private
    final
    class NestedSuccessCallTest
    implements TestCall
    {
        public void call() { complete( "nested-succ-call" ); }
    }

    @Test( expected = MarkerException.class )
    private
    final
    class NestedFailCallTest
    implements TestCall
    {
        public
        void
        call()
            throws Exception
        {
            complete( "nested-fail-call" );
            throw new MarkerException();
        }
    }

    private
    abstract
    class AbstractCall
    implements TestCall
    {}

    @Test
    private
    final
    class NestedDescendantTest
    extends AbstractCall
    {
        public void call() { complete( "nested-descendant" ); }
    }

    // Regression to ensure that befores run before invocation factories, making
    // before results available to InvocationFactory methods
    @InvocationFactory
    private
    List< ? >
    testBeforesRunBeforeInvocationFactory()
    {
        final boolean pass = beforeRan;

        return Lang.< LabeledTestCall >asList(
            new LabeledTestCall( "test-befores-precede-inv-fact" ) {
                public void call()
                {
                    state.isTrue( pass );
                    complete( "test-befores-precede-inv-fact" );
                }
            }
        );
    }

    private
    final
    class TestCallImpl
    implements TestCall
    {
        private final Object[] expct;

        private TestCallImpl( Object... expct ) { this.expct = expct; }

        public void call() { complete( expct ); }
    }

    @InvocationFactory
    private
    List< LabeledTestObject >
    invocables()
    {
        return
            Lang.< LabeledTestObject >asList(
                Tests.createLabeledTestObject( 
                    new TestCallImpl( "inv1-call" ), "inv1-call" )
            );
    }

    private
    final
    class ParameterizedTestObject
    {
        private final String objName;

        private
        ParameterizedTestObject( String objName ) 
        { 
            this.objName = objName; 
        }

        @Test private void test1() { complete( objName, "test1", getClass() ); }
    }

    private
    final
    class ParameterizedTestObject2
    {
        private final String objName;

        private
        ParameterizedTestObject2( String objName )
        {
            this.objName = objName;
        }

        @InvocationFactory
        private
        List< LabeledTestObject >
        invocations()
        {
            return
                Lang.asList(
                    Tests.createLabeledTestObject(
                        new TestCallImpl( objName, "call-1" ), "call-1" )
                );
        }
    }

    private
    Object
    testObj( String lbl )
    {
        Object o = new ParameterizedTestObject( lbl );

        return Tests.createLabeledTestObject( o, lbl );
    }

    @TestFactory
    private
    List< Object >
    getTestObjects()
    {
        state.isTrue( beforeRan );

        return Lang.< Object >asList( testObj( "obj1" ), testObj( "obj2" ) );
    }

    @TestFactory
    private
    List< LabeledTestObject >
    getTestObjects2()
    {
        return 
            Lang.< LabeledTestObject >singletonList(
                Tests.createLabeledTestObject(
                    new ParameterizedTestObject2( "obj2" ),
                    "obj2"
                )
            );
    }

    @Test 
    private 
    void 
    test1() 
    { 
        state.isTrue( beforeRan );
        complete( "test1" );
    }

    private 
    final 
    static 
    class MarkerException 
    extends Exception 
    {
        private MarkerException() {}
        private MarkerException( String msg ) { super( msg ); }
    }

    @Test( expected = MarkerException.class )
    private
    void
    testExpectedFailureNoMessage()
        throws Exception
    {
        complete( "testExpectedFailureNoMessage" );
        throw new MarkerException();
    }

    @Test( expected = MarkerException.class,
           expectedPattern = "^expect-this\\d+$" )
    private
    void
    testExpectedFailureMessageMatch()
        throws Exception
    {
        complete( "testExpectedFailureMessageMatch" );
        throw new MarkerException( "expect-this123" );
    }

    private
    final
    static
    class FailContext
    implements TestFailureExpector
    {
        private final String msg;
        private final String pat;
        private final Class< ? extends Throwable > expctCls;

        private
        FailContext( String msg,
                     String pat,
                     Class< ? extends Throwable > expctCls )
        { 
            this.msg = msg;
            this.pat = pat;
            this.expctCls = expctCls;
        }

        public 
        Class< ? extends Throwable > 
        expectedFailureClass()
        {
            return expctCls;
        }

        public String expectedFailurePattern() { return pat; }

        Exception
        createFailure()
        {
            if ( expctCls == null ) return null;
            else return new MarkerException( msg );
        }
    }

    private
    static
    CharSequence
    makeLabel( Object obj,
               FailContext ctx )
    {
        String expctClsStr =
            ctx.expctCls == null ? "null" : ctx.expctCls.getSimpleName();

        return
            new StringBuilder().
                append( obj.getClass().getSimpleName() ).
                append( "/" ).
                append(
                    Strings.crossJoin( "=", ",",
                        "msg", ctx.msg,
                        "pat", ctx.pat,
                        "expctCls", expctClsStr
                    )
                );
    }

    private
    static
    Object[]
    makeExpectName( Class< ? > cls,
                    FailContext ctx )
    {
        return new Object[] { cls, ctx.msg, ctx.pat, ctx.expctCls };
    }

    private
    final
    static
    class FailingCall
    extends LabeledTestCall
    implements TestFailureExpector
    {
        private final FailContext ctx;

        private
        FailingCall( FailContext ctx )
        { 
            super( makeLabel( FailingCall.class, ctx ) );

            this.ctx = ctx;
            expectFailure( ctx.expctCls, ctx.pat );
        }

        public
        void
        call()
            throws Exception
        {
            complete( makeExpectName( getClass(), ctx ) );

            Exception ex = ctx.createFailure();
            if ( ex != null ) throw ex;
        }
    }

    @InvocationFactory
    private
    List< LabeledTestObject >
    testExpectorFailures()
    {
        List< LabeledTestObject > res = Lang.newList();

        for ( FailContext ctx : failContexts )
        {
            res.add( new FailingCall( ctx ) );
        }

        return res;
    }

    @After
    private
    void
    assertTestsRan()
    {
        stopNames.remove( "assertTestsRan" );
        state.isTrue( expects.isEmpty(), "expects:", expects );
    }
}
