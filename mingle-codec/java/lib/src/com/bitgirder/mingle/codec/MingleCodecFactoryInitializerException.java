package com.bitgirder.mingle.codec;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class MingleCodecFactoryInitializerException
extends MingleCodecException
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    MingleCodecFactoryInitializerException( String msg ) { super( msg ); }
}
