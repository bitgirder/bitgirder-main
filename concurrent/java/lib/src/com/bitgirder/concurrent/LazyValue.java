package com.bitgirder.concurrent;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Completion;

import java.util.concurrent.Callable;

public
abstract
class LazyValue< V >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    // guarded by sync
    private Completion< V > comp;

    protected LazyValue() {}

    protected
    abstract
    V
    call()
        throws Exception;

    public
    synchronized
    V
    get()
        throws Exception
    {
        if ( comp == null )
        {
            try { comp = Lang.successCompletion( call() ); }
            catch ( Throwable th ) { comp = Lang.failureCompletion( th ); }
        }

        return comp.get();
    }

    public
    static
    < V >
    LazyValue< V >
    forCall( final Callable< V > call )
    {
        inputs.notNull( call, "call" );

        return new LazyValue< V >() {
            protected V call() throws Exception { return call.call(); }
        };
    }
}
