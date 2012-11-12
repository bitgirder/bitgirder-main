package com.bitgirder.mingle.bincodec;

import static com.bitgirder.mingle.bincodec.MingleBinaryCodecConstants.*;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.io.Charsets;
import com.bitgirder.io.CharsetHelper;
import com.bitgirder.io.IoUtils;

import com.bitgirder.parser.SyntaxException;

import com.bitgirder.mingle.model.MingleStruct;
import com.bitgirder.mingle.model.MingleStructure;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleSymbolMap;
import com.bitgirder.mingle.model.MingleSymbolMapBuilder;
import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.MingleBoolean;
import com.bitgirder.mingle.model.AtomicTypeReference;
import com.bitgirder.mingle.model.MingleNull;
import com.bitgirder.mingle.model.MingleEnum;
import com.bitgirder.mingle.model.MingleList;
import com.bitgirder.mingle.model.MingleTimestamp;

import com.bitgirder.mingle.codec.MingleDecoder;
import com.bitgirder.mingle.codec.MingleCodecException;

import java.util.Deque;

import java.nio.ByteBuffer;

final
class MingleBinaryDecoder< E >
implements MingleDecoder< E >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final Class< E > cls;

    private final Deque< Reader > readers = Lang.newDeque();
    private final ProgressCheck pc = new ProgressCheck( "decoder" );

    private int pos = -1; // indicates position in message of last byte read
    private Reader resRdr;

    private 
    MingleBinaryDecoder( Class< E > cls )
    {
        this.cls = cls;

        expect( new MingleValueReader( TYPE_CODE_STRUCT ) );
    }

    final
    byte
    get( ByteBuffer bb )
    {
        byte res = bb.get();
        ++pos;

        return res;
    }

    final
    int
    getInt( ByteBuffer bb )
    {
        int res = bb.getInt();
        pos += 4;

        return res;
    }

    final
    long
    getLong( ByteBuffer bb )
    {
        long res = bb.getLong();

        pos += 8;

        return res;
    }

    final
    float
    getFloat( ByteBuffer bb )
    {
        float res = bb.getFloat();

        pos += 4;
        
        return res;
    }

    final
    double
    getDouble( ByteBuffer bb )
    {
        double res = bb.getDouble();

        pos += 8;

        return res;
    }

    final
    int
    copy( ByteBuffer src,
          ByteBuffer dest )
    {
        int res = IoUtils.copy( src, dest );
        pos += res;

        return res;
    }

    private
    MingleCodecException
    codecException( int explicitPos,
                    Object... args )
    {
        return MingleBinaryCodecs.codecException( explicitPos, args );
    }

    private
    MingleCodecException
    codecException( Object... args )
    {
        return MingleBinaryCodecs.codecException( pos, args );
    }

    private
    MingleCodecException
    asCodecException( SyntaxException se,
                      int startPos,
                      Object... msg )
    {
        return
            codecException(
                startPos,
                Strings.join( " ", msg ).toString() + ": " + se.getMessage() 
            );
    }

    private
    CharSequence
    toHex( byte code )
    {
        return String.format( "0x%1$02x", code );
    }

    private
    MingleCodecException
    unrecognizedType( byte code )
    {
        return codecException( "Unrecognized type code:", toHex( code ) );
    }

    private
    RuntimeException
    unexpectedConsume( Object obj )
    {
        return 
            state.createFail( 
                "Unexpected consume result", obj, "of type", 
                ( obj == null ? null : obj.getClass().getName() )
            );
    }

    // Returns false to simplify reader code which will often be calling this
    // method to push a sub-reader and immediately return a false value to
    // indicate that it's processing should be suspended
    private 
    boolean 
    expect( Reader r ) 
    { 
        readers.push( r ); 
        return false;
    }

    private
    abstract
    class Reader
    {
        abstract
        Object
        getResult()
            throws Exception;

        // Called when this instance becomes the top of stack again with the
        // read result from the Reader that was the top immediately before
        boolean
        consumeResult( Object obj )
            throws Exception
        {
            throw unexpectedConsume( obj );
        }

        boolean
        implReadFrom( ByteBuffer bb )
            throws Exception
        {
            throw state.createFail( "Unimplemented" );
        }

        boolean
        implReadFrom( ByteBuffer bb,
                      boolean endOfInput )
            throws Exception
        {
            return implReadFrom( bb );
        }
        
        final
        boolean
        readFrom( ByteBuffer bb,
                  boolean endOfInput )
            throws Exception
        {
            return implReadFrom( bb, endOfInput );
        }
    }

    private
    abstract
    class FixedValueReader
    extends Reader
    {
        private final int recLen;

        private Object val;

        private FixedValueReader( int recLen ) { this.recLen = recLen; }

        final
        Object
        getResult()
        {
            state.isFalse( val == null, "val not read yet" );
            return val;
        }

        abstract
        Object
        getValue( ByteBuffer bb )
            throws Exception;

        final
        boolean
        implReadFrom( ByteBuffer bb )
            throws Exception
        {
            if ( bb.remaining() >= recLen ) 
            {
                val = getValue( bb );
                return true;
            }
            else return false;
        }
    }

    private
    boolean
    expectJvInt32()
    {
        return 
            expect( new FixedValueReader( 4 ) {
                Object getValue( ByteBuffer bb ) { return getInt( bb ); }
            });
    }

    private
    boolean
    expectBoolean()
    {
        return
            expect( new FixedValueReader( 1 ) {
                Object getValue( ByteBuffer bb ) throws Exception
                {
                    byte b = get( bb );

                    if ( b == (byte) 0 ) return MingleBoolean.FALSE;
                    else if ( b == (byte) 1 ) return MingleBoolean.TRUE;
                    else 
                    {
                        throw codecException( 
                            "Invalid boolean value:", toHex( b ) );
                    }
                }
            });
    }

    private
    boolean
    expectInt32()
    {
        return
            expect( new FixedValueReader( 4 ) {
                Object getValue( ByteBuffer bb ) {
                    return MingleModels.asMingleInt32( getInt( bb ) );
                }
            });
    }

    private
    boolean
    expectInt64()
    {
        return
            expect( new FixedValueReader( 8 ) {
                Object getValue( ByteBuffer bb ) {
                    return MingleModels.asMingleInt64( getLong( bb ) );
                }
            });
    }

    private
    abstract
    class SizedBufferReader
    extends Reader
    {
        // when not-null and completed, the following will all be true:
        // bb.position() == 0, bb.limit() == bb.capacity(), bb.hasArray() ==
        // true, bb.arrayOffset() == 0
        private ByteBuffer buf;

        private int startPos = -1;

        abstract
        Object
        getResult( ByteBuffer buf )
            throws Exception;

        final
        int
        bufferPosition()
        {
            state.isFalse( startPos < 0, "buffer position not yet known" );
            return startPos;
        }

        final
        Object
        getResult()
            throws Exception
        {
            state.isFalse( buf == null || buf.hasRemaining() );

            buf.flip();

            return getResult( buf );
        }

        final
        boolean
        implReadFrom( ByteBuffer bb )
        {
            if ( buf == null && bb.remaining() >= 4 )
            {
                byte[] arr = new byte[ getInt( bb ) ];
                buf = ByteBuffer.wrap( arr );
                startPos = pos + 1; // actual first byte is one ahead of pos
            }

            if ( buf != null ) copy( bb, buf );

            return buf != null && ( ! buf.hasRemaining() );
        }
    }

    private
    boolean
    expectFloat()
    {
        return
            expect( new FixedValueReader( 4 ) {
                Object getValue( ByteBuffer bb ) {
                    return MingleModels.asMingleFloat( getFloat( bb ) );
                }
            });
    }

    private
    boolean
    expectDouble()
    {
        return
            expect( new FixedValueReader( 8 ) {
                Object getValue( ByteBuffer bb ) {
                    return MingleModels.asMingleDouble( getDouble( bb ) );
                }
            });
    }

    private
    boolean
    expectUtf8String()
    {
        return
            expect( new AbstractStringReader( Charsets.UTF_8 ) {
                Object getResult( CharSequence str ) {
                    return MingleModels.asMingleString( str );
                }
            });
    }

    private
    boolean
    expectBuffer()
    {
        return
            expect( new SizedBufferReader() {
                Object getResult( ByteBuffer bb ) {
                    return MingleModels.asMingleBuffer( bb );
                }
            });
    }

    private
    boolean
    expectNull()
    {
        return
            expect( new FixedValueReader( 0 ) {
                Object getValue( ByteBuffer bb ) { 
                    return MingleNull.getInstance();
                }
            });
    }

    private
    final
    class MingleValueReader
    extends Reader
    {
        private final Byte expctCode;

        private MingleValue val;

        private 
        MingleValueReader( Byte expctCode )
        { 
            this.expctCode = expctCode; 
        }

        Object
        getResult()
        {
            state.notNull( val, "getResult() called but val is null" );
            return val;
        }

        @Override
        boolean
        consumeResult( Object obj )
        {
            val = state.cast( MingleValue.class, obj );
            return true;
        }

        private
        byte
        getCode( ByteBuffer bb )
            throws Exception
        {
            byte res = get( bb );

            if ( expctCode != null && res != expctCode.byteValue() )
            {
                throw codecException(
                    "Saw type code", toHex( res ), 
                    "but expected", toHex( expctCode ) );
            }

            return res;
        }

        boolean
        implReadFrom( ByteBuffer bb )
            throws Exception
        {
            byte code = getCode( bb );

            if ( code == TYPE_CODE_BOOLEAN ) expectBoolean();
            else if ( code == TYPE_CODE_INT32 ) expectInt32();
            else if ( code == TYPE_CODE_INT64 ) expectInt64();
            else if ( code == TYPE_CODE_FLOAT ) expectFloat();
            else if ( code == TYPE_CODE_DOUBLE ) expectDouble();
            else if ( code == TYPE_CODE_UTF8_STRING ) expectUtf8String();
            else if ( code == TYPE_CODE_BUFFER ) expectBuffer();
            else if ( code == TYPE_CODE_STRUCT ) expectStruct();
            else if ( code == TYPE_CODE_EXCEPTION ) expectException();
            else if ( code == TYPE_CODE_SYMBOL_MAP ) expectSymbolMap();
            else if ( code == TYPE_CODE_ENUM ) expectEnum();
            else if ( code == TYPE_CODE_LIST ) expectList();
            else if ( code == TYPE_CODE_RFC3339_STR ) expectRfc3339Str();
            else if ( code == TYPE_CODE_NULL ) expectNull();
            else throw unrecognizedType( code );

            return true;
        }
    }

    private
    abstract
    class AbstractStringReader
    extends Reader
    {
        private CharsetHelper cs;
        private CharSequence str;

        private int strPos;

        private AbstractStringReader( CharsetHelper cs ) { this.cs = cs; }
        private AbstractStringReader() { this( null ); }

        final
        MingleCodecException
        asCodecException( SyntaxException se,
                          Object... args )
        {
            return 
                MingleBinaryDecoder.this.
                    asCodecException( se, strPos, args );
        }

        @Override
        boolean
        consumeResult( Object obj )
        {
            state.isTrue( str == null );
            str = state.cast( CharSequence.class, obj );

            return true;
        }

        abstract
        Object
        getResult( CharSequence str )
            throws Exception;

        final
        Object
        getResult()
            throws Exception
        {
            return getResult( str );
        }

        private
        void
        initCharset( ByteBuffer bb )
            throws Exception
        {
            byte code = get( bb );

            if ( code == TYPE_CODE_UTF8_STRING ) cs = Charsets.UTF_8;
            else 
            {
                throw codecException( 
                    "Unrecognized string type:", toHex( code ) );
            }
        }

        final
        boolean
        implReadFrom( ByteBuffer bb )
            throws Exception
        {
            state.isTrue( str == null ); // we should never enter here otherwise

            if ( cs == null ) initCharset( bb );

            if ( cs != null )
            {
                expect( new SizedBufferReader() 
                {
                    Object 
                    getResult( ByteBuffer bb ) 
                        throws Exception 
                    {
                        strPos = bufferPosition();
                        return cs.asString( bb );
                    }
                });
            }

            return false;
        }
    }

    private
    boolean
    expectMingleValue()
    {
        return expect( new MingleValueReader( null ) );
    }

    private 
    boolean 
    expectStruct() 
    { 
        return 
            expect( 
                new StructureReader() 
                {
                    @Override
                    MingleStructure
                    buildStructure( AtomicTypeReference typ,
                                    MingleSymbolMap flds )
                    {
                        return MingleModels.asMingleStruct( typ, flds );
                    }
                }
            );
    }

    private 
    boolean 
    expectException() 
    { 
        return 
            expect(
                new StructureReader()
                {
                    @Override
                    MingleStructure
                    buildStructure( AtomicTypeReference typ,
                                    MingleSymbolMap flds )
                    {
                        return MingleModels.asMingleException( typ, flds );
                    }
                }
            );
    }

    private
    boolean
    expectSymbolMap()
    {
        return expect( new SymbolMapReader() );
    }

    private
    final
    class EnumReader
    extends Reader
    {
        private AtomicTypeReference type;
        private MingleIdentifier val;

        Object
        getResult()
        {
            state.isFalse( 
                type == null || val == null,
                "getResult() called before enum ready" );
            
            return MingleEnum.create( type, val );
        }

        @Override
        boolean
        consumeResult( Object obj )
            throws Exception
        {
            if ( type == null ) type = asAtomicType( obj );
            else if ( val == null ) 
            {
                val = state.cast( MingleIdentifier.class, obj );
            }
            else throw unexpectedConsume( obj );

            return type != null && val != null;
        }

        boolean
        implReadFrom( ByteBuffer bb )
        {
            if ( type == null ) return expect( new TypeReferenceReader() );
            else if ( val == null ) return expect( new IdentifierReader() );
            else throw state.createFail( "Enum reader is done" );
        }
    }

    private boolean expectEnum() { return expect( new EnumReader() ); }

    private
    final
    class ListReader
    extends Reader
    {
        private Integer listLen; // value of which is currently ignored
        private MingleList.Builder lb;

        Object
        getResult()
        {
            return lb == null ? MingleModels.getEmptyList() : lb.build();
        }

        @Override
        boolean
        consumeResult( Object obj )
        {
            if ( lb == null ) lb = new MingleList.Builder();
            lb.add( state.cast( MingleValue.class, obj ) );

            return false;
        }

        boolean
        implReadFrom( ByteBuffer bb )
        {
            if ( listLen == null && bb.remaining() >= 4 )
            {
                listLen = getInt( bb );
            }

            if ( listLen != null && bb.hasRemaining() )
            {
                if ( bb.get( bb.position() ) == TYPE_CODE_END )
                {
                    get( bb );
                    return true;
                }
                else return expectMingleValue();
            }
            else return false;
        }
    }

    private boolean expectList() { return expect( new ListReader() ); }

    private
    boolean
    expectRfc3339Str()
    {
        return 
            expect( new AbstractStringReader() {
                Object getResult( CharSequence str ) throws Exception 
                {
                    try { return MingleTimestamp.parse( str ); }
                    catch ( SyntaxException se )
                    {
                        throw asCodecException( se );
                    }
                }
            });
    }

    private
    final
    class TypeReferenceReader
    extends AbstractStringReader
    {
        Object
        getResult( CharSequence str )
            throws Exception
        {
            try { return MingleTypeReference.parse( str ); }
            catch ( SyntaxException se )
            {
                throw asCodecException( se, "Parsing type '" + str + "'" );
            }
        }
    }

    private
    AtomicTypeReference
    asAtomicType( Object obj )
        throws MingleCodecException
    {
        MingleTypeReference typ = state.cast( MingleTypeReference.class, obj );

        if ( obj instanceof AtomicTypeReference )
        {
            return (AtomicTypeReference) obj;
        }
        else throw codecException( "Expected atomic type, got:", typ );
    }

    private
    final
    class IdentifierReader
    extends AbstractStringReader
    {
        Object
        getResult( CharSequence str )
            throws Exception
        {
            try { return MingleIdentifier.parse( str ); }
            catch ( SyntaxException se )
            {
                throw 
                    asCodecException( se, "Parsing identifier '" + str + "'" );
            }
        }
    }

    private
    final
    class SymbolMapPairReader
    extends Reader
    {
        private MingleIdentifier id;
        private MingleValue mv;

        Object getResult() { return this; }

        @Override
        boolean
        consumeResult( Object obj )
        {
            if ( id == null ) id = state.cast( MingleIdentifier.class, obj );
            else if ( mv == null ) mv = state.cast( MingleValue.class, obj );
            else throw unexpectedConsume( obj );

            return mv != null;
        }

        boolean
        implReadFrom( ByteBuffer bb )
        {
            if ( id == null ) return expect( new IdentifierReader() );
            else if ( mv == null ) return expectMingleValue();
            else throw state.createFail( "sym map pair reader is done" );
        }
    }

    private
    final
    class SymbolMapReader
    extends Reader
    {
        private MingleSymbolMapBuilder bldr;

        Object
        getResult()
        {
            return 
                bldr == null ? MingleModels.getEmptySymbolMap() : bldr.build();
        }

        @Override
        boolean
        consumeResult( Object res )
        {
            SymbolMapPairReader r = (SymbolMapPairReader) res;
            
            if ( bldr == null ) bldr = MingleModels.symbolMapBuilder();

            state.notNull( r.id );
            state.notNull( r.mv );

            bldr.set( r.id, r.mv );

            return false;
        }

        boolean
        implReadFrom( ByteBuffer bb )
        {
            byte code = bb.get( bb.position() ); // don't remove byte yet
            
            if ( code == TYPE_CODE_END ) 
            {
                get( bb ); // okay to remove now
                return true;
            }
            else return expect( new SymbolMapPairReader() );
        }
    }

    private
    abstract
    class StructureReader
    extends Reader
    {
        private Integer sz;
        private AtomicTypeReference type;
        private MingleSymbolMap flds;

        abstract
        MingleStructure
        buildStructure( AtomicTypeReference type,
                        MingleSymbolMap flds );

        final Object getResult() { return buildStructure( type, flds ); }

        @Override
        boolean
        consumeResult( Object obj )
            throws Exception
        {
            if ( sz == null ) sz = state.cast( Integer.class, obj );
            else if ( type == null ) type = asAtomicType( obj );
            else if ( flds == null ) 
            {
                flds = state.cast( MingleSymbolMap.class, obj );
            }
            else throw unexpectedConsume( obj );

            return flds != null;
        }

        boolean
        implReadFrom( ByteBuffer bb )
        {
            if ( sz == null ) return expectJvInt32();
            else if ( type == null ) return expect( new TypeReferenceReader() );
            else if ( flds == null ) return expect( new SymbolMapReader() );
            else throw state.createFail( "Structure read is done" );
        }
    }

    public
    E
    getResult()
        throws Exception
    {
        state.isFalse( resRdr == null, "Call to getResult() but no value set" );

        return state.cast( cls, resRdr.getResult() );
    }

    private
    void
    popResults()
        throws Exception
    {
        for ( boolean loop = true; loop && ( ! readers.isEmpty() ); )
        {
            Reader rdr = readers.remove();
    
            if ( readers.isEmpty() ) resRdr = rdr;
            else loop = readers.peek().consumeResult( rdr.getResult() );
        }
    }

    private
    void
    runReadLoop( ByteBuffer bb,
                 boolean endOfInput )
        throws Exception
    {
        ByteBuffer bb2 = MingleBinaryCodecs.byteOrdered( bb );

        boolean loop;
        do
        {
            Reader rdr = readers.peek();
            loop = rdr.readFrom( bb2, endOfInput );

            // only pop results if rdr finished and did not add new readers atop
            // itself
            boolean rdrPushed = readers.peek() != rdr;
            if ( loop && ( ! rdrPushed ) ) popResults();

            if ( ! loop ) loop = rdrPushed;
        }
        while ( bb2.hasRemaining() && loop && ( ! readers.isEmpty() ) );

        bb.position( bb2.position() );
    }

    public
    boolean
    readFrom( ByteBuffer bb,
              boolean endOfInput )
        throws Exception
    {
        inputs.notNull( bb, "bb" );
        inputs.isTrue( bb.hasRemaining(), "Input buffer is empty" );

//        code( "Entering readFrom(), bb:", bb, "; readers:", readers );
        state.isFalse( 
            readers.isEmpty(), "Decoder is complete (reader stack is empty)" );

        pc.enter( bb );
        runReadLoop( bb, endOfInput );
//        code( "After read loop, bb:", bb, "; readers:", readers );
        pc.assertProgress( bb );

        return readers.isEmpty();
    }

    static
    < E >
    MingleBinaryDecoder< E >
    create( Class< E > cls )
    {
        inputs.notNull( cls, "cls" );

        inputs.isTrue(
            cls.equals( MingleStruct.class ),
            "Unexpected decode target class:", cls.getName()
        );

        return new MingleBinaryDecoder< E >( cls );
    }
}
