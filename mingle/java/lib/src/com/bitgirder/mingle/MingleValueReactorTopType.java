package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.util.EnumSet;

public
enum MingleValueReactorTopType
{
    VALUE( 
        EnumSet.of(
            MingleValueReactorEvent.Type.VALUE,
            MingleValueReactorEvent.Type.START_LIST,
            MingleValueReactorEvent.Type.START_MAP,
            MingleValueReactorEvent.Type.START_STRUCT
        )
    ),

    LIST( EnumSet.of( MingleValueReactorEvent.Type.START_LIST ) ),
    
    MAP( EnumSet.of( MingleValueReactorEvent.Type.START_MAP ) ),

    STRUCT( EnumSet.of( MingleValueReactorEvent.Type.START_STRUCT ) );

    private final static Inputs inputs = new Inputs();
    private final static State state = new State();
    
    private final EnumSet< MingleValueReactorEvent.Type > startSet;

    private
    MingleValueReactorTopType( 
        EnumSet< MingleValueReactorEvent.Type > startSet )
    {
        this.startSet = startSet;
    }

    public
    boolean
    couldStartWithEvent( MingleValueReactorEvent ev )
    {
        inputs.notNull( ev, "ev" );
        return startSet.contains( ev.type() );
    }
}
