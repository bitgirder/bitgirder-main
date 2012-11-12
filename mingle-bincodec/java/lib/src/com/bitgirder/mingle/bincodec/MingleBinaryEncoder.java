package com.bitgirder.mingle.bincodec;

import static com.bitgirder.mingle.bincodec.MingleBinaryCodecConstants.*;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.io.Charsets;
import com.bitgirder.io.IoUtils;

import com.bitgirder.mingle.model.MingleStruct;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleStructure;
import com.bitgirder.mingle.model.MingleSymbolMap;
import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleNull;
import com.bitgirder.mingle.model.MingleBoolean;
import com.bitgirder.mingle.model.MingleInt32;
import com.bitgirder.mingle.model.MingleInt64;
import com.bitgirder.mingle.model.MingleFloat;
import com.bitgirder.mingle.model.MingleDouble;
import com.bitgirder.mingle.model.MingleString;
import com.bitgirder.mingle.model.MingleEnum;
import com.bitgirder.mingle.model.MingleException;
import com.bitgirder.mingle.model.MingleBuffer;
import com.bitgirder.mingle.model.MingleTimestamp;
import com.bitgirder.mingle.model.MingleList;

import com.bitgirder.mingle.codec.MingleEncoder;

import java.util.Deque;
import java.util.Iterator;

import java.nio.ByteBuffer;

final
class MingleBinaryEncoder
implements MingleEncoder
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final Object obj;
    private final ProgressCheck pc = new ProgressCheck( "encoder" );

    private Deque< Writer > writers;

    MingleBinaryEncoder( Object obj )
    {
        this.obj = state.notNull( obj, "obj" );

        inputs.isTrue(
            obj instanceof MingleStruct,
            "Unrecognized encodable:", obj.getClass().getName()
        );
    }

    private
    static
    abstract
    class Writer
    {
        abstract
        boolean
        implWriteTo( ByteBuffer bb )
            throws Exception;

        final
        boolean
        writeTo( ByteBuffer bb )
            throws Exception
        {
            return implWriteTo( bb );
        }
    }

    private
    static
    abstract
    class FixedRecWriter
    extends Writer
    {
        private final int recLen;

        private FixedRecWriter( int recLen ) { this.recLen = recLen; }

        abstract
        void
        writeRec( ByteBuffer bb )
            throws Exception;

        final
        boolean
        implWriteTo( ByteBuffer bb )
            throws Exception
        {
            if ( bb.remaining() >= recLen ) 
            {
                writeRec( bb );
                return true;
            }
            else return false;
        }
    }

    private
    final
    class SymbolMapWriter
    extends Writer
    {
        private final MingleSymbolMap m;
        private Iterator< MingleIdentifier > flds;

        private
        SymbolMapWriter( MingleSymbolMap m )
        {
            this.m = m;
            this.flds = m.getFields().iterator();
        }

        private
        boolean
        pushNextPair()
            throws Exception
        {
            while ( flds.hasNext() )
            {
                MingleIdentifier id = flds.next();
                MingleValue mv = m.get( id );

                if ( mv != null && ( ! ( mv instanceof MingleNull ) ) )
                {
                    pushValue( mv );
                    pushIdentifier( id );

                    return true;
                }
            }

            return false;
        }

        final
        boolean
        implWriteTo( ByteBuffer bb )
            throws Exception
        {
            if ( flds == null ) return true;
            else
            {
                if ( ! pushNextPair() )
                {
                    pushEnd();
                    flds = null;
                }

                return false;
            }
        }
    }

    private
    final
    class ListWriter
    extends Writer
    {
        private final Iterator< MingleValue > it;

        private boolean sentEnd;

        private ListWriter( MingleList ml ) { it = ml.iterator(); }

        boolean
        implWriteTo( ByteBuffer bb )
            throws Exception
        {
            boolean res;

            if ( it.hasNext() ) 
            {
                pushValue( it.next() );
                return false;
            }
            else 
            {
                if ( sentEnd ) return true;
                else
                {
                    pushEnd();
                    sentEnd = true;

                    return false;
                }
            }
        }
    }

    private void pushWriter( Writer w ) { writers.push( w ); }

    private
    void
    pushInt32( final int i )
    {
        pushWriter( new FixedRecWriter( 4 ) {
            void writeRec( ByteBuffer bb ) { bb.putInt( i ); }
        });
    }

    private
    void
    pushTypeCode( final byte code )
    {
        pushWriter( new FixedRecWriter( 1 ) {
            void writeRec( ByteBuffer bb ) { bb.put( code ); }
        });
    }

    private void pushEnd() { pushTypeCode( TYPE_CODE_END ); }

    // Does not push the typed buffer value prefixed by TYPE_CODE_BUFFER; simply
    // pushes the contents of buf directly into the output
    private
    void
    pushBufferData( final ByteBuffer buf )
    {
        pushWriter( new Writer() {
            boolean implWriteTo( ByteBuffer bb )
            {
                IoUtils.copy( buf, bb );
                return ! buf.hasRemaining();
            }
        });
    }

    private
    void
    pushSizedBuffer( ByteBuffer bb )
    {
        pushBufferData( bb );
        pushInt32( bb.remaining() );
    }

    private
    void
    pushUtf8String( CharSequence str )
        throws Exception
    {
        ByteBuffer bb = Charsets.UTF_8.asByteBuffer( str );

        pushSizedBuffer( bb );
        pushTypeCode( TYPE_CODE_UTF8_STRING );
    }

    private
    void
    pushIdentifier( MingleIdentifier id )
        throws Exception
    {
        pushUtf8String( id.getExternalForm() );
    }

    private
    void
    pushTypeReference( MingleTypeReference typ )
        throws Exception
    {
        pushUtf8String( typ.getExternalForm() );
    }

    private
    void
    pushFields( MingleSymbolMap flds )
    {
        pushWriter( new SymbolMapWriter( flds ) );
    }

    private
    void
    pushSymbolMap( MingleSymbolMap m )
    {
        pushFields( m );
        pushTypeCode( TYPE_CODE_SYMBOL_MAP );
    }

    private
    void
    pushStructure( MingleStructure ms,
                   byte code )
        throws Exception
    {
        pushFields( ms.getFields() );
        pushTypeReference( ms.getType() );
        pushInt32( -1 );
        pushTypeCode( code );
    }

    private
    void
    pushStruct( MingleStruct ms )
        throws Exception
    {
        pushStructure( ms, TYPE_CODE_STRUCT );
    }

    private
    void
    pushException( MingleException me )
        throws Exception
    {
        pushStructure( me, TYPE_CODE_EXCEPTION );
    }

    private
    void
    pushEnum( MingleEnum me )
        throws Exception
    {
        pushIdentifier( me.getValue() );
        pushTypeReference( me.getType() );
        pushTypeCode( TYPE_CODE_ENUM );
    }

    private
    void
    pushList( MingleList ml )
    {
        pushWriter( new ListWriter( ml ) );
        pushInt32( -1 ); // not handling list size hints yet
        pushTypeCode( TYPE_CODE_LIST );
    }

    private
    void
    pushTimestamp( MingleTimestamp ts )
        throws Exception
    {
        pushUtf8String( ts.getRfc3339String( 9 ) );
        pushTypeCode( TYPE_CODE_RFC3339_STR );
    }

    private
    void
    doPushValue( byte code,
                 Writer w )
    {
        pushWriter( w );
        pushTypeCode( code );
    }

    private
    void
    pushBoolean( final MingleBoolean mb )
    {
        doPushValue( TYPE_CODE_BOOLEAN, new FixedRecWriter( 1 ) {
            void writeRec( ByteBuffer bb ) { 
                bb.put( mb.booleanValue() ? (byte) 1 : (byte) 0 );
            }
        });
    }

    private
    void
    pushInt32( final MingleInt32 mi )
    {
        doPushValue( TYPE_CODE_INT32, new FixedRecWriter( 4 ) {
            void writeRec( ByteBuffer bb ) { bb.putInt( mi.intValue() ); }
        });
    }

    private
    void
    pushInt64( final MingleInt64 mi )
    {
        doPushValue( TYPE_CODE_INT64, new FixedRecWriter( 8 ) {
            void writeRec( ByteBuffer bb ) { bb.putLong( mi.longValue() ); }
        });
    }

    private
    void
    pushFloat( final MingleFloat mf )
    {
        doPushValue( TYPE_CODE_FLOAT, new FixedRecWriter( 4 ) {
            void writeRec( ByteBuffer bb ) { bb.putFloat( mf.floatValue() ); }
        });
    }

    private
    void
    pushDouble( final MingleDouble md )
    {
        doPushValue( TYPE_CODE_DOUBLE, new FixedRecWriter( 8 ) {
            void writeRec( ByteBuffer bb ) { bb.putDouble( md.doubleValue() ); }
        });
    }

    private
    void
    pushString( MingleString ms )
        throws Exception
    {
        pushUtf8String( ms );
    }

    private
    void
    pushBuffer( MingleBuffer mb )
    {
        ByteBuffer bb = mb.getByteBuffer();

        pushSizedBuffer( bb );
        pushTypeCode( TYPE_CODE_BUFFER );
    }

    private void pushNull() { pushTypeCode( TYPE_CODE_NULL ); }

    private
    void
    pushValue( MingleValue mv )
        throws Exception
    {
        if ( mv instanceof MingleBoolean ) pushBoolean( (MingleBoolean) mv );
        else if ( mv instanceof MingleInt32 ) pushInt32( (MingleInt32) mv );
        else if ( mv instanceof MingleInt64 ) pushInt64( (MingleInt64) mv );
        else if ( mv instanceof MingleFloat ) pushFloat( (MingleFloat) mv );
        else if ( mv instanceof MingleDouble ) pushDouble( (MingleDouble) mv );
        else if ( mv instanceof MingleString ) pushString( (MingleString) mv );
        else if ( mv instanceof MingleBuffer ) pushBuffer( (MingleBuffer) mv );
        else if ( mv instanceof MingleStruct ) pushStruct( (MingleStruct) mv );
        else if ( mv instanceof MingleEnum ) pushEnum( (MingleEnum) mv );
        else if ( mv instanceof MingleList ) pushList( (MingleList) mv );
        else if ( mv instanceof MingleTimestamp )
        {
            pushTimestamp( (MingleTimestamp) mv );
        }
        else if ( mv instanceof MingleSymbolMap )
        {
            pushSymbolMap( (MingleSymbolMap) mv );
        }
        else if ( mv instanceof MingleException ) 
        {
            pushException( (MingleException) mv );
        }
        else if ( mv instanceof MingleNull ) pushNull();
        else state.fail( "Unhandled value type:", mv.getClass().getName() );
    }

    private
    void
    initWriters()
        throws Exception
    {
        writers = Lang.newDeque();

        if ( obj instanceof MingleStruct ) pushStruct( (MingleStruct) obj );
        else state.fail( "Unhandled encodable:", obj );
    }

    private
    boolean
    runWriteLoop( ByteBuffer bb )
        throws Exception
    {
        state.isFalse( 
            writers.isEmpty(), "Encoder is complete (writer stack is empty)" );

        ByteBuffer bb2 = MingleBinaryCodecs.byteOrdered( bb );

        boolean loop = true;
        do
        {
            Writer w = writers.peek();

            loop = w.writeTo( bb2 );

            // else branch resets loop to true if a new writer was pushed
            if ( loop ) state.equal( w, writers.pop() );
            else loop = writers.peek() != w;
        }
        while ( bb.hasRemaining() && loop && ( ! writers.isEmpty() ) );

        bb.position( bb2.position() );

        return loop;
    }

    public
    boolean
    writeTo( ByteBuffer bb )
        throws Exception
    {
        inputs.notNull( bb, "bb" );

        if ( writers == null ) initWriters();

        pc.enter( bb );
        boolean res = runWriteLoop( bb );
        pc.assertProgress( bb );
        
        return res && writers.isEmpty();
    }

    // Todo:
    //  
    //  - Add explicit type code for utf8 string
}
