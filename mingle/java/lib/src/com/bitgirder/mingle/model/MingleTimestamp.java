package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Range;

import com.bitgirder.mingle.parser.MingleParsers;

import com.bitgirder.parser.SyntaxException;

import java.util.TimeZone;
import java.util.GregorianCalendar;
import java.util.Locale;
import java.util.Date;

import java.sql.Timestamp;

public
final
class MingleTimestamp
implements MingleValue,
           Comparable< MingleTimestamp >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static int ONE_MILLION = 1000000;

    private final static Range< Integer > NANOS_RANGE =
        Range.openMax( 0, 1000000000 );
    
    private final static Range< Integer > MILLIS_RANGE =
        Range.openMax( 0, 1000 );

    private final static TimeZone DEFAULT_TIME_ZONE =
        TimeZone.getTimeZone( "UTC" );

    private final static Range< Integer > PRECISION_RANGE =
        Range.closed( 0, 9 );

    private final static TimeZone TZ_UTC = TimeZone.getTimeZone( "UTC" );

    private final GregorianCalendar cal;
    private final int nanos;

    private
    MingleTimestamp( Builder b )
    {
        cal = new GregorianCalendar( b.timeZone, Locale.US );

        // note the month subtracts 1 to adjust back to GregorianCalendar's
        // 0-indexed months
        cal.set( GregorianCalendar.YEAR, b.year );
        cal.set( GregorianCalendar.MONTH, b.month - 1 );
        cal.set( GregorianCalendar.DATE, b.date );
        cal.set( GregorianCalendar.HOUR_OF_DAY, b.hour );
        cal.set( GregorianCalendar.MINUTE, b.minute );
        cal.set( GregorianCalendar.SECOND, b.seconds );

        // Since we keep our own fractional precision, cancel out the one auto
        // set by the calendar constructor, since it will affect our comparisons
        cal.set( GregorianCalendar.MILLISECOND, 0 );

        this.nanos = b.nanos;
    }

    public int hashCode() { return cal.hashCode() | nanos; }

    public
    boolean
    equals( Object other )
    {
        if ( other == this ) return true;
        else if ( other instanceof MingleTimestamp )
        {
            MingleTimestamp ts2 = (MingleTimestamp) other;
            return nanos == ts2.nanos && cal.equals( ts2.cal );
        }
        else return false;
    }

    public
    int
    compareTo( MingleTimestamp ts2 )
    {
        if ( ts2 == null ) throw new NullPointerException();
        else
        {
            if ( ts2 == this ) return 0;
            else
            {
                int res = cal.compareTo( ts2.cal );
                return res == 0 ? nanos - ts2.nanos : res;
            }
        }
    }

    private
    boolean
    isUtc( TimeZone tz )
    {
        String id = tz.getID();

        return id.equalsIgnoreCase( "UTC" ) || id.equalsIgnoreCase( "Zulu" );
    }

    private
    void
    appendTwoDigitPad( StringBuilder sb,
                       int val )
    {
        if ( val < 10 ) sb.append( '0' );
        sb.append( val );
    }

    private
    void
    appendRfc3339TimeZone( StringBuilder sb )
    {
        TimeZone tz = cal.getTimeZone();

        if ( isUtc( tz ) ) sb.append( "Z" );
        else
        {
            // offset is in millis which we truncate to some number of minutes;
            // we leave the sign off the truncated form so that negative values
            // don't find their way into the string appends; offset retains the
            // sign for our checks
            int offset = cal.get( GregorianCalendar.ZONE_OFFSET );
            int trunc = Math.abs( ( offset / 1000 ) / 60 );

            sb.append( offset < 0 ? '-' : '+' );

            appendTwoDigitPad( sb, trunc / 60 );
            sb.append( ':' );
            appendTwoDigitPad( sb, trunc % 60 );
        }
    }

    private
    void
    appendPrecision( StringBuilder sb,
                     int precision )
    {
        if ( precision > 0 )
        {
            String decStr = String.format( "%1$09d", Math.max( nanos, 0 ) );
            decStr = decStr.substring( 0, precision );
            sb.append( '.' ).append( decStr );
        }
    }
 
    public
    CharSequence
    getRfc3339String( int precision )
    {
        inputs.inRange( precision, "precision", PRECISION_RANGE );

        StringBuilder res = new StringBuilder( 25 );

        res.append( cal.get( GregorianCalendar.YEAR ) );
        res.append( '-' );
        appendTwoDigitPad( res, cal.get( GregorianCalendar.MONTH ) + 1 );
        res.append( '-' );
        appendTwoDigitPad( res, cal.get( GregorianCalendar.DATE ) );
        res.append( 'T' );
        appendTwoDigitPad( res, cal.get( GregorianCalendar.HOUR_OF_DAY ) );
        res.append( ':' );
        appendTwoDigitPad( res, cal.get( GregorianCalendar.MINUTE ) );
        res.append( ':' );
        appendTwoDigitPad( res, cal.get( GregorianCalendar.SECOND ) );
 
        appendPrecision( res, precision );
        appendRfc3339TimeZone( res );

        return res;
    }

    public
    CharSequence
    getRfc3339String()
    {
        return getRfc3339String( PRECISION_RANGE.max() );
    }

    @Override public String toString() { return getRfc3339String().toString(); }

    public 
    long 
    getTimeInMillis() 
    { 
        // we still use getTimeInMillis() for convenience and then just add back
        // the millis which we store on our own
        return cal.getTimeInMillis() + ( nanos / ONE_MILLION );
    }

    public Date asJavaDate() { return new Date( getTimeInMillis() ); }

    public
    GregorianCalendar
    asJavaCalendar()
    {
        GregorianCalendar res = (GregorianCalendar) cal.clone();
        res.set( GregorianCalendar.MILLISECOND, nanos / ONE_MILLION );

        return res;
    }

    public
    Timestamp
    asSqlTimestamp()
    {
        // See javadocs for java.sql.Timestamp( long ) -- it ultimately stores
        // it's integral/frac part as this class does
        Timestamp res = new Timestamp( getTimeInMillis() );
        res.setNanos( nanos );
        
        return res;
    }

    public
    static
    MingleTimestamp
    create( CharSequence str )
    {
        return MingleParsers.createTimestamp( inputs.notNull( str, "str" ) );
    }

    public
    static
    MingleTimestamp
    parse( CharSequence str )
        throws SyntaxException
    {
        return MingleParsers.parseTimestamp( inputs.notNull( str, "str" ) );
    }

    public
    static
    MingleTimestamp
    fromMillis( long t )
    {
        GregorianCalendar c = new GregorianCalendar();
        c.setTimeZone( TZ_UTC );
        c.setTimeInMillis( t );

        return new Builder().setFromCalendar( c ).build();
    }

    public
    static
    MingleTimestamp
    fromDate( Date d )
    {
        return fromMillis( inputs.notNull( d, "d" ).getTime() );
    }

    public
    static
    MingleTimestamp
    now()
    {
        return fromMillis( System.currentTimeMillis() );
    }

    // We may also add flags such as setLenient( boolean ) to be passed through
    // during build to do extra validation; for now the default is the default
    // of Calendar (lenient).
    public
    final
    static
    class Builder
    {
        private int year;
        private int month;
        private int date;
        private int hour;
        private int minute;
        private int seconds;
        private int nanos;
        private TimeZone timeZone = DEFAULT_TIME_ZONE;

        public
        Builder
        setYear( int year )
        {
            this.year = year;
            return this;
        }

        // Note: unlike java.util.Calendar, this class takes a 1-based month (1
        // for January)
        public
        Builder
        setMonth( int month )
        {
            this.month = month;
            return this;
        }

        public
        Builder
        setDate( int date )
        {
            this.date = date;
            return this;
        }

        public
        Builder
        setHour( int hour )
        {
            this.hour = hour;
            return this;
        }

        public
        Builder
        setMinute( int minute )
        {
            this.minute = minute;
            return this;
        }

        public
        Builder
        setSeconds( int seconds )
        {
            this.seconds = seconds;
            return this;
        }

        public
        Builder
        setNanos( int nanos )
        {
            this.nanos = inputs.inRange( nanos, "nanos", NANOS_RANGE );
            return this;
        }

        public
        Builder
        setMillis( int millis )
        {
            inputs.inRange( millis, "millis", MILLIS_RANGE );
            this.nanos = millis * ONE_MILLION;

            return this;
        }

        public
        Builder
        setFraction( CharSequence frac )
        {
            int len = frac.length();

            inputs.isTrue( 
                len < 10, 
                "Fraction string is too long to represent nanos:", frac );

            char[] intStr = new char[ 9 ];
            int i;
            for ( i = 0; i < len; ++i ) intStr[ i ] = frac.charAt( i );
            while ( i < 9 ) intStr[ i++ ] = '0';

            return setNanos( Integer.parseInt( new String( intStr ) ) );
        }

        public
        Builder
        setTimeZone( TimeZone timeZone )
        {
            this.timeZone = inputs.notNull( timeZone, "timeZone" );
            return this;
        }

        public
        Builder
        setTimeZone( String tzStr )
        {
            inputs.notNull( tzStr, "tzStr" );
            return setTimeZone( TimeZone.getTimeZone( tzStr ) );
        }

        public
        Builder
        setFromCalendar( GregorianCalendar cal )
        {
            inputs.notNull( cal, "cal" );

            setYear( cal.get( GregorianCalendar.YEAR ) );
            setMonth( cal.get( GregorianCalendar.MONTH ) + 1 );
            setDate( cal.get( GregorianCalendar.DATE ) );
            setHour( cal.get( GregorianCalendar.HOUR_OF_DAY ) );
            setMinute( cal.get( GregorianCalendar.MINUTE ) );
            setSeconds( cal.get( GregorianCalendar.SECOND ) );
            setMillis( cal.get( GregorianCalendar.MILLISECOND ) );
            setTimeZone( cal.getTimeZone() );

            return this;
        } 

        public MingleTimestamp build() { return new MingleTimestamp( this ); }
    }
}
