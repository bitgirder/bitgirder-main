package com.bitgirder.mingle.reactor;

import com.bitgirder.mingle.MingleIdentifier;

import com.bitgirder.lang.path.ObjectPath;

public
interface ExceptionFactory
{
    public
    Exception
    createException( ObjectPath< MingleIdentifier > path,
                     String msg );
}        
