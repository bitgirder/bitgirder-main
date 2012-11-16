package com.bitgirder.crypto;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.PatternHelper;

import com.bitgirder.io.Charsets;
import com.bitgirder.io.IoUtils;

import java.util.Map;
import java.util.List;
import java.util.Enumeration;

import java.util.regex.Pattern;

import java.security.Key;
import java.security.KeyStore;
import java.security.SecureRandom;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.security.InvalidKeyException;
import java.security.GeneralSecurityException;

import java.security.spec.AlgorithmParameterSpec;

import java.io.Console;

import java.nio.ByteBuffer;

import javax.crypto.SecretKey;
import javax.crypto.Mac;
import javax.crypto.KeyGenerator;
import javax.crypto.Cipher;

import javax.crypto.spec.SecretKeySpec;
import javax.crypto.spec.IvParameterSpec;

// A general note about javax.crypto.Cipher, which we'll just put in this class
// for want of a better place. The javadocs for Cipher.update() indicate that it
// should be okay to use the same buffer reference for input/output. Experience
// developing the classes and tests in this package, as well as
// http://bugs.sun.com/bugdatabase/view_bug.do?bug_id=6582580, suggests
// otherwise. So, we do not currently attempt, anywhere in com.bitgirder.crypto,
// to optimize Cipher.update by using the same buf for in/out params in a call
// to that method. 

public
final
class CryptoUtils
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final static String DIG_ALG_MD5 = "md5";

    private final static ThreadLocal< SecureRandom > sr =
        new ThreadLocal< SecureRandom >() {
            @Override protected SecureRandom initialValue() {
                return new SecureRandom();
            }
        };

    public final static String ALG_HMAC_SHA1 = "HmacSha1";

    private CryptoUtils() {}

    private static SecureRandom secureRandom() { return sr.get(); }

    private
    static
    RuntimeException
    createRethrow( String msg,
                   Throwable th )
    {
        return new RuntimeException( msg, th );
    }

    private
    static
    Console
    expectConsole()
    {
        Console res = System.console();
        state.isFalse( res == null, "No console available for password entry" );

        return res;
    }

    // Used in situations where caller would consider it an assertion failure if
    // either the console is unavailable or if the readPassword returns null
    public
    static
    char[]
    readPassword( String fmt,
                  Object... args )
    {
        inputs.notNull( fmt, "fmt" );

        Console console = expectConsole();

        char[] res = console.readPassword( fmt, args );
        state.isFalse( res == null, "Password read returned null" );

        return res;
    }
    
    // Utility method to get an instance of the given MessageDigest from the
    // default provider, rethrowing any exceptions as runtime exception.
    public
    static
    MessageDigest
    getMessageDigest( String alg )
    {
        inputs.notNull( alg, "alg" );

        try { return MessageDigest.getInstance( alg ); }
        catch ( NoSuchAlgorithmException nsae )
        {
            throw createRethrow( "Couldn't get digest of type " + alg, nsae );
        }
    }

    public
    static
    MessageDigester
    createDigester( String alg )
    {
        return new MessageDigester( getMessageDigest( alg ) );
    }

    // Utility method which computes and returns a digest of the given src
    // buffer without changing the src buffers position, mark, limit, etc. See
    // getDigest() for note about exceptions encountered getting the
    // MessageDigest itself.
    public
    static
    ByteBuffer
    getDigest( ByteBuffer src,
               String alg )
    {
        inputs.notNull( src, "src" );
        inputs.notNull( alg, "alg" );

        MessageDigester dig = createDigester( alg );

        dig.update( src.slice() );
        return dig.digest();
    }

    public
    static
    ByteBuffer
    getMd5( ByteBuffer src )
    {
        return getDigest( src, DIG_ALG_MD5 );
    }

    public
    static
    SecretKey
    asSecretKey( CharSequence asciiKey,
                 String alg )
    {
        inputs.notNull( asciiKey, "asciiKey" );
        inputs.notNull( alg, "alg" );

        ByteBuffer bb = Charsets.US_ASCII.asByteBufferUnchecked( asciiKey );
        byte[] bytes = IoUtils.toByteArray( bb );
        
        return new SecretKeySpec( bytes, alg );
    }

    public
    static
    Mac
    expectMac( Key k,
               String alg )
    {
        inputs.notNull( k, "k" );
        inputs.notNull( alg, "alg" );

        try
        {
            Mac res = Mac.getInstance( alg );
            res.init( k );
        
            return res;
        }
        catch ( InvalidKeyException ike )
        {
            throw createRethrow( "Invalid key", ike );
        }
        catch ( NoSuchAlgorithmException nsae )
        {
            String msg = "Couldn't get mac for alg: " + alg;
            throw createRethrow( msg, nsae );
        }
    }

    public
    static
    ByteBuffer
    sign( ByteBuffer toSign,
          Mac mac )
    {
        inputs.notNull( toSign, "toSign" );
        inputs.notNull( mac, "mac" );

        mac.reset();
        mac.update( toSign );

        byte[] res = mac.doFinal();
        mac.reset();

        return ByteBuffer.wrap( res );
    }

    public
    static
    String
    getAlgorithm( String trans )
    {
        inputs.notNull( trans, "trans" );

        int indx = trans.indexOf( '/' );
        if ( indx < 0 ) indx = trans.length();

        if ( indx == 0 )
        {
            throw inputs.createFail(
                "Cannot determine algorithm name in transformation:", trans );
        }
        else return trans.substring( 0, indx );
    }

    public
    static
    int
    blockLengthOf( String trans )
    {
        inputs.notNull( trans, "trans" );

        String alg = getAlgorithm( trans );

        if ( alg.equalsIgnoreCase( "AES" ) ) return 16;
        else if ( alg.equalsIgnoreCase( "DESede" ) ) return 8;
        else if ( alg.equalsIgnoreCase( "Blowfish" ) ) return 8;
        else if ( alg.equalsIgnoreCase( "RC4" ) ) return 0;
        else if ( alg.equalsIgnoreCase( "ARCFOUR" ) ) return 0;
        else return -1;
    }

    public
    static
    int
    expectBlockLengthOf( String trans )
    {
        int res = blockLengthOf( trans );
        
        if ( res >= 0 ) return res;

        throw state.createFail( "Unrecognized transformation:", trans );
    }

    public
    static
    int
    ivLengthOf( String trans )
    {
        inputs.notNull( trans, "trans" );

        int res = blockLengthOf( trans );

        if ( res <= 0 ) return res;

        String[] parts = trans.split( "/" );
        if ( parts.length <= 1 ) return 0;

        String mode = parts[ 1 ];
        if ( mode == null || mode.equalsIgnoreCase( "ECB" ) ) return 0;

        return res;
    }

    public
    static
    KeyGenerator
    createKeyGenerator( String alg,
                        int keyLen )
        throws GeneralSecurityException
    {
        inputs.notNull( alg, "alg" );
        inputs.positiveI( keyLen, "keyLen" );

        KeyGenerator res = KeyGenerator.getInstance( alg );
        res.init( keyLen );

        return res;
    }

    public
    static
    IvParameterSpec
    createRandomIvSpec( int blockLen )
    {
        inputs.positiveI( blockLen, "blockLen" );

        byte[] arr = new byte[ blockLen ];
        secureRandom().nextBytes( arr );

        return new IvParameterSpec( arr );
    }

    public
    static
    Cipher
    createCipher( String trans )
        throws GeneralSecurityException
    {
        inputs.notNull( trans, "trans" );

        return Cipher.getInstance( trans );
    }

    public
    static
    Cipher
    initCipher( int opmode,
                Cipher c,
                Key k )
        throws GeneralSecurityException
    {
        inputs.notNull( c, "c" );
        inputs.notNull( k, "k" );

        c.init( opmode, k );
        return c;
    }

    public
    static
    Cipher
    initCipher( int opmode,
                Cipher c,
                Key k,
                AlgorithmParameterSpec aps )
        throws GeneralSecurityException
    {
        inputs.notNull( c, "c" );
        inputs.notNull( k, "k" );
        inputs.notNull( aps, "aps" );

        c.init( opmode, k, aps );
        return c;
    }

    public
    static
    Cipher
    initEncrypt( Cipher c,
                 Key k,
                 AlgorithmParameterSpec aps )
        throws GeneralSecurityException
    {
        return initCipher( Cipher.ENCRYPT_MODE, c, k, aps );
    }

    public
    static
    Cipher
    initEncrypt( Cipher c,
                 Key k )
        throws GeneralSecurityException
    {
        return initCipher( Cipher.ENCRYPT_MODE, c, k );
    }

    public
    static
    Cipher
    initDecrypt( Cipher c,
                 Key k,
                 AlgorithmParameterSpec aps )
        throws GeneralSecurityException
    {
        return initCipher( Cipher.DECRYPT_MODE, c, k, aps );
    }

    public
    static
    Cipher
    initDecrypt( Cipher c,
                 Key k )
        throws GeneralSecurityException
    {
        return initCipher( Cipher.DECRYPT_MODE, c, k );
    }

    public
    static
    ByteBuffer
    doFinal( Cipher c,
             ByteBuffer input )
        throws Exception
    {
        inputs.notNull( c, "c" );
        inputs.notNull( input, "input" );

        int outLen = c.getOutputSize( input.remaining() );
        ByteBuffer res = ByteBuffer.allocate( outLen );

        c.doFinal( input, res );
        res.flip();

        return res;
    }

    private
    static
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
    static
    SecretKey
    getSecretKey( KeyStore ks,
                  String id,
                  KeyStore.ProtectionParameter pp )
        throws GeneralSecurityException
    {
        KeyStore.Entry e = ks.getEntry( id, pp );

        if ( ! ( e instanceof KeyStore.SecretKeyEntry ) )
        {
            throw inputs.createFail( "Not a secret key:", id );
        }
        else return ( (KeyStore.SecretKeyEntry) e ).getSecretKey();
    }

    // This is very simplistic at the moment and assumes a fixed way to match
    // aliases (regex), that all keys are protected with the given password, and
    // that the type of key to be extracted is SecretKey.  We can generalize
    // these assumptions going forward as needed, rewriting this method on top
    // of the more generalized versions
    public
    static
    Map< String, SecretKey >
    getSecretKeys( KeyStore ks,
                   char[] pass,
                   String selectPat )
        throws GeneralSecurityException
    {
        inputs.notNull( ks, "ks" );
        // pass could be null
        inputs.notNull( selectPat, "selectPat" );

        KeyStore.ProtectionParameter pp = 
            new KeyStore.PasswordProtection( pass );

        Map< String, SecretKey > res = Lang.newMap();

        for ( String id : getMatchingAliases( ks, selectPat ) )
        {
            res.put( id, getSecretKey( ks, id, pp ) );
        }

        return res;
    } 
}
