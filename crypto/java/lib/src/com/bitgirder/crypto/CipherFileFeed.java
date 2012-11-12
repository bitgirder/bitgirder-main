package com.bitgirder.crypto;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.ProtocolCopy;

import java.nio.ByteBuffer;

public
final
class CipherFileFeed
extends CipherFileCopy
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private CipherFileFeed( Builder b ) { super( b ); }

    public static Builder encryptBuilder() { return new Builder( true ); }
    public static Builder decryptBuilder() { return new Builder( false ); }

    void
    completeBuild( ProtocolCopy.Builder< ByteBuffer > pcb,
                   CipherStreamProcessor.Builder< ?, ? > b )
        throws Exception
    {
        if ( isEncrypt() )
        {
            pcb.setSender( b.setProcessor( createFileSend() ).build() );
            pcb.setReceiver( processor() );
        }
        else
        {
            pcb.setSender( createFileSend() );
            pcb.setReceiver( b.setProcessor( processor() ).build() );
        }
    }
 
    public
    final
    static
    class Builder
    extends CipherFileCopy.Builder< CipherFileFeed, Builder >
    {
        private Builder( boolean isEncrypt ) { super( isEncrypt ); }

        public CipherFileFeed build() { return new CipherFileFeed( this ); }
    }
}
