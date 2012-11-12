package com.bitgirder.lang.path;

public
abstract
class ListPath< E >
extends ObjectPath< E >
{
    ListPath( ObjectPath< E > parent,
              String paramName )
    {
        super( parent, paramName );
    }

    abstract int getIndex();
}
