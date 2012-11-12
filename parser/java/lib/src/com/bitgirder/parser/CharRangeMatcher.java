package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class CharRangeMatcher
implements TerminalMatcher< Character >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final char minIncl;
    private final char maxIncl;

    private
    CharRangeMatcher( char minIncl,
                      char maxIncl )
    {
        inputs.isTrue( 
            minIncl <= maxIncl,
            "minIncl > maxIncl:", (int) minIncl, ">", (int) maxIncl );
 
        this.minIncl = minIncl;
        this.maxIncl = maxIncl;
    }

    char getMinInclusive() { return minIncl; }
    char getMaxInclusive() { return maxIncl; }

    static
    CharRangeMatcher
    forRange( char minIncl,
              char maxIncl )
    {
        return new CharRangeMatcher( minIncl, maxIncl );
    }

    public
    boolean
    isMatch( Character chObj )
    {
        char ch = chObj.charValue();

        return ch >= minIncl && ch <= maxIncl;
    }
}
