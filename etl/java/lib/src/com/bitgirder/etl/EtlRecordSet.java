package com.bitgirder.etl;

public
interface EtlRecordSet
extends Iterable< Object >
{
    public 
    int
    size();

    public
    boolean
    isFinal();
}
