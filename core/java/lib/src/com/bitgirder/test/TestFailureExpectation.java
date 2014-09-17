package com.bitgirder.test;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.PatternHelper;

import java.util.regex.Pattern;

public
final
class TestFailureExpectation
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Class< ? extends Throwable > cls;
    private final Pattern pat;

    TestFailureExpectation( Class< ? extends Throwable > cls,
                            Pattern pat )
    {
        this.cls = inputs.notNull( cls, "cls" );
        this.pat = inputs.notNull( pat, "pat" );
    }

    public Class< ? extends Throwable > expectedClass() { return cls; }
    public Pattern expectedPattern() { return pat; }

    @Override
    public
    final
    String
    toString()
    {
        return String.format( "[%s]: %s", expectedClass(), expectedPattern() );
    }

    public
    static
    TestFailureExpectation
    create( Class< ? extends Throwable > cls,
            CharSequence pat )
    {
        return 
            new TestFailureExpectation( 
                cls, PatternHelper.compile( inputs.notNull( pat, "pat" ) ) );
    }
}
