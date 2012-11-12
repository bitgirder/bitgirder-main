package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class QuantifierMatcher
implements ProductionMatcher
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final ProductionMatcher matcher;
    private final int minIncl;
    private final int maxIncl;

    private
    QuantifierMatcher( ProductionMatcher matcher,
                       int minIncl,
                       int maxIncl )
    {
        this.matcher = matcher;
        this.minIncl = minIncl;
        this.maxIncl = maxIncl;
        
        // base assertions
        state.nonnegativeI( minIncl, "minIncl" );
        state.isTrue( maxIncl > 0 );
        state.isTrue( maxIncl >= minIncl );
    }

    ProductionMatcher getMatcher() { return matcher; }
    int getMinInclusive() { return minIncl; }
    int getMaxInclusive() { return maxIncl; }

    static
    QuantifierMatcher
    unary( ProductionMatcher matcher )
    {
        inputs.notNull( matcher, "matcher" );
        return new QuantifierMatcher( matcher, 0, 1 );
    }

    static
    QuantifierMatcher
    atLeastOne( ProductionMatcher matcher )
    {
        inputs.notNull( matcher, "matcher" );
        return new QuantifierMatcher( matcher, 1, Integer.MAX_VALUE );
    }

    static
    QuantifierMatcher
    kleene( ProductionMatcher matcher )
    {
        inputs.notNull( matcher, "matcher" );
        return new QuantifierMatcher( matcher, 0, Integer.MAX_VALUE );
    }

    static
    QuantifierMatcher
    exactly( ProductionMatcher matcher,
             int count )
    {
        inputs.notNull( matcher, "matcher" );
        inputs.positiveI( count, "count" );

        return new QuantifierMatcher( matcher, count, count );
    }
}
