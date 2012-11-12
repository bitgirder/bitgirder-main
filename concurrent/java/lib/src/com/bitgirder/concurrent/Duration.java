package com.bitgirder.concurrent;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.PatternHelper;
import com.bitgirder.lang.Strings;
import com.bitgirder.lang.Lang;

import java.util.Map;

import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.TimeoutException;

import java.util.regex.Pattern;
import java.util.regex.Matcher;

import java.text.DecimalFormat;

/**
 * Class binding together {@link TimeUnit} and actual time value to represent
 * time durations.
 * <p>
 * This class is threadsafe.
 * <p>
 * <b>Note:</b> Unless otherwise stated, the behavior of methods in this class
 * follow the conventions of {@link TimeUnit#convert} when an over- or underflow
 * would occur.
 */
public
final
class Duration
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    // See static initializer block
    private final static Pattern DURATION_PARSE_PAT;
    private final static Map< String, Object > ABBREVS;

    // JDK5 TimeUnit doesn't have these values, so we create an internal enum to
    // use along with ABBREVS for mapping to these units
    private static enum Jdk5TimeUnitCompat { MINUTES, HOURS, DAYS; }

    // Sometimes you've gotta have a sense of humor about things
    private final static String FORTNIGHT = "FORTNIGHT";

    public final static int DEFAULT_FRACTION_DIGITS = 3;

    public final static Duration ZERO = Duration.fromNanos( 0 );

    // Just to help avoid typos with the number of zeroes
    private final static int MILLION = 1000000;

    private final TimeUnit tu;

    private final long dur;

    private
    Duration( long dur,
              TimeUnit tu )
    {
        this.dur = dur;
        this.tu = tu;
    }

    /**
     * Causes the calling thread to sleep for the amount of time represented by
     * this Duration.
     */
    public
    void
    sleep()
        throws InterruptedException
    {
        tu.sleep( dur );
    }

    public TimeUnit getTimeUnit() { return tu; }
    public long getDuration() { return dur; }

    void
    join( Thread th )
        throws InterruptedException
    {
        long nanos = asNanos();
        long millis = nanos / MILLION;
        int nanoPart = (int) ( nanos - ( millis * MILLION ) );

        inputs.notNull( th, "th" );
        th.join( millis, nanoPart );
    }

    void
    joinOrFail( Thread th )
        throws InterruptedException,
               TimeoutException
    {
        join( th );

        if ( ! th.getState().equals( Thread.State.TERMINATED ) )
        {
            throw new TimeoutException();
        }
    }

    /**
     * Gets the value of this object in seconds.
     */
    public
    long
    asSeconds()
    {
        return tu.toSeconds( dur );
    }

    /**
     * Gets the value of this object in milliseconds.
     */
    public
    long
    asMillis()
    {
        return tu.toMillis( dur );
    }

    /**
     * Gets the value of this object in nanoseconds.
     */
    public
    long
    asNanos()
    {
        return tu.toNanos( dur );
    }

    private
    long
    checkWraparound( long l )
    {
        return l < 0 ? Long.MAX_VALUE : l;
    }

    /**
     * Gets the time at which this Duration will be reached relative to
     * this method's invocation. This method can be used to set monitoring or
     * other alarm clock values, or for monitoring timeouts. The time value
     * returned is measured in seconds and is relative to the result of {@link
     * System#currentTimeMillis()}.
     */
    public
    long
    getExpirationAsUnixTime()
    {
        return checkWraparound(
            ( System.currentTimeMillis() / 1000 ) + asSeconds() );
    }

    public
    long
    getExpirationInMillis()
    {
        return checkWraparound( System.currentTimeMillis() + asMillis() );
    }

    // This version prevents the result from wrapping around, but doesn't do all
    // it could to deal with the situation: the slighlty more sophisticated
    // approach would be to first attempt to promote this instance to a more
    // coarse-grained (in terms of TimeUnit) version of itself, which would make
    // room to continue backing off. For now, this implementation will suffice,
    // but we can always change it later.
    public
    Duration
    backOff( int power )
    {
        inputs.positiveI( power, "power" );

        int fact = (int) Math.pow( 2, power );

        long newDur = Long.MAX_VALUE / fact > dur ? dur * fact : Long.MAX_VALUE;
        return new Duration( newDur, tu );
    }

    public Duration backOff() { return backOff( 1 ); }

    /**
     * Returns a human-readable String describing this duration.
     */
    public
    String
    toString()
    {
        return dur + " " + tu.toString().toLowerCase();
    }

    public
    String
    toStringSeconds( int precision )
    {
        inputs.nonnegativeI( precision, "precision" );

        DecimalFormat f = new DecimalFormat();
        f.setMaximumFractionDigits( precision );
        f.setMinimumFractionDigits( precision );
        
        // the number of seconds as a double at its most granular
        double secs = 
            ( (double) TimeUnit.NANOSECONDS.convert( dur, tu ) ) * 1e-9d;

        return f.format( secs ) + "s";
    }

    public
    String
    toStringSeconds() 
    { 
        return toStringSeconds( DEFAULT_FRACTION_DIGITS ); 
    }

    public
    < V >
    V
    getFuture( Future< V > fut )
        throws InterruptedException,
               ExecutionException,
               TimeoutException
    {
        inputs.notNull( fut, "fut" );
        return fut.get( dur, tu );
    }

    /**
     * Returns a Duration object lasting the specified number of minutes.
     *
     *  @param mins The number of minutes associated with the intended Duration.
     */
    public
    static
    Duration
    fromMinutes( long mins )
    {
        return fromSeconds( mins * 60 );
    }

    public
    static
    Duration
    fromHours( long hrs )
    {
        return fromMinutes( hrs * 60 );
    }

    public
    static
    Duration
    fromDays( long days )
    {
        return fromHours( days * 24 );
    }

    public
    static
    Duration
    fromFortnights( long fortnights )
    {
        return fromDays( 14 * fortnights );
    }

    /**
     * Creates a Duration object representing the specified duration in seconds.
     *
     * @param seconds The number of seconds associated with the intended
     * Duration.
     */
    public
    static
    Duration
    fromSeconds( long seconds )
    {
        return new Duration( seconds, TimeUnit.SECONDS );
    }
    
    /**
     * Creates a Duration object representing the specified duration in
     * milliseconds.
     *
     *  @param millis The number of milliseconds associated with the intended
     *  Duration.
     */
    public
    static
    Duration
    fromMillis( long millis )
    {
        return new Duration( millis, TimeUnit.MILLISECONDS );
    }

    /**
     * Creates a Duration object representing the specified duration in
     * nanoseconds.
     *
     * @param nanos The number of nanoseconds associated with the intended
     * Duration.
     */
    public
    static
    Duration
    fromNanos( long nanos )
    {
        return new Duration( nanos, TimeUnit.NANOSECONDS );
    }

    public static Duration nowNanos() { return fromNanos( System.nanoTime() ); }

    public
    static
    Duration
    sinceNanos( long nanos )
    {
        return fromNanos( System.nanoTime() - nanos );
    }

    public
    static
    Duration
    fromString( CharSequence cs )
    {
        Matcher m = inputs.matches( cs, "cs", DURATION_PARSE_PAT );
        
        long dur = Long.parseLong( m.group( 1 ) );

        Object u = ABBREVS.get( m.group( 2 ) );
        if ( u == null ) u = TimeUnit.MILLISECONDS;

        if ( u instanceof TimeUnit ) return new Duration( dur, (TimeUnit) u );
        else if ( u instanceof Jdk5TimeUnitCompat )
        {
            switch ( (Jdk5TimeUnitCompat) u )
            {
                case MINUTES: return fromMinutes( dur );
                case HOURS: return fromHours( dur );
                case DAYS: return fromDays( dur );
                default: throw state.createFail( "Unexpected unit:", u );
            }
        }
        else if ( u == FORTNIGHT ) return fromFortnights( dur );
        else throw state.createFail( "Unexpected unit:", u );
    }

    public
    static
    Duration
    create( long dur,
            TimeUnit tu )
    {
        return new Duration( dur, inputs.notNull( tu, "tu" ) );
    }

    static
    {
        ABBREVS = Lang.newMap();

        ABBREVS.put( "ns", TimeUnit.NANOSECONDS );
        ABBREVS.put( "nanoseconds", TimeUnit.NANOSECONDS );
        ABBREVS.put( "nanos", TimeUnit.NANOSECONDS );

        ABBREVS.put( "ms", TimeUnit.MILLISECONDS );
        ABBREVS.put( "milliseconds", TimeUnit.MILLISECONDS );
        ABBREVS.put( "millis", TimeUnit.MILLISECONDS );

        ABBREVS.put( "s", TimeUnit.SECONDS );
        ABBREVS.put( "seconds", TimeUnit.SECONDS );
        ABBREVS.put( "sec", TimeUnit.SECONDS );
        ABBREVS.put( "secs", TimeUnit.SECONDS );

        ABBREVS.put( "m", Jdk5TimeUnitCompat.MINUTES );
        ABBREVS.put( "min", Jdk5TimeUnitCompat.MINUTES );
        ABBREVS.put( "mins", Jdk5TimeUnitCompat.MINUTES );
        ABBREVS.put( "minute", Jdk5TimeUnitCompat.MINUTES );
        ABBREVS.put( "minutes", Jdk5TimeUnitCompat.MINUTES );

        ABBREVS.put( "h", Jdk5TimeUnitCompat.HOURS );
        ABBREVS.put( "hr", Jdk5TimeUnitCompat.HOURS );
        ABBREVS.put( "hrs", Jdk5TimeUnitCompat.HOURS );
        ABBREVS.put( "hour", Jdk5TimeUnitCompat.HOURS );
        ABBREVS.put( "hours", Jdk5TimeUnitCompat.HOURS );

        ABBREVS.put( "d", Jdk5TimeUnitCompat.DAYS );
        ABBREVS.put( "day", Jdk5TimeUnitCompat.DAYS );
        ABBREVS.put( "days", Jdk5TimeUnitCompat.DAYS );

        ABBREVS.put( "fortnight", FORTNIGHT );
        ABBREVS.put( "fortnights", FORTNIGHT );

        DURATION_PARSE_PAT = 
            PatternHelper.compileCaseInsensitive(
                "^\\s*(-?\\s*\\d+)\\s*(" + 
                Strings.join( "|", ABBREVS.keySet() ) + ")?\\s*$" );
    }
}
