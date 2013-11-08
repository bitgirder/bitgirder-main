package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.Base64Encoder;
import com.bitgirder.io.Base64Exception;
import com.bitgirder.io.IoUtils;

import java.nio.ByteBuffer;

public
final
class MingleBuffer
implements MingleValue
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static Base64Encoder enc = new Base64Encoder();

    private final ByteBuffer bb;

    // live reference is kept, not copied, not made readOnly()
    public 
    MingleBuffer( ByteBuffer bb ) 
    { 
        this.bb = inputs.notNull( bb, "bb" ); 
    }

    public
    MingleBuffer( byte[] arr ) 
    { 
        this( ByteBuffer.wrap( inputs.notNull( arr, "arr" ) ) );
    }

    public int hashCode() { return bb.hashCode(); }

    public
    boolean
    equals( Object o )
    {
        if ( o == this ) return true;
        if ( ! ( o instanceof MingleBuffer ) ) return false;

        return bb.equals( ( (MingleBuffer) o ).bb );
    }

    public ByteBuffer getByteBuffer() { return bb.slice(); }

    public CharSequence asBase64String() { return enc.encode( bb ); }
    public CharSequence asHexString() { return IoUtils.asHexString( bb ); }

    public
    static
    MingleBuffer
    fromBase64String( CharSequence str )
        throws Base64Exception
    {
        inputs.notNull( str, "str" );

        return new MingleBuffer( enc.asByteBuffer( str ) );
    }

    @Override
    public
    String
    toString()
    {
        return "[mingle buffer, length: " + bb.remaining() + "]";
    }
}
