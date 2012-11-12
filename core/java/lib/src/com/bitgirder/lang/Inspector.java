package com.bitgirder.lang;

public
interface Inspector
{
    // Interfaces should return self so callers can chain
    public
    Inspector
    add( CharSequence fieldName,
         Object fieldValue );
}
