package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;

public
abstract
class AbstractExceptionExchanger< E >
extends AbstractValueExchanger< E >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public final static MingleIdentifier ID_MESSAGE =
        MingleIdentifier.create( "message" );

    protected
    AbstractExceptionExchanger( AtomicTypeReference typ,
                                Class< E > exCls )
    {
        super( typ, exCls );
    }

    protected
    final
    String
    expectMessage( MingleSymbolMapAccessor acc )
    {
        return acc.expectString( ID_MESSAGE );
    }

    protected
    final
    String
    getMessage( MingleSymbolMapAccessor acc )
    {
        return acc.getString( ID_MESSAGE );
    }

    protected
    abstract
    E
    buildException( MingleSymbolMapAccessor acc );

    public
    final
    Object
    asJavaValue( MingleValue mv,
                 ObjectPath< MingleIdentifier > path )
    {
        MingleSymbolMap m = expectSymbolMap( getMingleType(), mv, path );

        return buildException( MingleSymbolMapAccessor.create( m, path ) );
    }

    protected
    final
    MingleExceptionBuilder
    exceptionBuilder()
    {
        return MingleModels.exceptionBuilder().setType( getMingleType() );
    }
}
