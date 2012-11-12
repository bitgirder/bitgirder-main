package com.bitgirder.crypto;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.AbstractFileAndProtoCopy;
import com.bitgirder.io.AbstractFileCopy;
import com.bitgirder.io.ProtocolCopy;
import com.bitgirder.io.FileWrapper;
import com.bitgirder.io.FileSend;
import com.bitgirder.io.FileReceive;
import com.bitgirder.io.IoProcessor;

import com.bitgirder.process.ProcessActivity;

import java.nio.ByteBuffer;

public
abstract
class CipherFileCopy
extends AbstractFileAndProtoCopy
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final boolean isEncrypt;
    private final CipherFactory cf;

    CipherFileCopy( Builder< ?, ? > b )
    {
        super( b );

        this.isEncrypt = b.isEncrypt;
        this.cf = inputs.notNull( b.cf, "cf" );
    }

    final boolean isEncrypt() { return isEncrypt; }

    private
    CipherStreamProcessor.Builder< ?, ? >
    createStreamProcBuilder()
    {
        CipherStreamProcessor.Builder< ?, ? > b = isEncrypt 
            ? new CipherStreamSender.Builder()
            : new CipherStreamReceiver.Builder();
 
        b.setCipherFactory( cf );

        return b;
    }

    abstract
    void
    completeBuild( ProtocolCopy.Builder< ByteBuffer > pcb,
                   CipherStreamProcessor.Builder< ?, ? > b )
        throws Exception;

    protected
    final
    void
    completeBuild( ProtocolCopy.Builder< ByteBuffer > pcb )
        throws Exception
    {
        completeBuild( pcb, createStreamProcBuilder() ); 
    }

    public
    abstract
    static
    class Builder< C extends CipherFileCopy, B extends Builder< C, B > >
    extends AbstractFileAndProtoCopy.Builder< B >
    {
        private final boolean isEncrypt;
        private CipherFactory cf;

        Builder( boolean isEncrypt ) { this.isEncrypt = isEncrypt; }

        public
        final
        B
        setCipherFactory( CipherFactory cf )
        {
            this.cf = inputs.notNull( cf, "cf" );
            return castThis();
        }

        public
        abstract
        C
        build();
    }

    public
    final
    static
    class CryptFileOp
    extends ProcessActivity
    {
        private final boolean isEncrypt;
        private final FileWrapper src;
        private final FileWrapper dest;
        private final CipherFactory cf;
        private final IoProcessor ioProc;
        private final CipherFileCopy.EventHandler eh;

        private boolean started;

        private 
        CryptFileOp( CryptFileOpBuilder b )
        {
            super( b );

            this.isEncrypt = b.isEncrypt;
            this.src = inputs.notNull( b.src, "src" );
            this.dest = inputs.notNull( b.dest, "dest" );
            this.cf = inputs.notNull( b.cf, "cf" );
            this.ioProc = inputs.notNull( b.ioProc, "ioProc" );
            this.eh = inputs.notNull( b.eh, "eh" );
        }

        private
        CipherFileCopy.Builder< ?, ? >
        initBuilder()
            throws Exception
        {
            IoProcessor.Client ioCli = 
                ioProc.createClient( getActivityContext() );

            if ( isEncrypt )
            {
                return 
                    CipherFileFeed.encryptBuilder().
                        setProcessor( FileReceive.create( dest, ioCli ) ).
                        setFile( src );
            }
            else
            {
                return
                    CipherFileFill.decryptBuilder().
                        setProcessor( FileSend.create( src, ioCli ) ).
                        setFile( dest );
            }
        }

        private
        CipherFileCopy.Builder< ?, ? >
        completeBuild( CipherFileCopy.Builder< ?, ? > b )
        {
            b.setCipherFactory( cf );
            b.setIoProcessor( ioProc );
            b.setActivityContext( getActivityContext() );
            b.setEventHandler( eh );

            return b;
        }

        public
        void
        start()
            throws Exception
        {
            state.isFalse( started, "start() already called" );
            started = true;

            CipherFileCopy.Builder< ?, ? > b = initBuilder();
            completeBuild( b ).build().start();
        }
    } 

    public
    static
    CryptFileOpBuilder
    fileEncryptBuilder()
    {
        return new CryptFileOpBuilder( true );
    }

    public
    static
    CryptFileOpBuilder
    fileDecryptBuilder()
    {
        return new CryptFileOpBuilder( false ); 
    }

    public
    final
    static
    class CryptFileOpBuilder
    extends ProcessActivity.Builder< CryptFileOpBuilder >
    {
        private final boolean isEncrypt;
        private FileWrapper src;
        private FileWrapper dest;
        private CipherFactory cf;
        private IoProcessor ioProc;
        private CipherFileCopy.EventHandler eh;

        private 
        CryptFileOpBuilder( boolean isEncrypt ) 
        { 
            this.isEncrypt = isEncrypt;
        }

        public
        CryptFileOpBuilder
        setSource( FileWrapper src )
        {
            this.src = inputs.notNull( src, "src" );
            return this;
        }

        public
        CryptFileOpBuilder
        setDestination( FileWrapper dest )
        {
            this.dest = inputs.notNull( dest, "dest" );
            return this;
        }

        public
        CryptFileOpBuilder
        setCipherFactory( CipherFactory cf )
        {
            this.cf = inputs.notNull( cf, "cf" );
            return this;
        }

        public
        CryptFileOpBuilder
        setIoProcessor( IoProcessor ioProc )
        {
            this.ioProc = inputs.notNull( ioProc, "ioProc" );
            return this;
        }

        public
        CryptFileOpBuilder
        setEventHandler( CipherFileCopy.EventHandler eh )
        {
            this.eh = inputs.notNull( eh, "eh" );
            return this;
        }
        
        public CryptFileOp build() { return new CryptFileOp( this ); }
    }
}
