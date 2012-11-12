package com.bitgirder.crypto;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.ProtocolProcessors;
import com.bitgirder.io.DelegateProcessContext;

import java.nio.ByteBuffer;

import javax.crypto.Cipher;

public
final
class CipherStreamReceiver
extends CipherStreamProcessor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private ByteBuffer plain;

    private 
    CipherStreamReceiver( Builder b ) 
    { 
        super( b, Cipher.DECRYPT_MODE ); 
    }

    private
    ByteBuffer
    decrypt( ProcessContext< ByteBuffer > ctx )
        throws Exception
    {
        ByteBuffer input = ctx.object();
        plain = ensureOutBuf( input, plain );

        return doCipher( input, plain, ctx.isFinal() );
    }

//    private
//    final
//    class PipelineAdapterImpl
//    implements ProtocolProcessors.PipelineAdapter< ByteBuffer, ByteBuffer >
//    {
//        public
//        ProtocolProcessorState
//        getResultState( ProtocolProcessorState ps,
//                        ProcessContext< ByteBuffer > upCtx, 
//                        ProcessContext< ByteBuffer > downCtx )
//        {
//            state.isFalse(
//                downCtx.object().hasRemaining() && ps.isData(),
//                "Some plaintext was not consumed by receiver"
//            );
//
//            return ps;
//        }
//    }

    protected
    void
    processImpl( ProcessContext< ByteBuffer > ctx )
        throws Exception
    {
        decrypt( ctx );
        plain.flip();

        ProtocolProcessors.process(
            processor(),
            new DelegateProcessContext< ByteBuffer, ByteBuffer >( ctx )
            {
                public ByteBuffer object() { return plain; }

                protected void completeImpl( boolean isDone )
                {
                    state.isFalse(
                        plain.hasRemaining() && ( ! isDone ),
                        "Some plaintext was not consumed by receiver"
                    );

                    context().complete( isDone );
                }
            }
        );
//        return 
//            ProtocolProcessors.process(
//                processor(),
//                plain,
//                ctx.isFinal(),
//                ctx,
//                new PipelineAdapterImpl()
//            );
    }

    public
    final
    static
    class Builder
    extends CipherStreamProcessor.Builder< CipherStreamReceiver, Builder >
    {
        public 
        CipherStreamReceiver 
        build() 
        { 
            return new CipherStreamReceiver( this );
        }
    }
}
