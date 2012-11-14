package com.bitgirder.crypto;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.io.IoTestFactory;

import com.bitgirder.test.Test;
import com.bitgirder.test.LabeledTestCall;
import com.bitgirder.test.InvocationFactory;

import java.util.List;

import java.nio.ByteBuffer;

import javax.crypto.KeyGenerator;
import javax.crypto.SecretKey;
import javax.crypto.Cipher;
import javax.crypto.IllegalBlockSizeException;
import javax.crypto.BadPaddingException;

import javax.crypto.spec.IvParameterSpec;

@Test
public
final
class CipherTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private
    static
    MessageDigester
    md5Digester()
    {
        return CryptoUtils.createDigester( "md5" );
    }

    static
    SecretKey
    generateKey( CipherTestContext ctx )
        throws Exception
    {
        state.notNull( ctx, "ctx" );

        KeyGenerator kg = 
            CryptoUtils.createKeyGenerator( 
                CryptoUtils.getAlgorithm( ctx.transformation() ), 
                ctx.keyLen() 
            );
 
        return kg.generateKey();
    }

    private
    boolean
    isStreamCipher( CipherTestContext ctx )
    {
        String alg = CryptoUtils.getAlgorithm( ctx.transformation() );

        return 
            alg.equalsIgnoreCase( "RC4" ) || alg.equalsIgnoreCase( "ARCFOUR" );
    }

    private
    static
    int
    getBlockLen( CipherTestContext ctx )
    {
        String alg = CryptoUtils.getAlgorithm( ctx.transformation() );

        if ( alg.equalsIgnoreCase( "AES" ) ) return 16;
        else if ( alg.equalsIgnoreCase( "DESede" ) ) return 8;
        else if ( alg.equalsIgnoreCase( "Blowfish" ) ) return 8;
        else if ( alg.equalsIgnoreCase( "RC4" ) ) return 0;
        else if ( alg.equalsIgnoreCase( "ARCFOUR" ) ) return 0;
        else throw state.createFail( "Unhandled alg:", alg );
    }

    private
    static
    boolean
    needsIv( CipherTestContext ctx )
    {
        if ( getBlockLen( ctx ) == 0 ) return false;
        else
        {
            String[] parts = ctx.transformation().split( "/" );
            String mode = parts.length > 1 ? parts[ 1 ] : null;

            return mode != null && ( ! mode.equalsIgnoreCase( "ECB" ) );
        }
    }

    private
    static
    IvParameterSpec
    nextIv( CipherTestContext ctx )
    {
        return CryptoUtils.createRandomIvSpec( getBlockLen( ctx ) );
    }

    public
    static
    CipherFactory
    createCipherFactory( CipherTestContext ctx )
        throws Exception
    {
        inputs.notNull( ctx, "ctx" );

        CipherFactory.Builder b = 
            new CipherFactory.Builder().
                setTransformation( ctx.transformation() ).
                setKey( generateKey( ctx ) );
        
        if ( needsIv( ctx ) ) b.setIv( nextIv( ctx ) );

        return b.build();
    }

    public
    static
    CipherFactory
    createCipherFactory( String transformation,
                         int keyLen )
        throws Exception
    {
        inputs.notNull( transformation, "transformation" );

        return 
            createCipherFactory( 
                CipherTestContext.create( transformation, keyLen ) );
    }

    // Used in other testing to get a list of cipher test contexts which we
    // consider representative of how files might be encrypted, in order to
    // allow other packages to get crypto test coverage without being concerned
    // with which ciphers to test
    public
    static
    List< CipherTestContext >
    getFileCipherContexts()
    {
        return
            Lang.asList(
                CipherTestContext.create( "AES/CBC/PKCS5Padding", 192 )
            );
    }

    private
    final
    class BufferCipherTest
    extends LabeledTestCall 
    {
        private final CipherTestContext ctx;

        private
        BufferCipherTest( CipherTestContext ctx )
        {
            super( ctx.getLabel() );

            this.ctx = ctx;
        }

        private
        void
        doRoundtrip( int len,
                     CipherFactory cf )
            throws Exception
        {
            ByteBuffer plainText = IoTestFactory.nextByteBuffer( len );

            ByteBuffer cipherText = 
                CryptoUtils.doFinal( cf.initEncrypt(), plainText.slice() );

            // maybe seems silly but we want to prevent some horrific bug in our
            // implementation such that we aren't actually doing any crypto at
            // all for some reason, which would lead to plainText == cipherText
            // == plainText2
            state.isFalse( plainText.equals( cipherText ) );

            ByteBuffer plainText2 = 
                CryptoUtils.doFinal( cf.initDecrypt(), cipherText );

            state.equal( plainText, plainText2 );
        }

        public
        void
        call()
            throws Exception
        {
            CipherFactory cf = createCipherFactory( ctx );

            int bl = getBlockLen( ctx );

            if ( bl == 0 ) doRoundtrip( 1024, cf );
            else
            {
                // cover all boundary conditions
                for ( int i = 0; i < bl; ++i ) doRoundtrip( 1024 + i, cf );
            }
        }
    }

    private
    static
    void
    addBlockContexts( List< CipherTestContext > res,
                      String alg,
                      int[] keyLens )
    {
        for ( int keyLen : keyLens )
        {
            // Sun JDK provider doesn't seem to support CFB1
            for ( String mode : 
                    new String[] { "CBC", "CFB", "CFB8", "OFB", "ECB" } )
            {
                for ( String padding : new String[] { "PKCS5Padding" } )
                {
                    res.add( 
                        CipherTestContext.create(
                            alg + "/" + mode + "/" + padding, keyLen ) );
                }
            }
        }
    }

    private
    static
    void
    addStreamContexts( List< CipherTestContext > res,
                       String alg,
                       int[] keyLens )
    {
        for ( int keyLen : keyLens )
        {
            res.add( CipherTestContext.create( alg, keyLen ) );
        }
    }

    static
    List< CipherTestContext >
    getBaseTestContexts()
    {
        List< CipherTestContext > res = Lang.newList();

        addBlockContexts( res, "AES", new int[] { 128, 192, 256 } );
        addBlockContexts( res, "DESede", new int[] { 168 } );
        addBlockContexts( res, "Blowfish", new int[] { 32, 128, 256, 448 } );
        addStreamContexts( res, "RC4", new int[] { 40, 256, 1024 } );
        addStreamContexts( res, "ARCFOUR", new int[] { 40, 256, 1024 } );

        return res;
    }

    @InvocationFactory
    private
    List< BufferCipherTest >
    testBufferCipher()
    {
        List< BufferCipherTest > res = Lang.newList();

        for ( CipherTestContext ctx : getBaseTestContexts() )
        {
            res.add( new BufferCipherTest( ctx ) );
        }

        return res;
    }
}
