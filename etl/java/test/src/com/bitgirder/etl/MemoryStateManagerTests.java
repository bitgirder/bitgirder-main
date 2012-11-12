package com.bitgirder.etl;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.test.Test;

@Test
final
class MemoryStateManagerTests
extends EtlStateManagerTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    protected
    EtlTestReactor
    createTestReactor()
    {
        return EtlTests.createMemoryTestReactor();
    }

    protected Object createStateObject( Integer i ) { return i; }

    protected Object createFeedPosition() { return Lang.randomUuid(); }
}
