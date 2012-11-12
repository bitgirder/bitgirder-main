package com.bitgirder.mingle.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.TypedString;

public
final
class CommentText
extends TypedString< CommentText >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    CommentText( CharSequence txt ) { super( txt, "txt" ); }
}
