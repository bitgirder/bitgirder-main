package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.io.IOException;

public
final
class FilePermissionException
extends IOException
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    FilePermissionException( String msg ) { super( msg ); }
}
