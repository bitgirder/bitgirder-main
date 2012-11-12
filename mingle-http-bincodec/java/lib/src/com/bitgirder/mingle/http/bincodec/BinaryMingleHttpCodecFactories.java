package com.bitgirder.mingle.http.bincodec;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.mingle.codec.MingleCodec;

import com.bitgirder.mingle.http.MingleHttpCodecContext;
import com.bitgirder.mingle.http.MingleHttpCodecFactory;
import com.bitgirder.mingle.http.MingleHttpCodecFactorySelector;
import com.bitgirder.mingle.http.MingleHttpCodecFactories;

import com.bitgirder.mingle.bincodec.MingleBinaryCodecs;

import com.bitgirder.http.HttpRequestMessage;

import java.util.List;

final
class BinaryMingleHttpCodecFactories
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static String CTYPE = "application/mingle-binary";

    public
    static
    MingleHttpCodecContext
    getCodecContext()
    {
        return 
            new MingleHttpCodecContext()
            {
                private final MingleCodec codec = MingleBinaryCodecs.getCodec();

                public MingleCodec codec() { return codec; }
                public CharSequence contentType() { return CTYPE; }
            };
    }

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
            return Lang.< MingleHttpCodecContext >asList( getCodecContext() );
        }
    }

    private
    final
    static
    class FactorySelectorInit
    implements MingleHttpCodecFactories.FactorySelectorInitializer
    {
        public
        void
        addFactories( MingleHttpCodecFactorySelector.Builder b )
        {
            final MingleHttpCodecContext codecCtx = getCodecContext();

            b.selectFullType( 
                CTYPE, 
                new MingleHttpCodecFactory() 
                {
                    public 
                    MingleHttpCodecContext
                    codecContextFor( HttpRequestMessage req )
                    {
                        return codecCtx;
                    }
                }
            );
        }
    }
}
