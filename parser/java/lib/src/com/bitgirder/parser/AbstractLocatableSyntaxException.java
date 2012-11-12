package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

abstract
class AbstractLocatableSyntaxException
extends SyntaxException
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final SourceTextLocation loc;

    AbstractLocatableSyntaxException( SourceTextLocation loc )
    {
        super( inputs.notNull( loc, "loc" ).toString() );

        this.loc = loc;
    }

    public SourceTextLocation getLocation() { return loc; }
}
