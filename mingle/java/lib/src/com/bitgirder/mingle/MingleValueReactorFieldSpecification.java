package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class MingleValueReactorFieldSpecification
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleIdentifier field;
    private final boolean required;

    public
    MingleValueReactorFieldSpecification( MingleIdentifier field,
                                          boolean required )
    {
        this.field = inputs.notNull( field, "field" );
        this.required = required;
    }

    public MingleIdentifier field() { return field; }
    public boolean required() { return required; }
}
