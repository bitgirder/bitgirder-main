package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Completion;
import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;
import com.bitgirder.lang.PatternHelper;
import com.bitgirder.lang.ObjectReceiver;

import com.bitgirder.lang.reflect.ReflectUtils;

import com.bitgirder.test.Test;
import com.bitgirder.test.Before;
import com.bitgirder.test.After;

import com.bitgirder.log.CodeLoggers;

import java.net.URL;

import java.util.Random;
import java.util.Iterator;
import java.util.Arrays;
import java.util.List;
import java.util.Set;
import java.util.Collection;

import java.util.zip.GZIPOutputStream;
import java.util.zip.CRC32;

import java.io.IOException;
import java.io.EOFException;
import java.io.FileOutputStream;
import java.io.File;
import java.io.FileFilter;
import java.io.FilenameFilter;
import java.io.InputStream;
import java.io.ByteArrayInputStream;
import java.io.Serializable;

import java.nio.ByteBuffer;
import java.nio.CharBuffer;

import java.nio.channels.FileChannel;

import java.nio.charset.CharacterCodingException;
import java.nio.charset.UnmappableCharacterException;
import java.nio.charset.MalformedInputException;
import java.nio.charset.CharsetDecoder;

import java.security.MessageDigest;

@Test
public
final
class IoTests
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final static String TEST_RSRC1 = "com/bitgirder/io/test-resource1";
    private final static String TEST_RSRC2 = "com/bitgirder/io/hello.txt";
    
    private static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final Random rand = new Random();

    private final Collection< DirWrapper > toWipe =
        Lang.newConcurrentCollection();

    @After
    private
    void
    wipeDirs()
        throws Exception
    {
        for ( DirWrapper dw : toWipe ) rmRf( dw );
    }

    private
    DirWrapper
    nextTempDir()
        throws Exception
    {
        DirWrapper res = IoTestFactory.createTempDir();
        toWipe.add( res );

        return res;
    }

    @Test
    private
    void
    testDataUnitForString()
    {
        Object[] pairs = {
            DataUnit.BYTE, "b",
            DataUnit.BYTE, "B",
            DataUnit.BYTE, " Bytes",
            DataUnit.BYTE, "BYte ",
            DataUnit.MEGABYTE, "MB ",
            DataUnit.KILOBYTE, " K",
            DataUnit.GIGABYTE, "   GIgabyTE      ",
            DataUnit.TERABYTE, "tb"
        };

        for ( int i = 0; i < pairs.length; i += 2 )
        {
            DataUnit expct = (DataUnit) pairs[ i ];
            DataUnit actual = DataUnit.forString( (String) pairs[ i + 1 ] );
            state.equal( expct, actual );
        }
    }

    @Test
    private
    void
    testDataSizeFromString()
    {
        // some shorthand for below
        DataUnit b = DataUnit.BYTE;
        DataUnit k = DataUnit.KILOBYTE;
        DataUnit m = DataUnit.MEGABYTE;
        DataUnit g = DataUnit.GIGABYTE;
        DataUnit t = DataUnit.TERABYTE;

        Object[] pairs = {
            1L, "1 byte",
            20L, " 20bytes",
            20L, "20b",
            20L, "20",
            b.convert( 1, k ), "1k",
            b.convert( 1, k ), "  1KB  ",
            b.convert( 2, k ), "2 kilobyTE",
            b.convert( 3, m ), "3mb",
            b.convert( 4, m ), "4 megabytes      ",
            b.convert( 3, g ), "3GB",
            b.convert( 3, t ), "3t"
        };

        for ( int i = 0; i < pairs.length; i += 2 )
        {
            long expct = ( (Long) pairs[ i ] ).longValue();
            long actual = 
                DataSize.fromString( (String) pairs[ i + 1 ] ).getByteCount();
            
            state.equal( expct, actual );
        }
    }

    @Test
    private
    void
    testDataSizeGetIntByteCountNormal()
    {
        state.equalInt( 0, DataSize.ZERO.getIntByteCount() );
        state.equalInt( 1, DataSize.ofBytes( 1 ).getIntByteCount() );
        state.equalInt( 4096, DataSize.ofKilobytes( 4 ).getIntByteCount() );

        state.equalInt(
            Integer.MAX_VALUE, 
            DataSize.ofBytes( Integer.MAX_VALUE ).getIntByteCount() );
    }

    @Test
    private
    void
    testDataSizeGetIntByteCountOverflow()
    {
        for ( DataSize sz :
                new DataSize[] {
                    DataSize.ofBytes( 1L << 50 ),
                    DataSize.ofBytes( (long) Integer.MAX_VALUE + 1L ),
                    DataSize.ofGigabytes( 100000000 )
                } )
        {
            try { sz.getIntByteCount(); }
            catch ( IllegalStateException ise )
            {
                state.equalString(
                    "Data size " + sz + 
                    " is too large to represent with an int byte count",
                    ise.getMessage()
                );
            }
        }
    }

    private
    ByteBuffer
    makeBuffer( int len )
    {
        state.isTrue( len <= Byte.MAX_VALUE + 1 );
        ByteBuffer res = ByteBuffer.allocate( len );

        for ( byte b = 0; b < len; ++b ) res.put( b, b );
        return res;
    }

    // Runs some copy tests: smaller --> larger and larger --> smaller (and
    // equal size in both cases too)
    @Test
    private
    void
    testIoUtilsCopyBufferToBuffer()
    {
        ByteBuffer src;

        ByteBuffer dest3 = IoUtils.allocateByteBuffer( DataSize.ofBytes( 3 ) );

        ByteBuffer dest10 =
            IoUtils.allocateByteBuffer( DataSize.ofBytes( 10 ) );

        src = makeBuffer( 3 );
        state.equalInt( 3, IoUtils.copy( src.duplicate(), dest3 ) );
        state.equalInt( 3, IoUtils.copy( src.duplicate(), dest10 ) );
        dest3.flip();
        dest10.flip();
        state.equal( src, dest3 );
        state.equal( src, dest10 );

        src = makeBuffer( 10 );
        dest3.clear();
        dest10.clear();
        state.equalInt( 3, IoUtils.copy( src.duplicate(), dest3 ) );
        state.equalInt( 10, IoUtils.copy( src.duplicate(), dest10 ) );
        dest10.flip();
        dest3.flip();
        state.equal( src, dest10 );
        state.equal( makeBuffer( 3 ), dest3 );
    }

    @Test
    private
    void
    testByteBufferCopyOf()
    {
        ByteBuffer b1 = IoTestFactory.nextByteBuffer( 1024 );
        ByteBuffer b1Slice = b1.slice();
        ByteBuffer copy = IoUtils.copyOf( b1 );

        state.equal( b1Slice, b1 ); // make sure copyOf didn't change b1
        state.equal( b1, copy ); // make sure copy contains the right data

        // check that the copy is really a copy of the original and not a view
        byte b = b1.get( 0 );
        b1.put( 0, (byte) ( b + 1 ) );
        state.equal( (byte) b, copy.get( 0 ) );
    }

    @Test
    private
    void
    testByteBufferExpand()
    {
        int baseSz = 400; // multiple of 4 since we get/put ints

        ByteBuffer bb = ByteBuffer.allocate( baseSz );
        while ( bb.hasRemaining() ) bb.putInt( bb.position() );
        bb.flip();

        ByteBuffer bb2 = IoUtils.expand( bb );
        state.equalInt( baseSz, bb.remaining() ); // check not modified

        // check props of expanded buffer
        state.equalInt( baseSz * 2, bb2.capacity() );
        state.equalInt( baseSz, bb2.remaining() );
        state.equalInt( baseSz, bb2.position() );

        // check content copied
        bb2.position( 0 );
        bb2.limit( baseSz );
        state.equal( bb, bb2 );
    }

    @Test
    private
    void
    testIoUtilsShiftPosAndLimit()
    {
        ByteBuffer bb = ByteBuffer.allocate( 10 );
        state.equal( bb, IoUtils.shiftPos( bb, 2 ) );
        state.equalInt( 2, bb.position() );
        IoUtils.shiftPos( bb, -1 );
        state.equalInt( 1, bb.position() );

        state.equal( bb, IoUtils.shiftLimit( bb, -3 ) );
        state.equalInt( 7, bb.limit() );
        IoUtils.shiftLimit( bb, 1 );
        state.equalInt( 8, bb.limit() );
    }

    private
    void
    assertZeroed( ByteBuffer bb,
                  int posExpct,
                  int limExpct )
    {
        state.equalInt( posExpct, bb.position() );
        state.equalInt( limExpct, bb.limit() );

        bb = bb.slice();
        while ( bb.hasRemaining() ) state.equalInt( 0, (int) bb.get() );
    }

    @Test
    private
    void
    testBzeroByteBuffer()
    {
        ByteBuffer bb = ByteBuffer.allocate( 10 );

        bb.put( 1, (byte) 1 );
        state.isTrue( bb == IoUtils.bzero( bb ) ); // check ref is returned
        assertZeroed( bb, 0, 10 );

        // make sure IoUtils.bzero() respects pos/limit
        bb.clear();
        bb.put( (byte) 1 );
        bb.put( 1, (byte) 1 );
        bb.put( 8, (byte) 1 );
        bb.put( 9, (byte) 1 );
        bb.limit( 9 );
        IoUtils.bzero( bb );
        assertZeroed( bb, 1, 9 );
        bb.clear(); // so we can access [0] and [9] below
        state.equalInt( 1, (int) bb.get( 0 ) );
        state.equalInt( 1, (int) bb.get( 9 ) );
    }

    // See http://bugs.sun.com/bugdatabase/view_bug.do?bug_id=4997655
    @Test
    private
    void
    testCharBufferSliceAndDuplicate()
    {
        // CharBuffer.wrap( "hello" ) here would fail the test, we check our
        // workaround
        CharBuffer orig = IoUtils.charBufferFor( "hello" );

        orig.get();

        state.equalInt( 1, orig.position() );
        state.equalInt( 4, orig.remaining() );
        state.equalInt( 'e', orig.get( 1 ) );

        CharBuffer dup = orig.duplicate();
        state.equalInt( 1, dup.position() );
        state.equalInt( 4, dup.remaining() );
        state.equalInt( 'e', dup.get( 1 ) );

        CharBuffer slice = orig.slice();
        state.equalInt( 0, slice.position() );
        state.equalInt( 4, slice.remaining() );
        state.equalInt( 'e', slice.get( 0 ) );
    }

    @Test
    private
    void
    testToByteBufferFromStream()
        throws Exception
    {
        InputStream is = 
            ReflectUtils.getResourceAsStream( getClass(), "hello.txt" );
        
        ByteBuffer bb = IoUtils.toByteBuffer( is, true );

        state.equalInt( 6, bb.remaining() );
        state.equalString( Charsets.UTF_8.asString( bb ), "Hello\n" );
    }

    @Test
    private
    void
    testToByteArrayFromByteBuffer()
    {
        ByteBuffer bb = IoTestFactory.nextByteBuffer( 813 );
        ByteBuffer expct = bb.slice(); // since we'll modify position of bb

        byte[] arr = IoUtils.toByteArray( bb );
        state.equalInt( arr.length, bb.capacity() );
        state.isFalse( bb.hasRemaining() );

        state.equal( expct, ByteBuffer.wrap( arr ) );
    }
    
    private
    void
    roundtripString( CharSequence expct,
                     CharsetHelper hlp )
        throws CharacterCodingException
    {
        ByteBuffer bb = hlp.asByteBuffer( expct );
        int pos = bb.position();
        int rem = bb.remaining();
        
        state.equal( expct, hlp.asString( bb ).toString() );

        // Make sure asString didn't modify the buffer
        state.equal( pos, bb.position() );
        state.equal( rem, bb.remaining() );
        
        byte[] arr = IoUtils.toByteArray( hlp.asByteBuffer( expct ) );
        state.equal( expct, hlp.asString( arr ).toString() );
    }

    @Test
    private
    void
    testUtf8StringRoundtrips()
        throws CharacterCodingException
    {
        roundtripString( "hello", Charsets.UTF_8 );
        roundtripString( "Copyright: \u00A9", Charsets.UTF_8 );
    }

    private
    void
    roundtripUrlEncode( CharSequence expct,
                        CharSequence expctUrlEnc,
                        CharsetHelper hlp )
    {
        CharSequence actualUrlEnc = hlp.urlEncode( expct );
        state.equal( expctUrlEnc.toString(), actualUrlEnc.toString() );

        CharSequence actual = hlp.urlDecode( actualUrlEnc );
        state.equal( expct.toString(), actual.toString() );
    }

    @Test
    private
    void
    testUtf8UrlEncode()
    {
        roundtripUrlEncode( "hello", "hello", Charsets.UTF_8 );
        roundtripUrlEncode( 
            "Copyright \u00A9", "Copyright+%C2%A9", Charsets.UTF_8 );
    }

    @Test
    private
    void
    testIso8859_1Roundtrips()
        throws CharacterCodingException
    {
        roundtripString( "hello", Charsets.ISO_8859_1 );
        roundtripString( "Copyright \u00A9", Charsets.ISO_8859_1 );
    }

    @Test
    private
    void
    testIso8859_1UrlEncode()
    {
        roundtripUrlEncode( "hello", "hello", Charsets.ISO_8859_1 );
        roundtripUrlEncode( 
            "Copyright \u00A9", "Copyright+%A9", Charsets.ISO_8859_1 );
    }

    @Test( expected = UnmappableCharacterException.class )
    private
    void
    testUnmappableAsByteBuffer()
        throws CharacterCodingException
    {
        Charsets.US_ASCII.asByteBuffer( "\u0080" );
    }

    @Test( expected = MalformedInputException.class )
    private
    void
    testUnmappableAsString()
        throws CharacterCodingException
    {
        Charsets.US_ASCII.asString( new byte[] { -1 } );
    }

    @Test
    private
    void
    testAsByteBufferUncheckedSuccess()
        throws Exception
    {
        state.equalString(
            "hello",
            Charsets.UTF_8.asString( 
                Charsets.UTF_8.asByteBufferUnchecked( "hello" ) ) );
    }

    @Test
    private
    void
    testAsByteBufferUncheckedFails()
    {
        try 
        { 
            Charsets.US_ASCII.asByteBufferUnchecked( "\u0080" );
            state.fail();
        }
        catch ( RuntimeException re )
        {
            state.equalString( 
                "Unexpected character coding exception", re.getMessage() );
            
            state.notNull( 
                state.cast( CharacterCodingException.class, re.getCause() ) );
        }
    }

    private
    void
    assertKnownPair( String ascii,
                     String base64 )
        throws Exception
    {
        byte[] data = ascii.getBytes( "US-ASCII" );
        
        Base64Encoder enc = new Base64Encoder();
        state.equal( enc.encode( data ).toString(), base64 );
        state.equal( enc.decode( base64 ), data );
    }

    // See http://www.ietf.org/rfc/rfc4648.txt
    @Test
    private
    void
    runRfc4648Tests()
        throws Exception
    {
        assertKnownPair( "", "" );
        assertKnownPair( "f", "Zg==" );
        assertKnownPair( "fo", "Zm8=" );
        assertKnownPair( "foo", "Zm9v" );
        assertKnownPair( "foob", "Zm9vYg==" );
        assertKnownPair( "fooba", "Zm9vYmE=" );
        assertKnownPair( "foobar", "Zm9vYmFy" );
    }

    @Test
    private
    void
    runRandomRoundtrips()
        throws Exception
    {
        Base64Encoder enc = new Base64Encoder();
        int reps = 5000;

        for ( int i = 0; i < reps; ++i )
        {
            byte[] data = new byte[ rand.nextInt( 2000 ) ];
            rand.nextBytes( data );

            state.equal( data, enc.decode( enc.encode( data ) ) );
        }
    }

    @Test( expected = Base64Exception.class )
    private
    void
    testOutOfRangeChecked()
        throws Exception
    {
        String invalid = new String( new char[] { 'a', 'b', 'c', 200 } );
        new Base64Encoder().decode( invalid );
            
    }

    @Test( expected = Base64Exception.class )
    private
    void
    testInvalidCharChecked()
        throws Exception
    {
        new Base64Encoder().decode( "Zg;=" );
    }

    @Test( expected = Base64Exception.class )
    private
    void
    testBadDecodeInputLength()
        throws Exception
    {
        new Base64Encoder().decode( "abcde" );
    }

    private
    CharSequence
    asString( Iterable< ByteBuffer > bufs )
        throws Exception
    {
        CharsetDecoder dec = Charsets.UTF_8.newDecoder();
        return IoUtils.asString( bufs, dec );
    }

    @Test
    private
    void
    testAsStringOnEmptyBufList()
        throws Exception
    {
        state.equalString( "", asString( Lang.< ByteBuffer >emptyList() ) );
    }

    @Test
    private
    void
    testRemainingMultiBufs()
    {
        ByteBuffer[] arr = 
            new ByteBuffer[] {
                    ByteBuffer.allocate( 3 ),
                    ByteBuffer.allocate( 0 ),
                    ByteBuffer.allocate( 0 ),
                    ByteBuffer.allocate( 7 )
            };
        
        state.equalInt( 10, IoUtils.remaining( arr ) );
        state.equalInt( 10, IoUtils.remaining( Lang.asList( arr ) ) );

        state.equalInt( 0, IoUtils.remaining() );

        state.equalInt( 
            0, IoUtils.remaining( Lang.< ByteBuffer >emptyList() ) );
    }

    private
    void
    assertIoUtilsAsLine( int... breaks )
        throws Exception
    {
        String str = "\ud834\uDD1e:Stuff before the colon is a g-clef";
        ByteBuffer buf = Charsets.UTF_8.asByteBuffer( str );

        List< ByteBuffer > bufs = Lang.newList();

        for ( int i = 0; i < breaks.length; ++i )
        {
            ByteBuffer bb = buf.duplicate();

            int pos = breaks[ i ];
            bb.limit( pos );
            buf.position( pos );

            bufs.add( bb );
        }

        bufs.add( buf ); // whatever is left, could be whole thing

        state.equalString( str, asString( bufs ) );
    }

    @Test
    private
    void
    testIoUtilsAsLineSingleBuffer()
        throws Exception
    {       
        assertIoUtilsAsLine();
    }

    @Test
    private
    void
    testIoUtilsAsLineMultiBufsNoMultiCharSplit()
        throws Exception
    {
        assertIoUtilsAsLine( 10, 15 );
    }

    @Test
    private
    void
    testIoUtilsAsLineMultiBufsMultiCharSplit()
        throws Exception
    {
        assertIoUtilsAsLine( 2, 5, 10, 15, 23 );
    }

    // To simplify the test, we rely on the fact that the internals of
    // IoUtils.asHexString() operate byte-by-byte, and so we only check that it
    // correctly operates on single-byte inputs
    @Test
    private
    void
    testAsHexString()
    {
        ByteBuffer bb = ByteBuffer.allocate( 1 );

        for ( int i = 0; i < 256; ++i )
        {
            bb.put( 0, (byte) i );
            CharSequence actual = IoUtils.asHexString( bb );
            String expct = ( i < 16 ? "0" : "" ) + Integer.toHexString( i );
            state.equalString( expct, actual );
        }
    }

    // Uses md5 output from command line 'md5 -s "hello"' as another sanity
    // check for IoUtils.asHexString. This is nice since it includes a sequence
    // (the leftmost) with a leading zero, which a naive implementation of
    // toHexString, for instance one based on converting the bytes to a
    // BigInteger and then using BigInteger.toString( 16 ), would get wrong (by
    // dropping the leading zeroes as numerically irrelevant).
    @Test
    private
    void
    testAsHexStringUsingMd5()
        throws Exception
    {
        // We could theoretically do this with com.bitgirder.crypto.CryptoUtils,
        // but don't want to introduce a crypto dependency in io just for this
        // test
        ByteBuffer text = Charsets.US_ASCII.asByteBuffer( "test" );

        MessageDigest md = MessageDigest.getInstance( "MD5" );
        md.update( text );
        ByteBuffer sig = ByteBuffer.wrap( md.digest() );

        state.equalString( 
            "098f6bcd4621d373cade4e832627b4f6", IoUtils.asHexString( sig ) );
    }

    private
    void
    assertHexRoundtrip( byte[] arr )
        throws Exception
    {
        String hex = IoUtils.asHexString( arr ).toString();
        assertEqual( arr, IoUtils.hexToByteArray( hex.toLowerCase() ) );
        assertEqual( arr, IoUtils.hexToByteArray( hex.toUpperCase() ) );
    }

    @Test
    private
    void
    testHexStringRoundtrips()
        throws Exception
    {
        assertHexRoundtrip( new byte[] {} );

        assertHexRoundtrip( 
            new byte[] { (byte) 0xff, (byte) 0x00, (byte) 0x02 } );

        assertHexRoundtrip( IoTestFactory.nextByteArray( 100 ) );
    }

    @Test( expected = IOException.class,
           expectedPattern = "\\QInvalid odd length for hex string: 3\\E" )
    private
    void
    testFromHexFailsOddInputLength()
        throws Exception
    {
        IoUtils.hexToByteArray( "a01" );
    }

    @Test( expected = IOException.class,
           expectedPattern = "\\QNot a hex char at index 4: '$' (0x24)\\E" )
    private
    void
    testFromHexFailsBadHexChar()
        throws Exception
    {
        IoUtils.hexToByteArray( "0a1b$c" );
    }

    // Regression put in upon discovering that obtaining a FileChannel from a
    // FileOutputStream opened for append does not necessarily mean that the
    // position will be at the end. FileWrapper.openAppendChannel() now ensures
    // this, as covered by this test
    @Test
    private
    void
    testFileOpenAppendChannel()
        throws Exception
    {
        FileWrapper fw = IoTestFactory.createTempFile();

        int len = 10;
        ByteBuffer bb = ByteBuffer.allocate( len );
        FileChannel fc = fw.openWriteChannel();
        while ( bb.hasRemaining() ) fc.write( bb );
        fc.close();

        fc = fw.openAppendChannel();
        state.equal( 10L, fc.position() );
    }

    private
    void
    assertFill( InputStream is,
                int len,
                ByteBuffer expct )
        throws Exception
    {
        byte[] dest = new byte[ len ];
        IoUtils.fill( is, dest );
        state.equal( expct, ByteBuffer.wrap( dest ) );
    }

    @Test
    private
    void
    testIoUtilsFillExact()
        throws Exception
    {
        byte[] src = IoTestFactory.nextByteArray( 1024 );
        ByteArrayInputStream is = new ByteArrayInputStream( src );

        // Test fill which exhausts stream exactly
        assertFill( is, src.length, ByteBuffer.wrap( src ) );

        // Test that fill of short buffer does not consume any more data than it
        // is supposed to
        int lim = 10;
        is = new ByteArrayInputStream( src );

        // read the first short buffer
        assertFill( is, lim, ByteBuffer.wrap( src, 0, lim ) );

        // now check that the remaining data was left in the stream
        int len = src.length - lim;
        assertFill( is, len, ByteBuffer.wrap( src, lim, len ) );
    }

    @Test( expected = EOFException.class )
    private
    void
    testIoUtilsFillEOFDetected()
        throws Exception
    {
        assertFill( 
            new ByteArrayInputStream( new byte[] {} ),
            10,
            IoUtils.emptyByteBuffer()
        );
    }

    // Other fill tests work on the beginning of the target array; we have this
    // also to get coverage of an explicit offset+len combo
    @Test
    private
    void
    testIoUtilsFillAtArrayMiddle()
        throws Exception
    {
        byte[] arr = new byte[ 30 ];

        byte[] buf = new byte[ 10 ];
        for ( int i = 0; i < buf.length; ++i ) buf[ i ] = 0x01;
        ByteArrayInputStream bis = new ByteArrayInputStream( buf );

        IoUtils.fill( bis, arr, 10, 10 );

        byte[] expct = new byte[ 30 ];
        for ( int i = 10; i < 20; ++i ) expct[ i ] = 0x01;

        assertEqual( expct, arr );
    }

    @Test
    private
    void
    testWhichFindsLs()
    {
        state.isTrue( state.notNull( IoUtils.which( "ls" ).endsWith( "ls" ) ) );
    }

    @Test
    private
    void
    testWhichNotFound()
    {
        state.equal( null, IoUtils.which( "asddfasfdsajfa" ) );
    }

    @Test
    private
    void
    testGetResources()
        throws Exception
    {
        List< URL > rsrcs = IoUtils.getResources( TEST_RSRC1 );
 
        state.equalInt( 2, rsrcs.size() );
    }

    // expectedPattern below uses the char class to catch the specific parts of
    // the file path that will be specific to the caller's env but still to
    // assert that the error message is of the correct length (2) and concerns
    // the correct files
    @Test( expected = IllegalStateException.class,
           expectedPattern = 
            "Multiple resources found matching name com/bitgirder/io/test-resource1: \\[(?:jar:)?file:[\\w/\\-]+(?:\\.jar!)?/com/bitgirder/io/test-resource1, (?:jar:)?file:[\\w/\\-]+(?:\\.jar!)?/com/bitgirder/io/test-resource1\\]" )
    private
    void
    testExpectSingleResourceFailsOnMultipleMatches()
        throws Exception
    {
        IoUtils.expectSingleResource( TEST_RSRC1 );
    }

    @Test( expected = IllegalStateException.class,
           expectedPattern = 
            "No resources found matching name /alksajfsklafjsadkljfsdklafjsdklfj" )
    private
    void
    testExpectSingleResourceFailsOnNoMatches()
        throws Exception
    {
        // this file probably doesn't exist
        IoUtils.expectSingleResource( "/alksajfsklafjsadkljfsdklafjsdklfj" );
    }

    @Test
    private
    void
    testExpectSingleResourceAsStream()
        throws Exception
    {
        state.equalString(
            "Hello\n",
            Charsets.UTF_8.asString(
                IoUtils.toByteBuffer(
                    IoUtils.expectSingleResourceAsStream( TEST_RSRC2 ), true
                )
            )
        );
    }

    // We also take this as coverage of IoUtils.loadProperties( InputStream ),
    // knowing that the implementation of IoUtils.loadProperties( URL ) is a
    // facade on top of that method
    @Test
    private
    void
    testLoadProperties()
        throws Exception
    {
        String rsrc = "com/bitgirder/io/test-props.properties";

        state.equalString(
            "val1",
            state.getProperty(
                IoUtils.loadProperties( IoUtils.expectSingleResource( rsrc ) ),
                "prop1",
                "test-props.properties"
            )
        );
    }

    private
    void
    assertReadLines( CharSequence expct,
                     CharSequence src )
        throws Exception
    {
        // This is our way of exposing a mutable primitive value to the anon
        // class's implementation of close() below
        final boolean[] closed = new boolean[] { false };

        InputStream is =
            new ByteArrayInputStream( src.toString().getBytes( "utf-8" ) ) {
                @Override public void close() { closed[ 0 ] = true; }
            };

        state.equalString(
            expct,
            Strings.join( "|", IoUtils.readLines( is, "utf-8" ) )
        );

        state.isTrue( closed[ 0 ] );
    }

    @Test 
    private 
    void 
    testReadLinesEmpty() 
        throws Exception
    { 
        assertReadLines( "", "" ); 
    }

    @Test
    private
    void
    testReadLinesUnterminatedLastLine()
        throws Exception
    {
        assertReadLines( "line1|line2", "line1\nline2" );
    }

    @Test
    private
    void
    testReadLinesStandard()
        throws Exception
    {
        assertReadLines( "line1|line2", "line1\nline2\n" );
    }

    // Fix for a bug in Crc32Digest that was incorrectly indexing into the
    // backing array based solely on the buffer's position without adding in
    // ByteBuffer.arrayOffset(), causing incorrect values to be returned for
    // slices not anchored at the original buffer's 0-position
    @Test
    private
    void
    testCrc32DigestOnArrayBackedSliceTestRegression()
        throws Exception
    {
        ByteBuffer bb = IoTestFactory.nextByteBuffer( 1024 );
        state.isTrue( bb.hasArray() ); // otherwise test makes not sense

        bb.position( 512 );
        ByteBuffer bb2 = bb.slice();
        
        assertDigests(
            IoUtils.digest( Crc32Digest.create(), true, bb ),
            IoUtils.digest( Crc32Digest.create(), true, bb2 ) );
    }

    private
    void
    assertAbcCharReader( CharReader cr )
        throws Exception
    {
        state.equal( (int) 'a', cr.peek() );
        state.equal( (int) 'a', cr.read() );
        state.equal( (int) 'b', cr.peek() );
        state.equal( (int) 'b', cr.read() );
        state.equal( (int) 'c', cr.read() );
        state.equal( -1, cr.peek() );
        state.equal( -1, cr.peek() ); // back-to-back calls should be okay
        state.equal( -1, cr.read() );
        state.equal( -1, cr.read() );
    }

    @Test
    private
    void
    testDefaultCharReader()
        throws Exception
    {
        assertAbcCharReader( IoUtils.charReaderFor( "abc" ) );
    }

    @Test
    private
    void
    testCountingCharReader()
        throws Exception
    {
        // first just check general behavior independent of counting ability
        assertAbcCharReader(
            new CountingCharReader( IoUtils.charReaderFor( "abc" ) ) );

        CountingCharReader ccr =
            new CountingCharReader( IoUtils.charReaderFor( "ab" ) );
        
        state.equal( 0L, ccr.position() );
        ccr.peek();
        state.equal( 0L, ccr.position() );
        ccr.read();
        state.equal( 1L, ccr.position() );
        ccr.peek();
        state.equal( 1L, ccr.position() );
        ccr.read();
        state.equal( 2L, ccr.position() );
        ccr.peek();
        ccr.read();
        state.equal( 2L, ccr.position() );
    }

    private
    final
    static
    class SerialTester
    implements Serializable
    {
        private final String s;

        private SerialTester( String s ) { this.s = s; }
    }

    private
    < V extends Serializable >
    V
    runSerialRoundtrip( V obj,
                        Class< V > cls )
        throws Exception
    {
        byte[] arr = IoUtils.toSerialByteArray( obj );

        return cls.cast( IoUtils.fromSerialByteArray( arr ) );
    }

    @Test
    private
    void
    testByteArraySerialization()
        throws Exception
    {
        SerialTester t1 = new SerialTester( "hello" );
        SerialTester t2 = runSerialRoundtrip( t1, SerialTester.class );

        state.equalString( t1.s, t2.s );
    }

    // Static util methods for tests doing IO stuff

    public
    static
    void
    assertEqual( byte[] expct,
                 byte[] actual,
                 boolean showHex )
    {
        if ( state.sameNullity( expct, actual ) )
        {
            state.equalf( 
                ByteBuffer.wrap( expct ), 
                ByteBuffer.wrap( actual ),
                "byte arrays differ: %s != %s",
                showHex ? IoUtils.asHexString( expct ) : expct,
                showHex ? IoUtils.asHexString( actual ) : actual
            );
        }
    }

    public
    static
    void
    assertEqual( byte[] expct,
                 byte[] actual )
    {
        assertEqual( expct, actual, true );
    }

    // does not modify positions of input buffers
    public
    static
    ByteBuffer
    concatenateBuffers( Iterable< ? extends ByteBuffer > bufs )
    {
        inputs.notNull( bufs, "bufs" );

        int len = 0;
        for ( ByteBuffer bb : bufs ) len += bb.remaining();

        ByteBuffer res = ByteBuffer.allocate( len );
        for ( ByteBuffer bb : bufs ) res.put( bb.slice() );

        return (ByteBuffer) res.flip();
    }

    public
    static
    ByteBuffer
    concatenateBuffers( ByteBuffer... bufs )
    {
        return concatenateBuffers( Lang.asList( bufs ) );
    }

    public
    static
    long
    crc32( Iterable< ? extends ByteBuffer > bufs )
    {
        try { return IoUtils.digest( Crc32Digest.create(), false, bufs ); }
        catch ( Exception ex ) { throw new RuntimeException( ex ); }
    }

    public
    static
    long
    crc32( ByteBuffer... bufs )
    {
        inputs.noneNull( bufs, "bufs" );
        return crc32( Lang.asList( bufs ) );
    }

    public
    static
    long
    crc32( byte[] arr )
    {
        inputs.notNull( arr, "arr" );
        return crc32( ByteBuffer.wrap( arr ) );
    }

    public
    static
    void
    assertDigests( long digExpct,
                   long digActual )
    {
//        code( "digExpct:", digExpct, "; digActual:", digActual );
        state.equal( digExpct, digActual );
    }

    // Blocking recursive delete
    public
    static
    void
    rmRf( DirWrapper dir,
          ObjectReceiver< ? super AbstractFileWrapper > postDelete )
        throws Exception
    {
        inputs.notNull( dir, "dir" );
        inputs.notNull( postDelete, "postDelete" );

        for ( File f : dir.getFile().listFiles() )
        {
            AbstractFileWrapper afw = AbstractFileWrapper.wrap( f );

            if ( afw instanceof DirWrapper ) 
            {
                rmRf( (DirWrapper) afw, postDelete );
            }
            else
            {
                afw.delete();
                postDelete.receive( afw );
            }
        }

        dir.delete();
        postDelete.receive( dir );
    }

    public
    static
    void
    rmRf( DirWrapper dir )
        throws Exception
    {
        rmRf(
            dir,
            new ObjectReceiver< AbstractFileWrapper >() {
                public void receive( AbstractFileWrapper afw ) {}
            }
        );
    }

    public
    static
    ByteBuffer
    toByteBuffer( Iterable< ? extends ByteBuffer > bufs )
    {
        inputs.noneNull( bufs, "bufs" );

        int tot = 0;
        for ( ByteBuffer buf : bufs ) tot += buf.remaining();

        ByteBuffer res = ByteBuffer.allocate( tot );

        for ( ByteBuffer buf : bufs ) res.put( buf );

        state.isFalse( res.hasRemaining() );

        res.flip();

        return res;
    }

    public
    static
    ByteBuffer
    toByteBuffer( ByteBuffer[] bufs )
    {
        inputs.notNull( bufs, "bufs" );
        return toByteBuffer( Lang.asList( bufs ) );
    }
}
