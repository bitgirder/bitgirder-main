package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.MingleTestMethods;

public
final
class ReactorTestMethods
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private ReactorTestMethods() {}

    // ev1.type known to equal ev2.type
    private
    static
    void
    assertEventDataEqual( MingleReactorEvent ev1,
                          String ev1Name,
                          MingleReactorEvent ev2,
                          String ev2Name )
    {
        Object o1 = null;
        Object o2 = null;
        String desc = null;

        switch ( ev1.type() ) {
        case VALUE: o1 = ev1.value(); o2 = ev2.value(); desc = "value"; break;
        case FIELD_START:
            o1 = ev1.field(); o2 = ev2.field(); desc = "field"; break;
        case STRUCT_START:
            o1 = ev1.structType(); o2 = ev2.structType(); desc = "structType";
            break;
        default: return;
        }

        state.equalf( o1, o2, "%s.%s != %s.%s (%s != %s)",
            ev1Name, desc, ev2Name, desc, o1, o2 );
    }

    public
    static
    void
    assertEventsEqual( MingleReactorEvent ev1,
                       String ev1Name,
                       MingleReactorEvent ev2,
                       String ev2Name )
    {
        if ( state.sameNullity( ev1, ev2 ) ) 
        {
            state.equalf( ev1.type(), ev2.type(), 
                "%s.type != %s.type (%s != %s)", ev1Name, ev2Name, ev1.type(),
                ev2.type() );
        }

        assertEventDataEqual( ev1, ev1Name, ev2, ev2Name );

        if ( state.sameNullity( ev1.path(), ev2.path() ) ) 
        {
            MingleTestMethods.assertIdPathsEqual( 
                ev1.path(), ev1Name + ".path()", 
                ev2.path(), ev2Name + ".path()" );
        }
    }
}
