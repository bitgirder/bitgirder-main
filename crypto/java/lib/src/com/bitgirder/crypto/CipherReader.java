package com.bitgirder.crypto;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.ProtocolProcessors;
import com.bitgirder.io.AbstractProcessContext;
import com.bitgirder.io.IoUtils;

import java.nio.ByteBuffer;

public
final
class CipherReader
extends CipherIoFilter
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private ByteBuffer inputBuf;
    private boolean inputDone;
    private boolean cipherBufReady;

    private 
    CipherReader( Builder b ) 
    {
        super( b.setOpMode( Ciphers.DECRYPT_MODE ) ); 
    }

    private
    void
    doCipher()
        throws Exception
    {
        inputBuf.flip(); 

//        code( "doing cipher, inputBuf:", inputBuf );
        doCipher( inputBuf, inputDone );
//        code( 
//            "after doCipher, inputBuf:", inputBuf, "; cipherBuf:", cipherBuf()
//        );

        // sanity check and reset inputBuf in case we read more data later
        state.isFalse( inputBuf.hasRemaining() );
        inputBuf.clear();

        cipherBufReady = true;
    }
 
    private
    final
    class ReadContext
    extends AbstractProcessContext< ByteBuffer >
    {
        private final ProcessContext< ByteBuffer > ctx;

        private
        ReadContext( ProcessContext< ByteBuffer > ctx )
        {
            super( getActivityContext() );

            this.ctx = ctx;
        }

        protected boolean isFinalImpl() { return false; }
        protected ByteBuffer objectImpl() { return inputBuf; }

        @Override protected void failImpl( Throwable th ) { ctx.fail( th ); }

//        private
//        ProtocolProcessorState
//        completeRead( ProtocolProcessorState ps )
//            throws Exception
//        {
//            state.isFalse( ps.isAsync() );
//
//            // important to set inputDone before a call to doCipher, since
//            // inputDone is used to indicate that this is the last cipher input
//            inputDone = ps.isComplete(); 
//
//            if ( ps.isComplete() || ( ! inputBuf.hasRemaining() ) ) doCipher();
//
//            return null;
//        }

        protected
        void
        completeImpl( boolean isDone )
            throws Exception
        {
            // important to set inputDone before a call to doCipher, since
            // inputDone is used to indicate that this is the last cipher input
            inputDone = isDone;

            if ( isDone || ( ! inputBuf.hasRemaining() ) ) doCipher();

//            ctx.asyncComplete( ps == null ? doProcessCipher( ctx ) : null );
            applyCipher( ctx );
        }
    }

    private
    void
    ensureInputBuf()
    {
        if ( inputBuf == null )
        {
            inputBuf = ByteBuffer.allocate( maxInputLen() );
        }
    }

    private
    void
    readCipherInput( ProcessContext< ByteBuffer > ctx )
        throws Exception
    {
//        code( "reading cipher input" );

        ensureInputBuf();
//        ReadContext rdCtx = new ReadContext( ctx );
        ProtocolProcessors.process( proc(), new ReadContext( ctx ) );
//        ProtocolProcessorState res =
//            ProtocolProcessors.process( proc(), rdCtx );
//
////        code( "read res:", res, "; inputBuf:", inputBuf );
// 
//        return res.isAsync() ? res : rdCtx.completeRead( res );
    }

    private
    void
    copyCipherOutputData( ProcessContext< ByteBuffer > ctx )
        throws Exception
    {
//        code( "Copying cipherBuf()", cipherBuf(), "to ctx:", inspect( ctx ) );
        IoUtils.copy( cipherBuf(), ctx.object() );

        if ( cipherBuf().hasRemaining() )
        {
            state.isFalse( ctx.object().hasRemaining() );
            ctx.complete( false );
        }
        else
        {
            cipherBufReady = false;

            if ( inputDone ) ctx.complete( true );
            else 
            {
                if ( ctx.object().hasRemaining() ) applyCipher( ctx );
                else ctx.complete( false );
            }
        }
    }

    private
    void
    applyCipher( ProcessContext< ByteBuffer > ctx )
        throws Exception
    {
        if ( cipherBufReady ) copyCipherOutputData( ctx );
        else readCipherInput( ctx );
    }

//    private
//    ProtocolProcessorState
//    doProcessCipher( ProcessContext< ByteBuffer > ctx )
//        throws Exception
//    {
//        ProtocolProcessorState ps;
//        while ( ( ps = applyCipher( ctx ) ) == null );
//
//        return ps;
//    }

    void
    processCipher( ProcessContext< ByteBuffer > ctx )
        throws Exception
    {
        state.isTrue( ctx.object().hasRemaining(), "context buffer is empty" );
        applyCipher( ctx );
    }

    public
    final
    static
    class Builder
    extends CipherIoFilter.Builder< Builder >
    {
        public CipherReader build() { return new CipherReader( this ); }
    }
}
