package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class NoSuchTypeDefinitionException
extends RuntimeException
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public
    NoSuchTypeDefinitionException( QualifiedTypeName qn )
    {
        super( inputs.notNull( qn, "qn" ).getExternalForm().toString() );
    }
}
