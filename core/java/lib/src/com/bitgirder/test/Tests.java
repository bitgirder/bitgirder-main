package com.bitgirder.test;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class Tests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private Tests() {}

    private
    final
    static
    class LabeledTestObjectImpl
    implements LabeledTestObject
    {
        private final Object invTarg;
        private final CharSequence label;

        private
        LabeledTestObjectImpl( Object invTarg,
                               CharSequence label )
        {
            this.invTarg = invTarg;
            this.label = label;
        }

        public Object getInvocationTarget() { return invTarg; }
        public CharSequence getLabel() { return label; }
    }

    public
    static
    LabeledTestObject
    createLabeledTestObject( Object invTarg,
                             CharSequence label )
    {
        inputs.notNull( invTarg, "invTarg" );
        inputs.notNull( label, "label" );

        return new LabeledTestObjectImpl( invTarg, label );
    }
}
