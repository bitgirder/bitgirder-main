package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Range;
import com.bitgirder.lang.Lang;
import com.bitgirder.lang.PatternHelper;
import com.bitgirder.lang.Strings;

import java.util.TimeZone;
import java.util.GregorianCalendar;
import java.util.Locale;
import java.util.Date;

import java.util.regex.Pattern;
import java.util.regex.Matcher;

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

    private final static Range< Integer > PRECISION_RANGE =
        Range.closed( 0, 9 );

    private final static TimeZone TZ_UTC = TimeZone.getTimeZone( "UTC" );

    // This regex is actually more permissive than the rfc since it does not
    // check things such as valid days of month, etc. We can tighten this up as
    // our needs dictate it
    private final static Pattern STRICT_RFC3339_TIMESTAMP_PATTERN =
        PatternHelper.compile(
            "(\\d{4})-(\\d{2})-(\\d{2})[Tt](\\d{2}):(\\d{2}):(\\d{2})" +
            "(?:\\.(\\d+){1,9})?(?:([zZ])|([+\\-]\\d{2}:\\d{2}))"
        );

    private final static int RFC3339_GROUP_YEAR = 1;
    private final static int RFC3339_GROUP_MONTH = 2;
    private final static int RFC3339_GROUP_DATE = 3;
    private final static int RFC3339_GROUP_HOUR = 4;
    private final static int RFC3339_GROUP_MINUTE = 5;
    private final static int RFC3339_GROUP_SECONDS = 6;
    private final static int RFC3339_GROUP_FRAC_PART = 7;
    private final static int RFC3339_GROUP_TIME_ZONE_ZULU = 8;
    private final static int RFC3339_GROUP_TIME_ZONE_UTC_OFFSET = 9;

    private final long secs;
    private final int nanos;

    private
    MingleTimestamp( long secs,
                     int nanos )
    {
        this.secs = secs;
        this.nanos = inputs.nonnegativeI( nanos, "nanos" );
    }

    public 
    int 
    hashCode() 
    { 
        return ( (int) secs ) | nanos; 
    }

    public
    boolean
    equals( Object other )
    {
        if ( other == this ) return true;
        if ( ! ( other instanceof MingleTimestamp ) ) return false;

        MingleTimestamp ts2 = (MingleTimestamp) other;
        return nanos == ts2.nanos && secs == ts2.secs;
    }

    public
    int
    compareTo( MingleTimestamp ts2 )
    {
        if ( ts2 == null ) throw new NullPointerException();
        if ( ts2 == this ) return 0;
        
        int res = secs < ts2.secs ? -1 : secs == ts2.secs ? 0 : 1;
        if ( res != 0 ) return res;
        
        res = nanos < ts2.nanos ? -1 : nanos == ts2.nanos ? 0 : 1;
        if ( secs < 0 ) res *= -1;

        return res;
    }

    public long seconds() { return secs; }
    public int nanos() { return nanos; }

    public
    long
    getTimeInMillis()
    {
        return ( secs * 1000L ) + ( nanos / ONE_MILLION );
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
        sb.append( "Z" );
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

        GregorianCalendar cal = new GregorianCalendar();
        cal.setTimeZone( TZ_UTC );
        cal.setTimeInMillis( secs * 1000L ); // discard any fractional part

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

//    public
//    Timestamp
//    asSqlTimestamp()
//    {
//        // See javadocs for java.sql.Timestamp( long ) -- it ultimately stores
//        // it's integral/frac part as this class does
//        Timestamp res = new Timestamp( getTimeInMillis() );
//        res.setNanos( nanos );
//        
//        return res;
//    }

    public
    static
    MingleTimestamp
    fromUnixNanos( long secs,
                   int ns )
    {
        return new MingleTimestamp( secs, ns );
    }

    public
    static
    MingleTimestamp
    fromMillis( long t )
    {
        long secs = t / 1000;
        int ns = Math.abs( (int) ( t % 1000L ) ) * ONE_MILLION;

        return fromUnixNanos( secs, ns );
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

    // We put the parsing in here rather than with the larger mingle parser code
    // since this is really just an rfc3339 parse and is well self-contained;
    // moreover, we want to be able to parse rfc3339 strings into instances of
    // this class as part of our larger parser test code (for Timestamp type
    // restrictions, namely).

    private
    static
    void
    setCalInt( GregorianCalendar c,
               Matcher m,
               int group,
               int calFld )
    {
        int val = Integer.parseInt( m.group( group ) );

        if ( calFld == GregorianCalendar.MONTH ) --val;
        c.set( calFld, val );
    }

    private
    static
    void
    setTimeZone( GregorianCalendar c,
                 Matcher m )
    {
        String tzStr;

        if ( m.group( RFC3339_GROUP_TIME_ZONE_ZULU ) == null )
        {
            tzStr = "GMT" + m.group( RFC3339_GROUP_TIME_ZONE_UTC_OFFSET );
        }
        else tzStr = "UTC";
            
        c.setTimeZone( TimeZone.getTimeZone( tzStr ) );
    }

    private
    static
    int
    getNanos( Matcher m )
    {
        String fracPart = m.group( RFC3339_GROUP_FRAC_PART );

        if ( fracPart == null ) return 0;
        
        int res = Integer.parseInt( fracPart );
        for ( int i = fracPart.length(); i < 9; ++i ) res *= 10;
        
        return res;
    }

    private
    static
    MingleTimestamp
    buildTimestamp( Matcher m )
    {
        GregorianCalendar c = new GregorianCalendar( Locale.US );

        setCalInt( c, m, RFC3339_GROUP_YEAR, GregorianCalendar.YEAR );
        setCalInt( c, m, RFC3339_GROUP_MONTH, GregorianCalendar.MONTH ); // -1
        setCalInt( c, m, RFC3339_GROUP_DATE, GregorianCalendar.DATE );
        setCalInt( c, m, RFC3339_GROUP_HOUR, GregorianCalendar.HOUR_OF_DAY );
        setCalInt( c, m, RFC3339_GROUP_MINUTE, GregorianCalendar.MINUTE );
        setCalInt( c, m, RFC3339_GROUP_SECONDS, GregorianCalendar.SECOND );
        
        int nanos = getNanos( m );
        c.set( GregorianCalendar.MILLISECOND, 0 );
        
        setTimeZone( c, m );

        return new MingleTimestamp( c.getTimeInMillis() / 1000, nanos ); 
    }

    public
    static
    MingleTimestamp
    parse( CharSequence str,
           int colOffset )
        throws MingleSyntaxException
    {
        inputs.notNull( str, "str" );
        inputs.nonnegativeI( colOffset, "colOffset" );

        Matcher m = STRICT_RFC3339_TIMESTAMP_PATTERN.matcher( str );

        if ( m.matches() ) return buildTimestamp( m );
        else
        {
            String msg = "Invalid timestamp: " + Lang.getRfc4627String( str );
            throw new MingleSyntaxException( msg, colOffset );
        }
    }

    public
    static
    MingleTimestamp
    parse( CharSequence str )
        throws MingleSyntaxException
    {
        return parse( str, 0 );
    }

    public
    static
    MingleTimestamp
    create( CharSequence str )
    {
        try { return parse( str ); }
        catch ( MingleSyntaxException se ) 
        { 
            throw new IllegalArgumentException( se.getMessage(), se );
        }
    }
}
