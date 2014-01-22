package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

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
}
