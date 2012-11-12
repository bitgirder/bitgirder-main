package com.bitgirder.json;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.io.Charsets;

import java.util.Deque;
import java.util.Map;
import java.util.List;
import java.util.Iterator;

import java.io.IOException;

import java.nio.charset.Charset;
import java.nio.charset.CoderResult;
import java.nio.charset.CharsetEncoder;

import java.nio.ByteBuffer;
import java.nio.CharBuffer;

// This may ultimately become an implementation of some more generalized
// Serializer interface, perhaps in bitgirder.io
public
final
class JsonSerializer
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static String START_OBJECT_STR = "{";
    private final static String END_OBJECT_STR = "}";
    private final static String START_LIST_STR = "[";
    private final static String END_LIST_STR = "]";
    private final static String NAME_SEPARATOR_STR = ":";
    private final static String VALUE_SEPARATOR_STR = ",";

    private
    static 
    enum Mode
    {
        FILL_CHAR_BUF, 
        DRAIN_CHAR_BUF, 
        FLUSH_ENCODER, 
        COMPLETE 
    }

    private final Deque< Object > serStack = Lang.newDeque();
    private final Deque< CharSequenceWriter > strQueue = Lang.newDeque();

    private final CharBuffer cb;

    private final CharsetEncoder encoder;

    private Mode mode = Mode.FILL_CHAR_BUF;

    private 
    JsonSerializer( JsonText text,
                    Options opts )
    {
        this.encoder = opts.charset.newEncoder();

        if ( opts.serialSuffix != null )
        {
            serStack.addFirst( new CharSequenceWriter( opts.serialSuffix ) );
        }

        // We do this manually rather than delegate to pushSerializer (which
        // would otherwise be correct) since pushSerializer throws a checked
        // exception that will not occur here
        if ( text instanceof JsonObject ) 
        {
            pushObjectSerializer( (JsonObject) text );
        }
        else pushArraySerializer( (JsonArray) text );

        this.cb = CharBuffer.allocate( opts.charBufLen );
    }

    public float maxBytesPerChar() { return encoder.maxBytesPerChar(); }

    private
    void
    dumpState( String msg )
    {
        StringBuilder sb = 
            new StringBuilder( msg ).
                append( ": " );
        
        sb.append(
            Strings.crossJoin( "=", ", ",
                "serStack", serStack,
                "strQueue", strQueue,
                "cb.position()", cb.position(),
                "cb.limit()", cb.limit(),
                "cb.capacity()", cb.capacity(),
                "mode", mode
            ) );
        
        code( sb );
    }

    private
    final
    static
    class CharSequenceWriter
    {
        private final CharSequence str;
        private final int end;

        private int indx;

        private 
        CharSequenceWriter( CharSequence str )
        {
            this.str = str;
            this.end = str.length();
        }
        
        private
        boolean
        writeTo( CharBuffer cb )
        {
            while ( cb.hasRemaining() && indx < end )
            {
                cb.put( str.charAt( indx++ ) );
            }

            return indx == end;
        }
    }

    private
    void
    addString( CharSequence str )
    {
        strQueue.add( new CharSequenceWriter( str ) );
    }

    private
    void
    addJsonString( JsonString jsonStr )
        throws IOException
    {
        addString( jsonStr.getExternalForm() );
    }

    private
    final
    static
    class ObjectSerializer
    {
        private final Iterator< Map.Entry< JsonString, List< JsonValue > > > 
            membersIt;

        private JsonString curKey;
        private Iterator< JsonValue > curVals;

        private
        ObjectSerializer(
            Iterator< Map.Entry< JsonString, List< JsonValue > > > membersIt )
        {
            this.membersIt = membersIt;
        }
    }

    private
    final
    static
    class ArraySerializer
    {
        private final Iterator< JsonValue > vals;

        private 
        ArraySerializer( Iterator< JsonValue > vals ) 
        { 
            this.vals = vals;
        }
    }

    private
    void
    pushObjectSerializer( JsonObject obj )
    {
        ObjectSerializer os = new ObjectSerializer( obj.entrySet().iterator() );
        addString( START_OBJECT_STR );
        
        if ( advanceState( os ) ) serStack.addFirst( os );
        else addString( END_OBJECT_STR );
    }

    private
    void
    pushArraySerializer( JsonArray arr )
    {
        ArraySerializer as = new ArraySerializer( arr.iterator() );

        addString( START_LIST_STR );

        if ( as.vals.hasNext() ) serStack.addFirst( as );
        else addString( END_LIST_STR );
    }

    private
    void
    pushSerializer( JsonValue val )
        throws IOException
    {
        if ( val instanceof JsonString || 
             val instanceof JsonNumber ||
             val instanceof JsonNull ||
             val instanceof JsonBoolean )
        {
            // no need to push to serStack; just add string to write queue
            addStringQueue( val ); 
        }
        else if ( val instanceof JsonObject )
        {
            pushObjectSerializer( (JsonObject) val );
        }
        else if ( val instanceof JsonArray ) 
        {
            pushArraySerializer( (JsonArray) val );
        }
        else
        {
            throw state.createFail(
                "Unexpected argument type to pushSerializer():", 
                val.getClass() );
        }
    }

    private
    boolean
    advanceState( ObjectSerializer os )
    {
        boolean res;

        if ( os.curVals != null && os.curVals.hasNext() ) res = true;
        else
        {
            os.curVals = null;
            os.curKey = null;

            if ( os.membersIt.hasNext() )
            {
                Map.Entry< JsonString, List< JsonValue > > e = 
                    os.membersIt.next();

                os.curKey = e.getKey();
                os.curVals = e.getValue().iterator();

                res = true;
            }
            else res = false;
        }

        return res;
    }

    private
    void
    addStringQueue( ObjectSerializer os )
        throws IOException
    {
        addJsonString( os.curKey );
        addString( NAME_SEPARATOR_STR );
 
        JsonValue val = os.curVals.next();

        if ( advanceState( os ) )
        {
            serStack.addFirst( os );
            serStack.addFirst( VALUE_SEPARATOR_STR );
        }
        else serStack.addFirst( END_OBJECT_STR );

        pushSerializer( val );
    }

    private
    void
    addStringQueue( ArraySerializer as )
        throws IOException
    {
        JsonValue val = as.vals.next();

        if ( as.vals.hasNext() ) 
        {
            serStack.addFirst( as );
            serStack.addFirst( VALUE_SEPARATOR_STR );
        }
        else serStack.addFirst( END_LIST_STR );
        
        pushSerializer( val );
    }

    private
    void
    addStringQueue( JsonNumber num )
    {
        addString( num.getNumber().toString() );
    }

    private 
    void 
    addStringQueue( JsonString str ) 
        throws IOException
    { 
        addJsonString( str ); 
    }

    private void addStringQueue( JsonBoolean b ) { addString( b.toString() ); }
    private void addStringQueue( JsonNull n ) { addString( "null" ); }

    // elt is something that was in or could be placed in serStack
    private
    void
    addStringQueue( Object elt )
        throws IOException
    {
        if ( elt == END_OBJECT_STR ) addString( END_OBJECT_STR );
        else if ( elt == END_LIST_STR ) addString( END_LIST_STR );
        else if ( elt == VALUE_SEPARATOR_STR ) addString( VALUE_SEPARATOR_STR );
        else if ( elt instanceof ObjectSerializer )
        {
            addStringQueue( (ObjectSerializer) elt );
        }
        else if ( elt instanceof ArraySerializer )
        {
            addStringQueue( (ArraySerializer) elt );
        }
        else if ( elt instanceof JsonNumber )
        {
            addStringQueue( (JsonNumber) elt );
        }
        else if ( elt instanceof JsonString ) 
        {
            addStringQueue( (JsonString) elt );
        }
        else if ( elt instanceof JsonBoolean )
        {
            addStringQueue( (JsonBoolean) elt );
        }
        else if ( elt instanceof JsonNull ) addStringQueue( (JsonNull) elt );
        else if ( elt instanceof CharSequenceWriter )
        {
            strQueue.add( (CharSequenceWriter) elt );
        }
        else 
        {
            throw state.createFail( 
                "Unexpected serializer stack element:", elt );
        }
    }

    private
    void
    drainStringQueue()
    {
        state.isFalse( strQueue.isEmpty() );

        while ( cb.hasRemaining() && ! strQueue.isEmpty() )
        {
            if ( strQueue.peekFirst().writeTo( cb ) ) strQueue.removeFirst();
            else state.isFalse( cb.hasRemaining() );
        }

        if ( ( ! cb.hasRemaining() ) ||
             ( strQueue.isEmpty() && serStack.isEmpty() ) )
        {
            cb.flip();
            mode = Mode.DRAIN_CHAR_BUF;
        }
    }

    private
    void
    fillCharBuf()
        throws IOException
    {
//        dumpState( "Entering fillCharBuf()" );

        while ( cb.hasRemaining() && mode == Mode.FILL_CHAR_BUF )
        {
            while ( strQueue.isEmpty() ) 
            {
                // Always remove the top element, which may be re-added in the
                // actual type-specific handler method
                Object top = serStack.removeFirst();
                addStringQueue( top );

                // Assert that the loop makes progress
                state.isFalse( 
                    serStack.peekFirst() == top && strQueue.isEmpty() );
            }

            drainStringQueue();
        }
    }

    private
    boolean
    drainCharBuf( ByteBuffer dest )
        throws IOException
    {
//        dumpState( "Entering drainCharBuf() with dest: " + dest );

        boolean endOfInput = serStack.isEmpty() && strQueue.isEmpty();

        CoderResult cr = encoder.encode( cb, dest, endOfInput );

        if ( cr.isError() ) cr.throwException();
        else if ( cr.isUnderflow() ) 
        {
            if ( endOfInput ) mode = Mode.FLUSH_ENCODER;
            else 
            {
                state.isFalse( cb.hasRemaining() );

                cb.clear();
                mode = Mode.FILL_CHAR_BUF;
            }
        }
        else state.isTrue( cr.isOverflow() );

        return cr.isOverflow();
    }

    private
    void
    flushEncoder( ByteBuffer dest )
        throws IOException
    {
        CoderResult cr = encoder.flush( dest );

        if ( cr.isError() ) cr.throwException();
        else if ( cr.isUnderflow() ) mode = Mode.COMPLETE;
        else
        {
            state.isTrue( cr.isOverflow() );
            state.isFalse( dest.hasRemaining() );
        }
    }

    private
    final
    static
    class LoopContext
    {
        private boolean isProgressing;
        private int remain;
        private Mode mode;

        private
        LoopContext( ByteBuffer dest,
                     Mode mode )
        {
            this.remain = dest.remaining();
            this.mode = mode;

            isProgressing = true; // to start
        }

        private
        void
        update( ByteBuffer dest,
                Mode mode )
        {
            isProgressing = remain > dest.remaining() || this.mode != mode;

            remain = dest.remaining();
            this.mode = mode;
        }
    }

    // After loop we make one call to flushEncoder if mode is FLUSH_ENCODER,
    // whether of not there is room in the dest buffer. It can happen that if
    // dest is exactly the size of the remaining output it will be full after
    // the loop and the encoder will be requiring a flush call, even though it
    // will ultimately produce no data. Note that we don't make this call inside
    // the loop since it could lead to an infinite loop in cases where the dest
    // buffer is empty but the flush actually does have data to write.
    public
    boolean
    writeTo( ByteBuffer dest )
        throws IOException
    {
        LoopContext lc = new LoopContext( dest, mode );
        boolean isOverflow = false;

        while ( dest.hasRemaining() && mode != Mode.COMPLETE && 
                lc.isProgressing && ( ! isOverflow ) )
        {
            switch ( mode )
            {
                case FILL_CHAR_BUF: fillCharBuf(); break;
                case DRAIN_CHAR_BUF: isOverflow = drainCharBuf( dest ); break;
                case FLUSH_ENCODER: flushEncoder( dest ); break;
                case COMPLETE: state.fail();
            }

            lc.update( dest, mode );
        }

        state.isTrue( lc.isProgressing, "Loop did not make progress" );

        if ( mode == Mode.FLUSH_ENCODER ) flushEncoder( dest );
        return mode == Mode.COMPLETE;
    }

    public
    static
    JsonSerializer
    create( JsonText text,
            Options opts )
    {
        inputs.notNull( text, "text" );
        inputs.notNull( opts, "opts" );

        return new JsonSerializer( text, opts );
    }

    public 
    static 
    JsonSerializer 
    create( JsonText text )
    {
        return create( text, Options.getDefault() );
    }
    
    public
    final
    static
    class Options
    {
        private final static int DEFAULT_CHAR_BUF_LEN = 256;
        private final static Charset DEFAULT_CHARSET = Charsets.UTF_8.charset();

        private final static Options DEFAULT_OPTS =
            new Builder().
                setCharset( Charsets.UTF_8.charset() ).
                build();

        private final Charset charset;
        private final int charBufLen;
        private final String serialSuffix;

        private
        Options( Builder b )
        {
            this.charset = inputs.notNull( b.charset, "charset" );
            this.charBufLen = b.charBufLen;
            this.serialSuffix = b.serialSuffix;
        }

        public static Options getDefault() { return DEFAULT_OPTS; }

        public
        final
        static
        class Builder
        {
            private Charset charset = DEFAULT_CHARSET;
            private int charBufLen = DEFAULT_CHAR_BUF_LEN;
            private String serialSuffix;
    
            public
            Builder
            setCharset( Charset charset )
            {
                this.charset = inputs.notNull( charset, "charset" );
                return this;
            }
    
            public
            Builder
            setCharBufLen( int len )
            {
                this.charBufLen = inputs.positiveI( len, "len" );
                return this;
            }

            public
            Builder
            setSerialSuffix( String serialSuffix )
            {
                this.serialSuffix = 
                    inputs.notNull( serialSuffix, "serialSuffix" );

                return this;
            }
     
            public Options build() { return new Options( this ); }
        }
    }
}
