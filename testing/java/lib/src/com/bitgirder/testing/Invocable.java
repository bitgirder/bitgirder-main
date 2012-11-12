package com.bitgirder.testing;

import com.bitgirder.test.TestFailureExpectation;

public
interface Invocable
{
    public
    void
    invoke( Object invTarg )
        throws Exception;

    public
    CharSequence
    getName();

    public
    TestFailureExpectation
    getFailureExpectation();
}
