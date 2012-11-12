package com.bitgirder.mingle.codec;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.model.MingleIdentifier;

public
final
class NoSuchMingleCodecException
extends MingleCodecException
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public NoSuchMingleCodecException() {}
    public NoSuchMingleCodecException( String msg ) { super( msg ); }

    public 
    NoSuchMingleCodecException( MingleIdentifier id )
    {
        this( inputs.notNull( id, "id" ).getExternalForm().toString() );
    }
}
