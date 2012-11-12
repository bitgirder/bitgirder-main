package com.bitgirder.systest.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.StandardThread;

import com.bitgirder.mingle.model.MingleStruct;
import com.bitgirder.mingle.model.MingleStructBuilder;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleValidation;
import com.bitgirder.mingle.model.MingleSymbolMapAccessor;
import com.bitgirder.mingle.model.MingleTypeReference;

import com.bitgirder.mingle.json.JsonMingleCodecs;

import com.bitgirder.mingle.codec.MingleCodec;
import com.bitgirder.mingle.codec.MingleCodecs;

import com.bitgirder.application.ApplicationProcess;

import com.bitgirder.io.IoUtils;
import com.bitgirder.io.IoTestFactory;
import com.bitgirder.io.Crc32Digest;
import com.bitgirder.io.GunzipProcessor;
import com.bitgirder.io.GzipProcessor;
import com.bitgirder.io.GzipHeaderContext;
import com.bitgirder.io.ByteBufferAccumulator;
import com.bitgirder.io.ProtocolProcessors;
import com.bitgirder.io.ProtocolProcessor;
import com.bitgirder.io.ProtocolCopies;
import com.bitgirder.io.ProtocolProcessorTests;

import java.io.DataInputStream;
import java.io.ByteArrayOutputStream;

import java.nio.ByteBuffer;

final
class GzipTester
extends ApplicationProcess
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static String NS_STR = "bitgirder:systest:io@v1/GzipTester";

    private final MingleCodec codec = JsonMingleCodecs.getJsonCodec( "\n" );

    private GzipTester( Configurator c ) { super( c ); }

    private static class Configurator extends ApplicationProcess.Configurator {}

    private
    static
    MingleTypeReference
    makeType( CharSequence nm )
    {
        return MingleTypeReference.create( NS_STR + "/" + nm );
    }

    private
    static
    boolean
    isInstance( MingleStruct ms,
                CharSequence typ )
    {
        return makeType( typ ).equals( ms.getType() );
    }

    private
    final
    static
    class GzipStreamSummary
    {
        private CharSequence fname;
        private CharSequence fcomment;
        private Integer compressedLen;
        private long crc32;
    }

    private
    abstract
    static
    class AbstractStreamXfer
    {
        GzipStreamSummary expct;
    }

    private final static class ReadGzipStream extends AbstractStreamXfer {}
    private final static class SendGzipStream extends AbstractStreamXfer {}

    private
    void
    writeObject( MingleStruct ms )
        throws Exception
    {
        byte[] ser = MingleCodecs.toByteArray( codec, ms );

        System.out.write( ser );
        System.out.flush();
    }

    private
    void
    writeObject( GzipStreamSummary s )
        throws Exception
    {
        MingleStructBuilder b = MingleModels.structBuilder();
        b.setType( NS_STR + "/GzipStreamSummary" );
        
        if ( s.fname != null ) b.f().setString( "fname", s.fname );
        if ( s.fcomment != null ) b.f().setString( "fcomment", s.fcomment );

        if ( s.compressedLen != null )
        {
            b.f().setInt64( "compressedLen", s.compressedLen );
        }

        b.f().setInt64( "crc32", s.crc32 );

        writeObject( b.build() );
    }

    private
    MingleStruct
    readObject()
        throws Exception
    {
        DataInputStream dis = new DataInputStream( System.in );

        int len = dis.readInt();
        state.isTrue( len > 0, "Read non-positive length:", len );

        byte[] arr = new byte[ len ];
        for ( int i = 0; i < len; i += dis.read( arr, i, len - i ) );

        return MingleCodecs.fromByteArray( codec, arr, MingleStruct.class );
    }

    // does not change positions of buffers in acc
    private
    GzipStreamSummary
    createGzipStreamSummary( GunzipProcessor gz,
                             ByteBufferAccumulator acc )
        throws Exception
    {
        GzipStreamSummary res = new GzipStreamSummary();

        res.fname = gz.getFileName();
        res.fcomment = gz.getFileComment();

        res.crc32 = 
            IoUtils.digest( Crc32Digest.create(), false, acc.getBuffers() );

        return res;
    }

    private
    final
    static
    class MutableBool
    {
        private boolean val;
    }

    private
    void
    copyStream( ProtocolProcessor< ByteBuffer > src,
                ProtocolProcessor< ByteBuffer > dest,
                final MutableBool b )
        throws Exception
    {
        ProtocolCopies.copyByteStream(
            src,
            dest,
            ByteBuffer.allocate( 1024 ),
            getActivityContext(),
            new AbstractTask() {
                protected void runImpl() 
                { 
                    synchronized ( b ) 
                    { 
                        b.val = true;
                        b.notify(); 
                    }
                }
            }
        );
    }

    private
    void
    copyBlocking( final ProtocolProcessor< ByteBuffer > src,
                  final ProtocolProcessor< ByteBuffer > dest )
        throws Exception
    {
        final MutableBool b = new MutableBool();

        submit( new AbstractTask() {
            protected void runImpl() throws Exception {
                copyStream( src, dest, b );
            }
        });

        synchronized ( b ) { while ( ! b.val ) b.wait(); }
    }

    private
    void
    runGunzip( final byte[] gzipped,
               final GunzipProcessor gz )
        throws Exception
    {
        copyBlocking(
            ProtocolProcessors.createBufferSend( ByteBuffer.wrap( gzipped ) ), 
            gz
        );
    }

    private
    void
    readGzipStream( ReadGzipStream req )
        throws Exception
    {
        byte[] strm = new byte[ req.expct.compressedLen ];
        IoUtils.fill( System.in, strm );

        ByteBufferAccumulator acc = ByteBufferAccumulator.create( 1024 );

        GunzipProcessor gz = 
            GunzipProcessor.create( acc, getActivityContext() );

//        ProtocolProcessors.
//            processImmediate( gz, ByteBuffer.wrap( strm ), true );
        runGunzip( strm, gz );

        writeObject( createGzipStreamSummary( gz, acc ) );
    }

    private
    GzipProcessor.Builder
    createGzipBuilder( SendGzipStream req )
    {
        GzipProcessor.Builder res = 
            new GzipProcessor.Builder().
                setActivityContext( getActivityContext() );

        GzipHeaderContext.Builder b = new GzipHeaderContext.Builder();

        if ( req.expct.fname != null ) b.setFileName( req.expct.fname );

        if ( req.expct.fcomment != null ) 
        {
            b.setFileComment( req.expct.fcomment );
        }

        return res.setHeaderContext( b.build() );
    }

    private
    ByteBuffer
    gzip( GzipProcessor gz )
        throws Exception
    {
        ByteBufferAccumulator acc = ByteBufferAccumulator.create( 512 );

        copyBlocking( gz, acc );

        return ProtocolProcessorTests.toByteBuffer( acc );
    }

    private
    void
    sendGzipStream( SendGzipStream req )
        throws Exception
    {
        ByteBuffer data = 
            IoTestFactory.nextByteBuffer( req.expct.compressedLen );

        GzipProcessor gz = 
            createGzipBuilder( req ).
            setProcessor( ProtocolProcessors.createBufferSend( data.slice() ) ).
            build();

        ByteBuffer gzipped = gzip( gz );
//        ByteBuffer gzipped = ProtocolProcessorTests.toByteBuffer( gz );

        GzipStreamSummary s = new GzipStreamSummary();

        s.crc32 = IoUtils.digest( Crc32Digest.create(), false, data );
        s.compressedLen = gzipped.remaining();
 
        writeObject( s );

        IoUtils.write( System.out, gzipped );
    }

    private
    GzipStreamSummary
    asSummary( MingleSymbolMapAccessor acc )
    {
        GzipStreamSummary res = new GzipStreamSummary();

        res.fname = acc.getString( "fname" );
        res.fcomment = acc.getString( "fcomment" );
        res.compressedLen = acc.expectInt( "compressedLen" );
        res.crc32 = acc.getLong( "crc32" );

        return res;
    }

    private
    void
    readGzipStream( MingleStruct ms )
        throws Exception
    {
        MingleSymbolMapAccessor acc = MingleSymbolMapAccessor.create( ms );

        ReadGzipStream rd = new ReadGzipStream();
        rd.expct = asSummary( acc.expectSymbolMapAccessor( "expct" ) );

        readGzipStream( rd );
    }

    private
    void
    sendGzipStream( MingleStruct ms )
        throws Exception
    {
        MingleSymbolMapAccessor acc = MingleSymbolMapAccessor.create( ms );

        SendGzipStream snd = new SendGzipStream();
        snd.expct = asSummary( acc.expectSymbolMapAccessor( "expct" ) );

        sendGzipStream( snd );
    }

    private
    boolean
    processNextOp()
        throws Exception
    {
        MingleStruct op = readObject();

        boolean res = true;

        if ( isInstance( op, "ReadGzipStream" ) ) readGzipStream( op );
        else if ( isInstance( op, "SendGzipStream" ) ) sendGzipStream( op );
        else if ( isInstance( op, "Complete" ) ) res = false;
        else state.fail( "Unexpected op:", op );
        
        return res;
    }

    private
    void
    runApp()
        throws Exception
    {
        writeObject( 
            MingleModels.structBuilder().
                setType( NS_STR + "/Start" ).
                build()
        );

        while ( processNextOp() );
        exit();
    }

    protected
    void
    startImpl()
    {
        new StandardThread( "app-run-%1$d" ) { 
            public void run() { 
                try { runApp(); } catch ( Throwable th ) { fail( th ); }
            }
        }.
        start();
    }
}
