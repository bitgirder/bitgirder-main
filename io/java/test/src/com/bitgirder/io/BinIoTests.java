package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.TypedString;

import com.bitgirder.test.Test;

import java.io.OutputStream;
import java.io.InputStream;
import java.io.ByteArrayOutputStream;
import java.io.ByteArrayInputStream;

@Test
final
class BinIoTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private
    final
    static
    class Utf8
    extends TypedString< Utf8 >
    {
        private Utf8( CharSequence s ) { super( s ); }
    }

    private
    BinWriter
    createWriter( OutputStream os,
                  boolean bigEndian )
    {
        return bigEndian ? 
            BinWriter.asWriterBe( os ) : BinWriter.asWriterLe( os );
    }

    private
    BinReader
    createReader( InputStream is,
                  boolean bigEndian )
    {
        return bigEndian ? 
            BinReader.asReaderBe( is ) : BinReader.asReaderLe( is );
    }

    private
    byte[]
    reverse( byte[] arr )
    {
        byte[] res = new byte[ arr.length ];

        for ( int i = 0, e = arr.length; i < e; ++i )
        {
            res[ i ] = arr[ e - 1 - i ];
        }

        return res;
    }

    private
    void
    writeVal( BinWriter wr,
              Object val )
        throws Exception
    {
        if ( val instanceof Object[] ) val = ( (Object[]) val )[ 0 ];

        if ( val instanceof Byte ) wr.writeByte( (Byte) val );
        else if ( val instanceof Integer ) wr.writeInt( (Integer) val );
        else if ( val instanceof Long ) wr.writeLong( (Long) val );
        else if ( val instanceof Float ) wr.writeFloat( (Float) val );
        else if ( val instanceof Double ) wr.writeDouble( (Double) val );
        else if ( val instanceof Boolean ) wr.writeBoolean( (Boolean) val );
        else if ( val instanceof byte[] ) wr.writeByteArray( (byte[]) val );
        else if ( val instanceof Utf8 ) wr.writeUtf8( (CharSequence) val );
        else state.fail( "Unrecognized write val:", val );
    }

    private
    Object
    readVal( BinReader rd,
             Class< ? > cls )
        throws Exception
    {
        if ( cls.equals( Byte.class ) ) return rd.readByte();
        else if ( cls.equals( Integer.class ) ) return rd.readInt();
        else if ( cls.equals( Long.class ) ) return rd.readLong();
        else if ( cls.equals( Float.class ) ) return rd.readFloat();
        else if ( cls.equals( Double.class ) ) return rd.readDouble();
        else if ( cls.equals( Boolean.class ) ) return rd.readBoolean();
        else if ( cls.equals( byte[].class ) ) return rd.readByteArray();
        else if ( cls.equals( Utf8.class ) ) return new Utf8( rd.readUtf8() );
        else throw state.createFail( "Unrecognized read type:", cls );
    }

    private
    void
    expectVal( Object expct,
               BinReader rd )
        throws Exception
    {
        if ( expct instanceof Object[] ) expct = ( (Object[]) expct )[ 1 ];

        Object act = readVal( rd, expct.getClass() );

        if ( expct instanceof byte[] )
        {
            IoTests.assertEqual( (byte[]) expct, (byte[]) act );
        } 
        else state.equal( expct, act );
    }

    private
    void
    assertByteOrdering( Object val,
                        byte[] expct,
                        boolean bigEndian )
        throws Exception
    {
        ByteArrayOutputStream bos = new ByteArrayOutputStream();

        BinWriter wr = createWriter( bos, bigEndian );
        writeVal( wr, val );

        byte[] arr = bos.toByteArray();
        IoTests.assertEqual( expct, arr );

        ByteArrayInputStream bis = new ByteArrayInputStream( arr );
        BinReader rd = createReader( bis, bigEndian );
        state.equal( val, readVal( rd, val.getClass() ) );
    }

    private
    void
    assertByteOrdering( Object val,
                        byte[] beExpct )
        throws Exception
    {
        assertByteOrdering( val, beExpct, true );
        assertByteOrdering( val, reverse( beExpct ), false );
    }

    @Test
    private
    void
    testByteOrdering()
        throws Exception
    {
        assertByteOrdering( 
            Integer.valueOf( 1 ), 
            new byte[] { 0x00, 0x00, 0x00, 0x01 }
        );

        assertByteOrdering(
            Long.valueOf( 1L ),
            new byte[] { 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01 }
        );

        assertByteOrdering(
            Float.valueOf( 1.0f ),
            new byte[] { 0x3f, (byte) 0x80, 0x00, 0x00 }
        );

        assertByteOrdering(
            Double.valueOf( 1.0d ),
            new byte[] { 0x3f, (byte) 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00 }
        );
    }

    private
    void
    assertRoundtrips( boolean bigEndian,
                      Object... vals )
        throws Exception
    {
        ByteArrayOutputStream bos = new ByteArrayOutputStream();
        BinWriter wr = createWriter( bos, bigEndian );

        for ( Object val : vals ) writeVal( wr, val );

        ByteArrayInputStream bis = 
            new ByteArrayInputStream( bos.toByteArray() );

        BinReader rd = createReader( bis, bigEndian );

        for ( Object val : vals ) expectVal( val, rd );
    }

    @Test
    private
    void
    testRoundtrips()
        throws Exception
    {
        for ( boolean bigEndian : new boolean[] { true, false } )
        {
            assertRoundtrips( bigEndian,
                Byte.valueOf( (byte) 0x01 ),
                Integer.valueOf( 1 ),
                Long.valueOf( 1L ),
                Float.valueOf( 1.0f ),
                Double.valueOf( 1.0d ),
                Boolean.TRUE,
                Boolean.FALSE,
                new Object[] { Byte.valueOf( (byte) 2 ), Boolean.TRUE },
                new Object[] { Byte.valueOf( (byte) 0 ), Boolean.FALSE },
                new Object[] { Byte.valueOf( (byte) -1 ), Boolean.TRUE },
                new byte[] {},
                new byte[] { 0x00, 0x01 },
                new Utf8( "" ),
                new Utf8( "hello" )
            );
        }
    }
}
