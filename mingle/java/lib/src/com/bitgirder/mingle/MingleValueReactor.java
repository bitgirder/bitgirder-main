package com.bitgirder.mingle;

public
interface MingleValueReactor
{
    // ev is only valid during this call and may be changed later by calling
    // code. processors which will use ev later should copy it
    public
    void
    processEvent( MingleValueReactorEvent ev )
        throws Exception;
}
