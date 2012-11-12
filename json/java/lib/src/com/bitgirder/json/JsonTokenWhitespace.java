package com.bitgirder.json;

// At some point we may find we need to actually retain the whitespace, and at
// that point we could change this token; for now it's just ignored anyway so
// there is no need to create new objects on each encounter
public
final
class JsonTokenWhitespace
implements JsonToken
{
    final static JsonTokenWhitespace INSTANCE = new JsonTokenWhitespace();

    JsonTokenWhitespace() {}
}
