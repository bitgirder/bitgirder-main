package com.bitgirder.json;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class JsonTokenString
implements JsonToken
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final char[] str;

    JsonTokenString( char[] str ) { this.str = str; }

    public String toJavaString() { return new String( str ); }
}
