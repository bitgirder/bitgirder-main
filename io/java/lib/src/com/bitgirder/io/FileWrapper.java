package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.io.File;
import java.io.IOException;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.RandomAccessFile;

import java.nio.ByteBuffer;

import java.nio.channels.FileChannel;

/**
 * Wrapper around File objects that are regular files. Methods that change the
 * underlying file (such as writes or deletes) are not threadsafe. Methods which
 * read but don't alter the file are, so long as they are called in the absence
 * of other methods which may alter the file.
 */
public
final
class FileWrapper
extends AbstractFileWrapper< FileWrapper >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final static String MODE_READ_WRITE = "rw";

    /**
     * Creates a FileWrapper to wrap the given file.
     *
     * @throws IllegalArgumentException If the file exists and is a directory.
     */
    public FileWrapper( File f ) { super( f ); }
 
    /**
     * See {@link #FileWrapper( File )}.
     */
    public FileWrapper( CharSequence cs ) { super( cs ); }

    /**
     * Creates a FileWrapper that is a child of the given parent directory.
     *
     * @throws IllegalArgumentException If the specified target exists and is a
     * directory.
     */
    public
    FileWrapper( DirWrapper parent,
                 CharSequence child )
    {
        super( parent, child );
    }

    public
    FileInputStream
    openReadStream()
        throws IOException
    {
        return new FileInputStream( getFile() );
    }

    public
    FileChannel
    openReadChannel()
        throws IOException
    {
        return openReadStream().getChannel();
    }

    public
    ByteBuffer
    mapRead()
        throws IOException
    {
        FileChannel fc = openReadChannel();

        try { return fc.map( FileChannel.MapMode.READ_ONLY, 0, fc.size() ); }
        finally { fc.close(); }
    }

    public
    FileOutputStream
    openWriteStream( boolean append )
        throws IOException
    {
        return new FileOutputStream( getFile(), append );
    }

    public
    FileOutputStream
    openWriteStream()
        throws IOException
    {
        return openWriteStream( false );
    }

    public
    FileOutputStream
    openAppendStream()
        throws IOException
    {
        return openWriteStream( true );
    }

    public
    FileChannel
    openWriteChannel()
        throws IOException
    {
        return openWriteStream().getChannel();
    }

    public
    FileChannel
    openAppendChannel()
        throws IOException
    {
        FileChannel res = openAppendStream().getChannel();
        res.position( res.size() );

        return res;
    }

    public
    FileChannel
    openRandomAccess( String mode )
        throws IOException
    {
        inputs.notNull( mode, "mode" );
        return new RandomAccessFile( getFile(), mode ).getChannel();
    }

    public
    FileChannel
    openRandomAccess()
        throws IOException
    {
        return openRandomAccess( MODE_READ_WRITE );
    }

    public
    FileChannel
    openChannel( FileOpenMode mode )
        throws IOException
    {
        inputs.notNull( mode, "mode" );

        switch ( mode )
        {
            case READ: return openReadChannel();
            case WRITE: return openWriteChannel();
            case TRUNCATE: return openWriteChannel();
            case READ_WRITE: return openRandomAccess();
            case APPEND: return openAppendChannel();
            case SYNC: return openRandomAccess( "rws" );
            case DSYNC: return openRandomAccess( "rwd" );

            default: throw state.createFail( "Unrecognized mode:", mode );
        }
    }

    /**
     * Deletes this file.
     *
     * @throws IOException If the underlying delete operation is not a success.
     */
    public
    void
    delete()
        throws IOException
    {
        if ( ! getFile().delete() )
        {
            throw new IOException( "Could not delete file: " + toString() );
        }
    }

    /**
     * Gets the size of the data in this file.
     */
    public
    DataSize
    getSize()
    {
        return DataSize.ofBytes( getFile().length() );
    }

    // existence checks are done outside of this method as appropriate to the
    // test
    private
    void
    assertEmptiness( boolean wantEmpty )
    {
        long sz = getSize().size(); // units not examined here

        if ( sz == 0 && ! wantEmpty ) state.fail( this + " is empty" );
        if ( sz != 0 && wantEmpty ) state.fail( this + " is not empty" );
    }

    public
    void
    assertNonEmpty()
    {
        assertExists();
        assertEmptiness( false );
    }

    // asserts that file exists, but is empty
    public
    void
    assertEmpty()
    {
        assertExists();
        assertEmptiness( true );
    }

    // asserts that file either does not exist, or if it does, is empty
    public void assertFresh() { if ( exists() ) assertEmptiness( true ); }
}
