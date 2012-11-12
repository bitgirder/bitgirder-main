package com.bitgirder.json;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.IoUtils;
import com.bitgirder.io.DataSize;

import java.io.IOException;

import java.nio.ByteBuffer;

public
final
class JsonSerialization
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static DataSize DEFAULT_SEED_BUF_SIZE =
        DataSize.ofBytes( 256 );

    private JsonSerialization() {}

    public
    static
    ByteBuffer
    toByteBuffer( JsonSerializer ser )
        throws IOException
    {
        inputs.notNull( ser, "ser" );

        ByteBuffer res = IoUtils.allocateByteBuffer( DEFAULT_SEED_BUF_SIZE );
 
        while ( ! ser.writeTo( res ) )
        {
            state.isFalse( 
                res.hasRemaining(), 
                "Serializer returned false from writeTo() but write buffer " +
                "still has remaining" );
            
            res.flip();
            res = IoUtils.expand( res );
        }

        res.flip();
        return res.asReadOnlyBuffer();
    }

    public
    static
    ByteBuffer
    toByteBuffer( JsonText text )
        throws IOException
    {
        inputs.notNull( text, "text" );
        return toByteBuffer( JsonSerializer.create( text ) );
    }
}
