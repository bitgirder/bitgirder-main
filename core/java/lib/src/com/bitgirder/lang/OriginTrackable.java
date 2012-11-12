package com.bitgirder.lang;

import java.util.List;

public
interface OriginTrackable
{
    /** 
     * Returns the stack trace from the origin of the implementing Object's
     * call. Implementations may return <code>null</code> to indicate that no
     * origin trace is available, and this behavior may change from call to call
     * or program run to program run. If not null this method will return a List
     * that represents the origin's best description of the origin of this
     * object.
     */
    public
    List< StackTraceElement >
    getOriginStackTrace();
}
