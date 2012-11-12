package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class PrototypeDefinition
extends TypeDefinition
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final OperationSignature sig;

    private
    PrototypeDefinition( Builder b )
    {
        super( b );
        this.sig = inputs.notNull( b.sig, "sig" );
    }

    public OperationSignature getSignature() { return sig; }

    public
    final
    static
    class Builder
    extends TypeDefinition.Builder< PrototypeDefinition, Builder >
    {
        private OperationSignature sig;

        public
        Builder
        setSignature( OperationSignature sig )
        {
            this.sig = inputs.notNull( sig, "sig" );
            return this;
        }

        public
        PrototypeDefinition
        build()
        {
            return new PrototypeDefinition( this ); 
        }
    }
}
