package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class DepthTracker
implements MingleReactor
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private int depth;

    private DepthTracker() {}

    public int depth() { return depth; }

    public
    void
    processEvent( MingleReactorEvent ev )
    {
        switch ( ev.type() ) {
        case MAP_START: depth++; break;
        case STRUCT_START: depth++; break;
        case LIST_START: depth++; break;
        case END: depth--; break;
        }
    }

    public static DepthTracker create() { return new DepthTracker(); }
}
