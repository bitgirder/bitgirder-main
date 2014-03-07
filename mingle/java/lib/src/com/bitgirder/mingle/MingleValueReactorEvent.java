package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.lang.path.ObjectPath;

import java.util.List;

public
final
class MingleValueReactorEvent
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public
    static
    enum Type
    {
        VALUE,
        START_LIST,
        START_MAP,
        START_STRUCT,
        START_FIELD,
        END;
    }

    private Type type;

    private MingleIdentifier fld; // valid for START_FIELD

    private QualifiedTypeName structType; // valid for START_STRUCT

    private MingleValue val; // valid for VALUE

    private ObjectPath< MingleIdentifier > path;

    public
    MingleValueReactorEvent
    copy( boolean copyPath )
    {
        MingleValueReactorEvent res = new MingleValueReactorEvent();
        res.type = this.type;
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
    void
    resetTo( Type type )
    {
        this.type = type;
    }

    public void setStartList() { resetTo( Type.START_LIST ); }

    public void setStartMap() { resetTo( Type.START_MAP ); }

    public void setEnd() { resetTo( Type.END ); }

    public
    void
    setStartStruct( QualifiedTypeName qn )
    {
        this.structType = inputs.notNull( qn, "qn" );
        resetTo( Type.START_STRUCT );
    }

    public
    void
    setStartField( MingleIdentifier fld )
    {
        this.fld = inputs.notNull( fld, "fld" );
        resetTo( Type.START_FIELD );
    }

    public
    void
    setValue( MingleValue val )
    {
        this.val = inputs.notNull( val, "val" );
        resetTo( Type.VALUE );
    }

    public
    MingleIdentifier
    field()
    {
        return checkType( fld, Type.START_FIELD, "field()" );
    }

    public
    QualifiedTypeName
    structType()
    {
        return checkType( structType, Type.START_STRUCT, "structType()" );
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
    // MingleValueReactor.processEvent() unless the reactor itself changes it.
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
        case START_FIELD: addPair( res, "field", fld ); break;
        case START_STRUCT: addPair( res, "structType", structType ); break;
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
