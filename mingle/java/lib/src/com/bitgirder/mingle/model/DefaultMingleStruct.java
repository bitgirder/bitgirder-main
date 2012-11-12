package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class DefaultMingleStruct
extends DefaultMingleStructure< MingleStruct >
implements MingleStruct
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    DefaultMingleStruct( MingleStructBuilder b ) { super( b ); }
}
