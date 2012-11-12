package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.parser.MingleParsers;

public
final
class MingleServiceResponse
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public final static MingleIdentifier IDENT_RESULT =
        MingleParsers.createIdentifier( "result" );

    public final static MingleIdentifier IDENT_EXCEPTION =
        MingleParsers.createIdentifier( "exception" );

    private final MingleValue res;
    private final MingleException ex;

    private
    MingleServiceResponse( MingleValue res,
                           MingleException ex )
    {
        this.res = res;
        this.ex = ex;
    }

    public boolean isOk() { return ex == null; }

    public
    MingleValue 
    getResult()
    {
        state.isTrue( isOk(), "Attempt to access result but isOk() is false" );
        return res;
    }

    public
    MingleException
    getException()
    {
        state.isFalse( 
            isOk(), "Attempt to access exception but isOk() is true" );

        return ex;
    }

    public
    static
    MingleServiceResponse
    createFailure( MingleException ex )
    {
        inputs.notNull( ex, "ex" );
        return new MingleServiceResponse( null, ex );
    }

    public
    static
    MingleServiceResponse
    createSuccess( MingleValue res )
    {
        inputs.notNull( res, "res" );
        return new MingleServiceResponse( res, null );
    }

    public
    static
    MingleServiceResponse
    createSuccess()
    {
        return new MingleServiceResponse( null, null );
    }
}
