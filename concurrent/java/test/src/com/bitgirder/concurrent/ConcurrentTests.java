package com.bitgirder.concurrent;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.test.Test;

@Test
final
class ConcurrentTests
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private
    void
    assertDuration( Duration d1,
                    Duration d2 )
    {
        // simple but all we need for now
        state.equalString( d1.toString(), d2.toString() );
    }

    private
    void
    assertDuration( Duration expct,
                    String... strs )
    {
        for ( String str : strs )
        {
            Duration actual = Duration.fromString( str );

            state.equal( expct.getTimeUnit(), actual.getTimeUnit() );
            state.equal( expct.getDuration(), actual.getDuration() );
        }
    }

    @Test
    private
    void
    testDurationFromString()
        throws Exception
    {
        assertDuration(
            Duration.fromNanos( 12 ), "12ns", "12 nanos", "12 nanoseconds" );

        assertDuration( 
            Duration.fromMillis( 13 ),
            "13ms", "13 ms", " 13 ms ", " 13ms ", " 13ms", "13millis",
            "13 milliseconds", "13"
        );

        assertDuration( 
            Duration.fromSeconds( 5 ), "5s", "5secs", "5sec", "5seconds" );

        assertDuration(
            Duration.fromMinutes( 3 ),
            "3m", "3min", "3mins", "3minute", "3 minutes" );
        
        assertDuration(
            Duration.fromHours( 7 ), "7h", "7 hr", "7 hrs", "7hour", "7hours" );

        assertDuration( Duration.fromDays( 2 ), "2d", "2day", "2 days" );

        assertDuration( Duration.fromDays( 28 ), "2 fortnights", "2fortnight" );
    }

    @Test( expected = IllegalArgumentException.class )
    private
    void
    testParseFails()
    {
        Duration.fromString( "12x" );
    }

    @Test
    private
    void
    testExpirationAsUnixTime()
    {
        long ttlSecs = 3000;
        long tolerance = 1; // allow +/- 1 seconds at the most 

        Duration ttl = Duration.fromSeconds( ttlSecs );

        long nowSecs = System.currentTimeMillis() / 1000;
        long exp = ttl.getExpirationAsUnixTime();

        long diff = exp - nowSecs;
        state.isTrue( diff <= ttlSecs + tolerance && diff >= tolerance );
    }

    @Test
    private
    void
    testBackoff()
    {
        Duration d = Duration.fromMillis( 50 );

        d = d.backOff();
        state.equal( 100L, d.asMillis() );

        d = d.backOff().backOff().backOff().backOff();
        state.equal( 1600L, d.asMillis() );
        state.equalInt( 1, (int) d.asSeconds() );
    }

    private final static class Exception1 extends Exception {}
    private final static class Exception2 extends Exception {}
    private final static class Exception3 extends Exception {}

    @Test
    private
    void
    testDefaultRetryImplNullBackoff()
    {
        Retry r = Concurrency.createRetry( 3, Throwable.class );

        for ( int i = 1; i <= 2; ++i )
        {
            state.isTrue( r.nextDelay() == null );
            state.equalInt( i, r.retryCount() );

            Throwable th = i % 2 == 0 ? new Throwable() : new Exception1();
            state.isTrue( r.shouldRetry( th ) );
        }

        r.nextDelay(); // retries now exhausted
        state.isFalse( r.shouldRetry( new Throwable() ) );
    }

    @Test
    private
    void
    testDefaultRetryImplNonNullBackoff()
    {
        Retry r = 
            Concurrency.createRetry( 
                3, 
                Duration.fromMillis( 200 ), 
                Exception1.class, Exception2.class );

        state.equalInt( 0, r.retryCount() );
        state.equalInt( 200, (int) r.nextDelay().asMillis() );

        state.equalInt( 1, r.retryCount() );
        state.equalInt( 400, (int) r.nextDelay().asMillis() );

        state.isTrue( r.shouldRetry( new Exception1() ) );
        state.isTrue( r.shouldRetry( new Exception2() ) );
        state.isFalse( r.shouldRetry( new Exception3() ) );

        r.nextDelay();
        state.isFalse( r.shouldRetry( new Exception1() ) );
    }
}
