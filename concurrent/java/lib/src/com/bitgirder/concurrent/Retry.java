package com.bitgirder.concurrent;

// Implementations are under no obligation to be threadsafe or resettable and
// callers should not inherently assume that a given implementation is either.
public
interface Retry
{
    public
    boolean
    shouldRetry( Throwable th );

    // Can return null to indicate that the next retry can proceed immediately,
    // otherwise returns the duration that should elapse before the next retry
    // begins.
    //
    // This is not an idempotent call -- each call to this method should
    // increment the result returned by subsequence calls to retryCount().
    // Depending on the delay calculation algorithm, different delays may be
    // returned by different calls to this method on the same instance, such as
    // in the case of exponential backoff.
    public
    Duration
    nextDelay();

    // The number of retries that have been begun thus far but which may or may
    // not have completed
    public
    int
    retryCount();
}
