package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ObjectPaths;

public
class MingleValidationException
extends RuntimeException
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final String desc; // maybe null
    private final ObjectPath< MingleIdentifier > location; // not null
    private final boolean wasSerialized;

    MingleValidationException( String desc,
                               ObjectPath< MingleIdentifier > location,
                               boolean wasSerialized,
                               Throwable cause )
    {
        super(
            makeMessage(
                desc,
                inputs.notNull( location, "location" )
            ),
            cause
        );

        this.location = location;
        this.desc = desc;
        this.wasSerialized = wasSerialized;
    }

    public
    MingleValidationException( String desc,
                               ObjectPath< MingleIdentifier > location,
                               Throwable cause )
    {
        this( desc, location, false, cause );
    }

    public
    MingleValidationException( String desc,
                               ObjectPath< MingleIdentifier > location )
    {
        this( desc, location, null );
    }

    final boolean wasSerialized() { return wasSerialized; }

    public 
    final 
    ObjectPath< MingleIdentifier > 
    getLocation() 
    { 
        return location; 
    }

    public final String getDescription() { return desc; }

    private
    static
    String
    makeMessage( String desc,
                 ObjectPath< MingleIdentifier > location )
    {
        StringBuilder res = new StringBuilder();
 
        boolean isRoot = location.getParent() == null;

        if ( ! isRoot )
        {
            ObjectPaths.appendFormat( 
                location,
                MingleModels.getIdentifierPathFormatter(),
                res
            );
        }

        if ( desc != null ) 
        {
            if ( ! isRoot ) res.append( ": " );
            res.append( desc );
        }

        return res.toString();
    }
}
