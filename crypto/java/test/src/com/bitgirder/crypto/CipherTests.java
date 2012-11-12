package com.bitgirder.crypto;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;
import com.bitgirder.lang.StandardThread;
import com.bitgirder.lang.ObjectReceiver;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.io.IoTestFactory;
import com.bitgirder.io.IoTestSupport;
import com.bitgirder.io.IoTests;
import com.bitgirder.io.IoProcessor;
import com.bitgirder.io.IoUtils;
import com.bitgirder.io.FileWrapper;
import com.bitgirder.io.FileSend;
import com.bitgirder.io.FileFill;
import com.bitgirder.io.ProtocolRoundtripTest;
import com.bitgirder.io.ProtocolProcessor;
import com.bitgirder.io.ProtocolProcessors;
import com.bitgirder.io.ProtocolProcessorTests;
import com.bitgirder.io.DigestProcessor;
import com.bitgirder.io.DataSize;
import com.bitgirder.io.Crc32Digest;
import com.bitgirder.io.FileStorage;
import com.bitgirder.io.ProtocolTestStorage;
import com.bitgirder.io.BufferAccumulatorStorage;
import com.bitgirder.io.ByteBufferAccumulator;
import com.bitgirder.io.DegenerateStorage;
import com.bitgirder.io.GzipReader;
import com.bitgirder.io.GzipWriter;
import com.bitgirder.io.ProtocolCopies;

import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.ProcessActivity;

import com.bitgirder.test.Test;
import com.bitgirder.test.TestRuntime;
import com.bitgirder.test.TestFailureExpector;
import com.bitgirder.test.LabeledTestCall;
import com.bitgirder.test.LabeledTestObject;
import com.bitgirder.test.InvocationFactory;

import java.util.List;
import java.util.Map;
import java.util.Iterator;

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

    private final static CipherTestContext AES_256_CBC_PKCS5;

    private final TestRuntime rt;

    private CipherTests( TestRuntime rt ) { this.rt = rt; }

    private
    static
    MessageDigester
    md5Digester()
    {
        return CryptoUtils.createDigester( "md5" );
    }

    private
    static
    SecretKey
    generateKey( CipherTestContext ctx )
        throws Exception
    {
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
    FileWrapper
    nextFile()
        throws Exception
    {
        return IoUtils.createTempFile( "cipher-tests-", true );
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

        protected
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

    private
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

    private
    final
    static
    class ImplChecker
    implements Ciphers.AllocationEventHandler
    {
        private final int maxAllocs;
        private List< Integer > allocs = Lang.newList();

        private
        ImplChecker( int maxAllocs )
        {
            this.maxAllocs = maxAllocs;
        }

        // assert that we only allocate one buf during the life of the
        // processor
        public
        void
        allocatingBuffer( int len )
        {
            state.isTrue( 
                allocs.size() < maxAllocs, 
                "Attempt to alloc buffer of length", len, 
                "but max allocs already made:", allocs );

            allocs.add( len );
        }

        private void assertState() { state.isFalse( allocs.isEmpty() ); }
    }

    private
    final
    class CipherStreamRoundtripTest
    extends ProtocolRoundtripTest
    implements LabeledTestObject
    {
        private final CipherTestContext ctx;

        private final int streamLen;
        private final int xferSize;

        private final ImplChecker senderImplChecker = new ImplChecker( 1 );
        private final ImplChecker recvImplChecker = new ImplChecker( 1 );

        private DigestProcessor< Long > plainSender;
        private DigestProcessor< Long > plainRecv;
        private DigestProcessor< Long > cipherCrc32Recv;

        private
        CipherStreamRoundtripTest( CipherTestContext ctx,
                                   int streamLen,
                                   int xferSize )
        {
            this.ctx = ctx;
            this.streamLen = streamLen;
            this.xferSize = xferSize;
        }

        public
        CharSequence
        getLabel()
        {
            return Strings.crossJoin( "=", ",",
                "ctx", "{" + ctx.getLabel() + "}",
                "streamLen", streamLen,
                "xferSize", xferSize
            );
        }

        protected
        void
        beginAssert()
        {
//            code(
//                "plainSrc:", plainSender.digest(),
//                "; plainRecv:", plainRecv.digest(),
//                "; cipherCrc32:", cipherCrc32Recv.digest()
//            );

            recvImplChecker.assertState();
            senderImplChecker.assertState();

            long plainCrc32Sent = plainSender.digest();
            long plainCrc32Recv = plainRecv.digest();
            long cipherCrc32 = cipherCrc32Recv.digest();

            state.equal( plainCrc32Sent, plainCrc32Recv );

            if ( streamLen == 0 && isStreamCipher( ctx ) )
            {
                state.equal( 0L, plainCrc32Sent );
                state.equal( 0L, cipherCrc32 );
            }
            else state.isFalse( plainCrc32Sent == cipherCrc32 );

            assertDone();
        }

        private
        void
        initSender( CipherFactory cf )
        {
            plainSender =
                DigestProcessor.create(
                    new ProtocolProcessorTests.Generator( streamLen ),
                    Crc32Digest.create()
                );
            
            setSender(
                new CipherStreamSender.Builder().
                    setCipherFactory( cf ).
                    setProcessor( plainSender ).
                    setAllocationEventHandler( senderImplChecker ).
                    build()
            );
        }

        private
        void
        initReceiver( CipherFactory cf )
        {
            plainRecv = IoTests.crc32Digester();
            
            CipherStreamReceiver cipherRecv =
                new CipherStreamReceiver.Builder().
                    setCipherFactory( cf ).
                    setProcessor( plainRecv ).
                    setAllocationEventHandler( recvImplChecker ).
                    build();
 
            cipherCrc32Recv =
                DigestProcessor.create( cipherRecv, Crc32Digest.create() );
 
            setReceiver( cipherCrc32Recv );
        }

        protected
        void
        startTest()
            throws Exception
        {
            CipherFactory cf = createCipherFactory( ctx );

            initSender( cf );
            initReceiver( cf );
            setXferSize( xferSize );

            testReady();
        }
    }

    private
    void
    addStandardRoundtrips( List< CipherStreamRoundtripTest > res,
                           CipherTestContext ctx )
    {
        res.addAll( 
            Lang.asList(
                new CipherStreamRoundtripTest( ctx, 0, 1024 ),
                new CipherStreamRoundtripTest( ctx, 1023, 1024 ),
                new CipherStreamRoundtripTest( ctx, 1024, 1024 ),
                new CipherStreamRoundtripTest( ctx, 10240, 1024 ),
                new CipherStreamRoundtripTest( ctx, 10241, 1024 )
            )
        );
    }

    @InvocationFactory
    private
    List< CipherStreamRoundtripTest >
    testCipherStreamRoundtrip()
    {
        List< CipherStreamRoundtripTest > res = Lang.newList();

        for ( CipherTestContext ctx : getBaseTestContexts() )
        {
            addStandardRoundtrips( res, ctx );
        }

        return res;
    }

    // Really we're testing 4 separate operations but doing so with a somewhat
    // overlapping pipeline:
    //
    // We test:
    //  - sending a plaintext stream to a ciphertext file
    //  - receiving a ciphertext file as a plaintext stream
    //  - sending a ciphertext stream to a plaintext file
    //  - receiveing a plaintext file as a ciphertext stream
    //
    // To assert all of this we:
    //  
    //  1. Generate a plaintext stream and save it to ciphertext file cf1,
    //  noting the crc32 of the plaintext and of the ciphertext file, verifying
    //  that actual encryption did take place
    //
    //  2. Read the ciphertext as plaintext and verify that the received
    //  plaintext crc32 is as expected
    //
    //  3. Read cf1 as a ciphertext stream and write it to the plaintext file
    //  pt1, noting the crc32 of pt1 and assert that it is the expected
    //  plaintext crc32
    //
    //  4. Read pt1 as a ciphertext stream and assert that the stream's crc32 is
    //  the expected ciphertext crc32
    //
    private
    final
    class CipherFileCopyTest
    extends AbstractVoidProcess
    implements LabeledTestObject
    {
        private final CipherTestContext ctx;
        private final DataSize streamLen = DataSize.ofMegabytes( 1 );
        private final int ioBufSz = 10240;

        private CipherFactory cf;

        private 
        CipherFileCopyTest( CipherTestContext ctx ) 
        { 
            super( IoTestSupport.create( rt ) );

            this.ctx = ctx; 
        }

        public Object getInvocationTarget() { return this; }

        public CharSequence getLabel() { return ctx.getLabel(); }

        private
        < B extends CipherFileCopy.Builder< ?, ? > >
        B
        initBuilder( B b )
        {
            b.setCipherFactory( cf );
            b.setActivityContext( getActivityContext() );
            b.setIoProcessor( behavior( IoTestSupport.class ).ioProcessor() );
            b.setIoBufferSize( ioBufSz );

            return b;
        }

        private
        void
        readAndEncryptPt1( FileWrapper pt1,
                           final long cipherCrc32 )
        {
            final DigestProcessor< Long > proc = IoTests.crc32Digester();

            initBuilder( CipherFileFeed.encryptBuilder() ).
                setFile( pt1 ).
                setProcessor( proc ).
                setEventHandler(
                    new CipherFileFeed.AbstractEventHandler( self() ) 
                    {
                        @Override protected void fileCopyCompleteImpl()
                        {
                            long crc32 = proc.digest();
                            state.equal( cipherCrc32, crc32 );
                            exit();
                        }
                    }
                ).
                build().
                start();
        }

        private
        void
        verifyPt1Write( final FileWrapper pt1,
                        final long plainCrc32,
                        final long cipherCrc32 )
        {
            behavior( IoTestSupport.class ).
                getDigest( 
                    new IoTestSupport.Crc32FileDigest( pt1 ) 
                    {
                        protected void digestReady( Long crc32 )
                        {
                            state.equal( plainCrc32, crc32.longValue() );
                            readAndEncryptPt1( pt1, cipherCrc32 );
                        }
                    }
                );
        }

        private
        void
        writePt1( FileWrapper cf1,
                  final long plainCrc32,
                  final long cipherCrc32 )
            throws Exception
        {
            final FileWrapper pt1 = nextFile();

            IoProcessor.Client ioCli = 
                behavior( IoTestSupport.class ).ioClient();

            initBuilder( CipherFileFill.decryptBuilder() ).
                setFile( pt1 ).
                setProcessor( FileSend.create( cf1, ioCli ) ).
                setEventHandler(
                    new CipherFileFill.AbstractEventHandler( self() ) {
                        @Override protected void fileCopyCompleteImpl() {
                            verifyPt1Write( pt1, plainCrc32, cipherCrc32 );
                        }
                    }
                ).
                build().
                start();
        }

        private
        void
        verifyCf1Decrypt( FileWrapper cf1,
                          long plainCrc32Expct,
                          long cipherCrc32,
                          long plainCrc32Actual )
            throws Exception
        {
            state.equal( plainCrc32Expct, plainCrc32Actual );
            writePt1( cf1, plainCrc32Expct, cipherCrc32 );
        }

        private
        void
        readAsPlaintext( final FileWrapper cf1,
                         final long plainCrc32,
                         final long cipherCrc32 )
            throws Exception
        {
            final DigestProcessor< Long > proc = IoTests.crc32Digester();

            initBuilder( CipherFileFeed.decryptBuilder() ).
                setFile( cf1 ).
                setProcessor( proc ).
                setEventHandler(
                    new CipherFileFeed.AbstractEventHandler( self() ) 
                    {
                        @Override 
                        protected void fileCopyCompleteImpl() throws Exception
                        {
                            long crc32 = proc.digest();
                            verifyCf1Decrypt( 
                                cf1, plainCrc32, cipherCrc32, crc32 );
                        }
                    }
                ).
                build().
                start();
        }

        private
        void
        cf1Written( final FileWrapper cf1,
                    final long plainCrc32 )
        {
            behavior( IoTestSupport.class ).
                getDigest( 
                    new IoTestSupport.Crc32FileDigest( cf1 ) 
                    {
                        protected void digestReady( Long cipherCrc32 ) 
                            throws Exception
                        {
                            state.isFalse( 
                                plainCrc32 == cipherCrc32.longValue() );

                            readAsPlaintext( cf1, plainCrc32, cipherCrc32 );
                        }
                    }
                );
        }

        private
        void
        generateCf1()
            throws Exception
        {
            final DigestProcessor< Long > proc =
                DigestProcessor.create(
                    new ProtocolProcessorTests.Generator( streamLen ),
                    Crc32Digest.create()
                );
 
            final FileWrapper cf1 = nextFile();

            initBuilder( CipherFileFill.encryptBuilder() ).
                setFile( cf1 ).
                setProcessor( proc ).
                setEventHandler(
                    new CipherFileFill.AbstractEventHandler( self() ) {
                        @Override protected void fileCopyCompleteImpl() { 
                            cf1Written( cf1, proc.digest() );
                        }
                    }
                ).
                build().
                start();
        }

        protected
        void
        startImpl()
            throws Exception
        {
            cf = createCipherFactory( ctx );
            
            generateCf1();
        }
    }

    @InvocationFactory
    private
    List< CipherFileCopyTest >
    testCipherFileCopy()
    {
        List< CipherFileCopyTest > res = Lang.newList();

        for ( CipherTestContext ctx : getBaseTestContexts() )
        {
            res.add( new CipherFileCopyTest( ctx ) );
        }

        return res;
    }

    // This is really just meant to give code-coverage for the copy builders and
    // to ensure that we're correctly wiring up the internals of the
    // encrypt/decrypt activities. We're not currently re-running the full array
    // of CipherTestContexts that we are in testCipherFileCopy(), although we
    // could. We could also have piggybacked the tests below along inside of
    // CipherFileCopyTest, but that class is already sufficiently complicated
    // that throwing more in there doesn't seem worth the trouble
    @Test
    private
    final
    class CipherFileCopyCryptOpsTest
    extends AbstractVoidProcess
    {
        private final CipherTestContext ctx;

        private final DataSize fileLen = DataSize.ofKilobytes( 500 );
        private final DataSize ioBufSz = DataSize.ofKilobytes( 10 );

        private CipherFactory cf;

        private
        CipherFileCopyCryptOpsTest()
        {
            super( IoTestSupport.create( rt ) );

            ctx = CipherTestContext.create( "AES/CBC/PKCS5Padding", 192 );
        }

        private
        void
        checkCrc32s( Map< String, Long > crc32s )
        {
            state.equal( 
                state.get( crc32s, "pt1", "crc32s" ),
                state.get( crc32s, "pt2", "crc32s" ) );
            
            state.isFalse(
                state.get( crc32s, "pt1", "crc32s" ).equals(
                    state.get( crc32s, "ct1", "crc32s" )
                )
            );

            exit();
        }

        private
        void
        getCrc32( FileWrapper f,
                  final String alias,
                  final Map< String, Long > crc32s )
        {
            behavior( IoTestSupport.class ).
                getDigest(
                    new IoTestSupport.Crc32FileDigest( f ) {
                        protected void digestReady( Long dig )
                        {
                            Lang.putUnique( crc32s, alias, dig );
                            if ( crc32s.size() == 3 ) checkCrc32s( crc32s );
                        }
                    }
                );
        }

        private
        void
        assertCrypto( FileWrapper ct1,
                      FileWrapper pt2,
                      Map< String, Long > crc32s )
        {
            getCrc32( ct1, "ct1", crc32s );
            getCrc32( pt2, "pt2", crc32s );
        }

        private
        CipherFileCopy.CryptFileOpBuilder
        initBuilder( CipherFileCopy.CryptFileOpBuilder b )
        {
            b.setCipherFactory( cf ).
              setIoProcessor( behavior( IoTestSupport.class ).ioProcessor() ).
              setActivityContext( getActivityContext() );
            
            return b;
        }

        private
        void
        decryptFile( final FileWrapper pt1,
                     final FileWrapper ct1,
                     final Map< String, Long > crc32s )
            throws Exception
        {
            final FileWrapper pt2 = nextFile();

            initBuilder( CipherFileCopy.fileDecryptBuilder() ).
                setSource( ct1 ).
                setDestination( pt2 ).
                setEventHandler(
                    new CipherFileCopy.AbstractEventHandler( self() ) {
                        @Override protected void fileCopyCompleteImpl() {
                            assertCrypto( ct1, pt2, crc32s );
                        }
                    }
                ).
                build().
                start();
        }

        private
        void
        encryptFile( final FileWrapper src,
                     final Map< String, Long > crc32s )
            throws Exception
        {
            final FileWrapper dest = nextFile();

            initBuilder( CipherFileCopy.fileEncryptBuilder() ).
                setSource( src ).
                setDestination( dest ).
                setEventHandler(
                    new CipherFileCopy.AbstractEventHandler( self() ) 
                    {
                        @Override 
                        protected void fileCopyCompleteImpl() throws Exception {
                            decryptFile( src, dest, crc32s );
                        }
                    }
                ).
                build().
                start();
        }

        private
        void
        writePlaintextFile()
            throws Exception
        {
            behavior( IoTestSupport.class ).
                fill(
                    new IoTestSupport.Crc32FileFill( nextFile(), fileLen ) {
                        protected void fileFilled() throws Exception 
                        {
                            Map< String, Long > crc32s = Lang.newMap();
                            crc32s.put( "pt1", digest() );
                            encryptFile( file(), crc32s );
                        }
                    }
                );
        }

        protected
        void
        startImpl()
            throws Exception
        {
            cf = createCipherFactory( ctx );

            writePlaintextFile();
        }
    }

    @Test
    private
    final
    class OpenSslCipherInteropTest
    extends AbstractVoidProcess
    {
        private final DataSize fileLen = DataSize.ofMegabytes( 1 );

        private final CipherTestContext ctx;
        private final String algName;

        private CipherFactory cf;

        private
        OpenSslCipherInteropTest()
        {
            super( IoTestSupport.create( rt ) );

            ctx = CipherTestContext.create( "AES/CBC/PKCS5Padding", 256 );
            algName = "-aes-256-cbc";
        }

        private
        CharSequence
        getTestScript( FileWrapper ptFile,
                       FileWrapper cipherFile )
            throws Exception
        {
            CharSequence keyStr = 
                IoUtils.asHexString( cf.getKey().getEncoded() );

            IvParameterSpec iv = 
                (IvParameterSpec) cf.getAlgorithmParameterSpec();

            CharSequence ivStr = IoUtils.asHexString( iv.getIV() );

            return Strings.join( "\n",

                "openssl enc -e " + algName + " -K " + keyStr + 
                    " -iv " + ivStr + " -in " + ptFile + " | " +
                    "cmp " + cipherFile + " || exit -1",
                
                "openssl enc -d " + algName + " -K " + keyStr + 
                    " -iv " + ivStr + " -in " + cipherFile + " | " +
                    "cmp " + ptFile + " || exit -1"
            ) +
            "\n";
        }

        private
        final
        class ScriptRun
        extends StandardThread
        {
            private final CharSequence script;

            private 
            ScriptRun( CharSequence script ) 
            { 
                super( "cipher-test-script-run-%1$d" );
                
                this.script = script; 
            }

            private
            void
            runImpl()
                throws Exception
            {
                Process p = new ProcessBuilder( "/bin/bash", "-s" ).start();
                
                try
                {
                    byte[] bytes = script.toString().getBytes( "UTF-8" );
                    p.getOutputStream().write( bytes );
                    p.getOutputStream().close();

                    state.equalInt( 0, p.waitFor() );
                }
                finally
                {
                    try { p.getInputStream().close(); }
                    finally { p.getErrorStream().close(); }
                }
            }

            public
            void
            run()
            {
                try 
                { 
                    runImpl(); 
                    exit();
                } 
                catch ( Throwable th ) { fail( th ); }
            }
        }

        private
        void
        runOpenSsl( FileWrapper ptFile,
                    FileWrapper cipherFile )
            throws Exception
        {
            CharSequence script = getTestScript( ptFile, cipherFile );
            new ScriptRun( script ).run();
        }

        private
        void
        encrypt( final FileWrapper ptFile )
            throws Exception
        {
            final FileWrapper cipherFile = nextFile();

            IoProcessor.Client ioCli = 
                behavior( IoTestSupport.class ).ioClient();

            CipherFileFill.encryptBuilder().
                setActivityContext( getActivityContext() ).
                setIoProcessor( behavior( IoTestSupport.class ).ioProcessor() ).
                setCipherFactory( cf ).
                setFile( cipherFile ).
                setProcessor( FileSend.create( ptFile, ioCli ) ).
                setEventHandler(
                    new CipherFileFill.AbstractEventHandler( self() ) 
                    {
                        @Override 
                        protected void fileCopyCompleteImpl() throws Exception {
                            runOpenSsl( ptFile, cipherFile );
                        }
                    }
                ).
                build().
                start();
        }

        private
        void
        writePlaintext()
            throws Exception
        {
            final FileWrapper ptFile = nextFile();

            ProtocolProcessor< ByteBuffer > proc =
                new ProtocolProcessorTests.Generator( fileLen );

            new FileFill.Builder().
                setActivityContext( getActivityContext() ).
                setFile( ptFile ).
                setIoProcessor( behavior( IoTestSupport.class ).ioProcessor() ).
                setProcessor( proc ).
                setEventHandler(
                    new FileFill.AbstractEventHandler( self() ) 
                    {
                        @Override 
                        protected void fileCopyCompleteImpl() throws Exception {
                            encrypt( ptFile );
                        }
                    }
                ).
                build().
                start();
        }

        protected
        void
        startImpl()
            throws Exception
        {
            cf = createCipherFactory( ctx );

            writePlaintext();
        }
    }

    private
    class CipherReaderWriterTest
    extends AbstractVoidProcess
    implements LabeledTestObject,
               TestFailureExpector
    {
        private ProtocolTestStorage ts;
        private DataSize streamLen;
        private boolean useDirectBufs;
        private CipherTestContext cipherCtx;
        private boolean useGzip;
        private Integer maxInputLen;
        private int testBufSize = 8192;
        private ImplChecker writeImplChecker = new ImplChecker( 1 );
        private ImplChecker readImplChecker = new ImplChecker( 1 );
        private Class< ? extends Throwable > errCls;
        private CharSequence errPat;

        private CipherFactory cf;
        private ByteBuffer ioBuf;

        private
        CipherReaderWriterTest()
        {
            super( IoTestSupport.create( rt ) );
        }

        public Object getInvocationTarget() { return this; }

        public
        CharSequence
        getLabel()
        {
            return 
                Strings.crossJoin( "=", ",",
                    "tsCls", ts.getClass().getSimpleName(),
                    "ts", "{" + ts.getLabel() + "}",
                    "streamLen", streamLen,
                    "useDirectBufs", useDirectBufs,
                    "cipherCtx", "{" + cipherCtx.getLabel() + " }",
                    "useGzip", useGzip,
                    "maxInputLen", maxInputLen,
                    "testBufSize", testBufSize,
                    "maxWriteAllocs", writeImplChecker.maxAllocs,
                    "maxReadAllocs", readImplChecker.maxAllocs,
                    "errCls", errCls,
                    "errPat", errPat
                );
        }

        public
        Class< ? extends Throwable >
        expectedFailureClass() 
        {
            return errCls; 
        }

        public CharSequence expectedFailurePattern() { return errPat; }

        private
        CipherReaderWriterTest
        setTestStorage( ProtocolTestStorage ts )
        {
            this.ts = inputs.notNull( ts, "ts" );
            return this;
        }

        private
        CipherReaderWriterTest
        setStreamLen( DataSize streamLen )
        {
            this.streamLen = inputs.notNull( streamLen, "streamLen" );
            return this;
        }

        private
        CipherReaderWriterTest
        setUseDirectBufs( boolean useDirectBufs )
        {
            this.useDirectBufs = useDirectBufs;
            return this;
        }

        private
        CipherReaderWriterTest
        setCipherTestContext( CipherTestContext cipherCtx )
        {
            this.cipherCtx = inputs.notNull( cipherCtx, "cipherCtx" );
            return this;
        }

        private
        CipherReaderWriterTest
        setUseGzip( boolean useGzip )
        {
            this.useGzip = useGzip;
            return this;
        }

        private
        CipherReaderWriterTest
        setMaxInputLen( Integer maxInputLen )
        {
            this.maxInputLen = inputs.notNull( maxInputLen, "maxInputLen" );
            return this;
        }

        private
        CipherReaderWriterTest
        setTestBufSize( int testBufSize )
        {
            this.testBufSize = inputs.positiveI( testBufSize, "testBufSize" );
            return this;
        }

        private
        CipherReaderWriterTest
        setWriteImplChecker( ImplChecker writeImplChecker )
        {
            this.writeImplChecker = 
                inputs.notNull( writeImplChecker, "writeImplChecker" );

            return this;
        }

        private
        CipherReaderWriterTest
        setReadImplChecker( ImplChecker readImplChecker )
        {
            this.readImplChecker = 
                inputs.notNull( readImplChecker, "readImplChecker" );

            return this;
        }

        private
        CipherReaderWriterTest
        expectError( Class< ? extends Throwable > errCls,
                     CharSequence errPat )
        {
            this.errCls = inputs.notNull( errCls, "errCls" );
            this.errPat = errPat;

            return this;
        }

        private
        < B extends CipherIoFilter.Builder< B > >
        B
        init( B b )
        {
            b.setCipherFactory( cf );
            b.setActivityContext( getActivityContext() );

            if ( maxInputLen != null ) b.setMaxInputLength( maxInputLen );

            return b;
        }

        private
        ProtocolProcessor< ByteBuffer >
        createReadProcessor()
            throws Exception
        {
            ProtocolProcessor< ByteBuffer > res =
                ts.createSend( getActivityContext() );

            if ( useGzip ) res = GzipReader.create( res, getActivityContext() );

            return res;
        }

        private
        void
        startRead( final ByteBuffer md5Out )
            throws Exception
        {
            CipherReader r =
                init( new CipherReader.Builder() ).
                    setProcessor( createReadProcessor() ).
                    setAllocationEventHandler( readImplChecker ).
                    build();
            
            ProtocolProcessorTests.readStream(
                r, 
                ioBuf,
                md5Digester(),
                getActivityContext(),
                new ObjectReceiver< ByteBuffer >() {
                    public void receive( ByteBuffer md5In ) 
                    {
                        state.equal( md5Out, md5In );
                        writeImplChecker.assertState();
                        readImplChecker.assertState();
                        exit();
                    }
                }
            );
        }

        private
        ProtocolProcessor< ByteBuffer >
        createWriteProcessor()
            throws Exception
        {
            ProtocolProcessor< ByteBuffer > res =
                ts.createReceive( getActivityContext() );

            if ( useGzip ) res = GzipWriter.create( res, getActivityContext() );

            return res;
        }

        private
        void
        startWrite()
            throws Exception
        {
            CipherWriter w =
                init( new CipherWriter.Builder() ).
                    setProcessor( createWriteProcessor() ).
                    setAllocationEventHandler( writeImplChecker ).
                    build();
            
            ProtocolProcessorTests.
                writeStream( 
                    w, 
                    streamLen, 
                    ioBuf,
                    new ProtocolProcessorTests.RandomFiller(),
                    md5Digester(),
                    getActivityContext(),
                    new ObjectReceiver< ByteBuffer >() {
                        public void receive( ByteBuffer md5Out )
                            throws Exception
                        {
                            startRead( md5Out );
                        }
                    }
                );
        }

        protected
        void
        startImpl()
            throws Exception
        {
            state.notNull( ts, "ts" );
            state.notNull( streamLen, "streamLen" );
            state.notNull( cipherCtx, "cipherCtx" );

            cf = createCipherFactory( cipherCtx );

            ioBuf = useDirectBufs
                ? ByteBuffer.allocateDirect( testBufSize ) 
                : ByteBuffer.allocate( testBufSize );

            startWrite();
        }
    }

    private
    List< CipherTestContext >
    getReaderWriterCipherTestContexts()
    {
        return getBaseTestContexts();
    }

    private
    static
    ProtocolTestStorage
    createStorage( boolean useFs )
    {
        return useFs ? FileStorage.create() : BufferAccumulatorStorage.create();
    }

    private
    void
    addBaseReaderWriterTests( boolean useFs,
                              int[] szArr,
                              List< CipherReaderWriterTest > l )
    {
        for ( int sz : szArr )
        for ( int b = 0; b < 2; ++ b )
        for ( CipherTestContext ctx : getReaderWriterCipherTestContexts() )
        {
            l.add(
                new CipherReaderWriterTest().
                    setTestStorage( createStorage( useFs ) ).
                    setStreamLen( DataSize.ofBytes( sz ) ).
                    setUseDirectBufs( b == 0 ).
                    setCipherTestContext( ctx )
            );
        }
    }

    private
    void
    addBaseReaderWriterTests( List< CipherReaderWriterTest > l )
    {
        addBaseReaderWriterTests( true, new int[] { 0, 1024, 500 * 1024 }, l );
        addBaseReaderWriterTests( false, new int[] { 0, 100 }, l );
    }

    private
    CipherReaderWriterTest
    createDegenerateReaderWriterTest( CipherTestContext ctx,
                                      boolean useFs,
                                      boolean useDirectBufs )
    {
        ProtocolTestStorage ts = 
            DegenerateStorage.create( createStorage( useFs ), true, true );

        return
            new CipherReaderWriterTest().
                setTestStorage( ts ).
                setStreamLen( DataSize.ofBytes( 100 ) ).
                setUseDirectBufs( useDirectBufs ).
                setCipherTestContext( ctx );
    }

    private
    CipherReaderWriterTest
    createGzipReaderWriterTest( CipherTestContext ctx,
                                boolean useFs,
                                boolean useDirectBufs )
    {
        return
            new CipherReaderWriterTest().
                setCipherTestContext( ctx ).
                setStreamLen( DataSize.ofKilobytes( 8 ) ).
                setTestStorage( createStorage( useFs ) ).
                setUseDirectBufs( useDirectBufs ).
                setUseGzip( true );
    }

    private
    CipherReaderWriterTest
    createSmallCipherBufTest( CipherTestContext ctx,
                              boolean useFs,
                              boolean useDirectBufs )
    {
        // we allow 2 buffer allocs since we're setting maxInputLen which is not
        // a divisor of the cipher block len, making it likely that we'll grow
        // the buffer once
        return
            new CipherReaderWriterTest().
                setTestStorage( createStorage( useFs ) ).
                setUseDirectBufs( useDirectBufs ).
                setCipherTestContext( ctx ).
                setStreamLen( DataSize.ofBytes( 4000 ) ).
                setMaxInputLen( 17 ).
                setTestBufSize( 1000 ).
                setWriteImplChecker( new ImplChecker( 2 ) ).
                setReadImplChecker( new ImplChecker( 2 ) );
    }

    private
    void
    addMiscReaderWriterTests( List< CipherReaderWriterTest > l )
    {
        for ( int s = 0; s < 2; ++s )
        for ( int b = 0; b < 2; ++b )
        {
            CipherTestContext ctx = AES_256_CBC_PKCS5;

            boolean useFs = s == 0;
            boolean useDirectBufs = b == 0;
 
            l.add( 
                createDegenerateReaderWriterTest( ctx, useFs, useDirectBufs ) );

            l.add( createGzipReaderWriterTest( ctx, useFs, useDirectBufs ) );

            l.add( createSmallCipherBufTest( ctx, useFs, useDirectBufs ) );
        }
    }

    private
    final
    static
    class FailStorage
    implements ProtocolTestStorage
    {
        private final ProtocolTestStorage ts;
        private final boolean failSend;
        private final boolean failRecv;
        private final int failAt;

        private
        FailStorage( ProtocolTestStorage ts,
                     boolean failSend,
                     boolean failRecv,
                     int failAt )
        {
            this.ts = ts;
            this.failSend = failSend;
            this.failRecv = failRecv;
            this.failAt = failAt;
        }
        
        public
        CharSequence
        getLabel()
        {
            return
                Strings.crossJoin( "=", ",",
                    "tsCls", ts.getClass().getSimpleName(),
                    "failSend", failSend,
                    "failRecv", failRecv,
                    "failAt", failAt
                );
        }

        private
        ProtocolProcessor< ByteBuffer >
        wrap( ProtocolProcessor< ByteBuffer > proc,
              boolean fail )
        {
            if ( fail )
            {
                return
                    ProtocolProcessorTests.
                        createFailingProcessor( failAt, proc );
            }
            else return proc;
        }

        public
        ProtocolProcessor< ByteBuffer >
        createSend( ProcessActivity.Context ctx )
            throws Exception
        {
            return wrap( ts.createSend( ctx ), failSend );
        }

        public
        ProtocolProcessor< ByteBuffer >
        createReceive( ProcessActivity.Context ctx )
            throws Exception
        {
            return wrap( ts.createReceive( ctx ), failRecv );
        }
    }

    private
    FailStorage
    createFailStorage( int failAt,
                       boolean useFs,
                       boolean failSend )
    {
        return 
            new FailStorage( 
                createStorage( useFs ), failSend, ! failSend, failAt );
    }

    private
    void
    addImplFailureTests( List< CipherReaderWriterTest > l )
    {
        for ( int s = 0; s < 2; ++s )
        for ( int i = 0; i < 2; ++i )
        {
            boolean useFs = s == 0;

            l.add(
                new CipherReaderWriterTest().
                    setStreamLen( DataSize.ofBytes( 100 ) ).
                    setTestStorage( createFailStorage( 20, useFs, i == 0 ) ).
                    setCipherTestContext( AES_256_CBC_PKCS5 ).
                    expectError( 
                        ProtocolProcessorTests.MarkerException.class, null )
            );
        }
    }

    private
    final
    static
    class ShortEndStorage
    implements ProtocolTestStorage
    {
        private final ProtocolTestStorage ts;
        private final int endAt;
        private final boolean doShortRead;

        private
        ShortEndStorage( boolean useFs,
                         int endAt,
                         boolean doShortRead )
        {
            this.ts = createStorage( useFs );
            this.endAt = endAt;
            this.doShortRead = doShortRead;
        }

        public 
        CharSequence 
        getLabel() 
        {
            return
                Strings.crossJoin( "=", ",",
                    "tsCls", ts.getClass().getSimpleName(),
                    "endAt", endAt,
                    "doShortRead", doShortRead
                );
        }

        public
        ProtocolProcessor< ByteBuffer >
        createSend( ProcessActivity.Context ctx )
            throws Exception
        {
            ProtocolProcessor< ByteBuffer > proc;
            
            if ( doShortRead )
            {
                // if ts is FileStorage we also need to create the receive in a
                // way that the backing file send won't complain about seeing an
                // incomplete consumer
                proc = ts instanceof FileStorage
                    ? ( (FileStorage) ts ).createSend( ctx, endAt )
                    : ts.createSend( ctx );
            
                return ProtocolProcessors.asFixedLengthProcessor( proc, endAt );
            }
            else return ts.createSend( ctx );
        }

        public
        ProtocolProcessor< ByteBuffer >
        createReceive( ProcessActivity.Context ctx )
            throws Exception
        {
            ProtocolProcessor< ByteBuffer > proc = ts.createReceive( ctx );

            if ( doShortRead ) return proc;
            else
            {
                return ProtocolProcessors.asFixedLengthProcessor( proc, endAt );
            }
        }
    }

    private
    void
    addUnexpectedReadEndTests( List< CipherReaderWriterTest > l )
    {
        for ( int s = 0; s < 2; ++s )
        for ( int i = 0; i < 2; ++i )
        {
            boolean doBlockAligned = i == 0;
            int end = doBlockAligned ? 32 : 31; // aes is 16 byte blocks

            Class< ? extends Throwable > errCls = doBlockAligned
                ? BadPaddingException.class : IllegalBlockSizeException.class;

            CharSequence errPat = doBlockAligned
                ? "Given final block not properly padded"
                : "Input length must be multiple of 16 when decrypting with " +
                  "padded cipher";

            l.add(
                new CipherReaderWriterTest().
                    setStreamLen( DataSize.ofBytes( 1000 ) ).
                    setTestStorage( new ShortEndStorage( s == 0, end, true ) ).
                    setCipherTestContext( AES_256_CBC_PKCS5 ).
                    expectError( errCls, errPat )
            );
        }
    }

    private
    void
    addUnexpectedWriteEndTests( List< CipherReaderWriterTest > l )
    {
        for ( int s = 0; s < 2; ++s )
        {
            l.add(
                new CipherReaderWriterTest().
                    setStreamLen( DataSize.ofBytes( 1000 ) ).
                    setTestStorage( new ShortEndStorage( s == 0, 50, false ) ).
                    setCipherTestContext( AES_256_CBC_PKCS5 ).
                    expectError(
                        IllegalStateException.class,
                        "Unexpected complete from proc"
                    )
            );
        }
    }

    private
    void
    addFailureTests( List< CipherReaderWriterTest > l )
    {
        addImplFailureTests( l );
        addUnexpectedReadEndTests( l );
        addUnexpectedWriteEndTests( l );
    }

    @InvocationFactory
    private
    List< CipherReaderWriterTest >
    testCipherReaderAndWriter()
    {
        List< CipherReaderWriterTest > res = Lang.newList();

        addBaseReaderWriterTests( res );
        addMiscReaderWriterTests( res );
        addFailureTests( res );

        return res;
    }

    @Test
    private
    final
    class EmptyFinalInputRegressionTest
    extends AbstractVoidProcess
    {
        private
        ByteBuffer
        expectSingleBuffer( ByteBufferAccumulator acc )
        {
            List< ByteBuffer > bufs = acc.getBuffers();
            state.equalInt( 1, bufs.size() );
    
            return bufs.get( 0 );
        }

        private
        ProtocolProcessor< ByteBuffer >
        emptyFinalInputSourceFor( ByteBuffer src )
        {
            return
                ProtocolProcessors.createBufferSend(
                    src,
                    IoUtils.emptyByteBuffer()
                );
        }
    
        private
        CipherWriter
        createCipherWriter( CipherFactory cf,
                            ByteBufferAccumulator acc )
        {
            return 
                new CipherWriter.Builder().
                    setActivityContext( getActivityContext() ).
                    setCipherFactory( cf ).
                    setProcessor( acc ).
                    build();
        }

        private
        void
        runEncrypt( ByteBuffer src,
                    CipherFactory cf,
                    final ObjectReceiver< ByteBuffer > recv )
        {
            // 40 is large enough that it will hold all expected outputs
            final ByteBufferAccumulator acc = 
                ByteBufferAccumulator.create( 40 );
            
            ProtocolCopies.copyByteStream(
                emptyFinalInputSourceFor( src ),
                createCipherWriter( cf, acc ),
                ByteBuffer.allocate( 1024 ),
                getActivityContext(),
                new AbstractTask() {
                    protected void runImpl() throws Exception {
                        recv.receive( expectSingleBuffer( acc ) );
                    }
                }
            );
        }

        private
        void
        assertEncryption( ByteBuffer src,
                          ByteBuffer enc,
                          CipherFactory cf )
            throws Exception
        {
            ByteBuffer src2 = ByteBuffer.allocate( enc.remaining() );
            Cipher c = cf.initDecrypt();
            c.doFinal( enc, src2 );
            src2.flip();
            state.equal( src, src2 );
        }

        private
        void
        runCipher( CipherTestContext ctx,
                   int blockLen,
                   int blockExtra,
                   final Runnable onComp )
            throws Exception
        {
            final CipherFactory cf = createCipherFactory( ctx );
    
            final ByteBuffer src = 
                IoTestFactory.nextByteBuffer( blockLen + blockExtra );

            runEncrypt( src.slice(), cf, new ObjectReceiver< ByteBuffer >() {
                public void receive( ByteBuffer enc ) throws Exception 
                {
                    assertEncryption( src, enc, cf );
                    onComp.run();
                }
            });
        }

        private
        void
        runCiphers( final CipherTestContext ctx,
                    final int blockLen,
                    final int blockExtra,
                    final Runnable onComp )
            throws Exception
        {
            runCipher( ctx, blockLen, blockExtra, new AbstractTask() {
                protected void runImpl() throws Exception
                {
                    int next = blockExtra + 1;

                    if ( next == blockLen ) onComp.run();
                    else runCiphers( ctx, blockLen, next, onComp );
                }
            });
        }

        private
        void
        assertNext( final Iterator< CipherTestContext > it )
            throws Exception
        {
            CipherTestContext ctx = it.next();
            int blockLen = getBlockLen( ctx );
        
            Runnable onComp = new AbstractTask() {
                protected void runImpl() throws Exception {
                    if ( it.hasNext() ) assertNext( it ); else exit();
                }
            };

            // mostly of interest is i == 0 and any single non-zero value, but
            // we sweep all values just to cover all bases
            if ( blockLen == 0 ) runCipher( ctx, 0, 10, onComp );
            else runCiphers( ctx, blockLen, 0, onComp );
        }

        protected
        void
        startImpl()
            throws Exception
        {
            assertNext( getBaseTestContexts().iterator() );
        }
    }

    static
    {
        AES_256_CBC_PKCS5 = 
            CipherTestContext.create( "AES/CBC/PKCS5Padding", 256 );
    }

    // To test:
    //
    //  - test interop with cipher input stream and cipher output stream
}
