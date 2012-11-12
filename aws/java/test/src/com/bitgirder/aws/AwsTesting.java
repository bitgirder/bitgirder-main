package com.bitgirder.aws;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.StandardThread;

import com.bitgirder.io.IoUtils;

import com.bitgirder.crypto.CryptoUtils;

import com.bitgirder.test.TestRuntime;

import com.bitgirder.testing.Testing;

import java.util.Properties;

import javax.crypto.SecretKey;

public
final
class AwsTesting
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public final static String KEY_AWS_TESTING_PROPS =
        AwsTesting.class.getName() + ".awsTestingProps";

    public final static String KEY_ACCESS_KEY_ID =
        AwsTesting.class.getName() + ".accessKeyId";

    public final static String KEY_SECRET_KEY =
        AwsTesting.class.getName() + ".secretKey";
    
    public final static String PROP_ACCESS_KEY = "accessKeyId";
    public final static String PROP_SECRET_KEY = "secretKey";

    private AwsTesting() {}

    public
    static
    AwsAccessKeyId
    expectAccessKeyId( TestRuntime rt )
    {
        return
            Testing.expectObject(
                inputs.notNull( rt, "rt" ),
                KEY_ACCESS_KEY_ID,
                AwsAccessKeyId.class
            );
    }

    public
    static
    SecretKey
    expectSecretKey( TestRuntime rt )
    {
        return
            Testing.expectObject(
                inputs.notNull( rt, "rt" ),
                KEY_SECRET_KEY,
                SecretKey.class
            );
    }

    private
    static
    void
    initAwsProps( Testing.RuntimeInitializerContext ctx )
        throws Exception
    {
        Properties props =
            IoUtils.loadProperties(
                IoUtils.expectSingleResource( "aws-testing.properties" ) );
 
        ctx.setObject( KEY_AWS_TESTING_PROPS, props );
        
        String accKey = state.getProperty( props, PROP_ACCESS_KEY, "props" );
        String secKey = state.getProperty( props, PROP_SECRET_KEY, "props" );

        ctx.setObject( KEY_ACCESS_KEY_ID, new AwsAccessKeyId( accKey ) );

        ctx.setObject(
            KEY_SECRET_KEY,
            CryptoUtils.asSecretKey( secKey, CryptoUtils.ALG_HMAC_SHA1 )
        );
    }

    @Testing.RuntimeInitializer
    private
    static
    void
    readAwsProps( final Testing.RuntimeInitializerContext ctx )
    {
        new StandardThread( "aws-testing-init-%1$d" )
        {
            public void run()
            {
                try
                {
                    initAwsProps( ctx );
                    ctx.complete();
                }
                catch ( Throwable th ) { ctx.fail( th ); }
            }
        }.
        start();
    }
}
