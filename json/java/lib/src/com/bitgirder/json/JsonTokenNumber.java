package com.bitgirder.json;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class JsonTokenNumber
implements JsonToken
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final boolean negated;
    private final CharSequence intPart; // not null
    private final CharSequence fracPart; // maybe null
    private final Exponent exp; // maybe null

    JsonTokenNumber( boolean negated,
                     CharSequence intPart,
                     CharSequence fracPart,
                     Exponent exp )
    {
        this.negated = negated;
        this.intPart = inputs.notNull( intPart, "intPart" );
        this.fracPart = fracPart;
        this.exp = exp;
    }

    public boolean isNegated() { return negated; }
    public CharSequence getIntPart() { return intPart; }
    public CharSequence getFracPart() { return fracPart; }
    public Exponent getExponent() { return exp; }

    public
    final
    static
    class Exponent
    {
        private final boolean negated;
        private final CharSequence num;

        Exponent( boolean negated,
                  CharSequence num )
        {
            this.negated = negated;
            this.num = inputs.notNull( num, "num" );
        }

        public boolean isNegated() { return negated; }
        public CharSequence getNumber() { return num; }
    }
}
