package com.bitgirder.log;

public
interface CodeLogger
extends CodeEventSink
{
    public
    void
    code( Throwable th,
          Object... msg );

    public
    void
    code( Object... msg );

    public
    void
    warn( Throwable th,
          Object... msg );

    public
    void
    warn( Object... msg );
}
