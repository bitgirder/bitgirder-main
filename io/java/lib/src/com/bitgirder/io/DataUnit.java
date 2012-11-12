package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.PatternHelper;
import com.bitgirder.lang.Strings;

import java.util.Map;
import java.util.HashMap;
import java.util.Arrays;

import java.util.regex.Pattern;
import java.util.regex.Matcher;

/**
 * An enum representing various units of data and providing methods for
 * converting between them. This implementation is modeled after that of {@link
 * java.util.concurrent.TimeUnit}.
 */
public
enum DataUnit
{
    // Note: The order of declaration here is important, so that Enum.compareTo
    // is consistent with ascending size of the unit

    /**
     * A byte.
     */
    BYTE( "B", 0 ),

    /**
     * 2^10 bytes.
     */
    KILOBYTE( "KB", 1 ),

    /**
     * 2^20 bytes.
     */
    MEGABYTE( "MB", 2 ),

    /**
     * 2^30 bytes.
     */
    GIGABYTE( "GB", 3 ),

    /**
     * 2^40 bytes.
     */
    TERABYTE( "TB", 4 );

    final static Map< String, DataUnit > ABBREVS = 
        new HashMap< String, DataUnit >();

    final static Pattern ABBREV_PATTERN;

    static
    {
        ABBREVS.put( "b", BYTE );
        ABBREVS.put( "byte", BYTE );
        ABBREVS.put( "bytes", BYTE );
        ABBREVS.put( "k", KILOBYTE );
        ABBREVS.put( "kb", KILOBYTE );
        ABBREVS.put( "kilobyte", KILOBYTE );
        ABBREVS.put( "kilobytes", KILOBYTE );
        ABBREVS.put( "m", MEGABYTE );
        ABBREVS.put( "mb", MEGABYTE );
        ABBREVS.put( "megabyte", MEGABYTE );
        ABBREVS.put( "megabytes", MEGABYTE );
        ABBREVS.put( "g", GIGABYTE );
        ABBREVS.put( "gb", GIGABYTE );
        ABBREVS.put( "gigabyte", GIGABYTE );
        ABBREVS.put( "gigabytes", GIGABYTE );
        ABBREVS.put( "t", TERABYTE );
        ABBREVS.put( "tb", TERABYTE );
        ABBREVS.put( "terabyte", TERABYTE );
        ABBREVS.put( "terabytes", TERABYTE );
    
        // Note that capture group 1 will contain just the text on a match
        ABBREV_PATTERN = PatternHelper.compileCaseInsensitive(
            "\\s*(" + Strings.join( "|", ABBREVS.keySet() ) + ")\\s*" );
    }

    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private static final long[] multipliers = new long[] { 
        1L << 10, 
        1L << 20,
        1L << 30, 
        1L << 40 
    };

    // The maximum value allowed for the given conversion which would not
    // induce overflow
    private static final long[] maximae = new long[]
    {
        Long.MAX_VALUE / ( 1L << 10 ), // B->KB, KB->MB, MB->GB, GB->TB
        Long.MAX_VALUE / ( 1L << 20 ), // B->MB, KB->GB, MB->TB
        Long.MAX_VALUE / ( 1L << 30 ), // B->GB, KB->TB
        Long.MAX_VALUE / ( 1L << 40 )  // B->TB
    };

    private final String abbrev;
    private final int index;

    private
    DataUnit( String abbrev,
              int index )
    {
        this.abbrev = abbrev;
        this.index = index;
    }

    /**
     * Converts the given value in the given units into a value in this
     * DataUnit. In the event that overflow would occur, this method returns
     * Long.MAX_VALUE. Also note that when converting from a fine-grained unit
     * to a more coarse grained one (such as byte to megabyte), some precision
     * may be lost due to integer arithmetic.
     * <p>
     * Examples:
     *  <ul>
     *      <li><code>BYTE.convert( 2, KILOBYTE )</code> returns 2048.</li>
     *      <li><code>KILOBYTE.convert( 2048, BYTE )</code> returns 2.</li>
     *      <li><code>KILOBYTE.convert( 2047, BYTE )</code> returns 1 (loss of
     *      precision due to integer division.</li>
     *  </ul>
     *
     * @throws IllegalArgumentException If count is negative.
     */
    // TODO: make sure unit tests assert the examples given above.
    public
    long
    convert( long count,
             DataUnit unit )
    {
        inputs.nonnegativeL( count, "count" );
        inputs.notNull( unit, "unit" );

        long res;

        int mult = unit.index - index;

        if ( mult == 0 ) res = count;
        else if ( mult < 0 ) res = count / multipliers[ -mult - 1 ];
        else 
        {
            if ( count > maximae[ mult - 1 ] ) res = Long.MAX_VALUE;
            else res = count * multipliers[ mult - 1 ];
        }

        return res;
    }

    /**
     * Returns a more programmer-friendly version of this unit's name.
     */
    public
    String
    toString()
    {
        return abbrev;
    }

    public
    static
    DataUnit
    forString( String s )
    {
        inputs.notNull( s, "s" );

        Matcher m = ABBREV_PATTERN.matcher( s );
 
        inputs.isTrue( m.matches(),
            "Unknown abbreviation for DataUnit: " + s + "(doesn't match " +
            "pattern '" + ABBREV_PATTERN + "')" );

        DataUnit res = ABBREVS.get( m.group( 1 ).toLowerCase() );

        // sanity check on our handcoded table above
        state.isTrue( res != null, 
            "No abbreviation for '" + m.group( 1 ) + "'!" );

        return res;
    }
}
