package com.bitgirder.mingle.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.StandardThread;
import com.bitgirder.lang.Lang;
import com.bitgirder.lang.ObjectReceiver;
import com.bitgirder.lang.Strings;
import com.bitgirder.lang.PatternHelper;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.io.IoUtils;
import com.bitgirder.io.Charsets;

import com.bitgirder.process.ProcessActivity;

import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleString;
import com.bitgirder.mingle.model.MingleSymbolMap;
import com.bitgirder.mingle.model.MingleSymbolMapBuilder;
import com.bitgirder.mingle.model.MingleSymbolMapAccessor;

import java.io.EOFException;
import java.io.InputStream;
import java.io.PrintStream;

import java.nio.ByteBuffer;

import java.util.List;

import java.util.regex.Matcher;
import java.util.regex.Pattern;

import java.util.concurrent.BlockingQueue;

public
final
class MingleMessagePipe
extends ProcessActivity
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();
    
    // Allows key strings which are not valid mingle idents, but we'll catch
    // that elsewhere
    private final static Pattern PAT_MODIFIER_LINE =
        PatternHelper.compile( "^([^:]+)\\s*:\\s*(.+)$" );

    // group ids matched by PAT_MODIFIER
    private final static int GRP_ID_MOD_KEY = 1;
    private final static int GRP_ID_MOD_VAL = 2;

    private final static MingleIdentifier ID_LENGTH =
        MingleIdentifier.create( "length" );

    private final InputStream in;
    private final PrintStream out;

    private final Receiver receiver = new Receiver();
    private final Sender sender = new Sender();

    MingleMessagePipe( InputStream in,
                       PrintStream out,
                       ProcessActivity.Context pCtx )
    {
        super( state.notNull( pCtx, "pCtx" ) );

        this.in = state.notNull( in, "in" );
        this.out = state.notNull( out, "out" );
    }

    void 
    start() 
    { 
        receiver.start();
        sender.start();
    }

    private
    void
    stopThreads()
    {
        receiver.interrupt();
        sender.interrupt();
    }
    
    void stop() { stopThreads(); }

    private
    void
    abort( Throwable th )
    {
        stopThreads();
        fail( th );
    }

    public
    final
    static
    class Message
    {
        private final MingleSymbolMapAccessor mods;
        private final ByteBuffer body;

        private
        Message( MingleSymbolMapAccessor mods,
                 ByteBuffer body )
        {
            this.mods = mods;
            this.body = body;
        }

        public MingleSymbolMapAccessor modifiers() { return mods; }
        public ByteBuffer body() { return body; }
    }

    private
    abstract
    class OpThread< V >
    extends StandardThread
    {
        private final BlockingQueue< V > ops = Lang.newBlockingQueue();

        private 
        OpThread( String opName ) 
        { 
            super( "mingle-message-pipe-" + opName + "-%1$d" ); 
        }
 
        final void enqueue( V op ) { ops.add( op ); }

        abstract
        void
        implProcessOp( V op )
            throws Exception;

        private
        void
        processNextOp()
            throws Exception
        {
            V op = ops.take();
            implProcessOp( op );
        }

        public
        final
        void
        run()
        {
            try { while ( ! isInterrupted() ) processNextOp(); }
            catch ( InterruptedException ignore ) {}
            catch ( Throwable th ) { abort( th ); }
        }
    }

    private
    final
    static
    class ReceiveContext
    {
        private final ObjectReceiver< Message > recv;

        private
        ReceiveContext( ObjectReceiver< Message > recv )
        {
            this.recv = recv;
        }
    }

    private
    final
    class Receiver
    extends OpThread< ReceiveContext >
    {
        private Receiver() { super( "receive" ); }

        private
        ByteBuffer
        accumulateLineByte( int i,
                            ByteBuffer dest,
                            List< ByteBuffer > bufs )
        {
            dest.put( (byte) i );

            if ( ! dest.hasRemaining() )
            {
                dest.flip();
                bufs.add( dest );
                
                return ByteBuffer.allocate( 50 );
            }
            else return dest;
        }

        private
        List< ByteBuffer >
        accumulateLine()
            throws Exception
        {
            List< ByteBuffer > bufs = Lang.newList();
            
            ByteBuffer bb = ByteBuffer.allocate( 50 );

            while ( bb != null )
            {
                int i = in.read();

                if ( i < 0 ) throw new EOFException();
                else if ( i == '\n' )
                {
                    bb.flip();
                    bufs.add( bb );
                    bb = null;
                }
                else bb = accumulateLineByte( i, bb, bufs );
            }

            return bufs;
        }

        private
        CharSequence
        readLine()
            throws Exception
        {
            List< ByteBuffer > bufs = accumulateLine();

            if ( bufs.isEmpty() ) return "";
            else return IoUtils.asString( bufs, Charsets.UTF_8.newDecoder() );
        }

        private
        Matcher
        matchModifierLine( CharSequence line )
        {
            Matcher m = PAT_MODIFIER_LINE.matcher( line );

            state.isTrue( m.matches(), "Unmatched modifier line:", line );
            return m;
        }

        private
        MingleSymbolMapAccessor
        readModifiers()
            throws Exception
        {
            MingleSymbolMapBuilder b = MingleModels.symbolMapBuilder();

            CharSequence line = null;
            while ( ( line = readLine() ).length() > 0 )
            {
                Matcher m = matchModifierLine( line );

                b.setString( 
                    m.group( GRP_ID_MOD_KEY ), m.group( GRP_ID_MOD_VAL ) );
            }

            return MingleSymbolMapAccessor.create( b.build() );
        }

        private
        ByteBuffer
        readBinBody( MingleSymbolMapAccessor mods )
            throws Exception
        {
            int len = mods.expectInt( ID_LENGTH );

            byte[] arr = new byte[ len ];

            for ( int off = 0; off < arr.length; )
            {
                int i = in.read( arr, off, arr.length - off );
                if ( i < 0 ) throw new EOFException(); else off += i;
            }

            return ByteBuffer.wrap( arr );
        }

        void
        implProcessOp( final ReceiveContext ctx )
            throws Exception
        {
            MingleSymbolMapAccessor mods = readModifiers();

            final Message msg = new Message( mods, readBinBody( mods ) );

            submit( new AbstractTask() {
                protected void runImpl() throws Exception {
                    ctx.recv.receive( msg );
                }
            });
        }
    }

    public
    void
    receive( ObjectReceiver< Message > recv )
    {
        inputs.notNull( recv, "recv" );
        receiver.enqueue( new ReceiveContext( recv ) );
    }

    private
    final
    class SendContext
    {
        private final MingleSymbolMap mods;
        private final ByteBuffer body;
        private final Runnable onComp;

        private
        SendContext( MingleSymbolMap mods,
                     ByteBuffer body,
                     Runnable onComp )
        {
            this.mods = mods;
            this.body = body;
            this.onComp = onComp;
        }
    }

    private
    final
    class Sender
    extends OpThread< SendContext >
    {
        private Sender() { super( "sender" ); }

        private
        void
        sendLine( CharSequence line )
            throws Exception
        {
            out.println( line );
        }

        private
        void
        sendHeader( MingleIdentifier key,
                    Object val )
            throws Exception
        {
            CharSequence valStr;

            if ( val instanceof MingleValue ) 
            {
                valStr = (MingleString)
                    MingleModels.asMingleInstance(
                        MingleModels.TYPE_REF_MINGLE_STRING,
                        (MingleValue) val,
                        ObjectPath.< MingleIdentifier >getRoot()
                    );
            }
            else valStr = val.toString();

            sendLine( key.getExternalForm() + ": " + valStr );
        }

        private
        void
        sendResponse( SendContext ctx )
            throws Exception
        {
            for ( MingleIdentifier fld : ctx.mods.getFields() )
            {
                sendHeader( fld, ctx.mods.get( fld ) );
            }

            sendHeader( ID_LENGTH, ctx.body.remaining() );
            sendLine( "" );

            if ( ctx.body.hasRemaining() )
            {
                out.write( IoUtils.toByteArray( ctx.body ) );
            }
        }

        void
        implProcessOp( SendContext ctx )
            throws Exception
        {
            sendResponse( ctx );
            submit( ctx.onComp );
        }
    }

    public
    void
    send( MingleSymbolMap mods,
          ByteBuffer body,
          Runnable onComp )
    {
        inputs.notNull( mods, "mods" );
        inputs.notNull( body, "body" );
        inputs.notNull( onComp, "onComp" );

        sender.enqueue( new SendContext( mods, body, onComp ) );
    }

    public
    void
    send( MingleSymbolMap mods,
          Runnable onComp )
    {
        send( mods, IoUtils.emptyByteBuffer(), onComp );
    }
}
