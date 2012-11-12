package com.bitgirder.test;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
abstract
class LabeledTestCall
implements TestCall,
           LabeledTestObject,
           TestFailureExpector
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final CharSequence label;

    private Class< ? extends Throwable > failCls;
    private CharSequence failPat;

    protected
    LabeledTestCall( CharSequence label )
    {
        this.label = inputs.notNull( label, "label" );
    }

    // failCls or failPat may be null, since those are either or both valid
    // return values for the failure expectation methods. It is not allowed
    // though to set a non-null failCls twice, since that is almost certainly
    // indicative of a programming error.
    private
    LabeledTestCall
    doExpectFailure( Class< ? extends Throwable > failCls,
                     CharSequence failPat )
    {
        state.isTrue( 
            this.failCls == null, "A failure expectation is already set" );

        this.failCls = failCls;
        this.failPat = failPat;

        return this;
    }

    public
    final
    LabeledTestCall
    expectFailure( Class< ? extends Throwable > failCls )
    {
        return doExpectFailure( failCls, null );
    }

    public
    final
    LabeledTestCall
    expectFailure( Class< ? extends Throwable > failCls,
                   CharSequence failPat )
    {
        return doExpectFailure( failCls, failPat );
    }

    public final CharSequence getLabel() { return label; }
    public final Object getInvocationTarget() { return this; }

    // expectedFailure(Class|Pattern) are not final and may be overridden if an
    // implementation wants to provide its own impl; in that case setting values
    // on this instance via expectFailure() will proceed without error but will
    // have no effect

    public 
    Class< ? extends Throwable > 
    expectedFailureClass() 
    {
        return failCls;
    }

    public
    CharSequence
    expectedFailurePattern()
    {
        return failPat;
    }
}
