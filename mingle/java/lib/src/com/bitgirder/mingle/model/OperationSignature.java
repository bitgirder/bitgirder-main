package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;

public
final
class OperationSignature
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final FieldSet fields;
    private final MingleTypeReference retType;
    private final List< MingleTypeReference > thrown; // maybe empty; not null

    private
    OperationSignature( Builder b )
    {
        this.fields = inputs.notNull( b.fields, "fields" );
        this.retType = inputs.notNull( b.retType, "retType" );

        this.thrown = Lang.unmodifiableCopy( b.thrown );
    }

    public FieldSet getFieldSet() { return fields; }
    public MingleTypeReference getReturnType() { return retType; }
    public List< MingleTypeReference > getThrown() { return thrown; }

    public
    final
    static
    class Builder
    {
        private FieldSet fields;
        private MingleTypeReference retType;
        private List< MingleTypeReference > thrown = Lang.emptyList();

        public
        Builder
        setFields( FieldSet fields )
        {
            this.fields = inputs.notNull( fields, "fields" );
            return this;
        }

        public
        Builder
        setReturnType( MingleTypeReference retType )
        {
            this.retType = inputs.notNull( retType, "retType" );
            return this;
        }

        public
        Builder
        setThrown( List< MingleTypeReference > thrown )
        {
            this.thrown = Lang.unmodifiableCopy( thrown, "thrown" );
            return this;
        }

        public
        OperationSignature
        build()
        {
            return new OperationSignature( this );
        }
    }
}
