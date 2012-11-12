package com.bitgirder.mingle.bincodec;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.nio.ByteBuffer;

final
class ProgressCheck
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final CharSequence targName;

    private int posOrig;
    private int prevRemain = -1;

    ProgressCheck( CharSequence targName ) 
    { 
        this.targName = inputs.notNull( targName, "targName" );
    }

    void
    enter( ByteBuffer bb )
    {
        state.notNull( bb, "bb" );
        posOrig = bb.position();
    }

    void
    assertProgress( ByteBuffer bb )
    {
        if ( bb.position() == posOrig )
        {
            if ( prevRemain < 0 ) prevRemain = bb.remaining();
            else 
            {
                throw state.createFail(
                    "Repeated calls to", targName, 
                    "with insufficient buffer capacity", prevRemain );
            }
        }
        else prevRemain = -1;
    }
}
