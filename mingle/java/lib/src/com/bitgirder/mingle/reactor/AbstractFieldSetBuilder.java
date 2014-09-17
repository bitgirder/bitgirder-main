package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.mingle.MingleIdentifier;

public
abstract
class AbstractFieldSetBuilder
implements BuildReactor.FieldSetBuilder
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
    BuildReactor.Factory
    startField()
        throws Exception
    {
        throw Instance.unimplemented( this, "startField" );
    }

    protected
    BuildReactor.Factory
    startField( MingleIdentifier fld )
        throws Exception
    {
        return startField();
    }

    public
    BuildReactor.Factory
    startField( MingleIdentifier fld,
                ObjectPath< MingleIdentifier > path )
        throws Exception
    {
        return startField( fld );
    }

    protected
    void
    setValue( MingleIdentifier fld,
              Object val )
        throws Exception
    {
        throw Instance.unimplemented( this, "setValue" );
    }

    public
    void
    setValue( MingleIdentifier fld,
              Object val,
              ObjectPath< MingleIdentifier > path )
        throws Exception
    {
        setValue( fld, val );
    }
}
