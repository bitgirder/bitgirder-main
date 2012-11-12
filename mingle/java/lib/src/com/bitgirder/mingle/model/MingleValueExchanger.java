package com.bitgirder.mingle.model;

import com.bitgirder.lang.path.ObjectPath;

public
interface MingleValueExchanger
{
    public
    Class< ? >
    getJavaClass();

    public
    MingleTypeReference
    getMingleType();

    public
    Object
    asJavaValue( MingleValue mv,
                 ObjectPath< MingleIdentifier > path );
    
    public
    MingleValue
    asMingleValue( Object obj,
                   ObjectPath< String > path );
}
