package com.bitgirder.test;

public
interface TestFailureExpector
{
    public
    Class< ? extends Throwable >
    expectedFailureClass();

    public
    CharSequence
    expectedFailurePattern();
}
