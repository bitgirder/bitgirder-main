package com.bitgirder.test;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.util.regex.Pattern;

public
abstract
class LabeledTestCall
implements TestCall,
           LabeledTestObject,
           TestFailureExpector
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private String label;

    private Class< ? extends Throwable > failCls;
    private CharSequence failPat;

    protected
    LabeledTestCall( CharSequence label )
    {
        this.label = inputs.notNull( label, "label" ).toString();
    }

    protected LabeledTestCall() {}

    // failCls or failPat may be null, since those are either or both valid
    // return values for the failure expectation methods. It is not allowed
    // though to set a non-null failCls twice, since that is almost certainly
    // indicative of a programming error.
    private
    LabeledTestCall
    doExpectFailure( Class< ? extends Throwable > failCls,
                     CharSequence failPat,
                     boolean isReset )
    {
        state.isTrue( 
            isReset || this.failCls == null, 
            "A failure expectation is already set" );

        this.failCls = failCls;
        this.failPat = failPat;

        return this;
    }

    public
    final
    LabeledTestCall
    expectFailure( Class< ? extends Throwable > failCls )
    {
        inputs.notNull( failCls, "failCls" );
        return doExpectFailure( failCls, null, false );
    }

    public
    final
    LabeledTestCall
    expectFailure( Class< ? extends Throwable > failCls,
                   CharSequence failPat )
    {
        inputs.notNull( failCls, "failCls" );
        inputs.notNull( failPat, "failPat" );

        return doExpectFailure( failCls, failPat, false );
    }

    private
    LabeledTestCall
    doExpectFailure( Throwable th,
                     boolean isReset )
    {
        inputs.notNull( th, "th" );

        String msg = th.getMessage();
        String pat = msg == null ? null : Pattern.quote( msg );
        
        return doExpectFailure( th.getClass(), pat, isReset );
    }

    public
    final
    LabeledTestCall
    expectFailure( Throwable th )
    {
        return doExpectFailure( th, false );
    }

    public
    final
    LabeledTestCall
    resetExpectFailure( Throwable th )
    {
        return doExpectFailure( th, true );
    }

    public 
    final 
    String
    getLabel() 
    { 
        if ( label == null ) return toString();
        return label; 
    }

    public
    final
    void
    setLabel( CharSequence label )
    {
        inputs.notNull( label, "label" );

        state.isTruef( this.label == null, 
            "label already set: %s", this.label );

        this.label = label.toString();
    }

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
