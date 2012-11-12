package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.nio.ByteOrder;
import java.nio.ByteBuffer;

import java.io.OutputStream;
import java.io.IOException;

public
final
class BinWriter
extends BinIoProcessor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final OutputStream os;

    private
    BinWriter( OutputStream os,
               ByteOrder bo ) 
    { 
        super( bo ); 

        this.os = inputs.notNull( os, "os" );
    }

    private
    void
    writeBuffer()
        throws IOException
    {
        IoUtils.write( os, flipBuffer() );
    }

    public
    void
    writeByte( byte b )
        throws IOException
    {
        os.write( b );
    }

    public
    void
    writeInt( int i )
        throws IOException
    {
        resetBuffer( 4 ).putInt( i );
        writeBuffer();
    }

    public
    void
    writeLong( long l )
        throws IOException
    {
        resetBuffer( 8 ).putLong( l );
        writeBuffer();
    }

    public
    void
    writeFloat( float f )
        throws IOException
    {
        resetBuffer( 4 ).putFloat( f );
        writeBuffer();
    }

    public
    void
    writeDouble( double d )
        throws IOException
    {
        resetBuffer( 8 ).putDouble( d );
        writeBuffer();
    }

    public
    void
    writeBoolean( boolean b )
        throws IOException
    {
        writeByte( (byte) ( b ? 1 : 0 ) );
    }

    public
    void
    writeByteArray( byte[] arr )
        throws IOException
    {
        inputs.notNull( arr, "arr" );

        writeInt( arr.length );
        os.write( arr );
    }

    public
    void
    writeUtf8( CharSequence cs )
        throws IOException
    {
        inputs.notNull( cs, "cs" );
        writeByteArray( cs.toString().getBytes( "utf-8" ) );
    }

    public
    static
    BinWriter
    asWriterBe( OutputStream os )
    {
        return new BinWriter( os, ByteOrder.BIG_ENDIAN );
    }

    public
    static
    BinWriter
    asWriterLe( OutputStream os )
    {
        return new BinWriter( os, ByteOrder.LITTLE_ENDIAN );
    }
}
