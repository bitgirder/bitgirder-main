package com.bitgirder.mingle.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.parser.SourceTextLocation;

import com.bitgirder.mingle.model.MingleValue;

public
final
class RangeRestrictionSyntax
extends RestrictionSyntax
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final boolean includesLeft;
    private final MingleValue leftVal;
    private final MingleValue rightVal;
    private final boolean includesRight;

    RangeRestrictionSyntax( boolean includesLeft,
                            MingleValue leftVal,
                            MingleValue rightVal,
                            boolean includesRight,
                            SourceTextLocation loc )
    {
        super( loc );

        this.includesLeft = includesLeft;
        this.leftVal = leftVal;
        this.rightVal = rightVal;
        this.includesRight = includesRight;
    }

    public boolean includesLeft() { return includesLeft; }
    public MingleValue leftValue() { return leftVal; }
    public MingleValue rightValue() { return rightVal; }
    public boolean includesRight() { return includesRight; }
}
