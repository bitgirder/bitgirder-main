package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class MingleExceptionBuilder
extends MingleStructureBuilder< MingleExceptionBuilder, MingleException >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    MingleExceptionBuilder() {}

    public
    MingleExceptionBuilder
    setMessage( MingleString msg )
    {
        inputs.notNull( msg, "msg" );
        fields().set( DefaultMingleException.FIELD_MESSAGE, msg );

        return this;
    }

    public
    MingleExceptionBuilder
    setMessage( CharSequence msg )
    {
        inputs.notNull( msg, "msg" );
        return setMessage( MingleModels.asMingleString( msg ) );
    }

    public 
    MingleException 
    build() 
    { 
        return new DefaultMingleException( this ); 
    }
}
