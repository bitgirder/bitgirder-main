package com.bitgirder.mingle.model;

import com.bitgirder.lang.path.ObjectPath;

public
interface MingleValidator
{
    public
    void
    isFalse( boolean val,
             Object... msg );
    
    public
    void
    isTrue( boolean val,
            Object... msg );
    
    public
    ObjectPath< MingleIdentifier >
    getPath();
}
