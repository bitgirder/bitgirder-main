package com.bitgirder.test;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
abstract
class AbstractLabeledTestObject
implements LabeledTestObject
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final String lbl;

    protected
    AbstractLabeledTestObject( CharSequence lbl )
    {
        this.lbl = inputs.notNull( lbl, "lbl" ).toString();
    }

    public final String getLabel() { return lbl; }

    // overridable
    public Object getInvocationTarget() { return this; }
}
