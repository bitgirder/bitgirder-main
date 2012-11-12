package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Range;
import com.bitgirder.lang.Lang;

final
class OctetRangeMatcher
implements TerminalMatcher< Byte >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final static Range< Integer > OCTET_RANGE = Range.closed( 0, 255 );

    private final int minIncl;
    private final int maxIncl;

    private
    OctetRangeMatcher( int minIncl,
                       int maxIncl )
    {
        this.minIncl = minIncl;
        this.maxIncl = maxIncl;
    }

    public
    boolean
    isMatch( Byte byteObj )
    {
        byte b = byteObj.byteValue();
        int octet = Lang.asOctet( b );

        return octet >= minIncl && octet <= maxIncl;
    }

    static
    OctetRangeMatcher
    forOctet( int octet )
    {
        inputs.inRange( octet, "octet", OCTET_RANGE );
        return new OctetRangeMatcher( octet, octet );
    }

    static
    OctetRangeMatcher
    forRange( int minIncl,
              int maxIncl )
    {
        inputs.inRange( minIncl, "minIncl", OCTET_RANGE );
        inputs.inRange( maxIncl, "maxIncl", OCTET_RANGE );

        inputs.isTrue(
            minIncl <= maxIncl,
            "minIncl > maxIncl:", minIncl, ">", maxIncl );
 
        return new OctetRangeMatcher( minIncl, maxIncl );
    }
}
