package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.parser.MingleParsers;

public
abstract
class MingleStructureBuilder< B extends MingleStructureBuilder,
                              S extends MingleStructure >
extends MingleTypedValueBuilder< B, S >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final MingleSymbolMapBuilder< MingleStructureBuilder< B, S > > symBld = 
        MingleSymbolMapBuilder.create( this );

    public 
    final 
    MingleSymbolMapBuilder< MingleStructureBuilder< B, S > > fields() 
    { 
        return symBld; 
    }

    public 
    final 
    MingleSymbolMapBuilder< MingleStructureBuilder< B, S > > 
    f() 
    { 
        return fields(); 
    }

    public abstract S build();
}
