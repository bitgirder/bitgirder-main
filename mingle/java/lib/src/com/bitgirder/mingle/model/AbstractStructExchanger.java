package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;

public
abstract
class AbstractStructExchanger< S >
extends AbstractValueExchanger< S >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    protected
    AbstractStructExchanger( AtomicTypeReference typ,
                             Class< S > exCls )
    {
        super( typ, exCls );
    }

    protected
    abstract
    S
    buildStruct( MingleSymbolMapAccessor acc );

    public
    final
    Object
    asJavaValue( MingleValue mv,
                 ObjectPath< MingleIdentifier > path )
    {
        MingleSymbolMap m = expectSymbolMap( getMingleType(), mv, path );

        return buildStruct( MingleSymbolMapAccessor.create( m, path ) );
    }

    protected
    final
    MingleStructBuilder
    structBuilder()
    {
        return MingleModels.structBuilder().setType( getMingleType() );
    }
}
