package com.bitgirder.lang;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
abstract
class AbstractCompletion< V >
implements Completion< V >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    // both of these could be null (indicating a successful response of 'null')
    private final V res;
    private final Throwable th;

    protected
    AbstractCompletion( V res,
                        Throwable th )
    {
        this.res = res;
        this.th = th;
    }

    public final boolean isOk() { return th == null; }

    private
    void
    checkState( String callName,
                boolean wantOk )
    {
        boolean isOk = isOk();

        if ( isOk != wantOk )
        {
            state.createFail( callName, "called with isOk() returning", isOk );
        }
    }

    public
    final
    Throwable
    getThrowable()
    {
        checkState( "getException()", false );
        return th;
    }

    public
    final
    V
    getResult()
    {
        checkState( "getResult()", true );
        return res;
    }

    public
    final
    V
    get()
        throws Exception
    {
        if ( isOk() ) return res; 
        else 
        {
            if ( th instanceof Exception ) throw (Exception) th;
            else throw (Error) th;
        }
    }

    // subclasses can override
    @Override
    public
    String
    toString()
    {
        return Strings.inspect( this, true, "res", res, "th", th ).toString();
    }
}
