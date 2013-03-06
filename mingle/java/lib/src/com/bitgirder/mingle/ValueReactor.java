package com.bitgirder.mingle;

public
interface ValueReactor
{
    public
    final
    static
    class Event
    {
        private Event() {}
    }

    public
    Event
    process( Event ev )
        throws Exception;
}
