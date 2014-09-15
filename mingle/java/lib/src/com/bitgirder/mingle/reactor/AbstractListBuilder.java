package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.mingle.MingleIdentifier;

public
abstract
class AbstractListBuilder
implements BuildReactor.ListBuilder
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    protected
    Object
    produceValue()
        throws Exception
    {
        throw Instance.unimplemented( this, "produceValue" );
    }

    public
    Object
    produceValue( ObjectPath< MingleIdentifier > path )
        throws Exception
    {
        return produceValue();
    }

    protected
    void
    addValue( Object val )
        throws Exception
    {
        throw Instance.unimplemented( this, "addValue" );
    }

    public
    void
    addValue( Object val,
              ObjectPath< MingleIdentifier > path )
        throws Exception
    {
        addValue( val );
    }
}
