package com.bitgirder.mingle.bind;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.AtomicTypeReference;

public
interface MingleBinding
{
    public
    Object
    asJavaValue( AtomicTypeReference typ,
                 MingleValue mv,
                 MingleBinder mb,
                 ObjectPath< MingleIdentifier > path );
 
    public
    MingleValue
    asMingleValue( Object obj,
                   MingleBinder mb,
                   ObjectPath< String > path );
}
