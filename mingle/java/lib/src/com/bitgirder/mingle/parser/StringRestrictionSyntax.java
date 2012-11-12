package com.bitgirder.mingle.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.parser.SourceTextLocation;

import com.bitgirder.mingle.model.MingleString;

public
final
class StringRestrictionSyntax
extends RestrictionSyntax
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleString str;

    StringRestrictionSyntax( MingleString str,
                             SourceTextLocation loc )
    {
        super( loc );

        this.str = state.notNull( str, "str" );
    }

    public MingleString getString() { return str; }
}
