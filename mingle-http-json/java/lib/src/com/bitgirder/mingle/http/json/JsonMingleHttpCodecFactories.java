package com.bitgirder.mingle.http.json;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.mingle.http.MingleHttpCodecContext;
import com.bitgirder.mingle.http.MingleHttpCodecFactories;
import com.bitgirder.mingle.http.MingleHttpCodecFactorySelector;

import java.util.List;

final
class JsonMingleHttpCodecFactories
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private
    final
    static
    class ClientCodecInit
    implements MingleHttpCodecFactories.ClientCodecInitializer
    {
        public
        List< MingleHttpCodecContext >
        getClientCodecContexts()
        {
            return
                Lang.< MingleHttpCodecContext >asList(
                    JsonMingleHttpCodecFactory.
                        getInstance().
                        getDefaultCodecContext()
                );
        }
    }

    private
    final
    static
    class ServerCodecInit
    implements MingleHttpCodecFactories.FactorySelectorInitializer
    {
        public
        void
        addFactories( MingleHttpCodecFactorySelector.Builder b )
        {
            b.selectSubType( "json", JsonMingleHttpCodecFactory.getInstance() );
        }
    }
}
