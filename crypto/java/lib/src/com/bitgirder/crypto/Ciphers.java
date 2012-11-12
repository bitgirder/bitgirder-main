package com.bitgirder.crypto;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.nio.ByteBuffer;

import javax.crypto.Cipher;
import javax.crypto.ShortBufferException;

final
class Ciphers
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final static int ENCRYPT_MODE = Cipher.ENCRYPT_MODE;
    final static int DECRYPT_MODE = Cipher.DECRYPT_MODE;

    private Ciphers() {}

    // Lazily allocate/enlarge the output buf, which may be null to indicate
    // that no outbut buf currently exists. In the common situations, in which
    // the feeding process uses the same buffer for each of its calls (more
    // formally, in which the first value of input.remaining() is the max we'll
    // ever end up seeing), we will only ever allocate one output buf -- large
    // enough to hold a crypt operation plus a block's worth of padding
    //
    // If output is already of acceptible capacity, it is cleared and returned
    //
    // allocEh may be null
    static
    ByteBuffer
    ensureOutBuf( int inputLen,
                  ByteBuffer output,
                  Cipher cipher,
                  AllocationEventHandler allocEh )
        throws Exception
    {
        inputs.nonnegativeI( inputLen, "inputLen" );
        inputs.notNull( cipher, "cipher" );

        int outLen = cipher.getOutputSize( inputLen );

        if ( output == null || output.capacity() < outLen )
        {
            int len = outLen + cipher.getBlockSize();

            if ( allocEh != null ) allocEh.allocatingBuffer( outLen );
            output = ByteBuffer.allocate( len );
        }
        else output.clear();

        return output;
    }

    static
    ByteBuffer
    ensureOutBuf( ByteBuffer input,
                  ByteBuffer output,
                  Cipher cipher,
                  AllocationEventHandler allocEh )
        throws Exception
    {
        inputs.notNull( input, "input" );
        return ensureOutBuf( input.remaining(), output, cipher, allocEh );
    }

    private
    static
    void
    doFinal( ByteBuffer input,
             ByteBuffer output,
             Cipher cipher )
        throws Exception
    {
        if ( input.hasRemaining() ) cipher.doFinal( input, output );
        else
        {
            byte[] arr = cipher.doFinal();

            if ( arr.length > output.remaining() )
            {
                throw new ShortBufferException(
                    "final data is of length " + arr.length + 
                    " but output has remaining: " + output.remaining() );
            }
            else output.put( arr );
        }
    }

    static
    ByteBuffer
    doCipher( ByteBuffer input,
              ByteBuffer output,
              boolean isFinal,
              Cipher cipher )
        throws Exception
    {
        inputs.notNull( input, "input" );
        inputs.notNull( output, "output" );
        inputs.notNull( cipher, "cipher" );

        if ( isFinal ) doFinal( input, output, cipher );
        else cipher.update( input, output );

        return output;
    }

    // just starting out as package-only interface since we're only using it for
    // testing; we could make it public though eventually
    static
    interface AllocationEventHandler
    {
        public
        void
        allocatingBuffer( int len );
    }
}
