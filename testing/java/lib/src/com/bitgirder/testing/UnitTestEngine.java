package com.bitgirder.testing;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.PatternHelper;
import com.bitgirder.lang.Lang;
import com.bitgirder.lang.ObjectReceiver;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ImmutableListPath;
import com.bitgirder.lang.path.ObjectPaths;

import com.bitgirder.lang.reflect.ReflectUtils;
import com.bitgirder.lang.reflect.MethodInvocation;

import com.bitgirder.test.Test;
import com.bitgirder.test.TestPhase;
import com.bitgirder.test.Before;
import com.bitgirder.test.After;
import com.bitgirder.test.TestCall;
import com.bitgirder.test.TestFactory;
import com.bitgirder.test.TestFailureExpector;
import com.bitgirder.test.TestUtils;
import com.bitgirder.test.LabeledTestObject;
import com.bitgirder.test.TestFailureExpectation;
import com.bitgirder.test.InvocationFactory;

import com.bitgirder.concurrent.Duration;
import com.bitgirder.concurrent.Concurrency;

import java.lang.annotation.Annotation;

import java.lang.reflect.Method;
import java.lang.reflect.Constructor;
import java.lang.reflect.AnnotatedElement;

import java.util.List;
import java.util.Collection;
import java.util.Iterator;

import java.util.regex.Pattern;

import java.util.concurrent.BlockingQueue;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

import java.util.concurrent.atomic.AtomicInteger;

public
final
class UnitTestEngine
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final static int DEFAULT_POOL_SIZE = 10;

    private final static Pattern DEFAULT_FILTER_PATTERN = 
        PatternHelper.compile( ".*" );

    private final Pattern filterPattern;
    private final List< String > classNames;
    private final int poolSize;
    private final InvocationEventHandler eh;
    private final InstantiationHandler instHndlr;

    private final AtomicInteger active = new AtomicInteger();

    private final BlockingQueue< TestInstance > advanceable = 
        Lang.newBlockingQueue();

    private final Duration poolStopWait = Duration.fromSeconds( 5 );

    private ExecutorService pool;

    private
    UnitTestEngine( Builder b )
    {
        this.filterPattern = inputs.notNull( b.filterPattern, "filterPattern" );

        this.classNames = inputs.noneNull( b.classNames, "classNames" );
        inputs.isFalse( classNames.isEmpty(), "Need at least one test class" );

        this.poolSize = b.poolSize;
        this.eh = inputs.notNull( b.eh, "eh" );
        this.instHndlr = b.instHndlr;
    }

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private
    List< Class< ? > >
    getTestClasses()
        throws Exception
    {
        return TestUtils.getClassesForNames( classNames );
    }

    private
    static
    abstract
    class AbstractInvocable
    implements Invocable
    {
        private final ObjectPath< CharSequence > path;

        private
        AbstractInvocable( ObjectPath< CharSequence > path )
        {
            this.path = path;
        }

        public
        final
        CharSequence 
        getName() 
        {
            return
                ObjectPaths.
                    format( path, ObjectPaths.SLASH_FORMATTER ).toString();
        }
    }

    private
    static
    MethodInvocation
    createTestMethodInvocation( Method m,
                                Object inst )
    {
        MethodInvocation.Builder b = new MethodInvocation.Builder();

        b.setIgnoreUnmatchedKeys( true );
        b.setTarget( m );
        if ( inst != null ) b.setInstance( inst );

        Class< ? >[] typs = m.getParameterTypes();

        state.isTrue( 
            typs.length == 0, "Method should have no parameters:", m );
        
        return b.build();
    }

    private
    final
    static
    class DirectMethodInvocable
    extends AbstractInvocable
    {
        private final Method m;

        private
        DirectMethodInvocable( Method m,
                               ObjectPath< CharSequence > path )
        {
            super( path );

            this.m = m;
        }

        public
        void
        invoke( Object invTarg )
            throws Exception
        {
            invokeTestMethod( m, invTarg );
        }

        public 
        TestFailureExpectation 
        getFailureExpectation() 
        { 
            Test t = m.getAnnotation( Test.class );
            return t == null ? null : TestUtils.failureExpectationFor( t );
        }
    }

    private
    final
    class NestedCallInvocable
    extends AbstractInvocable
    {
        private final Constructor< ? > cons;

        private
        NestedCallInvocable( Constructor< ? > cons,
                             ObjectPath< CharSequence > path )
        {
            super( path );

            this.cons = cons;
        }

        public
        void
        invoke( Object invTarg )
            throws Exception
        {
            Object call = ReflectUtils.invoke( cons, invTarg );

            ( (TestCall) call ).call();
        }

        public
        TestFailureExpectation
        getFailureExpectation()
        {
            return 
                TestUtils.failureExpectationFor( 
                    cons.getDeclaringClass().getAnnotation( Test.class ) );
        }
    }

    private
    static
    ObjectPath< CharSequence >
    getRootPath( Class< ? > cls )
    {
        return ObjectPath.getRoot( (CharSequence) cls.getName() );
    }

    private
    static
    Object
    invokeTestMethod( Method m,
                      Object targ )
        throws Exception
    {
        MethodInvocation mi = createTestMethodInvocation( m, targ );

        return mi.invoke( Lang.emptyMap() );
    }

    private
    void
    castAndAddLabeledTestObject( Object o,
                                 List< LabeledTestObject > l,
                                 Method src )
    {
        if ( o instanceof LabeledTestObject ) l.add( (LabeledTestObject) o );
        else
        {
            state.fail( 
                "List element", o, " at index", l.size(), "in result from",
                src.getName(), "is not a LabeledTestObject"
            );
        }
    }

    // targ will be null when invoking a static factory method
    private
    List< LabeledTestObject >
    expectLabeledTestObjects( Method m,
                              Object targ )
        throws Exception
    {
        state.notNull( m, "m" );

        m.setAccessible( true );
        List< ? > l1 = (List< ? >) invokeTestMethod( m, targ );

        List< LabeledTestObject > res = Lang.newList( l1.size() );

        for ( Object o : l1 )
        {
            castAndAddLabeledTestObject( o, res, m );
        }

        return res;
    }

    private
    static
    CharSequence
    format( ObjectPath< ? extends CharSequence > path )
    {
        return ObjectPaths.format( path, ObjectPaths.DOT_FORMATTER );
    }

    private
    < V >
    V
    notNull( V obj,
             String methName,
             ObjectPath< String > errPath )
    {
        state.isFalse( 
            obj == null, methName, "returned null for", format( errPath ) );

        return obj;
    }

    private
    LabeledTestObject
    validate( LabeledTestObject lto,
              ObjectPath< String > errPath )
    {
        notNull( lto.getLabel(), "getLabel()", errPath );
        notNull( lto.getInvocationTarget(), "getInvocationTarget()", errPath );
        
        return lto;
    }

    // invTarg will be null when invoking static test factories
    private
    List< LabeledTestObject >
    validateAndCollectTestFactories( List< Method > testFactories,
                                     Object invTarg )
        throws Exception
    {
        List< LabeledTestObject > res = Lang.newList();

        for ( Method m : testFactories )
        {
            List< LabeledTestObject > testObjs =
                expectLabeledTestObjects( m, invTarg );

            ImmutableListPath< String > errPath = 
                ObjectPath.getRoot( m.getName() ).startImmutableList();

            for ( LabeledTestObject lto : testObjs )
            {
                res.add( validate( lto, errPath ) );
                errPath = errPath.next();
            }
        }

        return res;
    }

    private
    void
    invocationStarted( InvocationDescriptor id )
    {
        if ( eh != null )
        {
            long tm = System.currentTimeMillis();
            synchronized ( eh ) { eh.invocationStarted( id, tm ); }
        }
    }

    private
    void
    invocationCompleted( InvocationDescriptor id,
                         Throwable failure )
    {
        if ( eh != null )
        {
            long tm = System.currentTimeMillis();
            synchronized ( eh ) { eh.invocationCompleted( id, failure, tm ); }
        }
    }
    
    private
    static
    void
    checkDoubleFailureExpectation( Object obj,
                                   TestFailureExpectation fe1,
                                   TestFailureExpectation fe2 )
    {
        if ( ! ( fe1 == null || fe2 == null ) )
        {
            throw state.createFail(
                "Object", obj,
                "has duplicate failure expectations (both from @Test and by " +
                "being a failure expector"
            );
        }
    }

    private
    final
    static
    class TestCallInvocable
    extends AbstractInvocable
    {
        private final TestCall c;

        private
        TestCallInvocable( TestCall c,
                           ObjectPath< CharSequence > path )
        {
            super( path );

            this.c = c;
        }

        public
        void
        invoke( Object invTarg )
            throws Exception 
        { 
            c.call();
        }

        public
        TestFailureExpectation
        getFailureExpectation()
        {
            TestFailureExpectation fe1 = null;
    
            if ( c instanceof TestFailureExpector )
            {
                TestFailureExpector tfe = (TestFailureExpector) c;
                fe1 = TestUtils.failureExpectationFor( tfe );
            }
    
            TestFailureExpectation fe2 = null;
            Test t = c.getClass().getAnnotation( Test.class );
            if ( t != null ) fe2 = TestUtils.failureExpectationFor( t );
    
            checkDoubleFailureExpectation( c, fe1, fe2 );
            return fe1 == null ? fe2 : fe1; // could be null result regardless
        }
    }

    private
    final
    class TestInstance
    {
        private final Object invTarg;
        private final ObjectPath< CharSequence > path;

        private final List< Invocable > befores = Lang.newList();
        private final List< Invocable > tests = Lang.newList();
        private final List< Invocable > afters = Lang.newList();
        private final List< Method > testFacts = Lang.newList();
        private final List< Method > invFacts = Lang.newList();

        private final AtomicInteger waitCount = new AtomicInteger();
        private volatile boolean hadBeforeFailure;

        private volatile boolean aftersStarted;

        private TestInstance parent;

        private
        TestInstance( Object invTarg,
                      ObjectPath< CharSequence > path )
        {
            this.invTarg = invTarg;
            this.path = path;
        }

        private Class< ? > cls() { return invTarg.getClass(); }

        private
        void
        instanceComplete()
        {
            if ( parent != null ) 
            {
                parent.complete( "[child: " + format( path ) + "]" );
            }
        } 

        private
        void
        complete( CharSequence dbg )
        {
            if ( waitCount.decrementAndGet() == 0 ) 
            {
                if ( aftersStarted ) instanceComplete();
                else reactivate( this );
            }
        }

        private
        final
        class Invoke
        implements Runnable
        {
            private final Invocable inv;
            private final InvocationDescriptor desc;

            private
            Invoke( Invocable inv,
                    TestPhase phase )
            {
                this.inv = inv;
                this.desc = new InvocationDescriptor( inv.getName(), phase );
            }

            private
            void
            sendCompletion( Throwable failure )
            {
                invocationCompleted( desc, failure );

                if ( desc.getPhase() == TestPhase.BEFORE && failure != null )
                {
                    hadBeforeFailure = true;
                }

                complete( desc.getName() );
            }

            public
            void
            run()
            {
                Throwable failure = null;

                try 
                {
                    invocationStarted( desc );
                    inv.invoke( invTarg );
                }
                catch ( Throwable th )
                {
                    TestFailureExpectation fe = inv.getFailureExpectation();

                    if ( fe == null ) failure = th;
                    else failure = TestUtils.getFinalThrowable( fe, th );
                }
 
                sendCompletion( failure );
            }
        }

        private
        void
        resetInvocationState()
        {
            state.isTrue( waitCount.get() == 0,
                "Attempt to reset invocation state when wait count is",
                waitCount
            );
        }

        private
        void
        advanceInvocations( List< Invocable > invs,
                            TestPhase phase )
            throws InterruptedException
        {
            // Need to add all wait counts before submitting any tasks, since
            // otherwise tasks might finish at the same rate we add them,
            // causing complete() to prematurely see waitCount == 0
            waitCount.addAndGet( invs.size() );

            for ( Invocable inv : invs ) 
            {
                pool.submit( new Invoke( inv, phase ) );
            }

            invs.clear();
        }

        private
        boolean
        advanceBefores()
            throws InterruptedException
        {
            resetInvocationState();
            advanceInvocations( befores, TestPhase.BEFORE );

            return false;
        }
        
        // Side effect: clears testFacts
        private
        List< TestInstance >
        getChildTests()
            throws Exception
        {
            List< TestInstance > res =
                invokeTestFactories( testFacts, path, invTarg );

            for ( TestInstance ti : res ) ti.parent = this;

            testFacts.clear();

            return res;
        }
        
        private
        Invocable
        asInvocable( LabeledTestObject lto,
                     ObjectPath< CharSequence > path )
        {
            Object obj = lto.getInvocationTarget();
 
            path = path.descend( lto.getLabel() );
    
            if ( obj instanceof TestCall )
            {
                return new TestCallInvocable( (TestCall) obj, path );
            }
            else 
            {
                throw state.createFail( 
                    "Unrecognized invocation object:", obj );
            }
        }

        private
        List< Invocable >
        getInvocationFactoryInvocables()
            throws Exception
        {
            List< Invocable > res = Lang.newList();

            for ( Method m : invFacts )
            {
                List< LabeledTestObject > l = 
                    expectLabeledTestObjects( m, invTarg );

                ObjectPath< CharSequence > methPath = 
                    path.descend( m.getName() );
                
                for ( LabeledTestObject lto : l )
                {
                    res.add( asInvocable( lto, methPath ) );
                }
            }

            return res;
        }

        // Side effect: clears tests and invFacts
        private
        List< Invocable >
        getTestInvocables()
            throws Exception
        {
            List< Invocable > res = Lang.newList();

            res.addAll( tests );
            tests.clear();

            res.addAll( getInvocationFactoryInvocables() );
            invFacts.clear();

            filterTests( res );
            
            return res;
        }

        private
        void
        filterTests( List< Invocable > invs )
        {
            if ( filterPattern == null ) return;

            Iterator< Invocable > it = invs.iterator();

            while ( it.hasNext() )
            {
                Invocable inv = it.next();

                if ( ! filterPattern.matcher( inv.getName() ).matches() ) 
                {
                    it.remove();
                }
            }
        }

        // We first build up a list of all invocables and test factories in
        // order to ensure that we properly increment waitCount before any tasks
        // are submitted, in order to avoid letting waitCount hit zero while we
        // are still advancing
        private
        boolean
        advanceTests()
            throws Exception
        {
            resetInvocationState();

            List< Invocable > invs = getTestInvocables();
            List< TestInstance > children = getChildTests();

            if ( children.isEmpty() && invs.isEmpty() ) reactivate( this );
            else
            {
                waitCount.addAndGet( children.size() );
                advanceInvocations( invs, TestPhase.TEST );

                for ( TestInstance chld : children ) activate( chld );
            }

            return false;
        }

        private
        boolean
        advanceAfters()
            throws InterruptedException
        {
            resetInvocationState();

            // order is important here, since we need aftersStarted marked true
            // before advanceInvocations(), which might submit after tasks that
            // will complete alongside execution of this method body.
            aftersStarted = true; 
            
            if ( afters.isEmpty() ) instanceComplete();
            else advanceInvocations( afters, TestPhase.AFTER );

            return true;
        }

        // returns true when this instance is done and should no longer be
        // considered active or advanced
        private
        boolean
        advance()
            throws Exception
        {
            if ( ! befores.isEmpty() ) return advanceBefores();

            if ( hadBeforeFailure ) return true; // don't go to TEST/AFTER

            if ( tests.size() + invFacts.size() + testFacts.size() > 0 )
            {
                return advanceTests();
            }

            return advanceAfters();
        }
    }

    // Returns 1-arg constructor for nested; the arg is the implicit enclosing
    // instance
    private
    Constructor< ? >
    validateNestedTestClass( Class< ? > nested )
        throws Exception
    {
        state.isTruef( 
            TestCall.class.isAssignableFrom( nested ),
            "%s is not a TestCall", nested );
        
        state.isFalsef(
            ReflectUtils.isStatic( nested ), "%s is static", nested );
        
        Class< ? > enc = nested.getEnclosingClass();
        return ReflectUtils.getDeclaredConstructor( nested, enc );
    } 

    private
    void
    setNestedTestCalls( TestInstance ti )
        throws Exception
    {
        Class< ? > cls = ti.invTarg.getClass();

        Collection< Class< ? > > coll = 
            ReflectUtils.getDeclaredAncestorClasses( cls, Test.class );

        for ( Class< ? > nested : coll )
        {
            Constructor< ? > cons = validateNestedTestClass( nested );

            ObjectPath< CharSequence > path = 
                ti.path.descend( nested.getSimpleName() );

            ti.tests.add( new NestedCallInvocable( cons, path ) );
        }
    }

    // Returned list can be modified by caller
    private
    List< Method >
    getMethods( Class< ? > cls,
                Class< ? extends Annotation > annCls )
    {
        Collection< Method > res =
            ReflectUtils.getDeclaredAncestorMethods( cls, annCls );

        for ( Method m : res ) m.setAccessible( true );
        
        return Lang.newList( res );
    }

    private
    void
    setTestFactories( List< Method > facts,
                      Class< ? > cls,
                      boolean wantStatic )
    {
        inputs.notNull( cls, "cls" );

        for ( Method m : getMethods( cls, TestFactory.class ) )
        {
            if ( wantStatic == ReflectUtils.isStatic( m ) ) facts.add( m );
        }
    }

    private
    void
    setInvocables( List< Invocable > l,
                   TestInstance ti,
                   Class< ? extends Annotation > annCls )
    {
        Collection< Method > methods = 
            ReflectUtils.getDeclaredAncestorMethods( ti.cls(), annCls );

        for ( Method m : methods )
        {
            m.setAccessible( true );
        
            ObjectPath< CharSequence > path = ti.path.descend( m.getName() );

            l.add( new DirectMethodInvocable( m, path ) );
        }
    }

    private
    void
    setInvocationFactories( TestInstance ti )
    {
        ti.invFacts.addAll( getMethods( ti.cls(), InvocationFactory.class ) );
    }
    
    private
    TestInstance
    initTestInstance( TestInstance ti )
        throws Exception
    {
        setInvocables( ti.tests, ti, Test.class );
        setNestedTestCalls( ti );
        setTestFactories( ti.testFacts, ti.cls(), false );
        setInvocationFactories( ti );
        setInvocables( ti.befores, ti, Before.class );
        setInvocables( ti.afters, ti, After.class );

        return ti;
    }

    private
    Constructor< ? >
    getTargetInstanceConstructor( Class< ? > cls )
    {
        Constructor< ? >[] consArr = cls.getDeclaredConstructors();

        state.isTrue( 
            consArr.length == 1 && consArr[ 0 ].getParameterTypes().length == 0,
            cls, "should have only a single no-arg constructor" );

        return consArr[ 0 ];
    }

    private
    List< TestInstance >
    initTestInstances( List< LabeledTestObject > testObjs,
                       ObjectPath< CharSequence > basePath )
        throws Exception
    {
        List< TestInstance > res = Lang.newList( testObjs.size() );

        for ( LabeledTestObject lto : testObjs )
        {
            Object invTarg = lto.getInvocationTarget();

            ObjectPath< CharSequence > ltoPath = 
                basePath.descend( lto.getLabel() );
            
            res.add( initTestInstance( new TestInstance( invTarg, ltoPath ) ) );
        }

        return res;
    }

    private
    List< TestInstance >
    invokeTestFactories( List< Method > facts,
                         ObjectPath< CharSequence > basePath,
                         Object invTarg )
        throws Exception
    {
        List< LabeledTestObject > testObjs =
            validateAndCollectTestFactories( facts, invTarg );

        return initTestInstances( testObjs, basePath );
    }

    private
    Object
    instantiateTarget( Class< ? > cls )
        throws Exception
    {
        Constructor< ? > cons = getTargetInstanceConstructor( cls );
        cons.setAccessible( true );

        Object res = ReflectUtils.invoke( cons, new Object[] {} );

        if ( instHndlr != null ) instHndlr.instantiated( res );

        return res;
    }

    private
    void
    initTestAnnotatedClasses( List< Class< ? > > classes )
        throws Exception
    {
        for ( Class< ? > cls : TestUtils.getTestClasses( classes ) )
        {
            Object invTarg = instantiateTarget( cls );

            ObjectPath< CharSequence > path = getRootPath( invTarg.getClass() );
            TestInstance ti = new TestInstance( invTarg, path );

            activate( initTestInstance( ti ) );
        }
    }

    private
    void
    initStaticFactories( List< Class< ? > > classes )
        throws Exception
    {
        for ( Class< ? > cls : classes )
        {
            List< Method > facts = Lang.newList();
            setTestFactories( facts, cls, true );
 
            ObjectPath< CharSequence > path = getRootPath( cls );

            List< TestInstance > l = invokeTestFactories( facts, path, null );
            for ( TestInstance ti : l ) activate( ti );
        }
    }

    private
    void
    activate( TestInstance ti )
    {
        active.incrementAndGet();
        advanceable.add( ti );
    }

    private
    void
    reactivate( TestInstance ti )
    {
        advanceable.add( ti );
    }

    private
    void
    processNext()
        throws Exception
    {
        state.isTrue( active.get() > 0 );
        
        TestInstance ti = advanceable.take();
        
        if ( ti.advance() ) active.decrementAndGet();
    }

    public
    void
    execute()
        throws Exception
    {
        List< Class< ? > > classes = getTestClasses();

        initTestAnnotatedClasses( classes );
        initStaticFactories( classes );

        pool = Executors.newFixedThreadPool( poolSize );

        while ( active.get() > 0 ) processNext();

        Concurrency.shutdownAndWait( pool, poolStopWait );
    }

    static
    interface InstantiationHandler
    {
        public
        void
        instantiated( Object obj )
            throws Exception;
    }

    public
    final
    static
    class Builder
    {
        private Pattern filterPattern = DEFAULT_FILTER_PATTERN;
        private List< String > classNames = Lang.emptyList();
        private List< Class< ? > > classes = Lang.emptyList();
        private int poolSize = DEFAULT_POOL_SIZE;
        private InvocationEventHandler eh;
        private InstantiationHandler instHndlr;

        public
        Builder
        setFilterPattern( Pattern filterPattern )
        {
            this.filterPattern = 
                inputs.notNull( filterPattern, "filterPattern" );

            return this;
        }

        public
        Builder
        setFilterPattern( CharSequence filterPattern )
        {
            return setFilterPattern( PatternHelper.compile( filterPattern ) );
        }

        public
        Builder
        setClassNames( List< String > classNames )
        {
            this.classNames = inputs.notNull( classNames, "classNames" );
            return this;
        }

        public
        Builder
        setPoolSize( int poolSize )
        {
            this.poolSize = inputs.positiveI( poolSize, "poolSize" );
            return this;
        }

        public
        Builder
        setEventHandler( InvocationEventHandler eh )
        {
            this.eh = inputs.notNull( eh, "eh" );
            return this;
        }

        // Only exposed currently for testing so package level
        Builder
        setInstantiationHandler( InstantiationHandler instHndlr )
        {
            this.instHndlr = inputs.notNull( instHndlr, "instHndlr" );
            return this;
        }

        public UnitTestEngine build() { return new UnitTestEngine( this ); }
    }
}
