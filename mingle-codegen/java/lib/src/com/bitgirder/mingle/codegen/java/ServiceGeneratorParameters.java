package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.mingle.model.MingleIdentifier;

import java.util.List;
import java.util.Map;

final
class ServiceGeneratorParameters
extends TypeMaskedGeneratorParameters
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final JvType svcBaseCls;
    private final List< OperationGeneratorParameters > opParams;

    private
    ServiceGeneratorParameters( Builder b )
    {
        super( b );

        this.svcBaseCls = b.svcBaseCls;
        this.opParams = Lang.unmodifiableCopy( b.opParams );
    }

    JvType getServiceBaseClass() { return svcBaseCls; }

    Map< MingleIdentifier, OperationGeneratorParameters >
    opParamsByName()
    {
        Map< MingleIdentifier, OperationGeneratorParameters > res = 
            Lang.newMap();

        for ( OperationGeneratorParameters p : opParams ) res.put( p.name, p );

        return res;
    }

    final
    static
    class Builder
    extends TypeMaskedGeneratorParameters.Builder< Builder >
    {
        private JvType svcBaseCls;
        private List< OperationGeneratorParameters > opParams = 
            Lang.emptyList();

        public
        Builder
        setServiceBaseClass( JvType svcBaseCls )
        {
            this.svcBaseCls = inputs.notNull( svcBaseCls, "svcBaseCls" );
            return this;
        }

        public
        Builder
        setOpParams( List< OperationGeneratorParameters > opParams )
        {
            this.opParams = inputs.noneNull( opParams, "opParams" );
            return this;
        }

        ServiceGeneratorParameters
        build()
        {
            return new ServiceGeneratorParameters( this );
        }
    }
}
