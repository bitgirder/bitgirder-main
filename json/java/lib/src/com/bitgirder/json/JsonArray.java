package com.bitgirder.json;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;
import java.util.Iterator;
import java.util.ArrayList;

public
final
class JsonArray
implements JsonText,
           JsonValue,
           Iterable< JsonValue >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final ArrayList< JsonValue > arr;

    private
    JsonArray( Builder b )
    {
        this.arr = new ArrayList< JsonValue >( b.arr );
    }

    public int size() { return arr.size(); }

    public Iterator< JsonValue > iterator() { return arr.iterator(); }

    public JsonValue get( int i ) { return arr.get( i ); }

    public JsonObject getObject( int i ) { return (JsonObject) get( i ); }
    public JsonArray getArray( int i ) { return (JsonArray) get( i ); }
    public JsonNull getNull( int i ) { return (JsonNull) get( i ); }

    public 
    boolean 
    getBool( int i ) 
    { 
        return ( (JsonBoolean) get( i ) ).booleanValue();
    }

    public JsonString getString( int i ) { return (JsonString) get( i ); }
    public JsonNumber getNumber( int i ) { return (JsonNumber) get( i ); }

    public
    final
    static
    class Builder
    {
        private final List< JsonValue > arr = Lang.newList();

        public
        Builder
        add( JsonValue val )
        {
            inputs.notNull( val, "val" );
            arr.add( val );

            return this;
        }

        public JsonArray build() { return new JsonArray( this ); }
    }
}
