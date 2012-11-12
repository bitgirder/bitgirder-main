package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;

public
final
class MingleTypeCastException
extends MingleValidationException
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleTypeReference expct;
    private final MingleTypeReference actual;
    
    MingleTypeCastException( MingleTypeReference expct,
                             MingleTypeReference actual,
                             ObjectPath< MingleIdentifier > location,
                             boolean wasSerialized )
    {
        super( 
            makeMessage( 
                inputs.notNull( expct, "expct" ),
                inputs.notNull( actual, "actual" ) ),
            inputs.notNull( location, "location" ),
            wasSerialized,
            null
        );
        
        this.expct = expct;
        this.actual = actual;
    }

    public
    MingleTypeCastException( MingleTypeReference expct,
                             MingleTypeReference actual,
                             ObjectPath< MingleIdentifier > location )
    {
        this( expct, actual, location, false );
    }

    public final MingleTypeReference getExpectedType() { return expct; }
    public final MingleTypeReference getActualType() { return actual; }

    private
    static
    String
    makeMessage( MingleTypeReference expct,
                 MingleTypeReference actual )
    {
        return 
            new StringBuilder().
                append( "Expected mingle value of type " ).
                append( expct.getExternalForm() ).
                append( " but found " ).
                append( actual.getExternalForm() ).
                toString();
    }
}
