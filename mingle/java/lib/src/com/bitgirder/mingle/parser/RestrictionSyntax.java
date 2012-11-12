package com.bitgirder.mingle.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.parser.SourceTextLocation;

public
abstract
class RestrictionSyntax
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final SourceTextLocation loc;

    RestrictionSyntax( SourceTextLocation loc )
    {
        this.loc = state.notNull( loc, "loc" );
    }

    public final SourceTextLocation getLocation() { return loc; }
}
