package com.bitgirder.io.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.AbstractFileAccessTests;
import com.bitgirder.io.IoProcessors;
import com.bitgirder.io.IoExceptionFactory;
import com.bitgirder.io.FileWrapper;

import com.bitgirder.test.Test;

@Test
final
class BitgirderV1IoFileAccessTests
extends AbstractFileAccessTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    protected
    IoExceptionFactory
    exceptionFactory()
    {
        return V1Io.getIoExceptionFactory();
    }

    protected
    void
    assertFailure( Throwable th,
                   String f,
                   IoProcessors.FileOpenMode mode,
                   FailMode fm )
    {
        switch ( fm )
        {
            case EXIST: 
                state.equal( f, ( (NoSuchPathException) th ).path() ); break;
            
            case PERM:
                state.equal( f, ( (PathPermissionException) th ).path() ); 
                break;
            
            default: state.fail( "fm:", fm );
        }
    }
}
