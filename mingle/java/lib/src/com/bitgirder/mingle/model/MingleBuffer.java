package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.Base64Encoder;

import java.io.IOException;

import java.nio.ByteBuffer;

public
final
class MingleBuffer
implements MingleValue
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static Base64Encoder enc = new Base64Encoder();

    private final ByteBuffer bb;

    // live reference is kept, not copied, not made readOnly()
    MingleBuffer( ByteBuffer bb ) { this.bb = state.notNull( bb, "bb" ); }

    public ByteBuffer getByteBuffer() { return bb.slice(); }

    public CharSequence asBase64String() { return enc.encode( bb ); }

    public
    static
    MingleBuffer
    fromBase64String( CharSequence str )
        throws IOException
    {
        inputs.notNull( str, "str" );

        return new MingleBuffer( enc.asByteBuffer( str ) );
    }

    @Override
    public
    String
    toString()
    {
        return "[mingle buffer, length: " + bb.remaining() + "]";
    }
}
