package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.nio.ByteOrder;
import java.nio.ByteBuffer;

import java.io.IOException;
import java.io.InputStream;

public
final
class BinReader
extends BinIoProcessor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final InputStream is;

    private
    BinReader( InputStream is,
               ByteOrder bo )
    {
        super( bo );

        this.is = inputs.notNull( is, "is" );
    }

    private
    void
    readBuffer( int len )
        throws IOException
    {
        ByteBuffer buf = resetBuffer( len );

        IoUtils.fill( is, buf.array(), IoUtils.arrayPosOf( buf ), len );
        buf.position( buf.limit() ); // set in accordance with a read
    }

    public
    byte
    readByte()
        throws IOException
    {
        return (byte) is.read();
    }

    public
    int
    readInt()
        throws IOException
    {
        readBuffer( 4 );
        return flipBuffer().getInt();
    }

    public
    long
    readLong()
        throws IOException
    {
        readBuffer( 8 );
        return flipBuffer().getLong();
    }

    public
    float
    readFloat()
        throws IOException
    {
        readBuffer( 4 );
        return flipBuffer().getFloat();
    }

    public
    double
    readDouble()
        throws IOException
    {
        readBuffer( 8 );
        return flipBuffer().getDouble();
    }

    public
    boolean
    readBoolean()
        throws IOException
    {
        return readByte() != 0;
    }

    public
    byte[]
    readByteArray()
        throws IOException
    {
        int sz = readInt();

        if ( sz < 0 ) 
        {
            long uSz = Lang.asUnsignedInt( sz );
            throw new IOException( "Array size too large: " + uSz );
        }

        byte[] res = new byte[ sz ];
        IoUtils.fill( is, res );

        return res;
    }

    public
    String
    readUtf8()
        throws IOException
    {
        return new String( readByteArray(), "utf-8" );
    }

    public
    static
    BinReader
    asReaderBe( InputStream is )
    {
        return new BinReader( is, ByteOrder.BIG_ENDIAN );
    }

    public
    static
    BinReader
    asReaderLe( InputStream is )
    {
        return new BinReader( is, ByteOrder.LITTLE_ENDIAN );
    }
}
