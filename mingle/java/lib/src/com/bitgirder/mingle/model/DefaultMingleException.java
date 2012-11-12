package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.parser.MingleParsers;

public
final
class DefaultMingleException
extends DefaultMingleStructure< MingleException >
implements MingleException
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public final static MingleIdentifier FIELD_MESSAGE =
        MingleParsers.createIdentifier( "message" );

    DefaultMingleException( MingleExceptionBuilder b ) { super( b ); }
}
