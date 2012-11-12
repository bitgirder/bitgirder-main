package com.bitgirder.mingle.bincodec;

import static com.bitgirder.mingle.bincodec.MingleBinaryCodecConstants.*;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Strings;

import com.bitgirder.mingle.codec.MingleCodec;
import com.bitgirder.mingle.codec.MingleCodecDetector;
import com.bitgirder.mingle.codec.MingleCodecDetectorFactory;
import com.bitgirder.mingle.codec.MingleCodecFactory;
import com.bitgirder.mingle.codec.MingleCodecFactoryInitializer;
import com.bitgirder.mingle.codec.MingleCodecException;

import java.nio.ByteBuffer;
import java.nio.ByteOrder;

public
final
class MingleBinaryCodecs
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final static ByteOrder BYTE_ORDER = ByteOrder.LITTLE_ENDIAN;

    private MingleBinaryCodecs() {}

    public static MingleCodec getCodec() { return new MingleBinaryCodec(); }

    static
    ByteBuffer
    byteOrdered( ByteBuffer bb )
    {
        state.notNull( bb, "bb" );

        if ( bb.order().equals( BYTE_ORDER ) ) return bb;
        else return bb.duplicate().order( BYTE_ORDER );
    }

    static
    ByteBuffer
    allocateBuffer( int cap )
    {
        return ByteBuffer.allocate( cap ).order( BYTE_ORDER );
    }

    static
    MingleCodecException
    codecException( int pos,
                    Object... args )
    {
        StringBuilder msg = 
            new StringBuilder().
                append( "[offset " ).
                append( pos ).
                append( "] " ).
                append( Strings.join( " ", args ) );
        
        return new MingleCodecException( msg.toString() );
    }

    private
    final
    static
    class CodecDetectorFactoryImpl
    implements MingleCodecDetectorFactory
    {
        public
        MingleCodecDetector
        createCodecDetector()
        {
            return new MingleCodecDetector() {
                public Boolean update( ByteBuffer bb ) {
                    return bb.get() == TYPE_CODE_STRUCT;
                }
            };
        }
    }

    private
    final
    static
    class FactoryInitializer
    implements MingleCodecFactoryInitializer
    {
        public
        void
        initialize( MingleCodecFactory.Builder b )
        {
            b.addCodec( "binary", getCodec(), new CodecDetectorFactoryImpl() );
        }
    }
}
