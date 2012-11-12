package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.mingle.parser.MingleParsers;

public
final
class MingleInvocationValidator
extends DefaultMingleValidator
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    MingleInvocationValidator( ObjectPath< MingleIdentifier > path )
    {
        super( path );
    }

    private
    ObjectPath< MingleIdentifier >
    descend( MingleIdentifier fld,
             String paramName )
    {
        return getPath().descend( inputs.notNull( fld, paramName ) );
    }

    private
    ObjectPath< MingleIdentifier >
    descend( CharSequence fld,
             String paramName )
    {
        MingleIdentifier id =
            MingleParsers.createIdentifier( inputs.notNull( fld, paramName ) );
        
        return getPath().descend( id );
    }

    private
    MingleValidator
    getFieldValidator( MingleIdentifier fld )
    {
        return MingleModels.createValidator( getPath().descend( fld ) );
    }

    public
    MingleValidator
    field( MingleIdentifier fld )
    {
        return MingleModels.createValidator( descend( fld, "fld" ) );
    }

    public
    MingleValidator
    field( CharSequence fld )
    {
        return MingleModels.createValidator( descend( fld, "fld" ) );
    }

    public
    < V >
    V
    expect( MingleIdentifier fld,
            V val )
    {
        inputs.notNull( fld, "fld" );

        if ( val == null ) 
        {
            throw new MingleValidationException( 
                "value is null", getPath().descend( fld ) );
        }
        else return val;
    }

    public
    < V >
    V
    expect( CharSequence fld,
            V val )
    {
        return 
            expect( 
                MingleIdentifier.create( inputs.notNull( fld, "fld" ) ), val );
    }
}
