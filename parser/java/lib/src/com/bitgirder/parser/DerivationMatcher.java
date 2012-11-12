package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class DerivationMatcher< N >
implements ProductionMatcher
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final N head;

    private DerivationMatcher( N head ) { this.head = head; }

    N getHead() { return head; }

    static
    < N >
    DerivationMatcher< N >
    forHead( N head )
    {
        inputs.notNull( head, "head" );
        return new DerivationMatcher< N >( head );
    }
}
