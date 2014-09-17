package com.bitgirder.mingle;

import static com.bitgirder.log.CodeLoggers.Statics.*;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.test.Test;

import java.io.InputStream;
import java.io.OutputStream;

@Test
final
class CoreIoTests
extends AbstractIoTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    protected
    MingleValue
    readMingleValue( InputStream is )
        throws Exception
    {
        return MingleBinReader.create( is ).readScalar();
    }

    protected
    void
    writeMingleValue( MingleValue mv,
                      OutputStream os )
        throws Exception
    {
        MingleBinWriter w = MingleBinWriter.create( os );
        w.writeScalar( mv );
    }

    @Override
    protected
    boolean
    shouldRunTestForObject( Object rep )
    {
        return ( ! ( rep instanceof MingleList ||
                     rep instanceof MingleSymbolMap ||
                     rep instanceof MingleStruct ) );
    }
}
