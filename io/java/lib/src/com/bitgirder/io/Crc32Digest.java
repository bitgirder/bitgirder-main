package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import java.nio.ByteBuffer;

import java.util.zip.CRC32;

public
final
class Crc32Digest
implements OctetDigest< Long >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final CRC32 crc = new CRC32();

    private Crc32Digest() {}

    // Controls the size of the byte array used to covert from direct byte
    // buffer to byte[] for use with crc. Could make this configurable for
    // callers later.
    private final int copyArrLen = 512;

    public
    void
    update( byte[] arr,
            int offset,
            int len )
    {
        inputs.notNull( arr, "arr" );
        crc.update( arr, offset, len );
    }

    public void update( int b ) { crc.update( b ); }

    public void reset() { crc.reset(); }

    public
    void
    update( ByteBuffer data )
    {
        if ( data.hasArray() ) 
        {
            crc.update( 
                data.array(), IoUtils.arrayPosOf( data ), data.remaining() );

            data.position( data.limit() );
        }
        else
        {
            byte[] buf = new byte[ copyArrLen ];

            while ( data.hasRemaining() )
            {
                int updateLen = Math.min( buf.length, data.remaining() );

                data.get( buf, 0, updateLen );
                crc.update( buf, 0, updateLen );
            }
        }
    }

    public Long digest() { return crc.getValue(); }

    public static Crc32Digest create() { return new Crc32Digest(); }
}
