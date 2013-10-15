package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class PipelineReactor
implements ValueReactor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();
 
    public
    Event
    process( Event ev )
        throws Exception
    {
        throw new UnsupportedOperationException( "Unimplemented" );
    }

    public
    static
    PipelineReactor
    create( ValueReactor... reactors )
    {
        inputs.noneNull( reactors, "reactors" );
        return new PipelineReactor();
    }
}
