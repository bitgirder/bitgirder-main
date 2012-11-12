package com.bitgirder.log;

public
interface CodeEventSink
{
    // ev should be non-null; impls may fail on receipt of null, discard null
    // silently, or basically do whatever else they want. Impls make no
    // guarantee as to when, if ever, the event will be logged or written to any
    // backing medium or handlers; some sinks may write asynchronously, some may
    // filter, some may drop events under heavy load. Programs which require
    // more specific guarantees should insist on using whichever specific impls
    // provide those guarantees.
    public
    void
    logCode( CodeEvent ev );
}
