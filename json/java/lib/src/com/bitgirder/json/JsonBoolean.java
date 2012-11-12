package com.bitgirder.json;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class JsonBoolean
implements JsonValue
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    public final static JsonBoolean TRUE = new JsonBoolean( true );
    public final static JsonBoolean FALSE = new JsonBoolean( false );

    private final boolean b;

    private JsonBoolean( boolean b ) { this.b = b; }

    public boolean booleanValue() { return b; }

    public static JsonBoolean valueOf( boolean b ) { return b ? TRUE : FALSE; }

    @Override public String toString() { return Boolean.toString( b ); }
}
