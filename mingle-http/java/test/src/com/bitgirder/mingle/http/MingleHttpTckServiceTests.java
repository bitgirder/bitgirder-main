package com.bitgirder.mingle.http;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.process.AbstractProcess;

import com.bitgirder.mingle.codec.MingleCodec;

import com.bitgirder.mingle.tck.v1.MingleTckServiceTests;
import com.bitgirder.mingle.tck.v1.MingleTckTests;

import com.bitgirder.mingle.bind.MingleBinder;

import com.bitgirder.net.SelectableChannelManager;

import com.bitgirder.test.TestFactory;
import com.bitgirder.test.Tests;
import com.bitgirder.test.AbstractLabeledTestObject;

import java.util.List;

public
final
class MingleHttpTckServiceTests
extends AbstractLabeledTestObject
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final AbstractProcess< ? > cliFact;
    private final MingleHttpTesting.ServerLocation testLoc;
    private final MingleHttpCodecContext codecCtx;
    private final MingleBinder mb;
    private final boolean useGzip;

    private
    MingleHttpTckServiceTests( Builder b )
    {
        super( inputs.notNull( inputs.notNull( b, "b" ).label, "label" ) );

        this.cliFact = inputs.notNull( b.cliFact, "cliFact" );
        this.testLoc = inputs.notNull( b.testLoc, "testLoc" );
        this.codecCtx = inputs.notNull( b.codecCtx, "codecCtx" );
        this.mb = inputs.notNull( b.mb, "mb" );
        this.useGzip = b.useGzip;
    }

    public
    final
    class CreateRpcClientFactory
    {
        public 
        MingleHttpTesting.ServerLocation 
        testLocation() 
        { 
            return testLoc; 
        }

        public MingleHttpCodecContext codecContext() { return codecCtx; }
        public boolean useGzip() { return useGzip; }
    }

    @TestFactory
    private
    List< ? >
    getTckServiceTest()
    {
        return Lang.singletonList(
            Tests.createLabeledTestObject(
                new MingleTckServiceTests.Builder().
                    setBackend( cliFact ).
                    setMingleBinder( mb ).
                    setRpcClientFactoryRequest( new CreateRpcClientFactory() ).
                    setClientBuilder(
                        new MingleTckTests.ClientBuilder() {
                            public void build( AbstractProcess.Builder b ) {
                                b.mixin( SelectableChannelManager.create() );
                            }
                        }
                    ).
                    build(),
                "baseTckServiceTests"
            )
        );
    }

    public
    final
    static
    class Builder
    {
        private AbstractProcess< ? > cliFact;
        private MingleHttpTesting.ServerLocation testLoc;
        private CharSequence label;
        private MingleHttpCodecContext codecCtx;
        private MingleBinder mb;
        private boolean useGzip;

        public
        Builder
        setClientFactory( AbstractProcess< ? > cliFact )
        {
            this.cliFact = inputs.notNull( cliFact, "cliFact" );
            return this;
        }

        public
        Builder
        setLocation( MingleHttpTesting.ServerLocation testLoc )
        {
            this.testLoc = inputs.notNull( testLoc, "testLoc" );
            return this;
        }

        public
        Builder
        setUseGzip( boolean useGzip )
        {
            this.useGzip = useGzip;
            return this;
        }

        public
        Builder
        setLabel( CharSequence label )
        {
            this.label = inputs.notNull( label, "label" );
            return this;
        }

        public
        Builder
        setCodecContext( MingleHttpCodecContext codecCtx )
        {
            this.codecCtx = inputs.notNull( codecCtx, "codecCtx" );
            return this;
        }

        public
        Builder
        setMingleBinder( MingleBinder mb )
        {
            this.mb = inputs.notNull( mb, "mb" );
            return this;
        }

        public 
        MingleHttpTckServiceTests 
        build() 
        { 
            return new MingleHttpTckServiceTests( this );
        }
    }
}
