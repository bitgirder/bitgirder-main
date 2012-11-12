package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

// An instance of MingleSymbolMapBuilder which can be used to build just a
// standalone symbol map, as opposed to one embedded in a structure as its
// fields -- see MingleStructureBuilder
public
final
class StandaloneMingleSymbolMapBuilder
extends MingleSymbolMapBuilder< StandaloneMingleSymbolMapBuilder >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private StandaloneMingleSymbolMapBuilder() {}
    
    public
    static
    StandaloneMingleSymbolMapBuilder
    create()
    {
        return new StandaloneMingleSymbolMapBuilder();
    }
}
