package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Inspector;
import com.bitgirder.lang.Inspectable;

final
class ObjectInstanceMatcher< T >
implements TerminalMatcher< T >,
           Inspectable
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final T obj;

    ObjectInstanceMatcher( T obj ) { this.obj = state.notNull( obj, "obj" ); }

    public boolean isMatch( Object inObj ) { return obj.equals( inObj ); }

    public Inspector accept( Inspector i ) { return i.add( "obj", obj ); }
}
