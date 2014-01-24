package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;

public
final
class MingleStructAccessor
extends MingleSymbolMapAccessor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleStruct ms;

    private
    MingleStructAccessor( MingleStruct ms,
                          ObjectPath< MingleIdentifier > path )
    {
        super( ms.getFields(), path );
        this.ms = ms;
    }

    public MingleStruct getStruct() { return ms; }

    public QualifiedTypeName getType() { return ms.getType(); }

    public
    static
    MingleStructAccessor
    forStruct( MingleStruct ms,
               ObjectPath< MingleIdentifier > path )
    {
        return new MingleStructAccessor( inputs.notNull( ms, "ms" ), path );
    }

    public
    static
    MingleStructAccessor
    forStruct( MingleStruct ms )
    {
        return forStruct( ms, ObjectPath.< MingleIdentifier >getRoot() );
    }
}
