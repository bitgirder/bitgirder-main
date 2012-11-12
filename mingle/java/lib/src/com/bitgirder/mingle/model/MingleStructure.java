package com.bitgirder.mingle.model;

public
interface MingleStructure
extends MingleValue
{
    public
    AtomicTypeReference
    getType();

    public
    MingleSymbolMap
    getFields();
}
