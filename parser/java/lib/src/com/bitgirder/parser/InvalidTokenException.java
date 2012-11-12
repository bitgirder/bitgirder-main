package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class InvalidTokenException
extends AbstractLocatableSyntaxException
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    InvalidTokenException( SourceTextLocation loc ) { super( loc ); }
}
