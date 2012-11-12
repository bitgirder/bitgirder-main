package com.bitgirder.crypto;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.ProtocolCopy;

import java.nio.ByteBuffer;

public
final
class CipherFileFill
extends CipherFileCopy
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private CipherFileFill( Builder b ) { super( b ); }

    public static Builder encryptBuilder() { return new Builder( true ); }
    public static Builder decryptBuilder() { return new Builder( false ); }

    void
    completeBuild( ProtocolCopy.Builder< ByteBuffer > pcb,
                   CipherStreamProcessor.Builder< ?, ? > b )
        throws Exception
    {
        if ( isEncrypt() )
        {
            pcb.setSender( b.setProcessor( processor() ).build() );
            pcb.setReceiver( createFileReceive() );
        }
        else
        {
            pcb.setSender( processor() );
            pcb.setReceiver( b.setProcessor( createFileReceive() ).build() );
        }
    }

    public
    final
    static
    class Builder
    extends CipherFileCopy.Builder< CipherFileFill, Builder >
    {
        private Builder( boolean isEncrypt ) { super( isEncrypt ); }

        public CipherFileFill build() { return new CipherFileFill( this ); }
    }
}
