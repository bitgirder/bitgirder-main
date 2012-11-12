package com.bitgirder.mingle.http;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.util.List;

public
final
class MingleHttpCodecFactories
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private MingleHttpCodecFactories() {}

    public
    static
    interface ClientCodecInitializer
    {
        public
        List< MingleHttpCodecContext >
        getClientCodecContexts()
            throws Exception;
    }

    public
    static
    interface FactorySelectorInitializer
    {
        public
        void
        addFactories( MingleHttpCodecFactorySelector.Builder b )
            throws Exception;
    }
}
