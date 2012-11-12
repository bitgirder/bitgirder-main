package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.PatternHelper;

import com.bitgirder.io.Charsets;

import java.util.regex.Pattern;

public
final
class S3Bucket
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static Pattern IPV4_NUMERIC_QUAD =
        PatternHelper.compile( "\\d{1,3}(\\.\\d{1,3}){3}" );

    private final String data;

    private S3Bucket( String data ) { this.data = data; }

    public CharSequence asString() { return data; }
    @Override public String toString() { return data; }

    // See
    // http://docs.amazonwebservices.com/AmazonS3/latest/dev/BucketRestrictions.html
    // for more on the various validations

    // Note: docs state bucket str len "Must be between 3 and 255 characters
    // long" which strictly speaking would seem to imply the open interval
    // (3,255), but it seems more that the closed [3,255] is meant, so that's
    // what we're going with until we find otherwise
    private
    static
    void
    assertStringLength( CharSequence str )
    {
        int len = str.length();

        inputs.isTrue( 
            len >= 3 && len <= 255, 
            "Bucket string must be between 3 and 255 chars, got:", str );
    }

    private
    static
    void
    assertFirstChar( CharSequence str )
    {
        char c1 = str.charAt( 0 );

        inputs.isTrue( 
            Character.isLetter( c1 ) || Character.isDigit( c1 ),
            "First character is not a letter or a digit:", c1 );
    }

    private
    static
    boolean
    isBucketChar( char ch )
    {
        return 
            Character.isLowerCase( ch ) ||
            Character.isDigit( ch ) ||
            ch == '.' ||
            ch == '-' ||
            ch == '_'
        ;
    }

    private
    static
    void
    assertBucketChars( CharSequence str,
                       int start )
    {
        for ( int i = start, e = str.length(); i < e; ++i )
        {
            char ch = str.charAt( i );
            inputs.isTrue( 
                isBucketChar( ch ),
                "Invalid bucket character at index", i + ":", ch
            );
        }
    }

    private
    static
    void
    assertNotIpv4DottedQuad( CharSequence str )
    {
        inputs.isFalse(
            IPV4_NUMERIC_QUAD.matcher( str ).matches(),
            "Invalid bucket format (ipv4 numeric quad):", str 
        );
    }

    private
    static
    void
    validateBucketString( CharSequence str )
    {
        assertStringLength( str );
        assertFirstChar( str );
        assertBucketChars( str, 1 );
        assertNotIpv4DottedQuad( str );
    }

    public
    static
    S3Bucket
    fromString( CharSequence str )
    {
        validateBucketString( inputs.notNull( str, "str" ) );
        return new S3Bucket( str.toString() );
    }
}
