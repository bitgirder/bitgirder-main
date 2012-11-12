package com.bitgirder.crypto;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.OctetDigest;

import java.security.MessageDigest;

import java.nio.ByteBuffer;

public
final
class MessageDigester
implements OctetDigest< ByteBuffer >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MessageDigest dig;

    MessageDigester( MessageDigest dig ) 
    { 
        this.dig = state.notNull( dig, "dig" );
    }

    public ByteBuffer digest() { return ByteBuffer.wrap( dig.digest() ); }

    public
    void
    update( ByteBuffer buf )
    {
        dig.update( inputs.notNull( buf, "buf" ) );
    }
}
