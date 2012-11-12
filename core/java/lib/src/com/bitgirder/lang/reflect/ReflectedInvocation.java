package com.bitgirder.lang.reflect;

import java.util.Map;

public
interface ReflectedInvocation
{
    public
    Object
    invoke( Map< Object, Object > params )
        throws Exception;
    
    public
    Object
    getTarget();

    public
    Iterable< Object >
    getKeys();

    public
    boolean
    hasKey( Object key );
}
