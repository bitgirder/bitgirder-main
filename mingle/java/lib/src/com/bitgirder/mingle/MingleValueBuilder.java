package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.Deque;
import java.util.List;

public
final 
class MingleValueBuilder
implements MingleValueReactor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Deque< Object > stack = Lang.newDeque();

    private MingleValue val;

    private MingleValueBuilder() {}

    private
    MingleValue
    valueForObject( Object obj )
    {
        if ( obj instanceof MingleSymbolMap.BuilderImpl ) {
            return ( (MingleSymbolMap.BuilderImpl< ?, ? >) obj ).build();
        } else if ( obj instanceof List< ? > ) {
            List< MingleValue > mvList = Lang.castUnchecked( obj );
            return MingleList.asList( mvList );
        }

        throw state.failf( "unhandled object value from stack top: %s", obj );
    }

    private
    void
    processIntermediateValue( MingleValue mv )
    {
        Object obj = stack.peek();

        if ( obj instanceof MingleIdentifier ) 
        {
            MingleIdentifier fld = (MingleIdentifier) stack.pop();

            MingleSymbolMap.BuilderImpl< ?, ? > b =
                (MingleSymbolMap.BuilderImpl< ?, ? >) stack.peek();

            b.set( fld, mv );
            return;
        }
        else if ( obj instanceof List ) {
            Lang.< List< Object > >castUnchecked( obj ).add( mv );
            return;
        }
        
        state.failf( "unexpected stack top %s for value %s", obj, mv );
    }

    private
    void
    processValue( MingleValue mv )
    {
        if ( stack.isEmpty() ) {
            val = mv;
            return;
        }

        processIntermediateValue( mv );
    }

    private
    void
    processEnd()
    {
        Object obj = stack.pop();

        MingleValue mv = valueForObject( obj );
        processValue( mv );
    }

    public
    void
    processEvent( MingleValueReactorEvent ev )
    {
        switch ( ev.type() ) {
        case START_MAP: stack.push( new MingleSymbolMap.Builder() ); return;
        case START_STRUCT: 
            stack.push( new MingleStruct.Builder().setType( ev.structType() ) );
            return;
        case START_FIELD: stack.push( ev.field() ); return;
        case START_LIST: stack.push( Lang.< Object >newList() ); return;
        case VALUE: processValue( ev.value() ); return;
        case END: processEnd(); return;
        }
        
        state.failf( "unhandled type: %s", ev.type() );
    }

    public
    MingleValue
    value()
    {
        state.isFalse( val == null, "value is not built yet" );
        return val;
    }

    public
    static
    MingleValueBuilder
    create()
    {
        return new MingleValueBuilder();
    }
}
