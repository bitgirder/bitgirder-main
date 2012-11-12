package com.bitgirder.testing;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.test.TestPhase;

public
final
class InvocationDescriptor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final String name;
    private final TestPhase phase;

    InvocationDescriptor( CharSequence name,
                          TestPhase phase )
    {
        this.name = inputs.notNull( name, "name" ).toString();
        this.phase = inputs.notNull( phase, "phase" );
    }

    public String getName() { return name; }
    public TestPhase getPhase() { return phase; }
}
