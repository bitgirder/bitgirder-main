package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.Mingle;
import com.bitgirder.mingle.MingleIdentifier;
import com.bitgirder.mingle.MingleValue;
import com.bitgirder.mingle.QualifiedTypeName;
import com.bitgirder.mingle.ListTypeReference;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.lang.path.ObjectPath;

import java.util.List;

public
final
class MingleReactorEvent
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public
    static
    enum Type
    {
        VALUE,
        LIST_START,
        MAP_START,
        STRUCT_START,
        FIELD_START,
        END;
    }

    private Type type;

    private MingleIdentifier fld; // valid for FIELD_START

    private QualifiedTypeName structType; // valid for STRUCT_START

    private ListTypeReference listType; // valid for START_LIST

    private MingleValue val; // valid for VALUE

    private ObjectPath< MingleIdentifier > path;

    public
    MingleReactorEvent
    copy( boolean copyPath )
    {
        MingleReactorEvent res = new MingleReactorEvent();
        res.type = this.type;
        res.listType = this.listType;
        res.fld = this.fld;
        res.structType = this.structType;
        res.val = this.val;
        if ( copyPath ) res.path = this.path;
        
        return res;
    }

    public Type type() { return type; }

    private
    < V >
    V
    checkType( V val,
               Type expct,
               String callDesc )
    {
        state.isTruef( expct == type,
            "attempt to call %s when type is %s (should be %s)", 
                callDesc, type, expct );

        return val;
    }

    private
    MingleReactorEvent
    resetTo( Type type )
    {
        this.type = type;
        return this;
    }

    public 
    MingleReactorEvent
    setStartList( ListTypeReference listType ) 
    { 
        this.listType = inputs.notNull( listType, "listType" );
        return resetTo( Type.LIST_START ); 
    }

    public 
    MingleReactorEvent 
    setStartMap() 
    { 
        return resetTo( Type.MAP_START ); 
    }

    public MingleReactorEvent setEnd() { return resetTo( Type.END ); }

    public
    MingleReactorEvent
    setStartStruct( QualifiedTypeName qn )
    {
        this.structType = inputs.notNull( qn, "qn" );
        return resetTo( Type.STRUCT_START );
    }

    public
    MingleReactorEvent
    setStartField( MingleIdentifier fld )
    {
        this.fld = inputs.notNull( fld, "fld" );
        return resetTo( Type.FIELD_START );
    }

    public
    MingleReactorEvent
    setValue( MingleValue val )
    {
        this.val = inputs.notNull( val, "val" );
        return resetTo( Type.VALUE );
    }

    public
    MingleIdentifier
    field()
    {
        return checkType( fld, Type.FIELD_START, "field()" );
    }

    public
    QualifiedTypeName
    structType()
    {
        return checkType( structType, Type.STRUCT_START, "structType()" );
    }

    public
    ListTypeReference
    listType()
    {
        return checkType( listType, Type.LIST_START, "listType()" );
    }

    public
    MingleValue
    value()
    {
        return checkType( val, Type.VALUE, "value()" );
    }

    // path can be null
    public
    void
    setPath( ObjectPath< MingleIdentifier > path )
    {
        this.path = path;
    }

    // Returns the path value most recently set by a call to setPath(), which
    // may be null.  The path returned from this call may be a mutable one, but
    // should never change while this instance is being used in an invocation of
    // MingleReactor.processEvent() unless the reactor itself changes it.
    public ObjectPath< MingleIdentifier > path() { return path; }

    private
    void
    addPair( List< Object > pairs,
             Object k,
             Object v )
    {
        pairs.add( k );
        pairs.add( v );
    }

    private
    List< Object >
    inspectionPairs()
    {
        List< Object > res = Lang.newList( 8 );
        
        addPair( res, "type", type );

        if ( path != null ) { 
            addPair( res, "path", Mingle.formatIdPath( path ) );
        }

        switch ( type ) {
        case FIELD_START: addPair( res, "field", fld ); break;
        case STRUCT_START: addPair( res, "structType", structType ); break;
        case LIST_START: addPair( res, "listType", listType ); break;
        case VALUE: addPair( res, "value", Mingle.inspect( val ) ); break;
        }

        return res;
    }

    public
    String
    inspect()
    {
        List< Object > pairs = inspectionPairs();

        return new StringBuilder().append( "[ " ).
            append( Strings.crossJoin( "=", ", ", pairs ) ).
            append( " ]" ).
            toString();
    }
}
