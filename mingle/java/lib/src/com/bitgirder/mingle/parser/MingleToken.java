package com.bitgirder.mingle.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.parser.SourceTextLocation;

public
final
class MingleToken
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Object obj;
    private final SourceTextLocation loc;

    MingleToken( Object obj,
                 SourceTextLocation loc )
    {
        this.obj = state.notNull( obj, "obj" );
        this.loc = state.notNull( loc, "loc" );
    }

    public Object getObject() { return obj; }
    public SourceTextLocation getLocation() { return loc; }

    @Override
    public
    String
    toString()
    {
        return 
            new StringBuilder().
                append( "[ loc: " ).
                append( loc ).
                append( ", type: " ).
                append( obj.getClass().getSimpleName() ).
                append( ", obj: " ).
                append( obj ).
                append( " ]" ).
                toString();
    }
}
