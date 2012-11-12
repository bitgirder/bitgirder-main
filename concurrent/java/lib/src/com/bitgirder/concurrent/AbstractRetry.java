package com.bitgirder.concurrent;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
abstract
class AbstractRetry
implements Retry
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final int maxRetries;

    private Duration nextDelay;
    private int retryCount;

    protected
    AbstractRetry( int maxRetries,
                   Duration backoffSeed )
    {
        this.maxRetries = inputs.positiveI( maxRetries, "maxRetries" );
        nextDelay = backoffSeed;
    }

    public 
    final
    Duration 
    nextDelay() 
    { 
        ++retryCount;

        if ( nextDelay == null ) return null;
        else
        {
            Duration res = nextDelay;
            nextDelay = nextDelay.backOff();

            return res;
        }
    }

    public final int retryCount() { return retryCount; }
    public final int maxRetries() { return maxRetries; }
}
