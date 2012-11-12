package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.ObjectReceiver;
import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.aws.AwsAccessKeyId;
import com.bitgirder.aws.AwsTesting;

import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.ProcessRpcClient;
import com.bitgirder.process.ProcessExit;
import com.bitgirder.process.ProcessActivity;

import com.bitgirder.concurrent.Duration;

import com.bitgirder.io.IoTestFactory;
import com.bitgirder.io.IoTestSupport;
import com.bitgirder.io.IoTests;
import com.bitgirder.io.IoUtils;
import com.bitgirder.io.IoProcessor;
import com.bitgirder.io.DataSize;
import com.bitgirder.io.FileWrapper;
import com.bitgirder.io.OctetDigest;
import com.bitgirder.io.ProtocolProcessor;
import com.bitgirder.io.ProtocolProcessors;
import com.bitgirder.io.Charsets;

import com.bitgirder.net.NetTests;
import com.bitgirder.net.NetProtocolTransportFactory;

import com.bitgirder.net.ssl.NetSslTests;

import com.bitgirder.test.Test;
import com.bitgirder.test.LabeledTestObject;
import com.bitgirder.test.InvocationFactory;
import com.bitgirder.test.TestRuntime;

import com.bitgirder.xml.Xpaths;
import com.bitgirder.xml.XmlIo;

import com.bitgirder.crypto.CryptoUtils;

import java.util.List;
import java.util.Map;
import java.util.Iterator;
import java.util.Set;

import java.nio.ByteBuffer;

import javax.crypto.SecretKey;

import com.amazonaws.s3.doc._2006_03_01.ListBucketResponse;
import com.amazonaws.s3.doc._2006_03_01.ListBucketResult;
import com.amazonaws.s3.doc._2006_03_01.ListEntry;
import com.amazonaws.s3.doc._2006_03_01.PrefixEntry;

@Test
final
class S3ClientTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final TestRuntime rt;

    private final AwsAccessKeyId accessKeyId;
    private final SecretKey secretKey;
    private final S3Client s3Cli;

    // hardoced for the moment
    private final String testBucket = "s3test.bitgirder.com" ;

    private 
    S3ClientTests( TestRuntime rt ) 
    { 
        this.rt = rt; 

        this.accessKeyId = AwsTesting.expectAccessKeyId( rt );
        this.secretKey = AwsTesting.expectSecretKey( rt );
        this.s3Cli = S3Testing.expectS3Client( rt );
    }

    private
    static
    void
    assertEtag( ByteBuffer md5,
                CharSequence etag )
    {
        CharSequence etagExpct = "\"" + IoUtils.asHexString( md5 ) + "\"";
        state.equalString( etagExpct, etag );
    }

    private
    abstract
    class AbstractTest
    extends AbstractVoidProcess
    {
        private S3Client s3Cli;
        Boolean useSsl;

        private final List< S3ObjectKey > toReap = Lang.newList();

        private 
        AbstractTest() 
        { 
            super( 
                ProcessRpcClient.create(),
                IoTestSupport.create( rt )
            ); 
        }

        final
        CharSequence
        makeLabel( Object... args )
        {
            return Strings.crossJoin( "=", ",", args );
        }

        public Object getInvocationTarget() { return this; }

        final
        IoProcessor
        ioProc()
        {
            return behavior( IoTestSupport.class ).ioProcessor();
        }

        @Override
        protected
        void
        childExited( AbstractProcess< ? > proc,
                     ProcessExit< ? > exit )
        {
            if ( ! exit.isOk() ) fail( exit.getThrowable() );
            if ( ! hasChildren() ) exit();
        }

        final S3Client s3Cli() { return s3Cli; }

        final
        < B extends S3Location.AbstractBuilder< ? > >
        B
        initSslLoc( B b )
        {
            if ( useSsl != null ) b.setUseSsl( useSsl );
            return b;
        }

        final
        S3BucketLocation
        bucketLoc()
        {
            return 
                initSslLoc( new S3BucketLocation.Builder() ).
                setBucket( testBucket ).
                build();
        }

        final
        S3ObjectLocation
        nextKeyLoc()
        {
            return
                initSslLoc( new S3ObjectLocation.Builder() ).
                setBucket( testBucket ).
                setKey( nextObjectKey() ).
                build();
        }

        final
        < B extends S3ObjectRequest.Builder< B > >
        B
        initRequest( B b )
        {
            return b.setLocation( nextKeyLoc() );
        }

        private
        void
        dumpRemoteException( GenericS3RemoteException ex )
        {
            try
            {
                // Assuming for now that xml is utf-8
                warn( "Failing due to remote s3 err with err doc:",
                    Charsets.UTF_8.asString( ex.errorXml() )
                );
            }
            catch ( Throwable th ) 
            {
                warn( th, "Couldn't dump remote error" );
            }
        }

        private
        void
        dumpErrAndFail( Throwable th )
        {
            if ( th instanceof GenericS3RemoteException )
            {
                dumpRemoteException( (GenericS3RemoteException) th );
            }

            fail( th );
        }

        final
        CharSequence
        nextUnencodedKeyString()
        {
            long millis = System.currentTimeMillis();
            String tmStr = String.format( "%1$016x", millis );

            return
                new StringBuilder().
                    append( "/test-path/" ).
                    append( tmStr ).
                    append( '/' ).
                    append( Lang.randomUuid() );
        } 

        final String nextObjectKey() { return S3Testing.nextUnencodedKey(); }

        final
        < V >
        void
        callS3( Object req,
                final Class< V > respCls,
                final ObjectReceiver< ? super V > recv )
        {
            beginRpc( 
                s3Cli(),
                req,
                new DefaultRpcHandler() 
                {
                    @Override protected void rpcSucceeded( Object resp ) 
                    {
                        try { recv.receive( respCls.cast( resp ) ); }
                        catch ( Throwable th ) { fail( th ); }
                    }

                    @Override protected void rpcFailed( Throwable th ) {
                        dumpErrAndFail( th );
                    }
                }
            );
        }

        abstract
        void
        startTest()
            throws Exception;

        private
        void
        stopS3CliConditional()
        {
            if ( s3Cli == S3ClientTests.this.s3Cli ) exit();
            else s3Cli.stop();
        }

        private
        void
        doTestDone( final Iterator< S3ObjectKey > it )
        {
            if ( it.hasNext() )
            {
                callS3( 
                    new S3ObjectDeleteRequest.Builder().
                        setLocation(
                            initSslLoc( new S3ObjectLocation.Builder() ).
                            setBucket( testBucket ).
                            setKey( it.next().decode().toString() ).
                            build()
                        ).
                        build(),
                    S3ObjectDeleteResponse.class,
                    new ObjectReceiver< S3ObjectDeleteResponse >() {
                        public void receive( S3ObjectDeleteResponse r ) {
                            doTestDone( it );
                        }
                    }
                );
            }
            else stopS3CliConditional();
        }

        final void testDone() { doTestDone( toReap.iterator() ); }

        final
        void
        reapKey( S3ObjectKey key )
        {
            toReap.add( state.notNull( key, "key" ) );
        }

        final
        void
        reapKey( S3ObjectResponse< ? > resp )
        {
            reapKey( S3ObjectKey.encodeAndCreate( resp.info().key() ) );
        }

        S3Client
        getS3Client()
            throws Exception
        {
            return S3ClientTests.this.s3Cli;
        }

        protected
        final
        void
        startImpl()
            throws Exception
        {
            s3Cli = getS3Client();
            if ( s3Cli != S3ClientTests.this.s3Cli ) spawn( s3Cli );

            startTest();
        }
    }

    // Basic coverage of parsing of an GenericS3RemoteException
    @Test
    private
    final
    class GenericS3RemoteExceptionTest
    extends AbstractTest
    {
        private
        void
        checkErrorDocument( GenericS3RemoteException ex )
            throws Exception
        {
            // simple check to assert that we correctly stored the err doc
            state.equalString(
                "NoSuchBucket",
                Xpaths.evaluate( 
                    "/Error/Code/text()", 
                    XmlIo.parseDocument( IoUtils.toByteArray( ex.errorXml() ) )
                )
            );
        }

        private
        void
        assertRemoteException( GenericS3RemoteException ex )
            throws Exception
        {
            state.equalString( 
                "The specified bucket does not exist", ex.message() );

            state.equalString( "NoSuchBucket", ex.code() );
            S3Testing.assertRequestIds( ex );
            checkErrorDocument( ex );

            testDone();
        }

        void
        startTest()
        {
            S3ObjectPutRequest req =
                new S3ObjectPutRequest.Builder().
                    setLocation(
                        initSslLoc( new S3ObjectLocation.Builder() ).
                            setBucket( "s3test-fail.bitgirder.com" ).
                            setKey( "/should-not-exist" ).
                            build()
                    ).
                    setBody( ByteBuffer.allocate( 1 ) ).
                    build();

            beginRpc( s3Cli(), req, 
                new DefaultRpcHandler() {
                    @Override protected void rpcFailed( Throwable th ) 
                    {
                        try
                        {
                            assertRemoteException( 
                                state.cast( 
                                    GenericS3RemoteException.class, th ) );
                        }
                        catch ( Throwable th2 ) { fail( th2 ); }
                    }
                }
            );
        }
    }

    private
    class BasicObjectRoundtripTest
    extends AbstractTest
    implements LabeledTestObject
    {
        private final Duration deleteCheckStall = Duration.fromSeconds( 3 );

        private String contentType;
        private boolean sendMd5;
        private boolean setAmzMeta;
        private int bufLen;

        public
        CharSequence
        getLabel()
        {
            return makeLabel(
                "useSsl", useSsl,
                "contentType", contentType,
                "sendMd5", sendMd5,
                "setAmzMeta", setAmzMeta,
                "bufLen", bufLen
            );
        }

        private
        < B extends S3ObjectRequest.Builder< B > >
        B
        initRequest( B b,
                     S3ObjectResponse< ? > resp )
        {
            return
                b.setLocation(
                    initSslLoc( new S3ObjectLocation.Builder() ).
                    setBucket( resp.info().bucket() ).
                    setKey( resp.info().key() ).
                    build()
                );
        }

        private
        void
        assertRespBase( S3ObjectResponse< ? > resp,
                        ByteBuffer md5,
                        boolean canHaveMeta )
        {
            state.notNull( resp.info().amazonRequestId() );
            state.notNull( resp.info().amazonId2() );

            if ( md5 != null ) assertEtag( md5, resp.h().expectEtagString() );

            if ( canHaveMeta && setAmzMeta )
            {
                state.equalString( 
                    "val1,val2", 
                    Strings.join( ",", resp.h().expect( "x-amz-meta-1" ) ) );
                
                state.equalString( 
                    "val1", resp.h().expectOne( "x-amz-meta-2" ) );
            }
        }

        private
        void
        assertObjectGone( Throwable th,
                          S3ObjectResponse< ? > resp )
        {
            NoSuchS3ObjectException ex = 
                state.cast( NoSuchS3ObjectException.class, th );
            
            state.equalString( resp.info().bucket(), ex.bucket() );
            state.equalString( resp.info().key(), ex.key() );
 
            exit();
        }

        private
        void
        startAssertObjectGone( final S3ObjectResponse resp )
        {
            beginRpc(
                s3Cli(),
                initRequest( new S3ObjectHeadRequest.Builder(), resp ).build(),
                new DefaultRpcHandler() {
                    @Override protected void rpcSucceeded() { state.fail(); }
                    @Override protected void rpcFailed( Throwable th ) {
                        assertObjectGone( th, resp );
                    }
                }
            );
        }

        private
        void
        assertDelete( final S3ObjectResponse resp )
        {
            submit(
                new AbstractTask() {
                    protected void runImpl() { startAssertObjectGone( resp ); }
                },
                deleteCheckStall
            );
        }

        private
        void
        startDelete( S3ObjectResponse resp )
        {
            callS3(
                initRequest( new S3ObjectDeleteRequest.Builder(), resp ).
                    build(),
                S3ObjectDeleteResponse.class,
                new ObjectReceiver< S3ObjectDeleteResponse >() {
                    public void receive( S3ObjectDeleteResponse resp ) {
                        assertDelete( resp );
                    }
                }
            );
        }

        // overridable along with completeBuild( S3ObjectGetRequest.Builder )
        void
        assertGetContents( ByteBuffer md5,
                           S3ObjectGetResponse resp )
            throws Exception
        {
            ByteBuffer data = (ByteBuffer) resp.getBodyObject();
            data.flip();
            state.equal( md5, CryptoUtils.getMd5( data ) );
        }

        private
        void
        assertGetResp( ByteBuffer md5,
                       S3ObjectGetResponse resp )
            throws Exception
        {
            assertRespBase( resp, md5, true );

            assertGetContents( md5, resp );
            
            startDelete( resp );
        }

        // overridable
        void
        completeBuild( S3ObjectGetRequest.Builder b )
        {
            b.setReceiveToByteBuffer();
        }

        private
        S3ObjectGetRequest
        createGetRequest( S3ObjectResponse resp )
        {
            S3ObjectGetRequest.Builder b =
                initRequest( new S3ObjectGetRequest.Builder(), resp );
            
            completeBuild( b );

            return b.build();
        }

        private
        void
        startGet( final ByteBuffer md5Expct,
                  S3ObjectResponse resp )
        {
            callS3(
                createGetRequest( resp ),
                S3ObjectGetResponse.class,
                new ObjectReceiver< S3ObjectGetResponse >() {
                    public void receive( S3ObjectGetResponse resp )
                        throws Exception
                    {
                        assertGetResp( md5Expct, resp ); 
                    }
                }
            );
        }

        private
        void
        assertHeadResp( ByteBuffer md5,
                        S3ObjectHeadResponse resp )
        {
            assertRespBase( resp, md5, true );
            startGet( md5, resp );
        }

        private
        void
        startHead( final ByteBuffer md5Expct,
                   final S3ObjectResponse resp )
        {
            callS3(
                initRequest( new S3ObjectHeadRequest.Builder(), resp ).build(),
                S3ObjectHeadResponse.class,
                new ObjectReceiver< S3ObjectHeadResponse >() {
                    public void receive( S3ObjectHeadResponse resp ) 
                    {
                        assertHeadResp( md5Expct, resp );
                    }
                }
            );
        }

        private
        void
        assertPut( S3ObjectPutResponse resp,
                   ByteBuffer md5 )
            throws Exception
        {
            assertRespBase( resp, md5, false );
            startHead( md5, resp );
        }

        private
        S3ObjectPutRequest
        buildPutReq( ByteBuffer src,
                     ByteBuffer md5 )
        {
            S3ObjectPutRequest.Builder b =
                initRequest( new S3ObjectPutRequest.Builder() ).
                    setBody( src );
            
            if ( contentType != null ) b.setContentType( contentType );
            if ( sendMd5 ) b.setContentMd5( md5 );

            if ( setAmzMeta )
            {
                b.addMetaField( "x-amz-meta-1", "val1" );
                b.addMetaField( "x-AMZ-Meta-1", "val2" );

                b.setMetaField( "x-amz-meta-2", "val1" );
            }

            return b.build();
        }

        void
        startTest()
            throws Exception
        {
            ByteBuffer src = bufLen > 0 
                ? IoTestFactory.nextByteBuffer( bufLen )
                : IoUtils.emptyByteBuffer();

            final ByteBuffer md5 = CryptoUtils.getMd5( src.slice() );

            callS3(
                buildPutReq( src, md5 ),
                S3ObjectPutResponse.class,
                new ObjectReceiver< S3ObjectPutResponse >() {
                    public void receive( S3ObjectPutResponse resp )
                        throws Exception
                    {
                        assertPut( resp, md5 );
                    }
                }
            );
        }
    }

    private
    void
    addNonEmptyBufferRoundtrips( List< BasicObjectRoundtripTest > res )
    {
        for ( int i = 0; i < 16; ++i )
        {
            BasicObjectRoundtripTest t = new BasicObjectRoundtripTest();

            t.useSsl = ( i & 1 ) > 0;
            if ( ( i & 2 ) > 0 ) t.contentType = "application/octet-stream";
            t.sendMd5 = ( i & 4 ) > 0;
            t.setAmzMeta = ( i & 8 ) > 0;
            t.bufLen = 8192;

            res.add( t );
        }
    }

    private
    void
    addEmptyBufferRoundtrips( List< BasicObjectRoundtripTest > res )
    {
        for ( int i = 0; i < 2; ++ i )
        {
            BasicObjectRoundtripTest t = new BasicObjectRoundtripTest();

            t.useSsl = i > 0;
            t.bufLen = 0;

            res.add( t );
        }
    }

    // Real-life code wouldn't accumulate the fixed-length GET Object response
    // into a ByteBufferAccumulator, but we use it here just so we have a known
    // place where that mechanism for body receive is used and some test
    // coverage of it which is independent of other codepaths.
    @Test
    private
    final
    class ByteBufferAccumulatorBodyReceiverTest
    extends BasicObjectRoundtripTest
    {
        @Override
        void
        assertGetContents( ByteBuffer md5,
                           S3ObjectGetResponse resp )
            throws Exception
        {
            @SuppressWarnings( "unchecked" )
            Iterable< ByteBuffer > bufs = 
                (Iterable< ByteBuffer >) resp.getBodyObject();

            ByteBuffer md5Recv =
                IoUtils.digest( 
                    CryptoUtils.createDigester( "md5" ), true, bufs );
 
            state.equal( md5, md5Recv );
        }
                    
        @Override
        void
        completeBuild( S3ObjectGetRequest.Builder b )
        {
            b.setAccumulateBody();
        }
    }

    // Test dimensions:
    //
    //  - ssl on/off
    //  - content type present/not
    //  - content md5 present/not
    //  - x-amz-headers present/not
    @InvocationFactory
    private
    List< BasicObjectRoundtripTest >
    testByteBufferRoundtrip()
    {
        List< BasicObjectRoundtripTest > res = Lang.newList();

        addNonEmptyBufferRoundtrips( res );
        addEmptyBufferRoundtrips( res );

        return res;
    }

    @Test( expected = IllegalArgumentException.class,
           expectedPattern = 
            "Meta headers should start with 'x-amz-meta-' " +
            "\\(case-insensitive\\); got: 'stuff'" )
    private
    void
    testInvalidXAmzMetaDetected()
    {
        new S3ObjectPutRequest.Builder().
            addMetaField( "stuff", "bad" );
    }

    private
    static
    FileWrapper
    nextFile()
    {
        try { return IoUtils.createTempFile( "s3-test-", true ); }
        catch ( Exception ex ) 
        { 
            throw new RuntimeException( "Couldn't create temp file", ex );
        }
    }

    private
    static
    abstract
    class Md5Fill
    extends IoTestSupport.FileFill< ByteBuffer >
    {
        private
        Md5Fill( DataSize len )
        {
            super( nextFile(), len, CryptoUtils.createDigester( "md5" ) );
        }
    }

    private
    final
    class FileTransferTest
    extends AbstractTest
    implements LabeledTestObject
    {
        private final DataSize flen;

        private
        FileTransferTest( boolean useSsl,
                          DataSize flen )
        {
            this.useSsl = useSsl;
            this.flen = flen;
        }

        public
        CharSequence
        getLabel()
        {
            return makeLabel( "useSsl", useSsl, "flen", flen );
        }

        private
        void
        assertFileReceive( final S3ObjectGetResponse resp,
                           final ByteBuffer md5 )
        {
            final FileWrapper f = (FileWrapper) resp.getBodyObject();

            final OctetDigest< ByteBuffer > dig =
                CryptoUtils.createDigester( "md5" );

            behavior( IoTestSupport.class ).getDigest(
                new IoTestSupport.FileDigest< ByteBuffer >( f, dig ) {
                    protected void digestReady( ByteBuffer md5Recv ) 
                    {
                        state.equal( md5, md5Recv );
                        reapKey( resp );
                        testDone();
                    }
                }
            );
        }

        private
        void
        receiveFile( S3ObjectResponse< ? > resp,
                     final ByteBuffer md5 )
        {
            callS3(
                new S3ObjectGetRequest.Builder().
                    setLocation(
                        initSslLoc( new S3ObjectLocation.Builder() ).
                        setBucket( resp.info().bucket() ).
                        setKey( resp.info().key() ).
                        build()
                    ).
                    setReceiveToFile( nextFile(), ioProc() ).
                    build(),
                S3ObjectGetResponse.class,
                new ObjectReceiver< S3ObjectGetResponse >() {
                    public void receive( S3ObjectGetResponse resp ) {
                        assertFileReceive( resp, md5 );
                    }
                }
            );
        }

        private
        void
        sendFile( FileWrapper src,
                  final ByteBuffer md5 )
        {
            callS3(
                new S3ObjectPutRequest.Builder().
                    setLocation( nextKeyLoc() ).
                    setContentMd5( md5 ).
                    setBody( src, ioProc() ).
                    build(),
                S3ObjectPutResponse.class,
                new ObjectReceiver< S3ObjectPutResponse >() {
                    public void receive( S3ObjectPutResponse resp ) {
                        receiveFile( resp, md5 );
                    }
                }
            );
        }

        void
        startTest()
        {
            behavior( IoTestSupport.class ).
                fill( 
                    new Md5Fill( flen ) {
                        protected void fileFilled() {
                            sendFile( file(), digest() );
                        }
                    }
                );
        }
    }

    @InvocationFactory
    private
    List< FileTransferTest >
    testFileTransfer()
    {
        List< FileTransferTest > res = Lang.newList();

        for ( int i = 0; i < 4; ++i )
        {
            res.add(
                new FileTransferTest(
                    ( i & 1 ) == 0,
                    DataSize.ofKilobytes( ( i & 2 ) == 0 ? 0 : 50 )
                )
            );
        }

        return res;
    }

    private
    final
    static
    class StallingProducer
    extends ProcessActivity
    implements ProtocolProcessor< ByteBuffer >
    {
        private final ProtocolProcessor< ByteBuffer > proc;
        private Duration stall;

        private
        StallingProducer( Duration stall,
                          int clen,
                          Context ctx )
        {
            super( ctx );

            proc = 
                ProtocolProcessors.
                    createBufferSend( IoTestFactory.nextByteBuffer( clen ) );
            
            this.stall = stall;
        }

        public
        void
        process( final ProcessContext< ByteBuffer > ctx )
        {
            if ( stall == null ) ProtocolProcessors.process( proc, ctx );
            else
            {
                submit(
                    new AbstractTask() {
                        protected void runImpl() {
                            ProtocolProcessors.process( proc, ctx );
                        }
                    },
                    stall
                );
            }
        }
    }

    // Somewhat redundant as of this writing, since S3Client's concurrency
    // throttling is nothing more than that of its underlying HttpClient; still
    // it's good to have this explicit test in place as a regression against
    // subtle changes that may cause the throttling not to be set correctly or
    // to otherwise take effect.
    @Test
    private
    final
    class MaxConcurrentConnectionThrottlingTest
    extends AbstractTest
    {
        private final int clen = (int) DataSize.ofKilobytes( 8 ).getByteCount();

        // the combination of stall * concurrency needs to be large enough that
        // it would be expected to outweigh any single request to S3. That is,
        // if it turns out that all requests actually do run in parallel, we
        // want the product stall * concurrency to be large enough that we will
        // fail our total runtime assertion below
        private final Duration stall = Duration.fromSeconds( 3 );
        private final int concurrency = 3;

        private long startTime;
        private int wait = concurrency;

        private
        S3ObjectPutRequest.BodyProducer
        createStallingBody()
        {
            return new S3ObjectPutRequest.BodyProducer() 
            {
                private StallingProducer sp;

                public
                void
                init( ProcessActivity.Context ctx )
                {
                    sp = new StallingProducer( stall, clen, ctx );
                }

                public ProtocolProcessor< ByteBuffer > getBody() { return sp; }

                public long getContentLength() { return clen; }
            };
        }

        private
        void
        putDone( S3ObjectPutResponse resp )
        {
            reapKey( resp );

            if ( --wait == 0 )
            {
                state.isTrue(
                    System.currentTimeMillis() - startTime >
                    stall.asMillis() * concurrency
                );

                testDone();
            }
        }

        void
        startTest()
        {
            startTime = System.currentTimeMillis();

            for ( int i = 0; i < concurrency; ++i )
            {
                callS3(
                    new S3ObjectPutRequest.Builder().
                        setLocation( nextKeyLoc() ).
                        setBody( createStallingBody() ).
                        build(),
                    S3ObjectPutResponse.class,
                    new ObjectReceiver< S3ObjectPutResponse >() {
                        public void receive( S3ObjectPutResponse r ) {
                            putDone( r );
                        }
                    }
                );
            }
        }

        @Override
        S3Client
        getS3Client()
            throws Exception
        {
            return
                new S3Client.Builder().
                    setNetworking( NetTests.expectSelectorManager( rt ) ).
                    setRequestFactory( S3Testing.createRequestFactory( rt ) ).
                    setMaxConcurrentConnections( 1 ).
                    build();
        }
    }

    private
    abstract
    class AbstractBucketListTest
    extends AbstractTest
    {
        int bodyLen = (int) DataSize.ofKilobytes( 4 ).getByteCount();
        int subDirs = 2;
        int filesPerDir = 2;

        CharSequence root;
        final Map< String, S3ObjectPutResponse > tree = Lang.newMap();

        abstract
        void
        treeReady()
            throws Exception;

        private
        final
        class PutHandler
        implements ObjectReceiver< S3ObjectPutResponse >
        {
            private final int numFiles;

            private PutHandler( int numFiles ) { this.numFiles = numFiles; }

            public
            void
            receive( S3ObjectPutResponse resp )
                throws Exception
            {
                // store them in the way that we'll get them back in list
                // results to simplify assertion mappings later
                tree.put( 
                    S3ObjectKey.encodeAndCreate( resp.info().key() ).
                        decode( false ).
                        toString(),
                    resp 
                );

                if ( tree.size() == numFiles ) treeReady();
            }
        }

        private
        void
        putFile( S3ObjectKey key,
                 PutHandler ph )
        {
            ByteBuffer body = IoTestFactory.nextByteBuffer( bodyLen );
            
            callS3(
                new S3ObjectPutRequest.Builder().
                    setLocation(
                        initSslLoc( new S3ObjectLocation.Builder() ).
                        setBucket( testBucket ).
                        setKey( key.decode().toString() ).
                        build()
                    ).
                    setBody( body ).
                    setContentMd5( CryptoUtils.getMd5( body.slice() ) ).
                    build(),
                S3ObjectPutResponse.class,
                ph
            );
        }

        final
        void
        startTest()
        {
            int fileCount = ( subDirs + 1 ) * filesPerDir;

            PutHandler ph = new PutHandler( fileCount );
            
            root = nextUnencodedKeyString();
            
            for ( int i = 0; i < fileCount; ++i )
            {
                int dirIdx = i % ( subDirs + 1 );

                CharSequence subDir = 
                    dirIdx == subDirs ? "" : "dir" + dirIdx + "/";
                
                S3ObjectKey key = 
                    S3ObjectKey.encodeAndCreate( 
                        root + "/" + subDir + "file" + ( i % filesPerDir ) );
                
                putFile( key, ph );
            }
        }
    }

    @Test
    private
    final
    class BasicBucketListTest
    extends AbstractBucketListTest
    {
        // All increments must happen before the first decrement
        private int testWait;

        private
        void
        listAssertDone()
        {
            if ( --testWait == 0 ) testDone();
        }

        private
        S3ListBucketRequest.Builder
        createReqBuilder()
        {
            return
                new S3ListBucketRequest.Builder().
                    setLocation( bucketLoc() ).
                    setPrefix( S3ObjectKey.encodeAndCreate( root ) );
        }

        private
        final
        class FlatListingAsserter
        implements ObjectReceiver< S3ListBucketResponse >
        {
            private Map< String, S3ObjectPutResponse > wait =
                Lang.copyOf( tree );

            private Integer maxKeys; 
            private int callsRemaining;

            private
            FlatListingAsserter( Integer maxKeys,
                                 int expctCalls )
            {
                this.maxKeys = maxKeys;
                this.callsRemaining = expctCalls;
            }
 
            private
            void
            assertPaging( ListBucketResult lbr )
            {
                if ( maxKeys != null )
                {
                    state.equalInt( maxKeys, lbr.getMaxKeys() );
                    state.isTrue( lbr.getContents().size() <= maxKeys );
                }

                state.equal( 
                    wait.size() > lbr.getMaxKeys(), lbr.isIsTruncated() );
            }

            private
            void
            assertEntry( ListEntry e )
            {
                S3ObjectPutResponse r = 
                    state.remove( wait, e.getKey(), "wait" );
                
                state.equal( (long) bodyLen, e.getSize() );
                
                state.equal( r.info().etag(), e.getETag() );
            }

            private
            ListEntry
            assertEntries( ListBucketResult lbr )
            {
                ListEntry last = null;

                for ( ListEntry e : lbr.getContents() ) 
                {
                    assertEntry( e );
                    last = e;
                }

                return last;
            }

            private
            void
            assertRemaining( ListEntry last )
            {
                if ( wait.isEmpty() )
                {
                    state.equalInt( 0, callsRemaining );
                    listAssertDone();
                }
                else 
                {
                    state.notNull( last );
                    start( S3ObjectKey.encodeAndCreate( last.getKey() ) );
                }
            }

            public
            void
            receive( S3ListBucketResponse resp )
            {
                ListBucketResult lbr = 
                    (ListBucketResult) resp.getResultObject();
    
                --callsRemaining;
                
                assertPaging( lbr );
                ListEntry e = assertEntries( lbr );

                assertRemaining( e );
            } 

            private
            void
            start( S3ObjectKey marker )
            {
                S3ListBucketRequest.Builder b = createReqBuilder();
                
                if ( maxKeys != null ) b.setMaxKeys( maxKeys );
                if ( marker != null ) b.setMarker( marker );

                callS3( b.build(), S3ListBucketResponse.class, this );
            }

            // Only intended to be called to kick off listing
            private 
            void 
            start() 
            { 
                ++testWait;
                start( null ); 
            }
        }

        private
        void
        assertEmptyListResult( S3ListBucketResponse resp )
        {
            state.isTrue( 
                ( (ListBucketResult) resp.getResultObject() ).
                getContents().
                isEmpty()
            );

            listAssertDone();
        }

        private
        void
        assertEmptyListing()
        {
            ++testWait;

            S3ObjectKey pref =
                S3ObjectKey.
                    encodeAndCreate( "//////this-prefix-does-not-exist" );

            callS3(
                new S3ListBucketRequest.Builder().
                    setLocation( bucketLoc() ).
                    setPrefix( pref ).
                    build(),
                S3ListBucketResponse.class,
                new ObjectReceiver< S3ListBucketResponse >() {
                    public void receive( S3ListBucketResponse resp ) {
                        assertEmptyListResult( resp );
                    }
                }
            );
        }

        private
        void
        assertDelimitedListing( S3ListBucketResponse resp )
        {
            ListBucketResult lbr = (ListBucketResult) resp.getResultObject();
            
            Set< String > expcts = 
                Lang.newSet( root + "/dir0/", root + "/dir1/" );

            for ( PrefixEntry pe : lbr.getCommonPrefixes() )
            {
                state.remove( expcts, "/" + pe.getPrefix(), "expcts" );
            }

            state.isTrue( expcts.isEmpty() );
            listAssertDone();
        }

        private
        void
        assertListingWithDelimiter()
        {
            ++testWait;

            callS3(
                createReqBuilder().
                    setDelimiter( "/" ).
                    setPrefix( S3ObjectKey.encodeAndCreate( root + "/" ) ).
                    build(),
                S3ListBucketResponse.class,
                new ObjectReceiver< S3ListBucketResponse >() {
                    public void receive( S3ListBucketResponse resp ) {
                        assertDelimitedListing( resp );
                    }
                }
            );
        }

        void
        treeReady()
        {
            new FlatListingAsserter( null, 1 ).start();
            new FlatListingAsserter( 7, 1 ).start();
            new FlatListingAsserter( 6, 1 ).start();
            new FlatListingAsserter( 5, 2 ).start();
            new FlatListingAsserter( 3, 2 ).start();
            
            assertEmptyListing();
            assertListingWithDelimiter();
        }
    }

    @Test
    private
    final
    class GetNonexistentObjectFailureTest
    extends AbstractTest
    {
        void
        startTest()
        {
            final S3ObjectLocation loc = nextKeyLoc();

            beginRpc(
                s3Cli(),
                new S3ObjectGetRequest.Builder().
                    setLocation( loc ).
                    setReceiveToByteBuffer().
                    build(),
                new DefaultRpcHandler() 
                {
                    @Override protected void rpcFailed( Throwable th )
                    {
                        S3Testing.assertNoSuchObject( loc, th, true );
                        testDone();
                    }

                    @Override protected void rpcSucceeded() { state.fail(); }
                }
            );
        }
    }
}
