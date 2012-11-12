package com.bitgirder.mingle.codec;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.ObjectReceiver;

import com.bitgirder.io.IoUtils;
import com.bitgirder.io.IoProcessor;
import com.bitgirder.io.IoProcessors;
import com.bitgirder.io.FileWrapper;
import com.bitgirder.io.FileReceive;
import com.bitgirder.io.FileSend;
import com.bitgirder.io.ProtocolCopies;
import com.bitgirder.io.ProtocolProcessor;
import com.bitgirder.io.AbstractProtocolProcessor;

import com.bitgirder.process.ProcessActivity;

import java.util.List;

import java.nio.ByteBuffer;

import java.nio.channels.FileChannel;

public
final
class MingleCodecs
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private MingleCodecs() {}

    private
    static
    ByteBuffer
    allocateAndAdd( List< ByteBuffer > bufs,
                    int sz )
    {
        ByteBuffer buf = ByteBuffer.allocate( sz );
        bufs.add( buf );

        return buf;
    }

    public
    static
    ByteBuffer
    toByteBuffer( MingleEncoder enc )
        throws Exception
    {
        inputs.notNull( enc, "enc" );

        int sz = 256;
        int total = 0;
        List< ByteBuffer > bufs = Lang.newList();

        ByteBuffer buf = allocateAndAdd( bufs, sz );
        while ( ! enc.writeTo( buf ) )
        {
            buf.flip();
            total += buf.remaining();
            buf = allocateAndAdd( bufs, sz *= 2 );
        }

        buf.flip();
        total += buf.remaining();

        ByteBuffer res = ByteBuffer.allocate( total );
        for ( ByteBuffer bb : bufs ) res.put( bb );

        res.flip();
        return res;
    }

    public
    static
    ByteBuffer
    toByteBuffer( MingleCodec codec,
                  Object me )
        throws Exception
    {
        return toByteBuffer(
            inputs.notNull( codec, "codec" ).createEncoder(
                inputs.notNull( me, "me" ) ) );
    }

    public
    static
    byte[]
    toByteArray( MingleCodec codec,
                 Object me )
        throws Exception
    {
        return IoUtils.toByteArray( toByteBuffer( codec, me ) );
    }

    public
    static
    < E >
    E
    fromByteBuffer( MingleDecoder< ? extends E > dec,
                    ByteBuffer buf )
        throws Exception
    {
        inputs.notNull( dec, "dec" );
        inputs.notNull( buf, "buf" );

        state.isTrue(
            dec.readFrom( buf, true ),
            "Decoder returned false() from readFrom() on end of input" );
 
        return dec.getResult();
    }

    public
    static
    < E >
    E
    fromByteBuffer( MingleCodec codec,
                    ByteBuffer buf,
                    Class< E > cls )
        throws Exception
    {
        inputs.notNull( codec, "codec" );
        // buf null checked in delegated call
        inputs.notNull( cls, "cls" );

        MingleDecoder< E > dec = codec.createDecoder( cls );

        return fromByteBuffer( codec.createDecoder( cls ), buf );
    }

    public
    static
    < E >
    E
    fromByteArray( MingleCodec codec,
                   byte[] arr,
                   Class< E > cls )
        throws Exception
    {
        inputs.notNull( arr, "arr" );
        inputs.notNull( cls, "cls" );

        return fromByteBuffer( codec, ByteBuffer.wrap( arr ), cls );
    }

    private
    final
    static
    class EncodableSend
    extends AbstractProtocolProcessor< ByteBuffer >
    {
        private final MingleCodec codec;
        private final Object me;

        private MingleEncoder enc;

        private
        EncodableSend( MingleCodec codec,
                       Object me )
        {
            this.codec = codec;
            this.me = me;
        }

        protected
        void
        processImpl( ProcessContext< ByteBuffer > ctx )
            throws Exception
        {
            if ( enc == null ) enc = codec.createEncoder( me );

            ctx.complete( enc.writeTo( ctx.object() ) );
        }
    }

    private
    final
    static
    class EncodableReceive< E >
    extends AbstractProtocolProcessor< ByteBuffer >
    {
        private final MingleDecoder< ? extends E > dec;
        private final ObjectReceiver< ? super E > recv;

        private
        EncodableReceive( MingleDecoder< ? extends E > dec,
                          ObjectReceiver< ? super E > recv )
        {
            this.dec = dec;
            this.recv = recv;
        }

        // Only calls dec.readFrom() if ctx contains actionable data
        private
        boolean
        readFrom( ProcessContext< ByteBuffer > ctx )
            throws Exception
        {
            ByteBuffer bb = ctx.object();
            boolean isFinal = ctx.isFinal();

            if ( bb.hasRemaining() || isFinal )
            {
                return dec.readFrom( bb, isFinal );
            }
            else return false;
        }

        protected
        void
        processImpl( ProcessContext< ByteBuffer > ctx )
            throws Exception
        {
            if ( readFrom( ctx ) )
            {
                if ( recv != null )
                {
                    E res = dec.getResult();
                    recv.receive( res );
                }

                ctx.complete( true );
            }
            else ctx.complete( false );
        }
    }

    public
    static
    ProtocolProcessor< ByteBuffer >
    createSendProcessor( MingleCodec c,
                         Object enc )
    {
        return 
            new EncodableSend( 
                inputs.notNull( c, "c" ),
                inputs.notNull( enc, "enc" )
            );
    }

    public
    static
    ProtocolProcessor< ByteBuffer >
    createReceiveProcessor( MingleDecoder< ? > dec )
    {
        inputs.notNull( dec, "dec" );
        return new EncodableReceive< Object >( dec, null );
    }

    public
    static
    < E >
    ProtocolProcessor< ByteBuffer >
    createReceiveProcessor( MingleDecoder< ? extends E > dec,
                            ObjectReceiver< ? super E > recv )
    {
        inputs.notNull( dec, "dec" );
        inputs.notNull( recv, "recv" );

        return new EncodableReceive< E >( dec, recv );
    }

    public
    static
    < E >
    ProtocolProcessor< ByteBuffer >
    createReceiveProcessor( MingleCodec c,
                            Class< E > cls,
                            ObjectReceiver< ? super E > recv )
        throws Exception
    {
        inputs.notNull( recv, "recv" );
        inputs.notNull( cls, "cls" );
        inputs.notNull( c, "c" );

        MingleDecoder< E > dec = c.createDecoder( cls );
        return createReceiveProcessor( dec, recv );
    }

    private
    static
    void
    copyToFile( MingleCodec c,
                Object me,
                FileChannel dest,
                ProcessActivity.Context ctx,
                IoProcessor.Client ioCli,
                Runnable onComp )
    {
        ProtocolCopies.copyByteStream(
            createSendProcessor( c, me ),
            new FileReceive.Builder().
                setChannel( dest ).
                setCloseOnComplete( true ).
                setClient( ioCli ).
                build(),
            ByteBuffer.allocate( 1024 ),
            ctx,
            onComp
        );
    }

    public
    static
    void
    toFile( final MingleCodec c,
            final Object me,
            FileWrapper dest,
            final ProcessActivity.Context ctx,
            IoProcessor ioProc,
            final Runnable onComp )
    {
        inputs.notNull( c, "c" );
        inputs.notNull( me, "me" );
        inputs.notNull( dest, "dest" );
        inputs.notNull( ctx, "ctx" );
        inputs.notNull( ioProc, "ioProc" );
        inputs.notNull( onComp, "onComp" );

        final IoProcessor.Client ioCli = ioProc.createClient( ctx );

        ioCli.openFile( 
            dest, 
            IoProcessors.FileOpenMode.SYNC, 
            new ObjectReceiver< FileChannel >() {
                public void receive( FileChannel fc ) throws Exception {
                    copyToFile( c, me, fc, ctx, ioCli, onComp );
                }
            }
        );
    }
    
    private
    static
    < E >
    void
    readFromFile( MingleCodec c,
                  FileChannel fc,
                  Class< E > encCls,
                  ProcessActivity.Context ctx,
                  IoProcessor.Client ioCli,
                  ObjectReceiver< ? super E > recv )
        throws Exception
    {
        ProtocolCopies.copyByteStream(
            new FileSend.Builder().
                setChannel( fc ).
                setClient( ioCli ).
                setCloseOnComplete( true ).
                build(),
            createReceiveProcessor( c, encCls, recv ),
            ByteBuffer.allocate( 1024 ),
            ctx,
            Lang.getNoOpRunnable()
        );
    }

    public
    static
    < E >
    void
    fromFile( final MingleCodec c,
              FileWrapper src,
              final Class< E > encCls,
              final ProcessActivity.Context ctx,
              IoProcessor ioProc,
              final ObjectReceiver< ? super E > recv )
    {
        inputs.notNull( c, "c" );
        inputs.notNull( src, "src" );
        inputs.notNull( encCls, "encCls" );
        inputs.notNull( ctx, "ctx" );
        inputs.notNull( ioProc, "ioProc" );
        inputs.notNull( recv, "recv" );

        final IoProcessor.Client ioCli = ioProc.createClient( ctx );

        ioCli.openFile(
            src, 
            IoProcessors.FileOpenMode.READ,
            new ObjectReceiver< FileChannel >() {
                public void receive( FileChannel fc ) throws Exception {
                    readFromFile( c, fc, encCls, ctx, ioCli, recv );
                }
            }
        );
    }

    static
    MingleCodecDetectionException
    createDetectionNotCompletedException()
    {   
        return 
            new MingleCodecDetectionException( "Detection has not completed" );
    }

    private
    final
    static
    class DetectProcessor
    extends AbstractProtocolProcessor< ByteBuffer >
    {
        private final MingleCodecDetection det;

        private DetectProcessor( MingleCodecDetection det ) { this.det = det; }

        protected
        void
        processImpl( ProcessContext< ByteBuffer > ctx )
            throws Exception
        {
            if ( det.update( ctx.object() ) ) ctx.complete( true );
            else doneOrData( ctx );
        }
    }

    public
    static
    void
    detectCodec( MingleCodecFactory fact,
                 ProtocolProcessor< ByteBuffer > src,
                 int xferBufLen,
                 final ProcessActivity.Context pCtx,
                 final ObjectReceiver< MingleCodecDetection > recv )
    {
        inputs.notNull( fact, "fact" );
        inputs.notNull( src, "src" );
        inputs.positiveI( xferBufLen, "xferBufLen" );
        inputs.notNull( pCtx, "pCtx" );
        inputs.notNull( recv, "recv" );

        final MingleCodecDetection det = fact.createCodecDetection();

        ProtocolCopies.copyByteStream(
            src,
            new DetectProcessor( det ),
            ByteBuffer.allocate( xferBufLen ),
            pCtx,
            new Runnable() {
                public void run() 
                {
                    try { recv.receive( det ); }
                    catch ( Throwable th ) { pCtx.fail( th ); }
                }
            }
        );
    }

    public
    static
    MingleCodec
    detectCodec( MingleCodecFactory fact,
                 ByteBuffer src )
        throws Exception
    {
        inputs.notNull( fact, "fact" );
        inputs.notNull( src, "src" );

        MingleCodecDetection det = fact.createCodecDetection();

        if ( det.update( src.duplicate() ) ) return det.getResult();
        else throw createDetectionNotCompletedException();
    }
}
