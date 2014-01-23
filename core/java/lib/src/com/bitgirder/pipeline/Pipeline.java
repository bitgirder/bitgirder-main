package com.bitgirder.pipeline;

public
interface Pipeline< V >
{
    public int size();

    public V get( int idx );
}
