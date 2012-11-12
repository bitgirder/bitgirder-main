package com.bitgirder.crypto;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.ProtocolProcessors;
import com.bitgirder.io.AbstractProcessContext;

import java.nio.ByteBuffer;

public
final
class CipherWriter
extends CipherIoFilter
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    // set initial value so we are ready to encrypt on first process call
    private boolean cipherReady = true;

    private boolean isLastDrain;

    private 
    CipherWriter( Builder b ) 
    { 
        super( b.setOpMode( Ciphers.ENCRYPT_MODE ) ); 
    }

    private
    final
    class ApplyContext
    extends AbstractProcessContext< ByteBuffer >
    {
        private final ByteBuffer bb;
        private final boolean isFinal;
        private final ProcessContext< ByteBuffer > ctx;

        private
        ApplyContext( ByteBuffer bb,
                      boolean isFinal,
                      ProcessContext< ByteBuffer > ctx )
        {
            super( getActivityContext() );

            this.bb = bb;
            this.isFinal = isFinal;
            this.ctx = ctx;
        }

        protected boolean isFinalImpl() { return isFinal; }
        protected ByteBuffer objectImpl() { return bb; }
        
        @Override protected void failImpl( Throwable th ) { ctx.fail( th ); }

//        private
//        ProtocolProcessorState
//        getResult( ProtocolProcessorState ps )
//        {
////            state.isFalse( 
////                bb.hasRemaining(), "processor did not process all data" );
// 
//            switch ( ps )
//            {
//                case DATA: 
//                    if ( cipherBuf().hasRemaining() ) return null;
//                    else
//                    {
//                        cipherReady = true;
//
//                        if ( ctx.object().hasRemaining() ) return null;
//                        else return doneOrData( ctx );
//                    }
// 
//                case COMPLETE: 
//                    state.isTrue( 
//                        isLastDrain && ( ! cipherBuf().hasRemaining() ), 
//                        "Unexpected complete from proc" );
//                    return ps;
// 
//                case ASYNC: throw state.createFail( "Got async" );
// 
//                default: throw state.createFail( "Unhandled:", ps );
//            }
//        }

        protected
        void
        completeImpl( boolean isDone )
            throws Exception
        {
            if ( isDone )
            {
                state.isTrue( 
                    isLastDrain && ( ! cipherBuf().hasRemaining() ), 
                    "Unexpected complete from proc" );

                ctx.complete( true );
            }
            else
            {
                if ( cipherBuf().hasRemaining() ) doProcessCipher( ctx );
                else
                {
                    cipherReady = true;

                    if ( ctx.object().hasRemaining() ) doProcessCipher( ctx );
                    else doneOrData( ctx );
                }
            } 
//            ps = getResult( ps );
//            ctx.asyncComplete( ps == null ? processCipher( ctx ) : ps );
        }
    }

    private
    void
    applyProcessor( ByteBuffer bb,
                    boolean isFinal,
                    ProcessContext< ByteBuffer > ctx )
//        throws Exception
    {
        ApplyContext apCtx = new ApplyContext( bb, isFinal, ctx );
        ProtocolProcessors.process( proc(), apCtx );
//
//        ProtocolProcessorState res = 
//            ProtocolProcessors.process( proc(), apCtx );
//
////        code( "proc res:", res, "; bb:", bb );
//        return res.isAsync() ? res : apCtx.getResult( res );
    }

    private
    void
    applyCipher( ProcessContext< ByteBuffer > ctx )
        throws Exception
    {
//        code( "In applyCipher, ctx:", inspect( ctx ) );
        doCipher( ctx.object(), ctx.isFinal() );
        cipherReady = false;
        isLastDrain = ctx.isFinal() && ( ! ctx.object().hasRemaining() );
//        code( "After doCipher, ctx:", inspect( ctx ), "; isLastDrain:",
//            isLastDrain, "; cipherBuf:", cipherBuf() );

//        return null;
        doProcessCipher( ctx );
    }

    private
    void
    drainCipherOutput( ProcessContext< ByteBuffer > ctx )
        throws Exception
    {
//        code(
//            "in drainCipherOutput, ctx:", inspect( ctx ), "; isLastDrain:",
//            isLastDrain, "; cipherBuf:", cipherBuf() );

        if ( cipherBuf().hasRemaining() || isLastDrain )
        {
            applyProcessor( cipherBuf(), isLastDrain, ctx );
        }
        else 
        {
            cipherReady = true;
//            return isLastDrain ? complete() : data();
//            return ctx.object().hasRemaining() ? null : data();
            if ( ctx.object().hasRemaining() ) doProcessCipher( ctx );
            else ctx.complete( false );
        }
    }

    private
    void
    doProcessCipher( ProcessContext< ByteBuffer > ctx )
        throws Exception
    {
        if ( cipherReady ) applyCipher( ctx ); else drainCipherOutput( ctx );
//        ProtocolProcessorState ps = null;
//
//        while ( ps == null )
//        {
////            code(
////                "in loop, cipherReady:", cipherReady, "; ctx:", inspect( ctx ),
////                "; cipherBuf:", cipherBuf(), "; isLastDrain:", isLastDrain );
//            ps = cipherReady ? applyCipher( ctx ) : drainCipherOutput( ctx );
////            code( "after loop, cipherReady:", cipherReady, "; ps:", ps );
//        }
//
//        return ps;
    }

    void
    processCipher( ProcessContext< ByteBuffer > ctx )
        throws Exception
    {
        doProcessCipher( ctx );
    }

    public
    final
    static
    class Builder
    extends CipherIoFilter.Builder< Builder >
    {
        public CipherWriter build() { return new CipherWriter( this ); }
    }
}
