package com.bitgirder.lang;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.util.Random;
import java.util.Map;
import java.util.NavigableMap;
import java.util.TreeMap;

// Currently this class is built around the notion of float frequencies given as
// fractions of some arbitrary range and adjusted internally to span the range
// [0,1). Also, this class is built around java.util.Random. None of these
// things are necessarily set in stone and we can evolve the implementation as
// needed.
public
final
class Randomizer< V >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final NavigableMap< Float, V > freqsAdj;
    private final Random rand;

    private
    Randomizer( NavigableMap< Float, V > freqsAdj,
                Random rand )
    {
        this.freqsAdj = freqsAdj;
        this.rand = rand;
    }

    // This is currently package-level only and used for testing; it's possible
    // that we could open it at some point, but not clear now that it makes
    // sense and is worth being tied to in a public API
    V
    forFloat( float f )
    {
        Map.Entry< Float, V > e = state.notNull( freqsAdj.floorEntry( f ) );
        return e.getValue();
    }

    public V next() { return forFloat( rand.nextFloat() ); }

    @Override
    public
    String
    toString()
    {
        return Strings.inspect( this, true,
            "rand", rand,
            "freqsAdj", freqsAdj
        ).toString();
    }

    // Things to know:
    //
    //  - all map keys/vals must be non-null and all vals must have floatValue()
    //  > 0. 
    //
    //  - freqs may not be empty
    //
    //  - the randomizer will assign frequency slots using the iteration order
    //  of freqs.entrySet().iterator(); using LinkedHashMap or SortedMap can
    //  allow callers explicit control over the position within the spectrum
    //  [0,1)
    public
    static
    < U >
    Randomizer< U >
    create( Map< U, ? extends Number > freqs,
            Random rand )
    {
        inputs.notNull( freqs, "freqs" );
        inputs.isFalse( freqs.isEmpty(), "Frequency map is empty" );
        inputs.notNull( rand, "rand" );

        float span = 0f;
        for ( Number n : freqs.values() )
        {
            inputs.isFalse( n == null, "Frequency map contains a null value" );
            span += n.floatValue();
        }

        NavigableMap< Float, U > freqsAdj = new TreeMap< Float, U >();

        float cur = 0f;
        for ( Map.Entry< U, ? extends Number > e : freqs.entrySet() )
        {
            freqsAdj.put( cur / span, e.getKey() );
            cur += e.getValue().floatValue();
        }

        return new Randomizer< U >( freqsAdj, rand );
    }
}
