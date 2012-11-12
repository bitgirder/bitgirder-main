package com.bitgirder.crypto;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.io.ProtocolProcessor;
import com.bitgirder.io.AbstractProtocolProcessor;

import com.bitgirder.process.ProcessActivity;

import java.nio.ByteBuffer;

import javax.crypto.Cipher;

abstract
class CipherIoFilter
extends AbstractProtocolProcessor< ByteBuffer >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static int DEFAULT_MAX_INPUT_LEN = 4096;

    private final ProtocolProcessor< ByteBuffer > proc;
    private final CipherFactory cf;
    private final int opMode;
    private final int maxInputLen;
    private final Ciphers.AllocationEventHandler allocEh;
    private final ProcessActivity.Context pCtx;

    private Cipher cipher;
    private ByteBuffer cipherBuf;

    CipherIoFilter( Builder< ? > b )
    {
        this.proc = inputs.notNull( b.proc, "proc" );
        this.cf = inputs.notNull( b.cf, "cf" );
        this.opMode = inputs.notNull( b.opMode, "opMode" );
        this.maxInputLen = b.maxInputLen;
        this.allocEh = b.allocEh;
        this.pCtx = inputs.notNull( b.pCtx, "pCtx" );
    }

    final ProtocolProcessor< ByteBuffer > proc() { return proc; }
    final ByteBuffer cipherBuf() { return cipherBuf; }
    final int maxInputLen() { return maxInputLen; }
    final ProcessActivity.Context getActivityContext() { return pCtx; }

    // at most maxInputLen bytes of input will be used; output is always
    // cipherBuf(); cipherBuf is clear()ed at the start of each call and flipped
    // before this method returns; cipherBuf() may reference a different buffer
    // at the end of this method than at the beginning
    final
    void
    doCipher( ByteBuffer input,
              boolean isFinal )
        throws Exception
    {
        int lim = Math.min( input.remaining(), maxInputLen );
        
        ByteBuffer bb = input.duplicate();
        bb.limit( bb.position() + lim );

        cipherBuf = 
            Ciphers.ensureOutBuf( maxInputLen, cipherBuf, cipher, allocEh );

        Ciphers.doCipher( 
            bb, 
            cipherBuf, 
            lim == input.remaining() && isFinal, 
            cipher 
        ); 

        cipherBuf.flip();
        input.position( bb.position() );
    }

    private
    void
    ensureInit()
        throws Exception
    {
        if ( cipher == null ) cipher = cf.initCipher( opMode );
    }

    abstract
    void
    processCipher( ProcessContext< ByteBuffer > ctx )
        throws Exception;

    protected
    final
    void
    processImpl( ProcessContext< ByteBuffer > ctx )
        throws Exception
    {
        ensureInit();
        processCipher( ctx );
    }

    static
    abstract
    class Builder< B extends Builder< B > >
    {
        private ProtocolProcessor< ByteBuffer > proc;
        private CipherFactory cf;
        private int maxInputLen = DEFAULT_MAX_INPUT_LEN;
        private Integer opMode;
        private Ciphers.AllocationEventHandler allocEh;
        private ProcessActivity.Context pCtx;

        final B castThis() { return Lang.< B >castUnchecked( this ); }

        public
        final
        B
        setProcessor( ProtocolProcessor< ByteBuffer > proc )
        {
            this.proc = inputs.notNull( proc, "proc" );
            return castThis();
        }

        public
        final
        B
        setCipherFactory( CipherFactory cf )
        {
            this.cf = inputs.notNull( cf, "cf" );
            return castThis();
        }
        
        public
        final
        B
        setMaxInputLength( int maxInputLen )
        {
            this.maxInputLen = inputs.positiveI( maxInputLen, "maxInputLen" );
            return castThis();
        }

        final
        B
        setOpMode( int opMode )
        {
            this.opMode = opMode;
            return castThis();
        }

        public
        final
        B
        setActivityContext( ProcessActivity.Context pCtx )
        {
            this.pCtx = inputs.notNull( pCtx, "pCtx" );
            return castThis();
        }

        final
        B
        setAllocationEventHandler( Ciphers.AllocationEventHandler allocEh )
        {
            this.allocEh = inputs.notNull( allocEh, "allocEh" );
            return castThis();
        }
    }
}
