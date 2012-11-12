package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.nio.ByteBuffer;

import java.util.Random;

public
final
class IoTestFactory
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final static byte[] EMPTY_BYTE_ARRAY = new byte[ 0 ];

    private final static Random rand = new Random();

    private IoTestFactory() {}

    public
    static
    byte[]
    nextByteArray( int numBytes )
    {
        inputs.nonnegativeI( numBytes, "numBytes" );

        if ( numBytes == 0 ) return EMPTY_BYTE_ARRAY;
        else
        {
            byte[] res = new byte[ numBytes ];
            rand.nextBytes( res );
            
            return res;
        }
    }

    public
    static
    ByteBuffer
    nextByteBuffer( int numBytes )
    {
        return ByteBuffer.wrap( nextByteArray( numBytes ) );
    }

    public
    static
    ByteBuffer
    nextByteBuffer( DataSize sz )
    {
        inputs.notNull( sz, "sz" );

        long numBytes = sz.getByteCount();

        inputs.isTrue( 
            numBytes <= Integer.MAX_VALUE, 
            "Can't allocate byte buffer larger than", Integer.MAX_VALUE );

        return nextByteBuffer( (int) numBytes );
    }

    // fill bb with random bytes between position and limit; does not affect
    // position of buffer; returns original buffer
    public
    static
    ByteBuffer
    nextBytes( ByteBuffer bb )
    {
        ByteBuffer work = inputs.notNull( bb, "bb" ).slice();

        byte[] arr = new byte[ Math.min( 10240, work.remaining() ) ];

        while ( work.hasRemaining() )
        {
            rand.nextBytes( arr );
            work.put( arr, 0, Math.min( arr.length, work.remaining() ) );
        }

        return bb;
    }
    
    public
    static
    FileWrapper
    createTempFile( String prefix )
        throws Exception
    {
        inputs.notNull( prefix, "prefix" );
        return IoUtils.createTempFile( prefix, true );
    }

    public
    static
    FileWrapper
    createTempFile()
        throws Exception
    {
        return createTempFile( "test-file-" );
    }

    public
    static
    DirWrapper
    createTempDir()
        throws Exception
    {
        FileWrapper fw = createTempFile( "test-dir-" );
        if ( fw.exists() ) fw.delete();

        DirWrapper res = new DirWrapper( fw.getFile() );
        res.getFile().mkdirs();

        return res;
    }
}
