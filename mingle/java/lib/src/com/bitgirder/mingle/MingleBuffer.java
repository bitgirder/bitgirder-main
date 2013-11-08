package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.Base64Encoder;
import com.bitgirder.io.Base64Exception;
import com.bitgirder.io.IoUtils;

import java.util.Arrays;

import java.nio.ByteBuffer;

public
final
class MingleBuffer
implements MingleValue
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static Base64Encoder enc = new Base64Encoder();

    private final byte[] arr;

    public 
    MingleBuffer( byte[] arr )
    { 
        this.arr = inputs.notNull( arr, "arr" );
    }

    // returns live array
    public byte[] array() { return arr; }

    public int hashCode() { return Arrays.hashCode( arr ); }

    public
    boolean
    equals( Object o )
    {
        if ( o == this ) return true;
        if ( ! ( o instanceof MingleBuffer ) ) return false;

        return Arrays.equals( arr, ( (MingleBuffer) o ).arr );
    }

    // returns live buffer
    public 
    ByteBuffer 
    asByteBuffer() 
    { 
        return ByteBuffer.wrap( arr );
    }

    public 
    CharSequence 
    asBase64String() 
    { 
        return enc.encode( asByteBuffer() ); 
    }

    public CharSequence asHexString() { return IoUtils.asHexString( arr ); }

    public
    static
    MingleBuffer
    fromBase64String( CharSequence str )
        throws Base64Exception
    {
        inputs.notNull( str, "str" );

        return new MingleBuffer( enc.decode( str ) );
    }

    @Override
    public
    String
    toString()
    {
        return "[mingle buffer, length: " + arr.length + "]";
    }
}
