package com.bitgirder.core;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Strings;
import com.bitgirder.lang.PatternHelper;
import com.bitgirder.lang.Lang;

import com.bitgirder.lang.reflect.ReflectUtils;

import com.bitgirder.test.Test;
import com.bitgirder.test.TestUtils;

import java.lang.reflect.Method;
import java.lang.reflect.AnnotatedElement;
import java.lang.reflect.InvocationTargetException;

import java.util.regex.Pattern;

import java.util.List;

public
final
class UnitTestRunner
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final static String ARG_FILTER_PATTERN = "--filter-pattern";

    private final static Pattern DEFAULT_FILTER = PatternHelper.compile( ".*" );

    private final Pattern filterPattern;
    private final List< String > classNames;

    private
    UnitTestRunner( Builder b )
    {
        this.filterPattern = b.filterPattern;
        this.classNames = b.classNames;
    }

    private
    static
    void
    code( Object... args )
    {
        System.out.println( Strings.join( " ", args ) );
    }

    private
    final
    static
    class TestFailure
    {
        private final Method m;
        private final Throwable th;

        private
        TestFailure( Method m,
                     Throwable th )
        {
            this.m = m;
            this.th = th;
        }
    }

    private
    final
    static
    class ResultAcc
    {
        private final List< Method > successes = Lang.newList();
        private final List< TestFailure > failed = Lang.newList();
    }

    private
    String
    fqName( Method m )
    {
        return m.getDeclaringClass().getName() + "." + m.getName();
    }

    private
    boolean
    acceptFilter( Method m )
    {
        return filterPattern.matcher( fqName( m ) ).matches();
    }

    private
    Test
    getTest( AnnotatedElement t )
    {
        return state.notNull( t.getAnnotation( Test.class ) );
    }

    private
    void
    checkTestSignature( Method m )
    {
        state.isTrue( 
            m.getParameterTypes().length == 0,
            "Test method has parameters:", m );
    }

    private
    void
    accumulateResult( Method m,
                      Throwable th,
                      ResultAcc acc )
    {
        th = TestUtils.getFinalThrowable( getTest( m ), th );

        if ( th == null ) acc.successes.add( m );
        else acc.failed.add( new TestFailure( m, th ) );
    }

    private
    void
    invokeTest( Object inst,
                Method m,
                ResultAcc acc )
        throws Exception
    {
        code( "Invoking", fqName( m ) );

        try 
        {
            m.invoke( inst, (Object[]) null );
            accumulateResult( m, null, acc );
        }
        catch ( InvocationTargetException ex ) 
        { 
            accumulateResult( m, ex.getCause(), acc );
        }
    }

    private
    void
    runTests( Class< ? > cls,
              ResultAcc acc )
        throws Exception
    {
        Object inst = ReflectUtils.newInstance( cls );

        for ( Method m : ReflectUtils.getDeclaredMethods( cls, Test.class ) )
        {
            m.setAccessible( true );
            checkTestSignature( m );
            if ( acceptFilter( m ) ) invokeTest( inst, m, acc );
        }
    }

    private
    int
    reportResults( ResultAcc acc )
    {
        for ( Method m : acc.successes )
        {
            System.out.println( fqName( m ) + " succeeded" );
        }

        for ( TestFailure tf : acc.failed )
        {
            System.out.println( fqName( tf.m ) + " failed:" );
            tf.th.printStackTrace( System.out );
        }

        System.out.println(
            acc.successes.size() + " tests succeeded; " + 
            acc.failed.size() + " failed." );

        return acc.failed.isEmpty() ? 0 : 1;
    }

    private
    void
    run()
        throws Exception
    {
        List< Class< ? > > classes = 
            TestUtils.getTestClasses( 
                TestUtils.getClassesForNames( classNames ) );
 
        ResultAcc acc = new ResultAcc();

        for ( Class< ? > cls : classes ) runTests( cls, acc );

        int exitRes = reportResults( acc );
        System.exit( exitRes );
    }

    private
    static
    Builder
    createBuilder( String[] args )
    {
        Builder res = new Builder();

        for ( int i = 0, e = args.length; i < e; )
        {
            String arg = args[ i++ ];

            if ( arg.startsWith( "--" ) )
            {
                if ( arg.equals( ARG_FILTER_PATTERN ) )
                {
                    state.isFalse( 
                        i == e, ARG_FILTER_PATTERN, "requires a value" );
                    res.filterPattern = PatternHelper.compile( args[ i++ ] );
                }
                else inputs.fail( "Unrecognized argument:", arg );
            }
            else res.classNames.add( arg );
        }

        return res;
    }

    private
    static
    void
    validate( Builder b )
    {
        inputs.isFalse( 
            b.classNames.isEmpty(), "Need at least one test class" );
    }

    public
    static
    void
    main( String[] args )
        throws Exception
    {
        Builder b = createBuilder( args );
        validate( b );
 
        UnitTestRunner utr = new UnitTestRunner( b );
        utr.run();
    }

    private
    final
    static
    class Builder
    {
        private Pattern filterPattern = DEFAULT_FILTER;
        private List< String > classNames = Lang.newList();
    }
}
