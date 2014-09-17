package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.mingle.MingleIdentifier;

import java.util.List;

public
final
class FieldOrder
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();
    
    public
    final
    static
    class FieldSpecification
    {
        private final static Inputs inputs = new Inputs();
        private final static State state = new State();
    
        private final MingleIdentifier field;
        private final boolean required;
    
        public
        FieldSpecification( MingleIdentifier field,
                            boolean required )
        {
            this.field = inputs.notNull( field, "field" );
            this.required = required;
        }
    
        public MingleIdentifier field() { return field; }
        public boolean required() { return required; }
    }

    private final List< FieldSpecification > fields;

    public
    FieldOrder( List< FieldSpecification > fields )
    {
        this.fields = Lang.unmodifiableCopy( fields, "fields" );
    }

    public 
    List< FieldSpecification > 
    fields() 
    {
        return fields;
    }
}
