package com.bitgirder.lang;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.test.Test;

// Offers public methods to help with computing test sums, but also has some
// internal tests of those methods themselves
@Test
public
final
class TestSums
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();
    
    private TestSums() {}

    // returns sum of sequence starting at start (incl) and ending at end
    // (exclusive); start must be <= end; either or both may be <= 0
    public
    static
    int
    ofSequence( int start,
                int end )
    {
        inputs.isFalse( end < start, "end < start:", end, "<", start );

        if ( end == start ) return start;
        else
        {
            long len = end - start;
    
            // compute the intermediate sum, that of the numbers 1 .. len,
            // including len. In case our shift causes temporary overflow, we
            // compute the intermediate result as a long
            long intermed = len * ( len + 1L ) / 2L;
    
            return (int) ( intermed + ( len * ( start - 1 ) ) );
        }
    }

    @Test
    private
    void
    testSums()
    {
        state.equalInt( 45, ofSequence( 0, 10 ) );
        state.equalInt( 45, ofSequence( 1, 10 ) );
        state.equalInt( 44, ofSequence( 2, 10 ) );
        state.equalInt( 44, ofSequence( -1, 10 ) );

        state.equalInt( 5, ofSequence( 5, 5 ) );
    }
}
