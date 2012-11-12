package com.bitgirder.json;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.TypedString;
import com.bitgirder.lang.Lang;

import java.io.IOException;

public
final
class JsonString
extends TypedString< JsonString >
implements JsonValue
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private JsonString( CharSequence cs ) { super( cs ); }

    public
    static
    CharSequence
    getExternalForm( CharSequence cs )
        throws IOException
    {
        return Lang.getRfc4627String( inputs.notNull( cs, "cs" ) );
    }

    // Could make this part of the public API at any point
    CharSequence 
    getExternalForm() 
        throws IOException
    { 
        return getExternalForm( this ); 
    }
 
    public
    static
    JsonString
    create( CharSequence str )
    {
        return new JsonString( inputs.notNull( str, "str" ) );
    }
}
