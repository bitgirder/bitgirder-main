package com.bitgirder.crypto;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.io.IoUtils;
import com.bitgirder.io.IoTests;
import com.bitgirder.io.Charsets;

import com.bitgirder.test.Test;
import com.bitgirder.test.TestCall;

import java.util.List;

import java.security.KeyStore;

@Test
final
class KeyedBlobManagerTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static CipherTestContext CTX_DEFAULT =
        CipherTestContext.create( "AES/CBC/PKCS5Padding", 192 );

    private
    void
    runBasicRoundtrip( CipherTestContext ctx )
        throws Exception
    {
        KeyedBlobManager km =
            new KeyedBlobManager.Builder().
                setTransformation( ctx.transformation() ).
                setKey( "key1", CipherTests.generateKey( ctx ) ).
                setKey( "key2", CipherTests.generateKey( ctx ) ).
                build();
 
        byte[] plain = Charsets.UTF_8.asByteArray( "hello" );
        String blob = km.encrypt( plain );

        // check that key2 was selected by default
        state.isFalse( blob.indexOf( "key=key2" ) < 0 );

        byte[] plain2 = km.decrypt( blob );
        IoTests.assertEqual( plain, plain2 );
    }

    @Test
    private
    void
    testBasicRoundtrip()
        throws Exception
    {
        List< CipherTestContext > l = CipherTests.getBaseTestContexts();
        for ( CipherTestContext ctx : l ) runBasicRoundtrip( ctx );
    }

    private
    abstract
    class TestImpl
    implements TestCall
    {
        KeyedBlobManager km;

        abstract
        void
        implCall()
            throws Exception;

        public
        final
        void
        call()
            throws Exception
        {
            km =
                new KeyedBlobManager.Builder().
                    setTransformation( CTX_DEFAULT.transformation() ).
                    setKey( "key1", CipherTests.generateKey( CTX_DEFAULT ) ).
                    build();
            
            implCall();
        }
    }

    @Test( expected = KeyedBlobManager.InvalidKeyException.class,
           expectedPattern = "\\QInvalid key id: bad\\E" )
    private
    final
    class UnrecognizedKeyIdFailsTest
    extends TestImpl
    {
        void
        implCall()
            throws Exception
        {
            km.decrypt( "key=bad,data=" );
        }
    }

    private
    KeyStore
    newKeyStore( char[] pass )
        throws Exception
    {
        KeyStore res = 
            KeyStore.getInstance( CryptoConstants.KEY_STORE_TYPE_JCEKS );

        res.load( null, pass );

        return res;
    }

    private
    void
    addKeys( KeyStore ks,
             char[] pass,
             int keyCount,
             CipherTestContext ctx,
             String tmpl )
        throws Exception
    {
        for ( int i = 0; i < keyCount; ++i )
        {
            ks.setEntry( 
                String.format( tmpl, i ),
                new KeyStore.SecretKeyEntry( CipherTests.generateKey( ctx ) ),
                new KeyStore.PasswordProtection( pass )
            );
        }
    }

    @Test
    private
    void
    testBuildFromKeyStore()
        throws Exception
    {
        char[] pass = "test".toCharArray();
        KeyStore ks = newKeyStore( pass );

        addKeys( ks, pass, 2, CTX_DEFAULT, "accept-%016x" );
        addKeys( ks, pass, 2, CTX_DEFAULT, "reject-%016x" );

        KeyedBlobManager.Builder b = new KeyedBlobManager.Builder().
            setTransformation( CTX_DEFAULT.transformation() ).
            setKeysFrom( ks, pass, "^accept-.*" );

        state.equalString(
            "accept-0000000000000000|accept-0000000000000001",
            Strings.join( "|", b.keys.keySet() )
        );
 
        KeyedBlobManager km = b.build();
        byte[] plain = Charsets.UTF_8.asByteArray( "hello" );
        String blob = km.encrypt( plain );
        state.isFalse( blob.indexOf( "key=accept-0000000000000001" ) < 0 );
        IoTests.assertEqual( plain, km.decrypt( blob ) );
    }

    @Test( expected = IllegalArgumentException.class,
           expectedPattern = "\\QNo keys set\\E" )
    private
    void
    testBuildFailsWithNoKeys()
    {
        new KeyedBlobManager.Builder().
            setTransformation( CTX_DEFAULT.transformation() ).
            build();
    }

    private
    abstract
    class TamperTest
    extends TestImpl
    {
        private final String toAlter;

        private TamperTest( String toAlter ) { this.toAlter = toAlter; }
    
        private
        String
        alterBlob( String blob,
                   String toAlter )
        {
            // Our search string relies on the fact that the key starting with
            // toAlter is preceded by a comma, which we assert here as a sanity
            // check
            state.isTrue( 
                blob.startsWith( "key=" ) && ( ! toAlter.equals( "key" ) ) );
    
            int idx = blob.indexOf( "," + toAlter + "=" );
            int chIdx = idx + toAlter.length() + 2;
            char ch = blob.charAt( chIdx );
            code( "blob:", blob, "; ch:", ch );
            char[] chars = blob.toCharArray();
            chars[ chIdx ] = ch == 'x' ? 'y' : 'x'; // change the char
            blob = new String( chars );
            code( "blob:", blob );
    
            return blob;
        }
    
        void
        implCall()
            throws Exception
        {
            byte[] plain = Charsets.UTF_8.asByteArray( "hello" );
            String blob = km.encrypt( plain );
            blob = alterBlob( blob, toAlter );

            km.decrypt( blob );
        }
    }

    @Test( expected = KeyedBlobManager.CorruptBlobException.class )
    private
    final
    class TamperedIvDetectedTest
    extends TamperTest
    {
        private TamperedIvDetectedTest() { super( "iv" ); }
    }

    @Test( expected = KeyedBlobManager.CorruptBlobException.class )
    private
    final
    class TamperedDataDetectedTest
    extends TamperTest
    {
        private TamperedDataDetectedTest() { super( "data" ); }
    }

    // To test:
    //
    //  - exception when decrypt called with corrupted cipher data (iv or
    //  ciphertext)
}
