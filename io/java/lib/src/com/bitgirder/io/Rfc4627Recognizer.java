package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

public
abstract
class Rfc4627Recognizer
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    public static enum Status { RECOGNIZING, FAILED, COMPLETED; }

    private Status st = Status.RECOGNIZING;

    private String errMsg;

    Rfc4627Recognizer() {}

    public final Status getStatus() { return st; }
    final void setStatus( Status st ) { this.st = st; }

    public final boolean failed() { return st == Status.FAILED; }
    public final boolean completed() { return st == Status.COMPLETED; }
    public final boolean recognizing() { return st == Status.RECOGNIZING; }

    public
    final
    String
    getErrorMessage()
    {
        state.isTrue( 
            failed(), 
            "Attempt to access error message when failed() is false" 
        );

        return errMsg;
    }

    final
    void
    setFailure( String msg )
    {
        this.errMsg = msg;
        st = Status.FAILED;
    }

    abstract
    int
    recognizeImpl( CharSequence input,
                   int indx,
                   boolean isEnd );

    public
    final
    int
    recognize( CharSequence input,
               int indx,
               boolean isEnd )
    {
        inputs.notNull( input, "input" );
        inputs.nonnegativeI( indx, "indx" );

        inputs.isTrue( input.length() > indx, "indx >= input length" );
        return recognizeImpl( input, indx, isEnd );
    }
}
