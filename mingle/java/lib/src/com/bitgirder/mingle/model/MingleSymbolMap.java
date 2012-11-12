package com.bitgirder.mingle.model;

public
interface MingleSymbolMap
extends MingleValue
{
    public Iterable< MingleIdentifier > getFields();

    public boolean hasField( MingleIdentifier fld );

    // should return java null if field is not in this map; should never return
    // MingleNull as a value
    public MingleValue get( MingleIdentifier fld );
}
