package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.Charsets;

// As described here:
// http://docs.amazonwebservices.com/AmazonS3/2006-03-01/index.html?UsingKeys.html,
// keys must have a UTF-8 encoding of at most 1024 bytes. For the moment, this
// class does not check that property on construction of an instance, since
// doing so would effectively force this class to hold an internal binary
// version of the key as well as a string version -- the former for asserting
// and representing the correct encoded form, the latter for
// debugging/toString() and for doing other common string-like things with keys.
//
// This class ensures that keys begin with a leading slash, whether or not one
// is provided. If one is then it is left as-is; if not, one is added. Although
// keys are stored url-encoded, the leading slash is stored as a raw '/' char,
// even if the caller supplies a leading %2[fF], to simplify logging and
// debugging of request uris using these keys.
// 
// We can alter the factory methods going forward to provide instantiation time
// checks, etc, and can alter the class at need be to allow callers to trade off
// encoding efficiency v. storage efficiency v. programmer friendliness.
public
final
class S3ObjectKey
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final String key;

    private S3ObjectKey( String key ) { this.key = key; }

    public
    CharSequence
    asString( boolean withLeadingSlash )
    {
        return withLeadingSlash ? key : key.subSequence( 1, key.length() );
    }

    @Override public String toString() { return asString( true ).toString(); }

    public
    CharSequence
    decode( boolean withLeadingSlash )
    {
        return Charsets.UTF_8.urlDecode( asString( withLeadingSlash ) );
    }

    public CharSequence decode() { return decode( true ); }

    private
    static
    void
    notEmpty( CharSequence key )
    {
        inputs.isFalse( key.length() <= 1, "Empty key" );
    }

    private
    static
    CharSequence
    processDirectLeadingSlash( CharSequence key )
    {
        if ( key.length() > 2 && 
             key.charAt( 0 ) == '%' &&
             key.charAt( 1 ) == '2' &&
             ( key.charAt( 2 ) == 'f' || key.charAt( 2 ) == 'F' ) )
        {
            return
                new StringBuilder( key.length() - 2 ). 
                    append( '/' ).
                    append( key.subSequence( 3, key.length() ) );
        }
        else if ( key.length() > 0 && key.charAt( 0 ) != '/' )
        {
            return
                new StringBuilder( key.length() + 1 ).
                    append( '/' ).
                    append( key );
        }
        else return key;
    }

    public
    static
    S3ObjectKey
    createDirect( CharSequence key )
    {
        inputs.notNull( key, "key" );

        key = processDirectLeadingSlash( key );

        notEmpty( key );

        return new S3ObjectKey( key.toString() );
    }

    public
    static
    S3ObjectKey
    encodeAndCreate( CharSequence key )
    {
        inputs.notNull( key, "key" );
        notEmpty( key );

        StringBuilder res = new StringBuilder().append( '/' );
        CharSequence toEncode = key;

        if ( key.length() > 0 && key.charAt( 0 ) == '/' )
        {
            toEncode = key.subSequence( 1, key.length() ); 
        }

        res.append( Charsets.UTF_8.urlEncode( toEncode ) );

        return new S3ObjectKey( res.toString() );
    }
}
