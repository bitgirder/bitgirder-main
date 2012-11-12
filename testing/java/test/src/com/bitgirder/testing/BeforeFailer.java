package com.bitgirder.testing;

import com.bitgirder.test.Before;
import com.bitgirder.test.After;
import com.bitgirder.test.Test;

@Test
final
class BeforeFailer
{
    // By default this class won't fail and will run successfully as part of a
    // general test run. Certain tests may set this value to true on specific
    // instances to exercise failure handling.
    boolean doFail;

    final static class MarkerException extends RuntimeException {}

    @Before 
    private 
    void 
    before1() 
    { 
        if ( doFail ) throw new MarkerException(); 
    }

    @Test private void test1() {}
    @After private void after1() {}
}
