package com.bitgirder.demo.mingle;

import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;

final
class DemoUtils
{
    private final static State state = new State();

    private DemoUtils() {}

    static
    String
    copyAndReverse( String str,
                    int copies,
                    boolean reverse )
    {
        state.notNull( str, "str" );
        state.positiveI( copies, "copies" );

        int strLen = str.length();
        int resLen = strLen * copies;

        char[] arr = new char[ resLen ];

        for ( int i = 0; i < resLen; ++i )
        {
            int strPos = i % strLen; 
            if ( reverse ) strPos = strLen - strPos - 1;

            arr[ i ] = str.charAt( strPos );
        }

        return new String( arr );
    }

    static
    List< Long >
    getFibRes( int fib1,
               int fib2,
               int seqLen )
    {
        List< Long > res = Lang.newList( seqLen );

        res.add( (long) fib1 );
        res.add( (long) fib2 );

        long prev = fib1;
        long next = fib2;
        
        while ( res.size() < seqLen )
        {
            long tmp = prev;
            prev = next;
            next += tmp;

            res.add( next );
        }

        return res;
    }
}
