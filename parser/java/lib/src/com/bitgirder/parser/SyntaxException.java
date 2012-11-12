package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
class SyntaxException
extends Exception
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final SourceTextLocation loc;
    private final String rawMsg;

    public 
    SyntaxException() 
    { 
        this.loc = null; 
        this.rawMsg = null;
    }

    public 
    SyntaxException( String msg ) 
    { 
        super( msg ); 
        this.loc = null;
        this.rawMsg = null;
    }

    public
    SyntaxException( String msg,
                     SourceTextLocation loc )
    {
        super( 
            new StringBuilder().
                append( inputs.notNull( loc, "loc" ).getFileName() ).
                append( " [" ).
                append( loc.getLine() ).
                append( ',' ).
                append( loc.getColumn() ).
                append( "]: " ).
                append( msg ).
                toString()
        );

        this.loc = loc;
        this.rawMsg = msg;
    }
    
    public 
    SyntaxException( String msg,
                     Throwable cause )
    {
        super( msg, cause );
        this.loc = null;
        this.rawMsg = null;
    }

    public SourceTextLocation getLocation() { return loc; }

    // The message without the location prefixed
    public 
    String 
    getRawMessage() 
    { 
        return rawMsg == null ? getMessage() : rawMsg;
    }
}
