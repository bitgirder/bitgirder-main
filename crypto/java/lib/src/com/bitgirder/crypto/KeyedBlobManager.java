package com.bitgirder.crypto;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

import com.bitgirder.io.IoUtils;
import com.bitgirder.io.BinReader;
import com.bitgirder.io.BinWriter;

import java.util.SortedMap;
import java.util.List;
import java.util.Map;

import java.io.IOException;
import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;

import java.nio.ByteBuffer;

import java.security.GeneralSecurityException;

import javax.crypto.SecretKey;
import javax.crypto.Cipher;

import javax.crypto.spec.IvParameterSpec;

// Currently always uses lexicographically highest key for encrypt; we can allow
// callers later to set a different behavior
//
// Message and its associated IO methods are pkg accessible to aid in testing
public
final
class KeyedBlobManager
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static byte FLD_END = (byte) 0x00;
    private final static byte FLD_KEY = (byte) 0x01;
    private final static byte FLD_DATA = (byte) 0x02;
    private final static byte FLD_IV = (byte) 0x03;

    private final String trans;
    private final int ivLen; // only used when > 0
    private final String digAlg = "SHA-256";

    private final SortedMap< String, SecretKey > keys;

    private
    KeyedBlobManager( Builder b )
    {
        this.trans = inputs.notNull( b.trans, "trans" );
        this.ivLen = b.ivLen; // valid since b.trans is checked

        inputs.isFalse( b.keys.isEmpty(), "No keys set" );
        this.keys = Lang.newSortedMap();
        this.keys.putAll( b.keys );
    }

    final
    static
    class Message
    {
        String keyId;
        byte[] data;
        IvParameterSpec iv;
    }

    private
    byte[]
    prependHash( byte[] data )
    {
        MessageDigester dig = CryptoUtils.createDigester( digAlg );
        int digLen = dig.expectDigestLength();

        byte[] res = new byte[ digLen + data.length ];

        dig.update( ByteBuffer.wrap( data ) );
        ByteBuffer hash = dig.digest();
        hash.get( res, 0, digLen );

        System.arraycopy( data, 0, res, digLen, data.length );

        return res;
    }

    private
    void
    writeBlob( Message msg,
               BinWriter wr )
        throws IOException
    {
        wr.writeByte( FLD_KEY );
        wr.writeUtf8( msg.keyId );
        wr.writeByte( FLD_DATA );
        wr.writeByteArray( msg.data );

        if ( msg.iv != null )
        {
            wr.writeByte( FLD_IV );
            wr.writeByteArray( msg.iv.getIV() );
        }

        wr.writeByte( FLD_END );
    }

    byte[]
    makeBlob( Message msg )
    {
        ByteArrayOutputStream bos = new ByteArrayOutputStream();
        BinWriter wr = BinWriter.asWriterLe( bos );

        try 
        { 
            writeBlob( msg, wr ); 
            bos.close();

            return bos.toByteArray();
        }
        catch ( Exception ex ) 
        {
            throw new RuntimeException( "Failed to write blob", ex );
        }
    }

    // Currently creates a new Cipher each time; may be better later to use a
    // ThreadLocal instance
    //
    // Written to be public later, but holding back on exposing it so until
    // needed.
    private
    byte[]
    encrypt( byte[] plain,
             String keyId )
        throws GeneralSecurityException
    {
        inputs.notNull( plain, "plain" );
        inputs.notNull( keyId, "keyId" );

        SecretKey key = inputs.get( keys, keyId, "keys" );
        Cipher c = CryptoUtils.createCipher( trans );

        Message msg = new Message();
        msg.keyId = keyId;

        if ( ivLen > 0 )
        {
            msg.iv = CryptoUtils.createRandomIvSpec( ivLen );
            CryptoUtils.initEncrypt( c, key, msg.iv );
        }
        else CryptoUtils.initEncrypt( c, key );

        msg.data = c.doFinal( prependHash( plain ) );

        return makeBlob( msg );
    }

    public
    byte[]
    encrypt( byte[] plain )
        throws GeneralSecurityException
    {
        return encrypt( plain, keys.lastKey() );
    }

    public
    final
    static
    class BlobFormatException
    extends Exception
    {
        private BlobFormatException( String msg ) { super( msg ); }

        private
        BlobFormatException( String msg,
                             Throwable th )
        {
            super( msg, th );
        }
    }

    public
    final
    static
    class InvalidKeyException
    extends Exception
    {
        private 
        InvalidKeyException( String keyId )
        {
            super( "Invalid key id: " + keyId );
        }
    }

    private
    SecretKey
    keyFor( String keyId )
        throws InvalidKeyException
    {
        SecretKey res = keys.get( keyId );

        if ( res != null ) return res;

        throw new InvalidKeyException( keyId );
    }

    public
    final
    static
    class CorruptBlobException
    extends Exception
    {
        private CorruptBlobException( Throwable th ) { super( th ); }
        private CorruptBlobException( String msg ) { super( msg ); }
        private CorruptBlobException() {}

        private
        CorruptBlobException( String msg,
                              Throwable th )
        {
            super( msg, th );
        }
    }

    private
    void
    readMessage( Message msg,
                 BinReader rd )
        throws IOException,
               BlobFormatException
    {
        while ( true )
        {
            byte b = rd.readByte();

            switch ( b )
            {
                case FLD_END: return;
                case FLD_KEY: msg.keyId = rd.readUtf8(); break;
                case FLD_DATA: msg.data = rd.readByteArray(); break;

                case FLD_IV: 
                    msg.iv = new IvParameterSpec( rd.readByteArray() ); break;

                default:
                    throw new BlobFormatException(
                        String.format( "Unrecognized field: 0x%02x", b ) );
            }
        }
    }

    private
    Message
    validate( Message msg )
        throws BlobFormatException
    {
        String err = null;

        if ( msg.keyId == null ) err = "Missing key id";
        else if ( msg.data == null ) err = "Missing data";

        if ( err == null ) return msg;
        throw new BlobFormatException( err );
    }

    Message
    readMessage( byte[] blob )
        throws BlobFormatException
    {
        Message res = new Message();

        BinReader rd = BinReader.asReaderLe( new ByteArrayInputStream( blob ) );

        try { readMessage( res, rd ); }
        catch ( IOException ioe )
        {
            throw new BlobFormatException( "Couldn't read blob", ioe );
        }

        return validate( res );
    }

    private
    byte[]
    checkHash( byte[] data )
        throws CorruptBlobException
    {
        MessageDigester dig = CryptoUtils.createDigester( digAlg );

        int digLen = dig.expectDigestLength();

        ByteBuffer expct = ByteBuffer.wrap( data, 0, digLen );
        
        ByteBuffer toHash = 
            ByteBuffer.wrap( data, digLen, data.length - digLen );

        dig.update( toHash.slice() );
        ByteBuffer act = dig.digest();

        if ( ! act.equals( expct ) ) throw new CorruptBlobException();
        return IoUtils.toByteArray( toHash );
    }

    public
    byte[]
    decrypt( byte[] blob )
        throws GeneralSecurityException,
               BlobFormatException,
               InvalidKeyException,
               CorruptBlobException
    {
        inputs.notNull( blob, "blob" );

        Message msg = readMessage( blob );

        SecretKey key = keyFor( msg.keyId );

        Cipher c = CryptoUtils.createCipher( trans );

        if ( msg.iv == null ) CryptoUtils.initDecrypt( c, key );
        else CryptoUtils.initDecrypt( c, key, msg.iv );

        try { return checkHash( c.doFinal( msg.data ) ); }
        catch ( GeneralSecurityException gse ) 
        { 
            throw new CorruptBlobException( gse );
        }
    }

    public
    final
    static
    class Builder
    {
        private String trans;
        private int ivLen;

        // package-level to help with testing
        final SortedMap< String, SecretKey > keys = Lang.newSortedMap();

        public
        Builder
        setTransformation( String trans )
        {
            this.trans = inputs.notNull( trans, "trans" );
            ivLen = CryptoUtils.ivLengthOf( trans );

            return this;
        }

        public
        Builder
        setKey( String id,
                SecretKey key )
        {
            inputs.notNull( id, "id" );
            inputs.notNull( key, "key" );

            Lang.putUnique( keys, id, key );

            return this;
        }

        public
        Builder
        setKeys( Map< String, SecretKey > keys )
        {
            inputs.noneNull( keys, "keys" );

            for ( Map.Entry< String, SecretKey > e : keys.entrySet() )
            {
                setKey( e.getKey(), e.getValue() );
            }

            return this;
        } 

        public
        KeyedBlobManager
        build()
        {
            return new KeyedBlobManager( this );
        }
    }
}
