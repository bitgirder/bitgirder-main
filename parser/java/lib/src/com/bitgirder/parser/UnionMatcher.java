package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;

final
class UnionMatcher
implements ProductionMatcher
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    static enum Mode { FIRST_WINS, LONGEST_WINS; }

    private final List< ProductionMatcher > matchers;
    private final Mode mode;

    private
    UnionMatcher( List< ProductionMatcher > matchers,
                  Mode mode )
    {
        this.matchers = Lang.unmodifiableCopy( matchers, "matchers" );
        this.mode = mode;
    }

    List< ProductionMatcher > getMatchers() { return matchers; }
    Mode getMode() { return mode; }

    static
    UnionMatcher
    forMatchers( List< ProductionMatcher > matchers,
                 Mode mode )
    {
        inputs.notNull( mode, "mode" );
        return new UnionMatcher( matchers, mode );
    }

    static
    UnionMatcher
    forMatchers( List< ProductionMatcher > matchers )
    {
        return forMatchers( matchers, Mode.FIRST_WINS );
    }
}
