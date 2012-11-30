package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.PatternHelper;

import java.net.URL;

import java.util.List;
import java.util.Enumeration;
import java.util.Properties;
import java.util.Iterator;

import java.util.regex.Pattern;

import java.io.InputStream;
import java.io.ByteArrayOutputStream;
import java.io.ByteArrayInputStream;
import java.io.ObjectOutputStream;
import java.io.ObjectInputStream;
import java.io.IOException;
import java.io.EOFException;
import java.io.File;
import java.io.Closeable;
import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.io.OutputStream;
import java.io.StringReader;
import java.io.Reader;
import java.io.Serializable;

import java.nio.ByteBuffer;
import java.nio.CharBuffer;

import java.nio.charset.CharsetDecoder;
import java.nio.charset.CharacterCodingException;

public
class IoUtils
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static char[] HEX_CHARS = "0123456789abcdef".toCharArray();
 
    private final static DataSize DEFAULT_BUF_SIZE = DataSize.ofKilobytes( 2 );

    private final static ByteBuffer EMPTY_BYTE_BUFFER =
        ByteBuffer.allocate( 0 ).asReadOnlyBuffer();

    private final static String ENV_PATH = "PATH";

    private final static Pattern PAT_PATH_SEP = 
        PatternHelper.compile( File.pathSeparator );

    public
    static
    int
    arrayPosOf( ByteBuffer bb )
    {
        inputs.notNull( bb, "bb" );
        return bb.arrayOffset() + bb.position();
    }

    public
    static
    ByteBuffer
    shiftPos( ByteBuffer bb,
              int diff )
    {
        inputs.notNull( bb, "bb" );
        bb.position( bb.position() + diff );

        return bb;
    }

    public
    static
    ByteBuffer
    shiftLimit( ByteBuffer bb,
                int diff )
    {
        inputs.notNull( bb, "bb" );
        bb.limit( bb.limit() + diff );

        return bb;
    }

    public
    static
    ByteBuffer
    bzero( ByteBuffer bb )
    {
        inputs.notNull( bb, "bb" );

        int start = bb.position();
        while ( bb.hasRemaining() ) bb.put( (byte) 0 );

        bb.position( start );

        return bb;
    }

    // Assumed that caller has checked hasArray() already
    public
    static
    void
    write( OutputStream os,
           ByteBuffer buf )
        throws IOException
    {
        inputs.notNull( os, "os" );
        inputs.notNull( buf, "buf" );

        os.write( buf.array(), arrayPosOf( buf ), buf.remaining() );
    }

    private
    static
    void
    checkAllocatable( DataSize sz,
                      String paramName )
    {
        inputs.notNull( sz, paramName );
        
        long szBytes = sz.getByteCount();

        inputs.isTrue( 
            szBytes <= Integer.MAX_VALUE,
            paramName, "is too large to allocate:", szBytes );
    }
 
    public
    static
    ByteBuffer
    allocateByteBuffer( DataSize sz )
    {
        checkAllocatable( sz, "sz" );
        return ByteBuffer.allocate( (int) sz.getByteCount() );
    }

    public static ByteBuffer emptyByteBuffer() { return EMPTY_BYTE_BUFFER; }

    public
    static
    byte[]
    dup( byte[] buf )
    {
        inputs.notNull( buf, "buf" );

        byte[] res = new byte[ buf.length ];
        System.arraycopy( buf, 0, res, 0, buf.length );

        return res;
    }

    public
    static
    byte[]
    toByteArray( InputStream is )
        throws IOException
    {
        return toByteArray( is, false );
    }

    public
    static
    byte[]
    toByteArray( InputStream is,
                 boolean close )
        throws IOException
    {
        inputs.notNull( is, "is" );
        
        try
        {
            ByteArrayOutputStream bos = new ByteArrayOutputStream();
            
            byte[] buf = new byte[ (int) DEFAULT_BUF_SIZE.getByteCount() ];
            for ( int i = is.read( buf ); i >= 0; i = is.read( buf ) )
            {
                bos.write( buf, 0, i );
            }
    
            return bos.toByteArray();
        }
        finally { if ( close ) closeQuietly( is, is.toString() ); }
    }

    public
    static
    ByteBuffer
    toByteBuffer( InputStream is,
                  boolean close )
        throws IOException
    {
        return ByteBuffer.wrap( toByteArray( is, close ) );
    }

    public
    static
    ByteBuffer
    toByteBuffer( InputStream is ) 
        throws IOException
    {
        return toByteBuffer( is, false ); 
    }

    // closes is regardless of how this method completes
    public
    static
    List< String >
    readLines( InputStream is,
               String charset )
        throws IOException
    {
        inputs.notNull( is, "is" );
        try
        {
            inputs.notNull( charset, "charset" );
    
            List< String > res = Lang.newList();
    
            BufferedReader br = 
                new BufferedReader( new InputStreamReader( is, charset ) );
            
            String line;
            while ( ( line = br.readLine() ) != null ) res.add( line );
    
            return res;
        }
        finally { is.close(); }
    }

    // Fills dest from is, throwing EOFException if there is not sufficient data
    // in the stream to do so
    public
    static
    void
    fill( InputStream is,
          byte[] dest,
          int offset,
          int len )
        throws IOException
    {
        inputs.notNull( is, "is" );
        inputs.notNull( dest, "dest" );
        inputs.nonnegativeI( offset, "offset" );
        inputs.nonnegativeI( len, "len" );

        inputs.isTruef( 
            offset + len <= dest.length,
            "offset (%d) + length (%d) > dest.length (%d)", 
            offset, len, dest.length );

        while ( len > 0 )
        {
            int i = is.read( dest, offset, len ); 
            if ( i < 0 ) throw new EOFException(); 
            else
            {
                offset += i;
                len -= i;
            }
        }
    }

    public
    static
    void
    fill( InputStream is,
          byte[] dest )
        throws IOException
    {
        inputs.notNull( dest, "dest" );
        fill( is, dest, 0, dest.length );
    }

    public
    static
    < V >
    V
    digest( OctetDigest< ? extends V > dig,
            boolean consumeBufs,
            Iterable< ? extends ByteBuffer > bufs )
        throws Exception
    {
        inputs.notNull( dig, "dig" );
        inputs.noneNull( bufs, "bufs" );

        for ( ByteBuffer buf : bufs )
        {
            if ( ! consumeBufs ) buf = buf.slice();
            dig.update( buf );
        }

        return dig.digest();
    }

    public
    static
    < V >
    V
    digest( OctetDigest< ? extends V > dig,
            boolean consumeBufs,
            ByteBuffer... bufs )
        throws Exception
    {
        inputs.notNull( bufs, "bufs" );
        return digest( dig, consumeBufs, Lang.asList( bufs ) );
    }

    static
    long
    dumpCrc32( ByteBuffer bb )
    {
        inputs.notNull( bb, "bb" );

        Crc32Digest dig = Crc32Digest.create();

        dig.update( bb.slice() );
        return dig.digest();
    }

    // Catches and warns about any Exception, not just IOException
    public
    static
    void
    closeQuietly( Closeable c,
                  String errMsgId )
    {
        inputs.notNull( c, "c" );
        inputs.notNull( errMsgId, "errMsgId" );

        try { c.close(); }
        catch ( Exception ioe )
        {
            CodeLoggers.warn( ioe, "Exception thrown from close: " + errMsgId );
        }
    }

    public
    static
    ByteBuffer
    copyOf( ByteBuffer buf )
    {
        inputs.notNull( buf, "buf" );

        ByteBuffer res = ByteBuffer.allocate( buf.remaining() );
        res.put( buf.slice() );
        res.flip();

        return res;
    }

    public
    static
    int
    copy( ByteBuffer src,
          ByteBuffer dest )
    {
        int copyLen = Math.min( src.remaining(), dest.remaining() );

        ByteBuffer src2 = src.slice();
        src2.limit( copyLen );
        dest.put( src2 );

        src.position( src.position() + copyLen );

        return copyLen;
    }

    // Note: drains bb itself (bb.remaining() will be false on return from this
    // method)
    public
    static
    byte[]
    toByteArray( ByteBuffer bb )
    {
        inputs.notNull( bb, "bb" );

        byte[] res = new byte[ bb.remaining() ];
        bb.get( res );

        return res;
    }
 
    // Returns a new buffer with the contents of bb.position() up to bb.limit()
    // copied into the beginning of the result buffer, which will have size
    // double the size of bb. The result buffer will have position() set to
    // bb.limit() - bb.position(), ie, just past the last byte of input, and
    // will have its limit set to its capacity. This is to allow for the most
    // common use of this method, which is to read data into a buffer, discover
    // that more space is needed, double the size of the buffer, and resume
    // reading:
    //
    //      ByteBuffer bb = ByteBuffer.allocate( 128 );
    // 
    //      // Assume readDataToBuffer() is some method which reads data into
    //      // its input buffer, returning true if there is still data to be
    //      // read
    //      while ( readDataToBuffer( bb ) )
    //      {
    //          bb.flip(); // to get position/limit correct for expand
    //          bb = IoUtils.expand( bb );
    //      }
    //          
    // The input buffer bb will itself be unchanged.
    //
    public
    static
    ByteBuffer
    expand( ByteBuffer bb )
    {
        inputs.notNull( bb, "bb" );

        long dblSz = ( (long) bb.capacity() ) * 2L;

        inputs.isTrue(
            dblSz <= Integer.MAX_VALUE,
            "Expanding buffer to size", dblSz,
            "would overflow representable buffer size" );
 
        return ByteBuffer.allocate( (int) dblSz ).put( bb.slice() );
    }

    // Workaround for http://bugs.sun.com/bugdatabase/view_bug.do?bug_id=4997655
    //
    // This method should in all cases be used wherever CharBuffer.wrap would be
    public
    static
    CharBuffer
    charBufferFor( CharSequence s )
    {
        inputs.notNull( s, "s" );
        
        CharBuffer res = CharBuffer.allocate( s.length() );
        res.append( s );
        res.flip();

        return res.asReadOnlyBuffer();
    }

    // suffix may be null, see javadocs for File.createTempFile
    public
    static
    FileWrapper
    createTempFile( String prefix,
                    String suffix,
                    boolean deleteOnExit )
        throws IOException
    {
        inputs.notNull( prefix, "prefix" );

        File tmp = File.createTempFile( prefix, suffix );
        if ( deleteOnExit ) tmp.deleteOnExit();

        return new FileWrapper( tmp );
    }
 
    public
    static
    FileWrapper
    createTempFile( String prefix,
                    boolean deleteOnExit )
        throws IOException
    {
        return createTempFile( prefix, null, deleteOnExit );
    }

    // Does not modify bb
    public
    static
    CharSequence
    asHexString( ByteBuffer bb )
    {
        inputs.notNull( bb, "bb" );
        ByteBuffer slice = bb.slice();

        CharBuffer res = CharBuffer.allocate( slice.remaining() * 2 );

        while ( slice.hasRemaining() )
        {
            byte b = slice.get();
 
            res.put( HEX_CHARS[ ( b >> 4 ) & 0x0f ] );
            res.put( HEX_CHARS[ b & 0x0f ] );
        }

        res.flip();
        return res.asReadOnlyBuffer();
    }

    public
    static
    CharSequence
    asHexString( byte[] arr )
    {
        return asHexString( ByteBuffer.wrap( inputs.notNull( arr, "arr" ) ) );
    }

    // could make this public at some point
    private
    static
    boolean
    isHexChar( char ch )
    {
        return ( ch >= '0' && ch <= '9' ) ||
               ( ch >= 'a' && ch <= 'f' ) ||
               ( ch >= 'A' && ch <= 'F' );
    }

    private
    static
    void
    validateHexString( CharSequence hex )
        throws IOException
    {
        int i = 0;

        for ( int e = hex.length(); i < e; ++i )
        {
            char ch = hex.charAt( i );

            if ( ! isHexChar( ch ) )
            {
                String tmpl = "Not a hex char at index %d: '%c' (0x%02x)";
                String msg = String.format( tmpl, i, ch, (int) ch );
                throw new IOException( msg );
            }
        }

        if ( i % 2 == 1 ) 
        {
            throw new IOException( "Invalid odd length for hex string: " + i );
        }
    }

    public
    static
    byte[]
    hexToByteArray( CharSequence hex )
        throws IOException
    {
        inputs.notNull( hex, "hex" );

        validateHexString( hex );

        byte[] res = new byte[ hex.length() / 2 ];

        for ( int i = 0, e = hex.length(); i < e; i += 2 )
        {
            String s = hex.subSequence( i, i + 2 ).toString();
            res[ i / 2 ] = (byte) Integer.parseInt( s, 16 );
        }

        return res;
    }

    public
    static
    int
    remaining( Iterable< ? extends ByteBuffer > bufs )
    {
        inputs.noneNull( bufs, "bufs" );

        int res = 0;
        for ( ByteBuffer buf : bufs ) res += buf.remaining();

        return res;
    }

    public
    static
    int
    remaining( ByteBuffer... bufs )
    {
        inputs.notNull( bufs, "bufs" );
        return remaining( Lang.asList( bufs ) );
    }

    public
    static
    CharSequence
    asString( Iterable< ? extends ByteBuffer > line,
              CharsetDecoder dec )
        throws CharacterCodingException
    {
        inputs.noneNull( line, "line" );
        inputs.notNull( dec, "dec" );

        int len = remaining( line );
        
        ByteBuffer acc = ByteBuffer.allocate( len );
        for ( ByteBuffer bb : line ) acc.put( bb );

        acc.flip();
        return dec.decode( acc );
    }

    // Only tested now for unix-like systems
    public
    static
    String
    which( CharSequence cmd )
    {
        String cmdStr = inputs.notNull( cmd, "cmd" ).toString();

        String path = System.getenv( ENV_PATH );

        if ( path != null ) 
        {
            for ( String dir : PAT_PATH_SEP.split( path ) )
            {
                File f = new File( dir, cmdStr );
                if ( f.exists() && f.canExecute() ) return f.toString();
            }
        }

        return null; // if we get here
    }

    private
    static
    ClassLoader
    defaultClassLoader()
    {
        return IoUtils.class.getClassLoader();
    }

    public
    static
    List< URL >
    getResources( String name,
                  ClassLoader cl )
        throws IOException
    {
        inputs.notNull( name, "name" );
        inputs.notNull( cl, "cl" );

        Enumeration< URL > en = cl.getResources( name );

        if ( en.hasMoreElements() )
        {
            List< URL > res = Lang.newList();
            while ( en.hasMoreElements() ) res.add( en.nextElement() );

            return res;
        }
        else return Lang.emptyList();
    }

    public
    static
    List< URL >
    getResources( String name )
        throws IOException
    {
        return getResources( name, defaultClassLoader() );
    }

    public
    static
    URL
    expectSingleResource( String name,
                          ClassLoader cl )
        throws IOException
    {
        List< URL > l = getResources( name, cl );

        switch ( l.size() )
        {
            case 0: 
                throw state.createFail( 
                    "No resources found matching name", name );

            case 1: return l.get( 0 );

            default:
                throw state.createFail(
                    "Multiple resources found matching name", name + ":", l );
        }
    }

    public
    static
    URL
    expectSingleResource( String name )
        throws IOException
    {
        return expectSingleResource( name, defaultClassLoader() );
    }

    public
    static
    InputStream
    expectSingleResourceAsStream( String name,
                                  ClassLoader cl )
        throws IOException
    {
        return expectSingleResource( name, cl ).openStream();
    }

    public
    static
    InputStream
    expectSingleResourceAsStream( String name )
        throws IOException
    {
        return expectSingleResourceAsStream( name, defaultClassLoader() );
    }
    
    public
    static
    Properties
    loadProperties( InputStream is )
        throws IOException
    {
        inputs.notNull( is, "is" );

        try
        {
            Properties res = new Properties();
            res.load( is );

            return res;
        }
        finally { is.close(); }
    }

    public
    static
    Properties
    loadProperties( URL url )
        throws IOException
    {
        return loadProperties( inputs.notNull( url, "url" ).openStream() );
    }

    public
    static
    interface ResourceLineVisitor
    {
        public
        void
        visitLine( String line )
            throws Exception;

        public
        Exception
        getResourcesFailed( String rsrcName,
                            Throwable th );
        
        public
        Exception
        openFailed( String rsrcName,
                    Throwable th );

        public
        Exception
        visitFailed( String rsrcName,
                     int line,
                     Throwable th );
    }    

    public
    static
    abstract
    class AbstractResourceLineVisitor
    implements ResourceLineVisitor
    {
        protected
        Exception
        getException( String msg,
                      Throwable th )
        {
            return new Exception( msg + ": " + th.getMessage(), th );
        }

        public
        Exception
        getResourcesFailed( String rsrcName,
                            Throwable th )
        {
            return 
                getException(
                    "Couldn't get resources with name '" + rsrcName + "'", th );
        }

        public
        Exception
        openFailed( String rsrcName,
                    Throwable th )
        {
            return getException( "Couldn't open resource " + rsrcName, th );
        }

        public
        Exception
        visitFailed( String rsrcName,
                     int line,
                     Throwable th )
        {
            return 
                getException( 
                    "Error at line " +  line + " of " + rsrcName, th );
        }
    }

    private
    static
    List< URL >
    getLineVisitURLs( String rsrcName,
                      ResourceLineVisitor v )
        throws Exception
    {
        try { return getResources( rsrcName ); }
        catch ( Throwable th ) { throw v.getResourcesFailed( rsrcName, th ); }
    }

    private
    static
    BufferedReader
    openResource( URL u,
                  String charset,
                  ResourceLineVisitor v )
        throws Exception
    {
        try
        {
            return
                new BufferedReader(
                    new InputStreamReader( u.openStream(), charset ) );
        }
        catch ( Throwable th ) { throw v.openFailed( u.toString(), th ); }
    }

    private
    static
    void
    visitLine( String line,
               ResourceLineVisitor v,
               URL u,
               int lineNum )
        throws Exception
    {
        try { v.visitLine( line ); }
        catch ( Throwable th )
        {
            throw v.visitFailed( u.toString(), lineNum, th ); 
        }
    }

    private
    static
    void
    visitLines( URL u,
                String charset,
                ResourceLineVisitor v )
        throws Exception
    {
        BufferedReader br = openResource( u, charset, v );
        try
        {
            String line = null;
            for ( int i = 1; ( line = br.readLine() ) != null; ++i )
            {
                visitLine( line, v, u, i );
            }
        }
        finally { closeQuietly( br, u.toString() ); }
    }

    public
    static
    void
    visitResourceLines( String rsrcName,
                        String charset,
                        ResourceLineVisitor v )
        throws Exception
    {
        inputs.notNull( rsrcName, "rsrcName" );
        inputs.notNull( charset, "charset" );
        inputs.notNull( v, "v" );

        for ( URL u : getLineVisitURLs( rsrcName, v ) )
        {
            visitLines( u, charset, v );
        }
    }

    private
    final
    static
    class CharReaderImpl
    implements CharReader
    {
        private final Reader rd;

        private int peekVal = Integer.MAX_VALUE;

        private CharReaderImpl( Reader rd ) { this.rd = rd; }

        public
        int
        peek()
            throws IOException
        {
            if ( peekVal == Integer.MAX_VALUE ) peekVal = rd.read();
            return peekVal;
        }

        public
        int
        read()
            throws IOException
        {
            int res = peek();
            
            if ( res >= 0 ) peekVal = Integer.MAX_VALUE; // save -1 when peekVal

            return res;
        }
    }

    public
    static
    CharReader
    charReaderFor( Reader rd )
    {
        return new CharReaderImpl( inputs.notNull( rd, "rd" ) );
    }

    public
    static
    CharReader
    charReaderFor( CharSequence cs )
    {
        inputs.notNull( cs, "cs" );

        return new CharReaderImpl( new StringReader( cs.toString() ) );
    }

    public
    static
    byte[]
    toSerialByteArray( Serializable ser )
        throws IOException
    {
        inputs.notNull( ser, "ser" );

        ByteArrayOutputStream bos = new ByteArrayOutputStream();
        ObjectOutputStream oos = new ObjectOutputStream( bos );
        oos.writeObject( ser );
        oos.close();

        return bos.toByteArray();
    }

    public
    static
    Object
    fromSerialByteArray( byte[] arr )
        throws IOException,
               ClassNotFoundException
    {
        inputs.notNull( arr, "arr" );

        ByteArrayInputStream bis = new ByteArrayInputStream( arr );

        try
        {
            ObjectInputStream ois = new ObjectInputStream( bis );
            return ois.readObject();
        }
        finally { bis.close(); }
    }

    // The methods which follow are support methods for converting java strings
    // to json strings. The method is included here instead of in the JSON libs
    // since JSON strings are also a very convenient, safe, and meaningful way
    // to serialize unicode strings in general

    private
    static
    void
    appendUnicodeEscape( char ch,
                         StringBuilder sb )
    {
        sb.append( "\\u" );

        String numStr = Integer.toString( ch, 16 );

        if ( ch <= '\u000F' ) sb.append( "000" );
        else if ( ch <= '\u00FF' ) sb.append( "00" );
        else if ( ch <= '\u0FFF' ) sb.append( "0" );

        sb.append( Integer.toString( ch, 16 ) );
    }

    private
    static
    void
    verifyAndAppendLowSurrogate( CharSequence cs,
                                 int indx,
                                 StringBuilder sb )
    {
        if ( indx < cs.length() )
        {
            char ch = cs.charAt( indx );

            if ( Character.isLowSurrogate( ch ) ) sb.append( ch );
            else
            {
                inputs.fail(
                    "Character at index", indx, "is not a low surrogate " +
                    "but preceding character was a high surrogate" );
            }
        }
        else
        {
            inputs.fail(
                "Unexpected end of string while expecting low surrogate" );
        }
    }

    private
    static
    void
    appendOrdinaryChar( char ch,
                        StringBuilder sb )
    {

        if ( ch == '\u0020' || ch == '\u0021' ||
             ( ch >= '\u0023' && ch <= '\u005b' ) || ch >= '\u005d' )
        {
            sb.append( ch );
        }
        else appendUnicodeEscape( ch, sb );
    }

    private
    static
    void
    appendChar( char ch,
                StringBuilder sb )
    {
        switch ( ch )
        {
            case '"': sb.append( "\\\"" ); break;
            case '\\': sb.append( "\\\\" ); break;
            case '\b': sb.append( "\\b" ); break;
            case '\f': sb.append( "\\f" ); break;
            case '\n': sb.append( "\\n" ); break;
            case '\r': sb.append( "\\r" ); break;
            case '\t': sb.append( "\\t" ); break;

            default: appendOrdinaryChar( ch, sb );
        }
    }

    public
    static
    StringBuilder
    appendRfc4627String( StringBuilder sb,
                         CharSequence str )
    {
        inputs.notNull( sb, "sb" );
        inputs.notNull( str, "str" );

        sb.append( '"' );

        for ( int i = 0, e = str.length(); i < e; )
        {
            char ch = str.charAt( i++ );

            if ( Character.isHighSurrogate( ch ) )
            {
                sb.append( ch );
                verifyAndAppendLowSurrogate( str, i++, sb );
            }
            else appendChar( ch, sb );
        }

        sb.append( '"' );

        return sb;
    }

    public
    static
    CharSequence
    getRfc4627String( CharSequence str )
    {
        return appendRfc4627String( new StringBuilder(), str );
    }
}
