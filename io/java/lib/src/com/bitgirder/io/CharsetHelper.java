package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.io.UnsupportedEncodingException;

import java.net.URLEncoder;
import java.net.URLDecoder;

import java.nio.ByteBuffer;
import java.nio.CharBuffer;

import java.nio.charset.Charset;
import java.nio.charset.CharsetDecoder;
import java.nio.charset.CharsetEncoder;
import java.nio.charset.CharacterCodingException;

public
final
class CharsetHelper
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();
    
    private final Charset charset;

    public
    CharsetHelper( Charset charset )
    {
        this.charset = inputs.notNull( charset, "charset" );
    }

    // Utility method that does all main work, including null checks, for public
    // methods
    private
    CharSequence
    urlCode( CharSequence cs,
             boolean isEncode )
    {
        inputs.notNull( cs, "cs" );

        try
        {
            return isEncode
                ? URLEncoder.encode( cs.toString(), charset.name() )
                : URLDecoder.decode( cs.toString(), charset.name() );
        }
        catch ( UnsupportedEncodingException uee )
        {
            throw new IllegalStateException(
                "Couldn't url " + ( isEncode ? "en" : "de" ) + "code using " +
                "charset " + charset, uee );
        }
    }

    public
    CharSequence
    urlEncode( CharSequence cs )
    {
        return urlCode( cs, true );
    }

    public
    CharSequence
    urlDecode( CharSequence cs )
    {
        return urlCode( cs, false );
    }

    // Can add a thread local encoder cache later
    public
    ByteBuffer
    asByteBuffer( CharSequence s )
        throws CharacterCodingException
    {
        inputs.notNull( s, "s" );

        CharBuffer cb = IoUtils.charBufferFor( s );
        return newEncoder().encode( cb );
    }

    public
    byte[]
    asByteArray( CharSequence s )
        throws CharacterCodingException
    {
        return IoUtils.toByteArray( asByteBuffer( s ) );
    }

    private
    RuntimeException
    asUncheckedException( CharacterCodingException cce )
    {
        return
            new RuntimeException( 
                "Unexpected character coding exception", cce );
    }

    public
    ByteBuffer
    asByteBufferUnchecked( CharSequence s )
    {
        try { return asByteBuffer( s ); }
        catch ( CharacterCodingException cce )
        {
            throw asUncheckedException( cce );
        }
    }

    public
    byte[]
    asByteArrayUnchecked( CharSequence s )
    {
        try { return asByteArray( s ); }
        catch ( CharacterCodingException cce )
        {
            throw asUncheckedException( cce );
        }
    }

    public
    CharSequence
    asString( byte[] buf )
        throws CharacterCodingException
    {
        inputs.notNull( buf, "buf" );
        return asString( ByteBuffer.wrap( buf ) );
    }

    public
    CharBuffer
    asCharBuffer( ByteBuffer bb )
        throws CharacterCodingException
    {
        inputs.notNull( bb, "bb" );
        return newDecoder().decode( bb.asReadOnlyBuffer() );
    }

    public
    CharSequence
    asString( ByteBuffer bb )
        throws CharacterCodingException
    {
        return asCharBuffer( bb );
    }

    public Charset charset() { return charset; }
    public CharsetDecoder newDecoder() { return charset.newDecoder(); }
    public CharsetEncoder newEncoder() { return charset.newEncoder(); }
}
