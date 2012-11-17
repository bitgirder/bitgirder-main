package com.bitgirder.concurrent;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Completion;

import java.util.concurrent.Callable;

public
final
class LazyValue< V >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Callable< V > call;

    // guarded by sync
    private Completion< V > comp;

    private LazyValue( Callable< V > call ) { this.call = call; }

    public
    synchronized
    V
    get()
        throws Exception
    {
        if ( comp == null )
        {
            try { comp = Lang.successCompletion( call.call() ); }
            catch ( Throwable th ) { comp = Lang.failureCompletion( th ); }
        }

        return comp.get();
    }

    public
    static
    < V >
    LazyValue< V >
    forCall( Callable< V > call )
    {
        inputs.notNull( call, "call" );
        return new LazyValue< V >( call );
    }
}
