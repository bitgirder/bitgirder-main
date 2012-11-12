package com.bitgirder.validation;

public
class Inputs
extends Validator
{
    private final static Inputs INSTANCE = new Inputs();

    public static Inputs getInstance() { return INSTANCE; }

    public
    IllegalArgumentException
    createException( CharSequence inputName,
                     CharSequence msg )
    {
        return new IllegalArgumentException( 
            getDefaultMessage( inputName, msg ) );
    }

    private
    boolean
    parseBoolean( CharSequence cs,
                  String location )
    {
        notNull( cs, "cs" );

        boolean res = false;

        String s = cs.toString().trim().toLowerCase();

        if ( s.equals( "true" ) || s.equals( "yes" ) ) res = true;
        else if ( s.equals( "false" ) || s.equals( "no" ) ) res = false;
        else fail( location, cs );

        return res;
    }

    public
    boolean
    parseBoolean( CharSequence cs )
    {
        return parseBoolean( cs, "Invalid boolean value:" );
    }

    public
    boolean
    hasBooleanSystemProperty( String propName )
    {
        notNull( propName, "propName" );

        String propVal = System.getProperty( propName );

        if ( propVal == null ) return false;
        else
        {
            String location = 
                "Invalid property value for property " + propName + ":";

            return parseBoolean( propVal, location );
        }
    }
}
