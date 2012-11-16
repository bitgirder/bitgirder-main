package com.bitgirder.crypto;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.io.IoUtils;
import com.bitgirder.io.IoTests;
import com.bitgirder.io.IoTestFactory;
import com.bitgirder.io.Charsets;

import com.bitgirder.test.Test;
import com.bitgirder.test.TestCall;

import java.util.List;
import java.util.Map;

import javax.crypto.SecretKey;

import javax.crypto.spec.IvParameterSpec;

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
    KeyedBlobManager
    createManager( Map< String, SecretKey > keys,
                   CipherTestContext ctx )
    {
        return
            new KeyedBlobManager.Builder().
                setTransformation( ctx.transformation() ).
                setKeys( keys ).
                build();
    }

    // We do the encrypt with both keys and the decrypt with only the
    // lexicographically greatest in order to assert that the encrypt chooses
    // the lex largest by default
    private
    void
    runBasicRoundtrip( CipherTestContext ctx )
        throws Exception
    {
        Map< String, SecretKey > keys = Lang.newMap();
        keys.put( "key1", CipherTests.generateKey( ctx ) );
        keys.put( "key2", CipherTests.generateKey( ctx ) );

        KeyedBlobManager km = createManager( keys, ctx );
 
        byte[] plain = Charsets.UTF_8.asByteArray( "hello" );
        byte[] blob = km.encrypt( plain );

        state.remove( keys, "key1", "keys" );
        km = createManager( keys, ctx );

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

//    @Test( expected = KeyedBlobManager.InvalidKeyException.class,
//           expectedPattern = "\\QInvalid key id: bad\\E" )
//    private
//    final
//    class UnrecognizedKeyIdFailsTest
//    extends TestImpl
//    {
//        void
//        implCall()
//            throws Exception
//        {
//            km.decrypt( "key=bad,data=" );
//        }
//    }

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

//    private
//    abstract
//    class TamperTest
//    extends TestImpl
//    {
//        private final String toAlter;
//
//        private TamperTest( String toAlter ) { this.toAlter = toAlter; }
//    
//        private
//        String
//        alterBlob( String blob,
//                   String toAlter )
//        {
//            // Our search string relies on the fact that the key starting with
//            // toAlter is preceded by a comma, which we assert here as a sanity
//            // check
//            state.isTrue( 
//                blob.startsWith( "key=" ) && ( ! toAlter.equals( "key" ) ) );
//    
//            int idx = blob.indexOf( "," + toAlter + "=" );
//            int chIdx = idx + toAlter.length() + 2;
//            char ch = blob.charAt( chIdx );
//            code( "blob:", blob, "; ch:", ch );
//            char[] chars = blob.toCharArray();
//            chars[ chIdx ] = ch == 'x' ? 'y' : 'x'; // change the char
//            blob = new String( chars );
//            code( "blob:", blob );
//    
//            return blob;
//        }
//    
//        void
//        implCall()
//            throws Exception
//        {
//            byte[] plain = Charsets.UTF_8.asByteArray( "hello" );
//            String blob = km.encrypt( plain );
//            blob = alterBlob( blob, toAlter );
//
//            km.decrypt( blob );
//        }
//    }

    @Test( expected = KeyedBlobManager.CorruptBlobException.class )
    private
    final
    class TamperedIvDetectedTest
    extends TestImpl
    {
        void
        implCall()
            throws Exception
        {
            byte[] blob = km.encrypt( new byte[] { 0 } );
            KeyedBlobManager.Message msg = km.readMessage( blob );

            byte[] iv = msg.iv.getIV();
            iv[ 0 ]++;
            msg.iv = new IvParameterSpec( iv );

            km.decrypt( km.makeBlob( msg ) );
        }
    }

    @Test( expected = KeyedBlobManager.CorruptBlobException.class )
    private
    final
    class TamperedDataChangedDetectedTest
    extends TestImpl
    {
        void
        implCall()
            throws Exception
        {
            KeyedBlobManager.Message msg =
                km.readMessage( km.encrypt( new byte[] { 0 } ) );
            
            msg.data[ 0 ]++;

            km.decrypt( km.makeBlob( msg ) );
        }
    }

    @Test( expected = KeyedBlobManager.CorruptBlobException.class )
    private
    final
    class TamperedDataDataTruncatedTest
    extends TestImpl
    {
        void
        implCall()
            throws Exception
        {
            KeyedBlobManager.Message msg =
                km.readMessage( 
                    km.encrypt( IoTestFactory.nextByteArray( 100 ) ) );
            
            byte[] arr = new byte[ msg.data.length - 1 ];
            System.arraycopy( msg.data, 0, arr, 0, arr.length );
            msg.data = arr;

            km.decrypt( km.makeBlob( msg ) );
        }
    }
}
