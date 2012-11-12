package com.bitgirder.concurrent;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import java.util.List;

import java.util.concurrent.ExecutorService;

public
final
class Concurrency
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private Concurrency() {}

    private
    final
    static
    class RetryImpl
    extends AbstractRetry
    {
        private final List< Class< ? extends Throwable > > retriables;

        // Does some input checks here on behalf of public frontend methods
        private
        RetryImpl( int maxRetries,
                   Duration backoffSeed,
                   Class< ? >[] retriables )
        {
            super( maxRetries, backoffSeed );

            this.retriables = 
                checkRetriables( inputs.noneNull( retriables, "retriables" ) );
        }

        public
        boolean
        shouldRetry( Throwable th )
        {
            if ( retryCount() < maxRetries() )
            {
                for ( Class< ? extends Throwable > c : retriables )
                {
                    if ( c.isInstance( th ) ) return true;
                }

                return false;
            }
            else return false;
        }

        // util method to check and convert unparamterized class objects as
        // throwable subclasses
        private
        static
        List< Class< ? extends Throwable > >
        checkRetriables( Class< ? >[] retriables )
        {
            List< Class< ? extends Throwable > > res = 
                Lang.newList( retriables.length );
            
            for ( Class< ? > c : retriables )
            {
                res.add( c.asSubclass( Throwable.class ) );
            }

            return Lang.unmodifiableList( res );
        }
    }

    public
    static
    Retry
    createRetry( int maxRetries,
                 Class< ? >... retriables )
    {
        return new RetryImpl( maxRetries, null, retriables );
    }

    public
    static
    Retry
    createRetry( int maxRetries,
                 Duration backoffSeed,
                 Class< ? >... retriables )
    {
        inputs.notNull( backoffSeed, "backoffSeed" );
        return new RetryImpl( maxRetries, backoffSeed, retriables );
    }

    public
    static
    boolean
    shutdownAndWait( ExecutorService es,
                     Duration wait )
        throws InterruptedException
    {
        inputs.notNull( es, "es" );
        inputs.notNull( wait, "wait" );

        es.shutdown();
        return es.awaitTermination( wait.getDuration(), wait.getTimeUnit() );
    }
}
