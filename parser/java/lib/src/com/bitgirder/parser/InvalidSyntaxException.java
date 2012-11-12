package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class InvalidSyntaxException
extends AbstractLocatableSyntaxException
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    InvalidSyntaxException( SourceTextLocation loc ) { super( loc ); }
}
