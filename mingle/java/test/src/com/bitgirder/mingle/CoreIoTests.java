package com.bitgirder.mingle;

import static com.bitgirder.log.CodeLoggers.Statics.*;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.test.Test;

import java.io.InputStream;

@Test
final
class CoreIoTests
extends AbstractIoTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    @Override
    protected
    Object
    readValue( InputStream is,
               Object rep )
        throws Exception
    {
        if ( rep instanceof MingleValue ) {
            MingleBinReader rd = MingleBinReader.create( is );
            return rd.readScalar();
        }

        return super.readValue( is, rep );
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
