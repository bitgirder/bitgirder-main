package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.PatternHelper;

import java.nio.ByteBuffer;

import java.util.regex.Pattern;
import java.util.regex.Matcher;

/**
 * Class to handle and disambiguate size values regardless of the {@link
 * DataUnit} used. This class arose out of frustrations in passing around int or
 * long arguments and not knowing whether those arguments represented bytes,
 * kilobytes, megs, gigs, etc. The result was this class which bundles a value
 * along with its units and provides mechanisms to convert that value to other
 * units. 
 * <p>
 * Instances of this class are immutable, threadsafe, and delicious.
 */
public
class DataSize
implements Comparable< DataSize >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    public final static DataSize ZERO = new DataSize( 0, DataUnit.BYTE );

    private final static Pattern PARSE_PATTERN =
        PatternHelper.compileCaseInsensitive(
            "\\s*(\\d+)\\s*(" + DataUnit.ABBREV_PATTERN.toString() + ")?" );

    private final long size;

    private final DataUnit unit;

    private
    DataSize()
    {
        size = 0L;
        unit = null;
    }

    private
    void
    validate()
    {
        inputs.nonnegativeL( size, "size" );
        inputs.notNull( unit, "unit" );
    }

    public
    DataSize( CharSequence cs )
    {
        inputs.notNull( cs, "cs" );
 
        Matcher m = inputs.matches( cs, "cs", PARSE_PATTERN );

        this.size = Long.parseLong( m.group( 1 ) );

        String unitStr = m.group( 2 );

        this.unit = 
            unitStr == null 
                ? DataUnit.BYTE : DataUnit.forString( m.group( 2 ) );

        validate();
    }

    public
    DataSize( long size,
              DataUnit unit )
    {
        this.size = size;
        this.unit = unit;

        validate();
    }

    public DataUnit unit() { return unit; }
    public long size() { return size; }

    // Utility method for doing both the public add and subtract methods. Input
    // checking is still done by the public wrappers to this method though.
    private
    DataSize
    doAdd( DataSize other,
           int signum )
    {
        inputs.notNull( other, "other" );

        DataUnit targetUnit =
            unit.compareTo( other.unit ) < 0 ? unit : other.unit;

        long targetSize =
            targetUnit.convert( size, unit ) +
            ( signum * targetUnit.convert( other.size, other.unit ) );
 
        return new DataSize( targetSize, targetUnit );
    }

    public
    DataSize
    add( DataSize other )
    {
        inputs.notNull( other, "other" );
        return doAdd( other, 1 );
    }

    public
    DataSize
    subtract( DataSize other )
    {
        inputs.notNull( other, "other" );

        if ( compareTo( other ) < 0 )
        {
            throw new ArithmeticException(
                "Call to subtract would produce a negative size: " +
                this + " - " + other );
        }

        return doAdd( other, -1 );
    }

    public
    DataSize
    times( int i )
    {
        inputs.nonnegativeI( i, "i" );
        return new DataSize( size * i, unit );
    }

    public
    int
    compareTo( DataSize o )
    {
        if ( o == null ) throw new NullPointerException();

        // The unit used for comparison will be the most fine grained of the two
        // (to avoid precision loss)
        DataUnit compUnit = unit.compareTo( o.unit ) > 0 ? o.unit : unit;

        long v1 = compUnit.convert( size, unit );
        long v2 = compUnit.convert( o.size, o.unit );

        return v1 < v2 ? -1 : ( v1 == v2 ? 0 : 1 );
    }

    /**
     * Returns the number of bytes represented by this DataSize.
     */
    public
    long
    getByteCount()
    {
        return DataUnit.BYTE.convert( size, unit );
    }

    public
    int
    getIntByteCount()
    {
        long res = getByteCount();

        if ( res > Integer.MAX_VALUE )
        {
            throw state.createFail(
                "Data size", toString(), 
                "is too large to represent with an int byte count" );
        }
        else return (int) res;
    }

    // More succinct alias for getByteCount() -- should probably just replace
    // getByteCount() with this only
    public long asBytes() { return getByteCount(); }
    public long asKB() { return DataUnit.KILOBYTE.convert( size, unit ); }

    public
    boolean
    equals( Object o )
    {
        return o == null 
            ? false
            : o instanceof DataSize 
                ? compareTo( (DataSize) o ) == 0
                : false;
    }

    public
    int
    hashCode()
    {
        return new Long( getByteCount() ).hashCode();
    }

    /**
     * Returns a human readable string indicating the value and unit of this
     * DataSize.
     */
    public
    String
    toString()
    {
        return Long.toString( size ) + unit;
    }

    public
    static
    DataSize
    ofTerabytes( int size )
    {
        return new DataSize( size, DataUnit.TERABYTE );
    }

    public
    static
    DataSize
    ofGigabytes( long size )
    {
        return new DataSize( size, DataUnit.GIGABYTE );
    }

    /**
     * Constructs a DataSize object representing the given number of Megabytes.
     */
    public
    static
    DataSize
    ofMegabytes( long size )
    {
        return new DataSize( size, DataUnit.MEGABYTE );
    }

    /**
     * Constructs a DataSize object representing the given number of Kilobytes.
     */
    public
    static
    DataSize
    ofKilobytes( long size )
    {
        return new DataSize( size, DataUnit.KILOBYTE );
    }

    /**
     * Constructs a DataSize object representing the given number of Bytes.
     */
    public
    static
    DataSize
    ofBytes( long size )
    {
        return new DataSize( size, DataUnit.BYTE );
    }

    public
    static
    DataSize
    of( byte[] data )
    {
        return ofBytes( inputs.notNull( data, "data" ).length );
    }

    // Note: there was previously a method of( ByteBuffer ) which keyed off off
    // ByteBuffer.capacity. It was a confusing method since what generally was
    // intended was to key off of remaining() instead. So, the of( ByteBuffer )
    // method is removed, since it wasn't really used anywhere anyway

    public
    static
    DataSize
    remaining( ByteBuffer bb )
    {
        return ofBytes( inputs.notNull( bb, "bb" ).remaining() );
    }

    public
    static
    DataSize
    fromString( CharSequence s )
    {
        return new DataSize( inputs.notNull( s, "s" ) );
    }

    private
    static
    DataSize
    doMinOrMax( DataSize ds1,
                DataSize ds2,
                int signum )
    {
        inputs.notNull( ds1, "ds1" );
        inputs.notNull( ds2, "ds2" );

        return ( signum * ds1.compareTo( ds2 ) ) < 0 ? ds1 : ds2;
    }

    public
    static
    DataSize
    min( DataSize ds1,
         DataSize ds2 )
    {
        return doMinOrMax( ds1, ds2, 1 );
    }

    public
    static
    DataSize
    max( DataSize ds1,
         DataSize ds2 )
    {
        return doMinOrMax( ds1, ds2, -1 );
    }
}
