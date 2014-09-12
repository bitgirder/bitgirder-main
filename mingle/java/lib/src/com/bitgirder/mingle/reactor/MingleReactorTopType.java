package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.util.EnumSet;

public
enum MingleReactorTopType
{
    VALUE( 
        EnumSet.of(
            MingleReactorEvent.Type.VALUE,
            MingleReactorEvent.Type.LIST_START,
            MingleReactorEvent.Type.MAP_START,
            MingleReactorEvent.Type.STRUCT_START
        )
    ),

    LIST( EnumSet.of( MingleReactorEvent.Type.LIST_START ) ),
    
    MAP( EnumSet.of( MingleReactorEvent.Type.MAP_START ) ),

    STRUCT( EnumSet.of( MingleReactorEvent.Type.STRUCT_START ) );

    private final static Inputs inputs = new Inputs();
    private final static State state = new State();
    
    private final EnumSet< MingleReactorEvent.Type > startSet;

    private
    MingleReactorTopType( 
        EnumSet< MingleReactorEvent.Type > startSet )
    {
        this.startSet = startSet;
    }

    public
    boolean
    couldStartWithEvent( MingleReactorEvent ev )
    {
        inputs.notNull( ev, "ev" );
        return startSet.contains( ev.type() );
    }
}
