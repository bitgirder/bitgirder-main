package com.bitgirder.io;

import com.bitgirder.validation.Inputs;

import java.nio.CharBuffer;
import java.nio.ByteBuffer;

// See http://tools.ietf.org/html/rfc4648 for the general overview of base64.
// This implementation is very limited, but can grow if needed to accomodate
// other usages (custom alphabets, stream-based operations, etc).
//
// Instances are safe for concurrent use.
public
class Base64Encoder
{
    private static Inputs inputs = new Inputs();

    // The alphabet is fixed now (thus the static init block below), but we
    // could open things up later to custom impls. Also, we'll likely want to
    // add the urlsafe alphabet at some point too as a standard option. 
    //
    // Regardless, the assumption is that the alphabet is based on 65 ascii
    // characters, one of which is padding. The encoding table is 64 elements
    // (the alphabet) and maps to chars (which are equiv to bytes since we're
    // only accepting ascii). The decoding table is sparse and keyed on chars in
    // the encoded string, with unused elements set to NOT_IN_ALPHABET (so we
    // can check that the string supplied is appropriate to the alphabet). The
    // table is stored this way (half unused) so the decode algorithm can just
    // do direct array access.

    private final static int NOT_IN_ALPHABET = -1;

    private final static char PAD_CHAR = '=';
    private final static String BASE_64_ALPHABET =
        "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";

    private final static char[] BASE_64 = new char[ 64 ];
    private final static int[] BASE_64_REV = new int[ 128 ];

    static 
    {
        for ( int i = 0; i < 127; ++i ) BASE_64_REV[ i ] = NOT_IN_ALPHABET;

        for ( int i = 0; i < 64; ++i )
        {
            BASE_64[ i ] = BASE_64_ALPHABET.charAt( i );
            
            BASE_64_REV[ BASE_64_ALPHABET.charAt( i ) ] = i;
            BASE_64_REV[ PAD_CHAR ] = 0;
        }
    }

    private final char[] alphaTable;
    private final int[] revTable;
    private final char padChar;

    public
    Base64Encoder()
    {
        this.alphaTable = BASE_64;
        this.revTable = BASE_64_REV;
        this.padChar = PAD_CHAR;
    }

    // This is a helper struct to make the encode/decode a bit easier to break
    // apart and pass from one method to another. For each of encode/decode, a
    // method will create one of these and set the appropriate lengths,
    // paddings, etc.
    //
    // The encode/decode method will then send the struct to a loop that moves
    // through the supplied input data/chars and fills the output chars/data
    // respectively.
    //
    private
    static
    class Context
    {
        byte[] data; // raw data
        int pad; // 0,1 or 2: the pad count for the final block
        int i; // the current pointer into data (whether for decode/encode)
        CharBuffer chars; // encoded text
    }

    private
    Context
    createDecodeContext( CharSequence base64 )
    {
        Context ctx = new Context();
        ctx.chars = IoUtils.charBufferFor( base64 );

        int len = base64.length();

        if ( len > 1 && base64.charAt( len - 2 ) == padChar ) ctx.pad = 2;
        else if ( len > 0 && base64.charAt( len - 1 ) == padChar ) ctx.pad = 1;

        ctx.data = new byte[ ( ( len / 4 ) * 3 ) - ctx.pad ];
 
        return ctx;
    }

    // Check that the specified character is a valid char for our alphabet, and
    // if it is, return the corresponding numeric value
    private
    int
    getChar( Context ctx,
             int indx )
        throws Base64Exception
    {
        char c = ctx.chars.get( indx );

        if ( c > revTable.length ) 
        {
            throw new Base64Exception( 
                "Character " + (byte) c + " is out of range" );
        }

        int res = revTable[ c ];

        if ( res == NOT_IN_ALPHABET )
        {
            throw new Base64Exception( 
                "Character " + (byte) c + " is not in this encoder's alphabet"
            );
        }

        return res;
    }
 
    // Main loop body for decode. Take a set of characters 'abcd' ('c' or 'd'
    // could be pad chars on the last block, which are just treated as zero) and
    // turn them back into a 24 bit string (stored in the int 'accum'), which is
    // then broken apart into three bytes in left-to-right order and stored in
    // the array. 
    // 
    // Unless this is the last invocation of this method AND there was padding,
    // all three bytes will be stored, otherwise only those that are required to
    // fill the array will be (the remaining 1 or 2 are just ignored).
    //
    // This method advances the ctx.i pointer as it fills bytes into the output
    // array.
    private
    void
    decodeNextTrio( Context ctx )
        throws Base64Exception
    {
        // Unpack the characters, checking that they are valid. Set ci to the
        // index of the first of our 4 chars
        int ci = ( ctx.i / 3 ) * 4;
        int accum = getChar( ctx, ci++ ) << 18;
        accum |= getChar( ctx, ci++ ) << 12;
        accum |= getChar( ctx, ci++ ) << 6;
        accum |= getChar( ctx, ci++ );

        // Break them into 3 bytes
        byte b1 = (byte) ( 0xff & ( accum >> 16 ) );
        byte b2 = (byte) ( 0xff & ( accum >> 8 ) );
        byte b3 = (byte) ( 0xff & accum );

        // Add the relevant bytes to the output array. 
        if ( ctx.i < ctx.data.length ) ctx.data[ ctx.i++ ] = b1;
        if ( ctx.i < ctx.data.length ) ctx.data[ ctx.i++ ] = b2;
        if ( ctx.i < ctx.data.length ) ctx.data[ ctx.i++ ] = b3;
    } 

    public
    byte[]
    decode( CharSequence base64 )
        throws Base64Exception
    {
        inputs.notNull( base64, "base64" );

        if ( base64.length() % 4 != 0 )
        {
            throw new Base64Exception(
                "Length of input '" + base64 + "' (" + base64.length() + 
                ") is not a multiple of 4" 
            );
        }

        Context ctx = createDecodeContext( base64 );
        while ( ctx.i < ctx.data.length ) decodeNextTrio( ctx );

        return ctx.data;
    }

    public
    ByteBuffer
    asByteBuffer( CharSequence base64 )
        throws Base64Exception
    {
        return ByteBuffer.wrap( decode( base64 ) );
    }

    private
    Context
    createEncodeContext( byte[] data )
    {
        Context ctx = new Context();

        ctx.data = data;
        int trios = (int) Math.ceil( (double) data.length / 3 );
        ctx.pad = ( 3 - ( data.length % 3 ) ) % 3;
        ctx.chars = CharBuffer.allocate( trios * 4 );
 
        return ctx;
    }

    private
    void
    encodeNextTrio( Context ctx )
    {
        // The character index of the rightmost char we'll produce. initialized
        // here since the lines following it will advance the pointer ctx.i. We
        // set it to the rightmost so we can more naturally pop 6 bits off the
        // right of accum in the final lines below.
        int ci = ( ( ctx.i / 3 ) * 4 ) + 3; 

        int i1 = 0xff0000 & ( ctx.data[ ctx.i++ ] << 16 );
        int i2 = ctx.i == ctx.data.length ?  
            0 : 0x00ff00 & ( ctx.data[ ctx.i++ ] << 8 );
        int i3 = ctx.i == ctx.data.length ? 0 : 0x0000ff & ctx.data[ ctx.i++ ];
 
        int accum = i1 | i2 | i3;

        ctx.chars.put( ci--, alphaTable[ accum & 63 ] );
        ctx.chars.put( ci--, alphaTable[ ( accum >>= 6 ) & 63 ] );
        ctx.chars.put( ci--, alphaTable[ ( accum >>= 6 ) & 63 ] );
        ctx.chars.put( ci--, alphaTable[ ( accum >>= 6 ) & 63 ] );
    }

    public
    CharSequence
    encode( byte[] data )
    {
        inputs.notNull( data, "data" );
        Context ctx = createEncodeContext( data );

        // fill all the 4-char blocks, ignoring padding for now
        while ( ctx.i < ctx.data.length ) encodeNextTrio( ctx );

        // now overwrite 0,1 or 2 chars with the padChar as appropriate
        if ( ctx.pad > 0 ) ctx.chars.put( ctx.chars.length() - 1, padChar );
        if ( ctx.pad > 1 ) ctx.chars.put( ctx.chars.length() - 2, padChar );
 
        return ctx.chars.toString();
    }

    // Right now, this class is byte[] based, so this method converts to that.
    // Later it would probably behoove us to switch the impl of this class
    // around, so that the byte[] method calls into this one.
    public
    CharSequence
    encode( ByteBuffer bb )
    {
        inputs.notNull( bb, "bb" );

        bb = bb.slice(); // only want the part that's in play
        byte[] data = new byte[ bb.remaining() ];
        bb.get( data );

        return encode( data );
    }
}
