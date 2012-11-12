package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class OperationDefinition
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleIdentifier name;
    private final OperationSignature sig;

    private
    OperationDefinition( Builder b )
    {
        this.name = inputs.notNull( b.name, "name" );
        this.sig = inputs.notNull( b.sig, "sig" );
    }

    public MingleIdentifier getName() { return name; }
    public OperationSignature getSignature() { return sig; }

    public
    final
    static
    class Builder
    {
        private MingleIdentifier name;
        private OperationSignature sig;

        public
        Builder
        setName( MingleIdentifier name )
        {
            this.name = inputs.notNull( name, "name" );
            return this;
        }

        public
        Builder
        setSignature( OperationSignature sig )
        {
            this.sig = inputs.notNull( sig, "sig" );
            return this;
        }

        public
        OperationDefinition
        build()
        {
            return new OperationDefinition( this );
        }
    }
}
