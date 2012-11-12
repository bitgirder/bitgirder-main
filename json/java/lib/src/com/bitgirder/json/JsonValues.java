package com.bitgirder.json;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ImmutableListPath;
import com.bitgirder.lang.path.ObjectPaths;

import java.util.Map;
import java.util.List;

public
final
class JsonValues
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private JsonValues() {}

    private
    static
    < T >
    ObjectPath< T >
    descend( ObjectPath< T > path,
             T next )
    {
        return path == null ? null : path.descend( next );
    }

    private
    static
    < T >
    ImmutableListPath< T >
    startList( ObjectPath< T > path )
    {
        return path == null ? null : path.startImmutableList();
    }

    private
    static
    < T >
    ImmutableListPath< T >
    next( ImmutableListPath< T > path )
    {
        return path == null ? null : path.next();
    }

    private
    static
    RuntimeException
    createFail( ObjectPath< ? > path,
                Object... msg )
    {
        CharSequence tail = Strings.join( " ", msg );

        String msgStr;

        if ( path == null ) msgStr = tail.toString();
        else
        {
            StringBuilder sb = new StringBuilder();
            ObjectPaths.appendFormat( path, ObjectPaths.DOT_FORMATTER, sb );

            msgStr = sb.append( ' ' ).append( tail ).toString();
        }

        return new RuntimeException( msgStr );
    }

    private
    static
    String
    getKeyString( Map.Entry< ?, ? > e,
                  boolean checkedKeyType,
                  ObjectPath< String > path )
    {
        Object res = e.getKey();

        if ( checkedKeyType || res instanceof String ) return (String) res;
        else
        {
            throw 
                createFail( 
                    path, "map contains an element of type",
                    res.getClass().getName() + ":", res );
        }
    }

    private
    static
    JsonObject
    asJsonObject( Map< ?, ? > m,
                  boolean checkedKeyType,
                  ObjectPath< String > path )
    {
        if ( m.containsKey( null ) )
        {
            throw createFail( path, "Input contains the null key" );
        }

        JsonObject.Builder b = new JsonObject.Builder();

        for ( Map.Entry< ?, ? > e : m.entrySet() )
        {
            String keyStr = getKeyString( e, checkedKeyType, path );

            JsonString key = JsonString.create( keyStr );

            JsonValue val = 
                asJsonValue( e.getValue(), descend( path, keyStr ) );

            b.addMember( key, val );
        }

        return b.build();
    }

    private
    static
    JsonArray
    asJsonArray( Iterable< ? > coll,
                 ObjectPath< String > path )
    {
        ImmutableListPath< String > lstPath = startList( path );

        JsonArray.Builder b = new JsonArray.Builder();

        for ( Object obj : coll )
        {
            JsonValue val = asJsonValue( obj, lstPath );
            b.add( val );

            lstPath = next( lstPath );
        }

        return b.build();
    }

    private
    static
    JsonValue
    asJsonValue( Object obj,
                 ObjectPath< String > path )
    {
        if ( obj == null ) return JsonNull.INSTANCE;
        else if ( obj instanceof CharSequence ) 
        {
            return JsonString.create( (CharSequence) obj );
        }
        else if ( obj instanceof Number ) 
        {
            return JsonNumber.forNumber( (Number) obj );
        }
        else if ( obj instanceof Boolean )
        {
            return JsonBoolean.valueOf( ( (Boolean) obj ).booleanValue() );
        }
        else if ( obj instanceof Iterable )
        {
            return asJsonArray( (Iterable< ? >) obj, path );
        }
        else if ( obj instanceof Map )
        {
            return asJsonObject( (Map< ?, ? >) obj, false, path );
        }
        else 
        {
            throw createFail( 
                path, "Unconvertible java type:", obj.getClass().getName() );
        }
    }

    public
    static
    JsonValue
    asJsonValue( Object obj )
    {
        return asJsonValue( obj, null );
    }

    public
    static
    JsonObject
    asJsonObject( Map< String, Object > m )
    {
        inputs.notNull( m, "m" );

        return asJsonObject( m, true, ObjectPath.< String >getRoot() );
    }

    private
    static
    JsonValue
    expectSingleValue( List< JsonValue > vals,
                       ObjectPath< JsonString > path )
    {
        state.isFalse( vals.isEmpty() );

        if ( vals.size() == 1 ) return vals.get( 0 );
        else throw createFail( path, "multiple values for entry" );
    }

    private
    static
    Object
    asJavaList( JsonArray arr,
                ImmutableListPath< JsonString > path )
    {
        List< Object > res = Lang.newList( arr.size() );

        for ( JsonValue jv : arr )
        {
            res.add( asJavaValue( jv, path ) );
            path = next( path );
        }

        return Lang.unmodifiableList( res );
    }

    private
    static
    Object
    asJavaValue( JsonValue val,
                 ObjectPath< JsonString > path )
    {
        if ( val instanceof JsonNull ) return null;
        else if ( val instanceof JsonString ) return val.toString();
        else if ( val instanceof JsonNumber )
        {
            return ( (JsonNumber) val ).getNumber();
        }
        else if ( val instanceof JsonBoolean )
        {
            return Boolean.valueOf( ( (JsonBoolean) val ).booleanValue() );
        }
        else if ( val instanceof JsonArray )
        {
            return asJavaList( (JsonArray) val, startList( path ) );
        }
        else if ( val instanceof JsonObject )
        {
            return asJavaMap( (JsonObject) val, path );
        }
        else 
        {
            throw
                createFail( 
                    path, 
                    "Unconvertible json type:", val.getClass().getName() );
        }
    }

    private
    static
    Map< String, Object >
    asJavaMap( JsonObject json,
               ObjectPath< JsonString > path )
    {
        Map< String, Object > res = Lang.newMap( json.size() );

        for ( Map.Entry< JsonString, List< JsonValue > > e : json.entrySet() )
        {
            JsonString key = e.getKey();
            ObjectPath< JsonString > keyPath = descend( path, key );

            JsonValue val = expectSingleValue( e.getValue(), path );
            Object javaVal = asJavaValue( val, keyPath );

            if ( javaVal != null ) res.put( key.toString(), javaVal );
        }

        return Lang.unmodifiableMap( res );
    }

    // current default is to fail if a duplicate key is encountered when
    // traversing json object, but we could add an overload that takes some
    // control option (such as merge, keep-newest, keep-oldest, fail) instead
    // and have this one call with the fail opt.
    public
    static
    Map< String, Object >
    asJavaMap( JsonObject json )
    {
        inputs.notNull( json, "json" );
        return asJavaMap( json, ObjectPath.< JsonString >getRoot() );
    }
}
