package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;
import com.bitgirder.lang.ObjectReceiver;

import com.bitgirder.mingle.bind.AbstractBoundServiceTests;

import com.bitgirder.mingle.service.MingleServiceEndpoint;

import com.bitgirder.io.IoTestSupport;
import com.bitgirder.io.DataSize;
import com.bitgirder.io.FileWrapper;

import com.bitgirder.io.v1.NoSuchPathException;
import com.bitgirder.io.v1.PathPermissionException;

import com.bitgirder.crypto.CryptoUtils;
import com.bitgirder.crypto.MessageDigester;

import com.bitgirder.process.ProcessFailureTarget;

import com.bitgirder.test.Test;
import com.bitgirder.test.InvocationFactory;
import com.bitgirder.test.LabeledTestObject;
import com.bitgirder.test.TestRuntime;

import java.nio.ByteBuffer;

import java.util.List;
import java.util.Map;

@Test
final
class S3FileServiceTests
extends AbstractBoundServiceTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static String TEST_CTYPE = "application/binary";

    private final S3TestContext s3TestCtx;
    private final S3Client s3Cli;

    private 
    S3FileServiceTests( TestRuntime rt ) 
    { 
        super( rt ); 

        this.s3TestCtx = S3Testing.expectTestContext( rt );
        this.s3Cli = S3Testing.expectS3Client( rt );
    }

    private
    MessageDigester
    md5Digester()
    {
        return CryptoUtils.createDigester( "md5" );
    }

    private
    S3ObjectLocation
    nextObjectLoc( boolean useSsl )
    {
        return S3Testing.nextObjectLocation( s3TestCtx, useSsl );
    }

    private
    abstract
    class TestImpl
    extends ServiceTest
    {
        private
        TestImpl()
        {
            super( IoTestSupport.create( testRuntime() ) );
        }

        final IoTestSupport sprt() { return behavior( IoTestSupport.class ); }

        final
        FileWrapper
        nextNoPermFile()
            throws Exception
        {
            FileWrapper res = sprt().createTempFile();
            res.getFile().setReadable( false );
            res.getFile().setWritable( false );

            return res;
        }

        protected
        final
        void
        initServices( MingleServiceEndpoint.Builder b )
        {
            spawnService( 
                initBuilder( new S3FileService.Builder() ).
                    setS3Client( s3Cli ).
                    setIoProcessor( 
                        behavior( IoTestSupport.class ).ioProcessor() ).
                    build(), 
                b 
            );
            
            initDone( b );
        }

        final
        S3FileServiceClient
        cli()
        {
            return createClient( new S3FileServiceClient.Builder() );
        }
    }

    private
    final
    class FileUploadTest
    extends TestImpl
    implements LabeledTestObject
    {
        private Boolean withMeta;
        private Boolean withCtype;
        private Boolean withMd5;
        private Boolean useSsl;
        private DataSize fileSize;

        public
        CharSequence
        getLabel()
        {
            return
                Strings.crossJoin( "=", ",",
                    "withMeta", withMeta,
                    "withCtype", withCtype,
                    "withMd5", withMd5,
                    "useSsl", useSsl,
                    "fileSize", fileSize
                );
        }

        private
        void
        validate()
        {
            state.notNull( withMeta, "withMeta" );
            state.notNull( withCtype, "withCtype" );
            state.notNull( withMd5, "withMd5" );
            state.notNull( useSsl, "useSsl" );
            state.notNull( fileSize, "fileSize" );
        }

        private
        void
        assertMeta( Map< String, List< String > > m )
        {
            state.equalInt( 2, m.size() );

            state.equalString(
                "key1=val1,val2|key2=val3",
                Strings.crossJoin( "=", "|",
                    "key1", Strings.join( ",", state.get( m, "key1", "m" ) ),
                    "key2", Strings.join( ", ", state.get( m, "key2", "m" ) )
                )
            );
        }

        private
        void
        checkMeta( FileDownloadInfo dl )
        {
            Map< String, List< String > > m = Lang.newMap();

            for ( S3ObjectMetaData md : dl.s3Response().meta() )
            {
                Lang.putAppend( m, md.key(), md.value() );
            }

            if ( withMeta ) assertMeta( m ); else state.isTrue( m.isEmpty() );
        }

        private
        void
        verifyDownload( FileDownloadInfo dl,
                        FileUploadInfo ul,
                        ByteBuffer srcMd5 )
        {
            checkMeta( dl );
            state.equal( dl.md5(), srcMd5 );

            sprt().assertFileData(
                new FileWrapper( dl.file() ),
                new FileWrapper( ul.file() ),
                md5Digester(),
                md5Digester(),
                new AbstractTask() { protected void runImpl() { testDone(); } }
            );
        }

        private
        void
        beginDownload( final FileUploadInfo info,
                       final ByteBuffer srcMd5 )
            throws Exception
        {
            cli().
                downloadFile().
                setPath( sprt().createTempFile().toString() ).
                setLocation( info.location() ).
                receiveWith(
                    new ObjectReceiver< FileDownloadInfo >() {
                        public void receive( FileDownloadInfo dl ) {
                            verifyDownload( dl, info, srcMd5 );
                        }
                    }
                ).
                call();
        }

        private
        void
        verifyUpload( FileUploadInfo info,
                      ByteBuffer srcMd5 )
            throws Exception
        {
            state.equal( info.location().key(), info.s3Response().key() );
            state.equal( info.location().bucket(), info.s3Response().bucket() );
            state.notNull( info.s3Response().amazonId2() );
            state.notNull( info.s3Response().amazonRequestId() );
            state.notNull( info.s3Response().etag() );

            beginDownload( info, srcMd5 );
        }

        private
        S3FileServiceClient.UploadFileCall
        initUpload( S3FileServiceClient.UploadFileCall call,
                    FileWrapper file,
                    ByteBuffer md5 )
        {
            call.setPath( file.toString() );
            
            if ( withMd5 ) call.setMd5( md5.slice() );
            if ( withCtype ) call.setContentType( TEST_CTYPE );

            if ( withMeta )
            {
                call.setMeta(
                    Lang.asList(
                        S3ObjectMetaData.create( "key1", "val1" ),
                        S3ObjectMetaData.create( "key1", "val2" ),
                        S3ObjectMetaData.create( "key2", "val3" )
                    )
                );
            }

            return call;
        }

        private
        void
        beginUpload( FileWrapper file,
                     final ByteBuffer md5 )
        {
            initUpload( cli().uploadFile(), file, md5 ).
                setLocation( nextObjectLoc( useSsl ) ).
                receiveWith(
                    new ObjectReceiver< FileUploadInfo >() {
                        public void receive( FileUploadInfo info )
                            throws Exception
                        {
                            verifyUpload( info, md5 );
                        }
                    }
                ).
                call();
        } 
 
        protected
        void
        startTest()
            throws Exception
        {
            validate();

            sprt().fill(
                new IoTestSupport.FileFill< ByteBuffer >(
                    sprt().createTempFile(),
                    fileSize,
                    md5Digester() )
                {
                    protected void fileFilled() throws Exception {
                        beginUpload( file(), digest() );
                    }
                }
            );
        }
    }

    @InvocationFactory
    private
    List< FileUploadTest >
    testFileUpload()
    {
        List< FileUploadTest > res = Lang.newList();

        for ( int sz : new int[] { 0, 1 << 10 } )
        for ( int i = 0; i < 16; ++i )
        {
            FileUploadTest t = new FileUploadTest();
            t.withMeta = ( i & 1 ) > 0;
            t.withCtype = ( i & 2 ) > 0;
            t.withMd5 = ( i & 8 ) > 0;
            t.useSsl = ( i & 4 ) > 0;
            t.fileSize = DataSize.ofBytes( sz );

            res.add( t );
        }

        return res;
    }

    private
    final
    static
    class FailureExpector
    implements ObjectReceiver< Object >
    {
        public 
        void 
        receive( Object res )
        {
            state.fail( "Expected failure but got", res );
        }
    }

    @Test
    private
    final
    class DownloadNoSuchObjectFailureTest
    extends TestImpl
    {
        private final S3ObjectLocation loc = nextObjectLoc( false );

        protected
        void
        startTest()
            throws Exception
        {
            cli().downloadFile().
                setPath( sprt().createTempFile().toString() ).
                setLocation( loc ).
                receiveWith( new FailureExpector() ).
                setFailureTarget(
                    new ProcessFailureTarget() {
                        public void fail( Throwable th )
                        {
                            S3Testing.assertNoSuchObject( loc, th, true );
                            testDone();
                        }
                    }
                ).
                call();
        }
    }

    @Test
    private
    final
    class UploadOfNonexistentFileFailTest
    extends TestImpl
    {
        protected
        void
        startTest()
        {
            final String path = "/this/should/not/exist";

            cli().uploadFile().
                setPath( path ).
                setLocation( nextObjectLoc( false ) ).
                receiveWith( new FailureExpector() ).
                setFailureTarget(
                    new ProcessFailureTarget() {
                        public void fail( Throwable th ) 
                        {
                            NoSuchPathException ex = (NoSuchPathException) th;
                            state.equal( path, ex.path() );
                            testDone();
                        }
                    }
                ).
                call();
        }
    }

    // By using a key which doesn't exist (with very high probability) in
    // combination with our assertion that the failure we receive is a path
    // permission failure, we are also testing the fact that the implementation
    // does not even go to S3 if the local path is not valid (otherwise
    // the test would fail with the NoSuchS3ObjectException). This is something
    // we want to be true about the impl, in the spirit of resource efficiency,
    // and we take this test as our coverage of it.
    private
    final
    class PathPermissionFailureTest
    extends TestImpl
    implements LabeledTestObject
    {
        private final String mode;

        private PathPermissionFailureTest( String mode ) { this.mode = mode; }

        public CharSequence getLabel() { return mode; }

        private
        S3FileServiceClient.AbstractCall< ?, ? >
        initCall( String path )
        {
            if ( mode.equals( "upload" ) )
            {
                return 
                    cli().uploadFile().
                        setLocation( nextObjectLoc( false ) ).
                        setPath( path );
            }
            else if ( mode.equals( "download" ) )
            {
                return
                    cli().downloadFile().
                        setLocation( nextObjectLoc( false ) ).
                        setPath( path );
            }
            else throw state.createFail( "mode:", mode );
        }

        protected
        void
        startTest()
            throws Exception
        {
            final String path = nextNoPermFile().toString();

            initCall( path ).
                receiveWith( new FailureExpector() ).
                setFailureTarget(
                    new ProcessFailureTarget() {
                        public void fail( Throwable th )
                        {
                            state.equal( 
                                path, 
                                ( (PathPermissionException) th ).path() 
                            );

                            testDone();
                        }
                    }
                ).
                call();
        }
    }

    @InvocationFactory
    private
    List< PathPermissionFailureTest >
    testPathPermissionFailure()
    {
        return
            Lang.asList(
                new PathPermissionFailureTest( "upload" ),
                new PathPermissionFailureTest( "download" )
            );
    }
}
