package com.bitgirder.mingle.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.AbstractIoTests;
import com.bitgirder.mingle.MingleValue;

import com.bitgirder.mingle.reactor.MingleReactor;
import com.bitgirder.mingle.reactor.MingleReactors;
import com.bitgirder.mingle.reactor.BuildReactor;
import com.bitgirder.mingle.reactor.ValueBuildFactory;

import com.bitgirder.test.Test;

import java.io.InputStream;
import java.io.OutputStream;

@Test
final
class IoTests
extends AbstractIoTests
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    protected
    MingleValue
    readMingleValue( InputStream is )
        throws Exception
    {
        BuildReactor br = new BuildReactor.Builder().
            setFactory( new ValueBuildFactory() ).
            build();
        
        MingleIo.feedValue( is, br );

        return (MingleValue) br.value();
    }

    protected
    void
    writeMingleValue( MingleValue mv,
                      OutputStream os )
        throws Exception
    {
        MingleReactor rct = MingleIo.createWriteReactor( os );
        MingleReactors.visitValue( mv, rct );
    }
}
