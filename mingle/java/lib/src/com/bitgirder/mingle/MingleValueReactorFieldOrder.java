package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;

public
final
class MingleValueReactorFieldOrder
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final List< MingleValueReactorFieldSpecification > fields;

    public
    MingleValueReactorFieldOrder(
        List< MingleValueReactorFieldSpecification > fields )
    {
        this.fields = Lang.unmodifiableCopy( fields, "fields" );
    }

    public 
    List< MingleValueReactorFieldSpecification > 
    fields() 
    {
        return fields;
    }
}
