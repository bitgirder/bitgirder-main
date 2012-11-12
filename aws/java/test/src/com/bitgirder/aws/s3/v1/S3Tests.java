package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.test.Test;
import com.bitgirder.test.InvocationFactory;
import com.bitgirder.test.LabeledTestCall;
import com.bitgirder.test.TestFailureExpector;

import java.util.List;

@Test
final
class S3Tests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    // Label is bucket name being tested
    private
    final
    static
    class BucketFromStringTest
    extends LabeledTestCall
    implements TestFailureExpector
    {
        private final CharSequence errMsgExpct;
        
        private
        BucketFromStringTest( CharSequence bucket,
                              CharSequence errMsgExpct )
        {
            super( bucket );

            this.errMsgExpct = errMsgExpct;
        }

        private
        BucketFromStringTest( CharSequence bucket )
        {
            this( bucket, null );
        }

        public
        Class< ? extends Throwable >
        expectedFailureClass()
        {
            return errMsgExpct == null ? null : IllegalArgumentException.class;
        }

        public CharSequence expectedFailurePattern() { return errMsgExpct; }

        @Override
        protected
        void
        call()
        {
            S3Bucket bucket = S3Bucket.fromString( getLabel() );
            state.equalString( getLabel(), bucket.asString() );
        } 
    }

    @InvocationFactory
    private
    List< BucketFromStringTest >
    testBucketFromString()
    {
        StringBuilder tooLongBucket = new StringBuilder();
        for ( int i = 0; i < 256; ++i ) tooLongBucket.append( 'a' );

        return Lang.asList(
            
            new BucketFromStringTest( "okay" ),
            new BucketFromStringTest( "0kay" ),
            new BucketFromStringTest( "okay_123" ),
            new BucketFromStringTest( "o-_kay123...2" ),
            new BucketFromStringTest( "host.at.some.place" ),

            new BucketFromStringTest( 
                "aa", "Bucket string must be between 3 and 255 chars, got: aa"
            ),

            new BucketFromStringTest( 
                tooLongBucket, 
                "Bucket string must be between 3 and 255 chars, got: " +
                    tooLongBucket 
            ),

            new BucketFromStringTest( 
                "a-bad-n@me", "Invalid bucket character at index 7: @" ),

            new BucketFromStringTest( 
                ".cannot-start-with-period", 
                "First character is not a letter or a digit: ."
            ),

            new BucketFromStringTest( 
                "127.126.125.124", 
                "Invalid bucket format \\(ipv4 numeric quad\\): " +
                    "127\\.126\\.125\\.124"
            )
        );
    }

    private
    final
    static
    class S3ObjectKeyCreateTest
    extends LabeledTestCall
    implements TestFailureExpector
    {
        private final CharSequence rawKey;
        private final CharSequence keyExpct;
        private final CharSequence errExpct;
        private final boolean isEncoded;

        private
        S3ObjectKeyCreateTest( CharSequence rawKey,
                               CharSequence keyExpct,
                               CharSequence errExpct,
                               boolean isEncoded )
        {
            super(
                Strings.crossJoin( "=", ",",
                    "rawKey", rawKey,
                    "isEncoded", isEncoded,
                    "errExpct", errExpct
                )
            );

            this.rawKey = rawKey;
            this.keyExpct = keyExpct;
            this.errExpct = errExpct;
            this.isEncoded = isEncoded;
        }
        
        private
        S3ObjectKeyCreateTest( CharSequence rawKey,
                               CharSequence keyExpct,
                               boolean isEncoded )
        {
            this( rawKey, keyExpct, null, isEncoded );
        }
        
        private
        S3ObjectKeyCreateTest( CharSequence rawKey,
                               boolean isEncoded,
                               CharSequence errExpct )
        {
            this( rawKey, null, errExpct, isEncoded );
        }

        public
        Class< ? extends Throwable >
        expectedFailureClass()
        {
            return errExpct == null ? null : IllegalArgumentException.class;
        }

        public CharSequence expectedFailurePattern() { return errExpct; }

        @Override
        protected
        void
        call()
        {
            S3ObjectKey key = isEncoded
                ? S3ObjectKey.createDirect( rawKey )
                : S3ObjectKey.encodeAndCreate( rawKey );
            
            state.equalString( keyExpct, key.toString() );
        }
    }

    @InvocationFactory
    private
    List< S3ObjectKeyCreateTest >
    testCreateS3ObjectKey()
    {
        return Lang.asList(

            new S3ObjectKeyCreateTest( "/key1", "/key1", true ),
            new S3ObjectKeyCreateTest( "%2Fkey1", "/key1", true ),
            new S3ObjectKeyCreateTest( "%2fkey1", "/key1", true ),
            new S3ObjectKeyCreateTest( "/key1", "/key1", false ),
            new S3ObjectKeyCreateTest( "key1", "/key1", true ),
            new S3ObjectKeyCreateTest( "key1", "/key1", false ),
            new S3ObjectKeyCreateTest( "//key1", "//key1", true ),
            new S3ObjectKeyCreateTest( "//key1", "/%2Fkey1", false ),

            // This would likely represent a programmer error in real life,
            // since the call has what appears to be a pre-encoded key with
            // leading slash. Even so, we test that the library handles it
            // correctly and doubly-encodes the '%' sign and adds the leading
            // '/'
            new S3ObjectKeyCreateTest( "%2fkey1", "/%252fkey1", false ),

            new S3ObjectKeyCreateTest( "%2f", true, "Empty key" ),
            new S3ObjectKeyCreateTest( "/", true, "Empty key" ),
            new S3ObjectKeyCreateTest( "/", false, "Empty key" ),
            new S3ObjectKeyCreateTest( "", true, "Empty key" ),
            new S3ObjectKeyCreateTest( "", false, "Empty key" )
        );
    }

    @Test
    private
    void
    testKeyToStringLeadingSlashControl()
    {
        S3ObjectKey k = S3ObjectKey.encodeAndCreate( "/test" );

        state.equalString( "/test", k.toString() );
        state.equalString( "/test", k.asString( true ) );
        state.equalString( "test", k.asString( false ) );
    }
}
