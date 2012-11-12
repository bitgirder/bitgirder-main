package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;

final
class SequenceMatcher
implements ProductionMatcher
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final List< ProductionMatcher > seq;

    private SequenceMatcher( List< ProductionMatcher > seq ) { this.seq = seq; }

    List< ProductionMatcher > getSequence() { return seq; }
    
    static
    SequenceMatcher
    forSequence( List< ProductionMatcher > seq )
    {
        return 
            new SequenceMatcher(
                Lang.unmodifiableCopy( inputs.noneNull( seq, "seq" ) ) );
    }
}
