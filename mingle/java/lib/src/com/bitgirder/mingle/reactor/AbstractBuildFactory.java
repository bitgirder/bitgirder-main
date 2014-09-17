package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.mingle.MingleIdentifier;
import com.bitgirder.mingle.MingleValue;
import com.bitgirder.mingle.QualifiedTypeName;
import com.bitgirder.mingle.ListTypeReference;

public
abstract
class AbstractBuildFactory
implements BuildReactor.Factory
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    protected
    AbstractBuildFactory()
    {}

    protected
    Object
    buildValue()
        throws Exception
    {
        throw Instance.unimplemented( this, "buildValue" );
    }

    protected
    Object
    buildValue( MingleValue mv )
        throws Exception
    {
        return buildValue();
    }

    public
    Object
    buildValue( MingleValue mv,
                ObjectPath< MingleIdentifier > path )
        throws Exception
    {
        return buildValue( mv );
    }

    protected
    BuildReactor.FieldSetBuilder
    startMap()
        throws Exception
    {
        throw Instance.unimplemented( this, "startMap" );
    }

    public
    BuildReactor.FieldSetBuilder
    startMap( ObjectPath< MingleIdentifier > path )
        throws Exception
    {
        return startMap();
    }

    protected
    BuildReactor.FieldSetBuilder
    startStruct()
        throws Exception
    {
        throw Instance.unimplemented( this, "startStruct" );
    }

    protected
    BuildReactor.FieldSetBuilder
    startStruct( QualifiedTypeName typ )
        throws Exception
    {
        return startStruct();
    }

    public
    BuildReactor.FieldSetBuilder
    startStruct( QualifiedTypeName typ,
                 ObjectPath< MingleIdentifier > path )
        throws Exception
    {
        return startStruct( typ );
    }

    protected
    BuildReactor.ListBuilder
    startList()
        throws Exception
    {
        throw Instance.unimplemented( this, "startList" );
    }

    protected
    BuildReactor.ListBuilder
    startList( ListTypeReference lt )
        throws Exception
    {
        return startList();
    }

    public
    BuildReactor.ListBuilder
    startList( ListTypeReference lt,
               ObjectPath< MingleIdentifier > path )
        throws Exception
    {
        return startList( lt );
    }
}
