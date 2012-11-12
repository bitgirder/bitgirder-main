package com.bitgirder.log;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
abstract
class AbstractCodeLogger
implements CodeLogger
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    // Could let impls supply this result in their own way at some point if
    // needed
    private long time() { return System.currentTimeMillis(); }

    // ev will be not-null when this is called
    protected
    abstract
    void
    logCodeImpl( CodeEvent ev );

    public
    final
    void
    logCode( CodeEvent ev )
    {
        if ( ev != null ) logCodeImpl( ev );
    }

    public
    final
    void
    code( Throwable th,
          Object... msg )
    {
        logCode( 
            CodeEvents.create( CodeEventType.CODE, msg, th, null, time() ) );
    }

    public
    final
    void
    code( Object... msg )
    {
        logCode( 
            CodeEvents.create( CodeEventType.CODE, msg, null, null, time() ) );
    }

    public
    final
    void
    warn( Throwable th,
          Object... msg )
    {
        logCode( 
            CodeEvents.create( CodeEventType.WARN, msg, th, null, time() ) );
    }

    public
    final
    void
    warn( Object... msg )
    {
        logCode( 
            CodeEvents.create( CodeEventType.WARN, msg, null, null, time() ) );
    }
}
