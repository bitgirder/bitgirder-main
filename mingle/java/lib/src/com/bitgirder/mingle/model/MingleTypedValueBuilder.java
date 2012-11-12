package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.parser.MingleParsers;

public
abstract
class MingleTypedValueBuilder< B extends MingleTypedValueBuilder,
                               V extends MingleValue >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    AtomicTypeReference typeRef;

    final
    B
    castThis()
    {
        @SuppressWarnings( "unchecked" )
        B res = (B) this;
        return res;
    }

    public
    final
    B
    setType( AtomicTypeReference typeRef )
    {
        this.typeRef = inputs.notNull( typeRef, "typeRef" );
        return castThis();
    }

    public
    final
    B
    setType( CharSequence typeRefStr )
    {
        inputs.notNull( typeRefStr, "typeRefStr" );

        MingleTypeReference ref = 
            MingleParsers.createTypeReference( typeRefStr );

        if ( ref instanceof AtomicTypeReference )
        {
            return setType( (AtomicTypeReference) ref );
        }
        else 
        {
            throw 
                inputs.createFail( 
                    "Not an atomic type reference:", typeRefStr );
        }
    }

    public abstract V build();
}
