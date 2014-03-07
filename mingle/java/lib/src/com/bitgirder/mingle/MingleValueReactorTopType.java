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
            MingleValueReactorEvent.Type.LIST_START,
            MingleValueReactorEvent.Type.MAP_START,
            MingleValueReactorEvent.Type.STRUCT_START
        )
    ),

    LIST( EnumSet.of( MingleValueReactorEvent.Type.LIST_START ) ),
    
    MAP( EnumSet.of( MingleValueReactorEvent.Type.MAP_START ) ),

    STRUCT( EnumSet.of( MingleValueReactorEvent.Type.STRUCT_START ) );

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
