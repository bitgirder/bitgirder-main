package com.bitgirder.crypto;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.io.AbstractProtocolProcessor;
import com.bitgirder.io.ProtocolProcessor;

import java.nio.ByteBuffer;

import javax.crypto.Cipher;

public
abstract
class CipherStreamProcessor
extends AbstractProtocolProcessor< ByteBuffer >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final CipherFactory cf;
    private final ProtocolProcessor< ByteBuffer > proc;
    private final int opmode;
    private final Ciphers.AllocationEventHandler allocEh;

    private Cipher cipher;

    CipherStreamProcessor( Builder< ?, ? > b,
                           int opmode )
    {
        this.cf = inputs.notNull( b.cf, "cf" );
        this.proc = inputs.notNull( b.proc, "proc" );
        this.opmode = opmode;
        this.allocEh = b.allocEh;
    }

    final ProtocolProcessor< ByteBuffer > processor() { return proc; }

    final
    Cipher
    cipher()
        throws Exception
    {
        if ( cipher == null ) cipher = cf.initCipher( opmode );
        return cipher;
    }

    final
    ByteBuffer
    ensureOutBuf( ByteBuffer input,
                  ByteBuffer output )
        throws Exception
    {
        return Ciphers.ensureOutBuf( input, output, cipher(), allocEh );
    }

    final
    ByteBuffer
    doCipher( ByteBuffer input,
              ByteBuffer output,
              boolean isFinal )
        throws Exception
    {
        return Ciphers.doCipher( input, output, isFinal, cipher() );
    }

    public
    abstract
    static
    class Builder< P extends CipherStreamProcessor, B extends Builder< P, B > >
    {
        private String trans;
        private CipherFactory cf;
        private ProtocolProcessor< ByteBuffer > proc;
        private Ciphers.AllocationEventHandler allocEh;

        final B castThis() { return Lang.< B >castUnchecked( this ); }

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
        setProcessor( ProtocolProcessor< ByteBuffer > proc )
        {
            this.proc = inputs.notNull( proc, "proc" );
            return castThis();
        }

        final
        B
        setAllocationEventHandler( Ciphers.AllocationEventHandler allocEh )
        {
            this.allocEh = inputs.notNull( allocEh, "allocEh" );
            return castThis();
        }

        public
        abstract
        P
        build();
    }
}
