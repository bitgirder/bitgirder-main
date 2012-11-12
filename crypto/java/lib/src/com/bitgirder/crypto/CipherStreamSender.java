package com.bitgirder.crypto;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.ProtocolProcessors;
import com.bitgirder.io.DelegateProcessContext;
import com.bitgirder.io.IoUtils;

import java.nio.ByteBuffer;

import javax.crypto.Cipher;

public
final
class CipherStreamSender
extends CipherStreamProcessor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private ByteBuffer cipherBuf;
    private Boolean drainRes;

    private CipherStreamSender( Builder b ) { super( b, Cipher.ENCRYPT_MODE ); }

    private
    void
    drainCipherBuffer( Boolean isDone,
                       ProcessContext< ByteBuffer > ctx )
    {
        IoUtils.copy( cipherBuf, ctx.object() );

        if ( cipherBuf.hasRemaining() ) 
        {
            drainRes = isDone; // could be a reassign of the same value
            ctx.complete( false );
        }
        else 
        {
            drainRes = null;
            ctx.complete( isDone );
        }
    }

    private
    void
    checkResultState( boolean isDone,
                      boolean wasFinal )
    {
        state.isFalse( 
            wasFinal && ( ! isDone ),
            "Processor returned isDone false result while processing a " +
            "context that was final"
        );
    }

    public
    void
    gotPlaintext( boolean isDone,
                  ProcessContext< ByteBuffer > ctx,
                  ByteBuffer plainBuf )
        throws Exception
    {
        checkResultState( isDone, ctx.isFinal() );

        plainBuf.flip();

        cipherBuf = ensureOutBuf( plainBuf, cipherBuf );

        doCipher( plainBuf, cipherBuf, isDone );
        cipherBuf.flip();
        
        drainCipherBuffer( isDone, ctx );
    }

//    private
//    final
//    class PipelineAdapterImpl
//    implements ProtocolProcessors.PipelineAdapter< ByteBuffer, ByteBuffer >
//    {
//        private
//        void
//        checkResultState( ProtocolProcessorState ps,
//                          boolean wasFinal )
//        {
//            state.isFalse( 
//                wasFinal && ( ! ps.isComplete() ),
//                "Processor returned non-complete result while processing a " +
//                "context that was final:", ps
//            );
//        }
//
//        public
//        ProtocolProcessorState
//        getResultState( ProtocolProcessorState ps,
//                        ProcessContext< ByteBuffer > upCtx,
//                        ProcessContext< ByteBuffer > downCtx )
//            throws Exception
//        {
//            checkResultState( ps, downCtx.isFinal() );
//
//            ByteBuffer downBuf = downCtx.object();
//            downBuf.flip();
//
//            cipherBuf = ensureOutBuf( downBuf, cipherBuf );
//
//            doCipher( downBuf, cipherBuf, ps.isComplete() );
//            cipherBuf.flip();
//            
//            return drainCipherBuffer( ps, upCtx );
//        }
//    }
    
    private
    void
    getNextPlaintext( ProcessContext< ByteBuffer > ctx )
    {
        ProtocolProcessors.process(
            processor(),
            new DelegateProcessContext< ByteBuffer, ByteBuffer >( ctx )
            {
                private final ByteBuffer obj = context().object().slice();

                public ByteBuffer object() { return obj; }

                protected void completeImpl( boolean isDone ) throws Exception {
                    gotPlaintext( isDone, context(), obj );
                }
            }
        );
//            return ProtocolProcessors.process(
//                processor(),
//                ctx.object().slice(),
//                ctx.isFinal(),
//                ctx,
//                new PipelineAdapterImpl()
//            );
//        }
    }

    protected
    void
    processImpl( ProcessContext< ByteBuffer > ctx )
        throws Exception
    {
        if ( drainRes == null ) getNextPlaintext( ctx );
        else drainCipherBuffer( drainRes, ctx );
    }

    public
    final
    static
    class Builder
    extends CipherStreamProcessor.Builder< CipherStreamSender, Builder >
    {
        public 
        CipherStreamSender 
        build() 
        { 
            return new CipherStreamSender( this );
        }
    }
}
