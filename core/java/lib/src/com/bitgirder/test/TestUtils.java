package com.bitgirder.test;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.PatternHelper;

import java.util.List;

import java.util.regex.Pattern;

public
final
class TestUtils
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private TestUtils() {}

    public
    static
    List< Class< ? > >
    getClassesForNames( List< ? extends CharSequence > classNames )
        throws ClassNotFoundException
    {
        inputs.noneNull( classNames, "classNames" );

        List< Class< ? > > res = Lang.newList( classNames.size() );

        for ( CharSequence nm : classNames ) 
        {
            res.add( Class.forName( nm.toString() ) );
        }

        return res;
    }

    public
    static
    List< Class< ? > >
    getTestClasses( List< Class< ? > > classes )
        throws ClassNotFoundException
    {
        inputs.noneNull( classes, "classes" );
 
        List< Class< ? > > res = Lang.newList();

        for ( Class< ? > cls : classes )
        {
            if ( cls.isAnnotationPresent( Test.class ) ) res.add( cls );
        }

        return res;
    }

    private
    static
    Throwable
    invocationSucceeded( Test test )
    {
        Class< ? extends Throwable > expctCls = test.expected();

        if ( expctCls.equals( Test.NoneExpectedException.class ) ) 
        {
            return null;
        }
        else
        {
            return
                new IllegalStateException( 
                    "Expected exception of type " + expctCls.getName() + 
                    " to be thrown" );
        }
    }

    private
    static
    Throwable
    invocationThrewExpectedThrowable( Test test,
                                      Throwable ex )
    {
        String pat = test.expectedPattern();

        String msg = ex.getMessage();
        if ( msg == null ) msg = "";

        if ( Pattern.matches( pat, msg ) ) return null;
        else
        {
            return
                new IllegalStateException(
                    "Throwable message does not match pattern '" + pat + "': " +
                    msg + " (See cause for actual exception)", ex );
        }
    }

    private
    static
    Throwable
    invocationFailed( Test test,
                      Throwable ex )
    {
        Class< ? extends Throwable > expctCls = test.expected();

        if ( expctCls.equals( Test.NoneExpectedException.class ) )
        {
            return ex;
        }
        else
        {
            if ( expctCls.isInstance( ex ) ) 
            {
                return invocationThrewExpectedThrowable( test, ex );
            }
            else
            {
                return
                    new IllegalStateException(
                        "Expected throwable of type " + expctCls.getName() +
                        " but got one of type " + ex.getClass().getName() + 
                        " (see cause)", ex );
            }
        }
    }

    public
    static
    Throwable
    getFinalThrowable( Test test,
                       Throwable ex )
    {
        inputs.notNull( test, "test" );

        return ex == null
            ? invocationSucceeded( test ) : invocationFailed( test, ex );
    }

    private
    static
    Throwable
    invocationSucceeded( TestFailureExpectation tfe )
    {
        if ( tfe == null ) return null;
        else return state.createFailf( "Expected failure: %s", tfe );
    }

    private
    static
    Throwable
    invocationThrewExpectedThrowable( TestFailureExpectation tfe,
                                      Throwable th )
    {
        String msg = th.getMessage();
        if ( msg == null ) msg = "";

        if ( tfe.expectedPattern().matcher( msg ).matches() ) return null;
        else
        {
            return
                new IllegalStateException(
                    "Throwable message does not match pattern '" + 
                    tfe.expectedPattern() + "': " + msg + 
                    " (See cause for actual exception)", th );
        }
    }

    private
    static
    Throwable
    invocationFailed( TestFailureExpectation tfe,
                      Throwable th )
    {
        if ( tfe == null ) return th;
        else
        {
            if ( tfe.expectedClass().isInstance( th ) ) 
            {
                return invocationThrewExpectedThrowable( tfe, th );
            }
            else
            {
                return
                    new IllegalStateException(
                        "Expected throwable of type " + 
                        tfe.expectedClass().getName() +
                        " but got one of type " + th.getClass().getName() + 
                        " (see cause)", th );
            }
        }
    }

    // either param may be null; the former if no failure is expected, the
    // latter if none was encountered
    public
    static
    Throwable
    getFinalThrowable( TestFailureExpectation tfe,
                       Throwable th )
    {
        return th == null
            ? invocationSucceeded( tfe ) : invocationFailed( tfe, th );
    }

    public
    static
    TestFailureExpectation
    failureExpectationFor( Test test )
    {
        inputs.notNull( test, "test" );

        Class< ? extends Throwable > expctCls = test.expected();

        if ( expctCls.equals( Test.NoneExpectedException.class ) ) return null;
        else
        {
            Pattern pat = PatternHelper.compile( test.expectedPattern() );
            return new TestFailureExpectation( expctCls, pat );
        }
    }

    public
    static
    TestFailureExpectation
    failureExpectationFor( TestFailureExpector tfe )
    {
        inputs.notNull( tfe, "tfe" );

        Class< ? extends Throwable > cls = tfe.expectedFailureClass();

        if ( cls == null ) return null;
        else
        {
            CharSequence patStr = tfe.expectedFailurePattern();
            if ( patStr == null ) patStr = ".*";
 
            Pattern pat = PatternHelper.compile( patStr );

            return new TestFailureExpectation( cls, pat );
        }
    }
}
