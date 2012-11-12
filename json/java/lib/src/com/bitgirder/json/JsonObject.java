package com.bitgirder.json;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;
import java.util.Map;
import java.util.Set;

public
final
class JsonObject
implements JsonText,
           JsonValue
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    // this map itself, as well as the lists contained as values, are
    // unmodifiable and may be returned from public methods
    private final Map< JsonString, List< JsonValue > > members;

    private
    JsonObject( Builder b )
    {
        this.members = buildImmutableMembersMap( b.members );
    }

    public int size() { return members.size(); }

    private
    static
    JsonString
    makeName( CharSequence name,
              String errName )
    {
        inputs.notNull( name, errName );

        return 
            name instanceof JsonString
                ? (JsonString) name : JsonString.create( name.toString() );
    }

    public
    Set< Map.Entry< JsonString, List< JsonValue > > >
    entrySet()
    {
        return members.entrySet();
    }

    public
    boolean 
    hasMember( JsonString name )
    {
        inputs.notNull( name, "name" );
        return members.containsKey( name );
    }

    public
    boolean
    hasMember( CharSequence name )
    {
        return hasMember( makeName( name, "name" ) );
    }

    public
    List< JsonValue >
    getValues( CharSequence name )
    {
        return members.get( makeName( name, "name" ) );
    }

    public
    List< JsonValue >
    getValues( JsonString name )
    {
        return members.get( inputs.notNull( name, "name" ) );
    }

    private
    JsonValue
    doGet( JsonString name )
    {
        List< JsonValue > vals = getValues( name );

        if ( vals == null ) return null;
        else
        {
            inputs.isTrue( 
                vals.size() == 1, 
                "Object has multiple values for name:", name );

            return vals.get( 0 );
        }
    }

    public
    JsonValue
    getValue( JsonString name )
    {
        return doGet( inputs.notNull( name, "name" ) );
    }

    public
    JsonValue
    getValue( CharSequence name )
    {
        return doGet( makeName( name, "name" ) );
    }

    public
    JsonNumber
    getNumber( CharSequence name )
    {
        return (JsonNumber) getValue( name );
    }

    public
    boolean
    getBool( CharSequence name )
    {
        JsonBoolean b = (JsonBoolean) getValue( name );
        return b != null && b.booleanValue();
    }

    public
    JsonNull
    getNull( CharSequence name )
    {
        return (JsonNull) getValue( name );
    }

    public
    JsonString
    getString( CharSequence name )
    {
        return (JsonString) getValue( name );
    }

    public
    JsonArray
    getArray( CharSequence name )
    {
        return (JsonArray) getValue( name );
    }

    public
    JsonObject
    getObject( CharSequence name )
    {
        return (JsonObject) getValue( name );
    }

    public
    final
    static
    class Builder
    {
        private final List< Member > members = Lang.newList();

        public
        Builder
        addMember( JsonString key,
                   JsonValue val )
        {
            inputs.notNull( key, "key" );
            inputs.notNull( val, "val" );

            members.add( new Member( key, val ) );

            return this;
        }

        public
        Builder
        addMember( CharSequence key,
                   Object val )
        {
            return 
                addMember( 
                    JsonString.create( inputs.notNull( key, "key" ) ),
                    JsonValues.asJsonValue( val )
                );
        }

        public JsonObject build() { return new JsonObject( this ); }
    }

    private
    final
    static
    class Member
    {
        private final JsonString key;
        private final JsonValue value;

        private
        Member( JsonString key,
                JsonValue value )
        {
            this.key = key;
            this.value = value;
        }
    }

    private
    static
    Map< JsonString, List< JsonValue > >
    buildImmutableMembersMap( List< Member > members )
    {
        Map< JsonString, List< JsonValue > > res = Lang.newMap();

        for ( Member m : members ) Lang.putAppend( res, m.key, m.value );

        // reset lists as immutable
        for ( Map.Entry< JsonString, List< JsonValue > > e : res.entrySet() )
        {
            e.setValue( Lang.unmodifiableList( e.getValue() ) );
        }

        return Lang.unmodifiableMap( res );
    }
}
