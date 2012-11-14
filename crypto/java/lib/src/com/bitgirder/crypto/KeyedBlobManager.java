package com.bitgirder.crypto;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;
import com.bitgirder.lang.PatternHelper;

import com.bitgirder.io.Base64Encoder;
import com.bitgirder.io.IoUtils;

import java.util.SortedMap;
import java.util.List;
import java.util.Map;
import java.util.Enumeration;

import java.util.regex.Pattern;

import java.io.IOException;

import java.nio.ByteBuffer;

import java.security.GeneralSecurityException;
import java.security.KeyStore;

import javax.crypto.SecretKey;
import javax.crypto.Cipher;

import javax.crypto.spec.IvParameterSpec;

// Currently always uses lexicographically highest key for encrypt; we can allow
// callers later to set a different behavior
public
final
class KeyedBlobManager
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static Base64Encoder b64 = new Base64Encoder();
    
    private final static String KEY_KEY = "key";
    private final static String KEY_IV = "iv";
    private final static String KEY_DATA = "data";

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
    String
    makeString( String keyId,
                IvParameterSpec ivSpec,
                byte[] data )
    {
        List< CharSequence > toks = Lang.newList( 6 );

        toks.add( KEY_KEY );
        toks.add( keyId );
        toks.add( KEY_DATA );
        toks.add( b64.encode( data ) );

        if ( ivSpec != null )
        {
            toks.add( KEY_IV );
            toks.add( b64.encode( ivSpec.getIV() ) );
        }

        return Strings.crossJoin( "=", ",", toks ).toString();
    }

    // Currently creates a new Cipher each time; may be better later to use a
    // ThreadLocal instance
    //
    // Written to be public later, but holding back on exposing it so until
    // needed.
    private
    String
    encrypt( byte[] plain,
             String keyId )
        throws GeneralSecurityException
    {
        inputs.notNull( plain, "plain" );
        inputs.notNull( keyId, "keyId" );

        SecretKey key = inputs.get( keys, keyId, "keys" );
        Cipher c = CryptoUtils.createCipher( trans );
        IvParameterSpec ivSpec = null;

        if ( ivLen > 0 )
        {
            ivSpec = CryptoUtils.createRandomIvSpec( ivLen );
            CryptoUtils.initEncrypt( c, key, ivSpec );
        }
        else CryptoUtils.initEncrypt( c, key );

        byte[] data = c.doFinal( prependHash( plain ) );

        return makeString( keyId, ivSpec, data );
    }

    public
    String
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
    }

    private
    Map< String, String >
    parseBlob( String blob )
        throws BlobFormatException
    {
        Map< String, String > res = Lang.newMap();

        String[] pairStrs = blob.split( "," );

        for ( String pairStr : pairStrs )
        {
            int eqIdx = pairStr.indexOf( '=' );

            if ( eqIdx < 0 )
            {
                throw new BlobFormatException( "Bad pair: " + pairStr );
            }

            // for now we will accept an empty value from a pair like 'key='
            res.put( 
                pairStr.substring( 0, eqIdx ), pairStr.substring( eqIdx + 1 ) );
        }

        return res;
    }

    private
    final
    static
    class DecryptParts
    {
        private String keyId;
        private IvParameterSpec ivSpec;
        private byte[] data;
    }

    private
    < V >
    V
    getPair( Map< String, String > pairs,
             Class< V > cls,
             String key )
        throws BlobFormatException
    {
        String val = pairs.get( key );
        if ( val == null ) return null;

        if ( cls.equals( String.class ) ) return cls.cast( val );
        
        state.isTrue( cls.equals( byte[].class ) );

        try { return cls.cast( b64.decode( val ) ); }
        catch ( IOException ioe )
        {
            throw new BlobFormatException( 
                "Bad base64 data for key " + key + ": " + ioe.getMessage() );
        }
    }

    private
    < V >
    V
    expectPair( Map< String, String > pairs,
                Class< V > cls,
                String key )
        throws BlobFormatException
    {
        V res = getPair( pairs, cls, key );
        
        if ( res != null ) return res;
        throw new BlobFormatException( "No value for key: " + key );
    }

    private
    DecryptParts
    getDecryptParts( String blob )
        throws BlobFormatException
    {
        DecryptParts res = new DecryptParts();

        Map< String, String > pairs = parseBlob( blob );
        res.keyId = expectPair( pairs, String.class, KEY_KEY );
        res.data = expectPair( pairs, byte[].class, KEY_DATA );

        byte[] iv = getPair( pairs, byte[].class, KEY_IV );
        if ( iv != null ) res.ivSpec = new IvParameterSpec( iv );
        
        return res;
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
        private CorruptBlobException() {}
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
    decrypt( String blob )
        throws GeneralSecurityException,
               BlobFormatException,
               InvalidKeyException,
               CorruptBlobException
    {
        inputs.notNull( blob, "blob" );

        DecryptParts dp = getDecryptParts( blob );
        SecretKey key = keyFor( dp.keyId );

        Cipher c = CryptoUtils.createCipher( trans );

        if ( dp.ivSpec == null ) CryptoUtils.initDecrypt( c, key );
        else CryptoUtils.initDecrypt( c, key, dp.ivSpec );

        try { return checkHash( c.doFinal( dp.data ) ); }
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

        private
        List< String >
        getMatchingAliases( KeyStore ks,
                            String selectPat )
            throws GeneralSecurityException
        {
            Pattern pat = PatternHelper.compile( selectPat );

            List< String > res = Lang.newList();

            for ( Enumeration< String > e = ks.aliases(); e.hasMoreElements(); )
            {
                String id = e.nextElement();

                if ( pat.matcher( id ).matches() ) res.add( id );
            }

            return res;
        }

        private
        void
        setSecretKey( KeyStore ks,
                      String id,
                      KeyStore.ProtectionParameter pp )
            throws GeneralSecurityException
        {
            KeyStore.Entry e = ks.getEntry( id, pp );

            if ( ! ( e instanceof KeyStore.SecretKeyEntry ) )
            {
                inputs.fail( "Not a secret key:", id );
            }
            else 
            {
                SecretKey key = ( (KeyStore.SecretKeyEntry) e ).getSecretKey();
                setKey( id, key );
            }
        }

        // This is very simplistic at the moment and assumes a fixed way to
        // match aliases (regex) and that all keys are protected with the given
        // password. We can generalize these assumptions going forward as
        // needed, rewriting this method on top of the more generalized versions
        public
        Builder
        setKeysFrom( KeyStore ks,
                     char[] pass,
                     String selectPat )
            throws GeneralSecurityException
        {
            inputs.notNull( ks, "ks" );
            // pass could be null
            inputs.notNull( selectPat, "selectPat" );

            KeyStore.ProtectionParameter pp = 
                new KeyStore.PasswordProtection( pass );

            for ( String id : getMatchingAliases( ks, selectPat ) )
            {
                setSecretKey( ks, id, pp );
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
