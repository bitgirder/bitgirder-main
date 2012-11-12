package com.bitgirder.lang;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.Validator;
import com.bitgirder.validation.State;

public
final
class Range< C extends Comparable< C > >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final C min;
    private final boolean includesMin;

    private final C max;
    private final boolean includesMax;

    // validity checking happens in here, so factory methods can skip it so long
    // as params are named min/max too
    private
    Range( C min,
           boolean includesMin,
           C max,
           boolean includesMax )
    {
        this.includesMin = includesMin;
        this.min = includesMin ? inputs.notNull( min, "min" ) : min;

        this.includesMax = includesMax;
        this.max = includesMax ? inputs.notNull( max, "max" ) : max;

        inputs.isFalse( 
            min != null && max != null && max.compareTo( min ) < 0, 
            "max < min (", max, "<", min, ")" 
        );
    }

    public C min() { return min; }
    public boolean includesMin() { return includesMin; }

    public C max() { return max; }
    public boolean includesMax() { return includesMax; }

    public
    boolean
    includes( C val )
    {
        inputs.notNull( val, "val" );

        boolean res;

        int minComp = min == null ? 1 : val.compareTo( min );
 
        if ( minComp == 0 ) res = includesMin;
        else if ( minComp < 0 ) res = false;
        else
        {
            int maxComp = max == null ? -1 : val.compareTo( max );
            res = ( maxComp == 0 && includesMax ) || maxComp < 0;
        }

        return res;
    }

    @Override
    public
    String
    toString()
    {
        return new StringBuilder( 12 ).
            append( includesMin ? "[" : "(" ).
            append( min == null ? "" : min ).
            append( ", " ).
            append( max == null ? "" : max ).
            append( includesMax ? "]" : ")" ).
            toString();
    }

    public
    static
    < V extends Comparable< V > >
    Range< V >
    of( V val )
    {
        inputs.notNull( val, "val" );
        return new Range< V >( val, true, val, true );
    }

    public
    static
    < V extends Comparable< V > >
    Range< V >
    closed( V min,
            V max )
    {
        return new Range< V >( min, true, max, true );
    }

    public
    static
    < V extends Comparable< V > >
    Range< V >
    open( V min,
          V max )
    {
        return new Range< V >( min, false, max, false );
    }

    public
    static
    < V extends Comparable< V > >
    Range< V >
    openMin( V min,
             V max )
    {
        return new Range< V >( min, false, max, true );
    }

    public
    static
    < V extends Comparable< V > >
    Range< V >
    openMax( V min,
             V max )
    {
        return new Range< V >( min, true, max, false );
    }

    public
    static
    < V extends Comparable< V > >
    Range< V >
    open()
    {
        return new Range< V >( null, false, null, false );
    }
}
