package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;

public
final
class ServiceDefinition
extends TypeDefinition
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final List< OperationDefinition > opDefs;
    private final QualifiedTypeName secRef;

    private
    ServiceDefinition( Builder b )
    {
        super( b );

        this.opDefs = Lang.unmodifiableCopy( b.opDefs );
        this.secRef = b.secRef;
    }

    public List< OperationDefinition > getOperations() { return opDefs; }
    public QualifiedTypeName getSecurity() { return secRef; }

    public
    final
    static
    class Builder
    extends TypeDefinition.Builder< ServiceDefinition, Builder >
    {
        private List< OperationDefinition > opDefs = Lang.newList();
        private QualifiedTypeName secRef;

        public
        Builder
        setOperations( List< OperationDefinition > opDefs )
        {
            this.opDefs = Lang.unmodifiableCopy( opDefs, "opDefs" );
            return this;
        }

        public
        Builder
        setSecurity( QualifiedTypeName secRef )
        {
            this.secRef = inputs.notNull( secRef, "secRef" );
            return this;
        }

        public
        ServiceDefinition
        build()
        {
            return new ServiceDefinition( this ); 
        }
    }
}
