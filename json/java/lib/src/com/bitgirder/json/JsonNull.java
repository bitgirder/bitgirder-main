package com.bitgirder.json;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class JsonNull
implements JsonValue
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    public final static JsonNull INSTANCE = new JsonNull();

    private JsonNull() {}
}
