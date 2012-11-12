package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.ObjectReceiver;

import com.bitgirder.aws.AwsAccessKeyId;
import com.bitgirder.aws.AwsTesting;

import com.bitgirder.crypto.CryptoUtils;

import com.bitgirder.io.IoUtils;

import com.bitgirder.mingle.bind.MingleBindTests;
import com.bitgirder.mingle.bind.MingleBinder;
import com.bitgirder.mingle.bind.MingleBinders;

import com.bitgirder.mingle.codec.MingleCodec;
import com.bitgirder.mingle.codec.MingleCodecs;

import com.bitgirder.mingle.json.JsonMingleCodecs;

import com.bitgirder.mingle.model.MingleStruct;

import com.bitgirder.net.NetTests;

import com.bitgirder.test.TestRuntime;
import com.bitgirder.testing.Testing;

import java.nio.ByteBuffer;

import java.util.Properties;

public
final
class S3Testing
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public final static Object KEY_S3_CLIENT =
        S3Testing.class.getName() + ".s3Client";

    private final static String RSRC_TEST_CTX = 
        "com/bitgirder/aws/s3/v1/s3-test-context.json";

    private final static Object KEY_TEST_CTX =
        S3Testing.class.getName() + ".testContext";

    private final static MingleCodec codec = JsonMingleCodecs.getJsonCodec();
 
    public
    static
    S3HttpRequestFactory
    createRequestFactory( Properties props )
        throws Exception
    {
        String accKey = 
            state.getProperty( props, AwsTesting.PROP_ACCESS_KEY, "props" );

        String secKey =
            state.getProperty( props, AwsTesting.PROP_SECRET_KEY, "props" );

        return S3HttpRequestFactories.
            create( 
                new AwsAccessKeyId( accKey ), 
                CryptoUtils.asSecretKey( secKey, CryptoUtils.ALG_HMAC_SHA1 )
            );
    } 

    public
    static
    S3HttpRequestFactory
    createRequestFactory( TestRuntime rt )
        throws Exception
    {
        return
            createRequestFactory(
                Testing.
                    expectObject( 
                        rt, 
                        AwsTesting.KEY_AWS_TESTING_PROPS, 
                        Properties.class 
                    ) 
            );
    }

    private
    static
    void
    spawnS3Cli( Testing.RuntimeInitializerContext ctx,
                Properties props )
        throws Exception
    {
        S3HttpRequestFactory fact = createRequestFactory( props );

        TestRuntime rt = ctx.getRuntime();

        S3Client.Builder b =
            new S3Client.Builder().
                setNetworking( NetTests.expectSelectorManager( rt ) ).
                setRequestFactory( fact ).
                setMaxConcurrentConnections( 10 );
        
        ctx.spawnAndSetStoppable( b.build(), KEY_S3_CLIENT );
        ctx.complete();
    }

    @Testing.RuntimeInitializer
    private
    static
    void
    initS3Client( final Testing.RuntimeInitializerContext ctx )
    {
        Testing.awaitTestObject(
            ctx,
            AwsTesting.KEY_AWS_TESTING_PROPS,
            Properties.class,
            new ObjectReceiver< Properties >() {
                public void receive( Properties props ) 
                    throws Exception
                {
                    spawnS3Cli( ctx, props );
                }
            }
        );
    }

    public
    static
    S3Client
    expectS3Client( TestRuntime rt )
    {
        return Testing.expectObject(
            inputs.notNull( rt, "rt" ), KEY_S3_CLIENT, S3Client.class );
    }

    public
    static
    String
    nextUnencodedKey()
    {
        long millis = System.currentTimeMillis();
        String tmStr = String.format( "%1$016x", millis );

        return
            new StringBuilder().
                append( "/test-path/" ).
                append( tmStr ).
                append( '/' ).
                append( Lang.randomUuid() ).
                toString();
    } 

    public
    static
    S3ObjectLocation
    nextObjectLocation( S3TestContext ctx,
                        Boolean useSsl )
    {
        inputs.notNull( ctx, "ctx" );

        S3ObjectLocation.Builder b = 
            new S3ObjectLocation.Builder().
                setBucket( ctx.bucket() ).
                setKey( nextUnencodedKey() );
        
        if ( useSsl != null ) b.setUseSsl( useSsl );

        return b.build();
    }

    static
    void
    assertRequestIds( S3RemoteException ex )
    {
        inputs.notNull( ex, "ex" );

        S3AmazonRequestIds ids = state.notNull( ex.amazonRequestIds() );
        state.isTrue( ids.requestId().matches( "^\\p{XDigit}{16}$" ) );
        state.isTrue( ids.id2().matches( "^[a-zA-Z0-9+/]{64}$" ) );
    }

    static
    void
    assertNoSuchObject( S3ObjectLocation locExpct,
                        Throwable th,
                        boolean expectReqIds )
    {
        NoSuchS3ObjectException ex =
            state.cast( NoSuchS3ObjectException.class, th );
        
        state.equal( locExpct.key(), ex.key() );
        state.equal( locExpct.bucket(), ex.bucket() );

        if ( expectReqIds ) assertRequestIds( ex );
    }

    static
    S3TestContext
    expectTestContext( TestRuntime rt )
    {
        inputs.notNull( rt, "rt" );

        return Testing.expectObject( rt, KEY_TEST_CTX, S3TestContext.class );
    }

    private
    final
    static
    class InitTask
    extends Testing.AbstractInitTask
    {
        private final MingleBinder mb;

        private
        InitTask( MingleBinder mb,
                  Testing.RuntimeInitializerContext ctx )
        {
            super( ctx );
            
            this.mb = mb;
        }
        
        protected
        void
        runImpl()
            throws Exception
        {
            ByteBuffer bb =
                IoUtils.toByteBuffer(
                    IoUtils.expectSingleResourceAsStream( RSRC_TEST_CTX ), 
                    true 
                );
            
            MingleStruct ms =
                MingleCodecs.fromByteBuffer( codec, bb, MingleStruct.class );
    
            context().setObject(
                KEY_TEST_CTX,
                MingleBinders.asJavaValue( mb, S3TestContext.class, ms )
            );

            context().complete();
        }
    }

    @Testing.RuntimeInitializer
    private
    static
    void
    init( final Testing.RuntimeInitializerContext ctx )
    {
        Testing.awaitTestObject(
            ctx, 
            MingleBindTests.KEY_BINDER, 
            MingleBinder.class,
            new ObjectReceiver< MingleBinder >() {
                public void receive( MingleBinder mb ) throws Exception {
                    Testing.submitInitTask( ctx, new InitTask( mb, ctx ) );
                }
            }
        );
    }
}
