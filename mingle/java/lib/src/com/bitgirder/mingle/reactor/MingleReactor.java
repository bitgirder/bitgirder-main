package com.bitgirder.mingle.reactor;

public
interface MingleReactor
{
    // ev is only valid during this call and may be changed later by calling
    // code. processors which will use ev later should copy it
    public
    void
    processEvent( MingleReactorEvent ev )
        throws Exception;
}
