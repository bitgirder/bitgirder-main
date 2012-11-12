package com.bitgirder.lang;

public
interface ObjectReceiver< V >
{
    public
    void
    receive( V obj )
        throws Exception;
}
