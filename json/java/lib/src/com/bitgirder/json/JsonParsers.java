package com.bitgirder.json;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.Charsets;

import com.bitgirder.parser.SyntaxException;

import java.nio.ByteBuffer;

import java.nio.charset.Charset;

public
final
class JsonParsers
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private JsonParsers() {}

    // first four bytes will be at positions 0-3; See rfc4627 for algorithm used
    // here
    private
    static
    Charset
    detectCharsetFrom4Bytes( ByteBuffer data )
        throws SyntaxException
    {
        byte b0 = data.get( 0 );
        byte b1 = data.get( 1 );
        byte b2 = data.get( 2 );
        byte b3 = data.get( 3 );

        if ( b3 == ( b0 + b1 + b2 + b3 ) ) return Charsets.UTF_32BE.charset();
        else if ( b0 == 0 && b1 != 0 && b2 == 0 && b3 != 0 )
        {
            return Charsets.UTF_16BE.charset();
        }
        else if ( b0 == ( b0 + b1 + b2 + b3 ) )
        {
            return Charsets.UTF_32LE.charset();
        }
        else if ( b0 != 0 && b1 == 0 && b2 != 0 && b3 == 0 )
        {
            return Charsets.UTF_16LE.charset();
        }
        else if ( b0 != 0 && b1 != 0 && b2 != 0 && b3 != 0 )
        {
            return Charsets.UTF_8.charset();
        }
        else
        {
            throw new SyntaxException(
                "Can't detect charset from first four octets of json text" );
        }
    }

    // If this method sees 2 bytes exactly, UTF-8 is assumed, since the only
    // valid option would be that the data represents the text '{}' or '[]' (or
    // some 2-char invalid json text); if this is not the case, or if the bytes
    // are not valid UTF-8, we let the parse catch that.
    //
    // If at least 4 bytes are available, we apply the algorithm from rfc4627;
    // if not, we fail here
    //
    public
    static
    Charset
    detectCharset( ByteBuffer data )
        throws SyntaxException
    {
        inputs.notNull( data, "data" );

        switch ( data.remaining() )
        {
            case 0:
            case 1:
            case 3:
                throw new SyntaxException(
                    "Insufficient or malformed data in buffer" );

            case 2: return Charsets.UTF_8.charset();
 
            default:
                return detectCharsetFrom4Bytes( data.slice() );
        }
    }
}
