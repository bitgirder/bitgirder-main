package com.bitgirder.lang;

import java.util.List;

public
interface OriginTraceSettable
extends OriginTrackable
{
    /**
     * Implementing classes should accept but ignore all but the first trace set
     * on an object, based on the heuristic that the first one set will be the
     * one most useful in debugging (and that it would be very rare to attempt
     * to set more than one anyway on the same object). Callers should not
     * accept null as a valid input, but should allow the empty list.
     */
    public
    void
    setOriginStackTrace( List< StackTraceElement > trace );
}
