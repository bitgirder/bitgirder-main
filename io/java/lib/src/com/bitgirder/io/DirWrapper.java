package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.io.File;
import java.io.IOException;

import java.util.List;

/**
 * Wrapper around File objects that are directories. Methods which alter the
 * filesystem (such as creating a directory) are not necessarily threadsafe.
 * Methods which read data only (such as doing directory listings) are
 * threadsafe.
 */
public
class DirWrapper
extends AbstractFileWrapper< DirWrapper >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    /**
     * Creates a DirWrapper backed by the given File object, which either must
     * not yet exist or which must be a directory.
     *
     * @throws IllegalArgumentException If f is not a directory.
     */
    public DirWrapper( File f ) { super( f ); }

    /**
     * See {@link #DirWrapper( File )}.
     */
    public 
    DirWrapper( CharSequence cs )
    { 
        super( inputs.notNull( cs, "cs" ) ); 
    }

    /**
     * Creates a DirWrapper for the child directory of the given parent. The
     * target directory must either not yet exist, or if it does, must be a
     * directory (as opposed to a regular file).
     *
     * @throws IllegalArgumentException If the target directory exists and is a
     * file.
     */
    public
    DirWrapper( DirWrapper parent,
                CharSequence child )
    {
        super( parent, child );
    }

    /**
     * See {@link #DirWrapper( DirWrapper, String )}.
     */
    public
    DirWrapper( File parent,
                String child )
    {
        this( new DirWrapper( parent ), child );
    }

    public
    DirWrapper( DirWrapper parent,
                AbstractFileWrapper child )
    {
        this( 
            inputs.notNull( parent, "parent" ) + "/" + 
            inputs.notNull( child, "child" ) );
    }

    /**
     * Creates the directory represented by this DirWrapper object. This method
     * does nothing if this directory already exists. This method requires that
     * the parent of this directory exist. See {@link #mkdirs} for a version
     * which creates missing parent directories as well as the target directory
     * itself.
     *
     * @throws IOException If the parent directory of the target directory does
     * not exist, or if there is a problem creating the directory.
     */
    public
    DirWrapper
    mkdir()
        throws IOException
    {
        if ( ! getFile().exists() )
        {
            dirName().assertExists(); // check that we have an existing parent
 
            state.isTrue( 
                getFile().mkdir(), "Couldn't create directory", this );
        }

        return this;
    }

    /**
     * Creates the directory represented by this DirWrapper object, as well as
     * any missing parent directories. This method is a no-op if this directory
     * already exists.
     *
     * @throws IOException If there are problems creating the directory or any
     * parent directory. In cases when this exception is thrown, it may be the
     * case that some or all of this directory's parent directories were
     * created.
     */
    // Implementation note: the double existence check, one upon entry and one
    // before possibly throwing an exception, is here as a concurrency design
    // decision. A very common calling pattern for this method is from code
    // which is receiving or rebuilding a directory tree from the network or an
    // archive file. It seems that File.mkdirs() will return false if called
    // when the directory in question already exists, thus the initial exists()
    // check. However, we expect that in many cases concurrent threads will be
    // calling this method as part of their file receiving logic. Rather than
    // make this method synchronized, we just allow the calls to proceed in
    // parallel, but realize that mkdirs() may in fact return false to all but
    // the first-to-complete of any concurrent calls, and so re-check the
    // existence before deciding to interpret this false return as an error
    // condition.
    public
    DirWrapper
    mkdirs()
        throws IOException
    {
        if ( ! exists() )
        {
            if ( ! getFile().mkdirs() ) 
            {
                state.isTrue( exists(), "mkdirs() failed for", this );
            }
        }

        return this;
    }

    public
    static
    DirWrapper
    systemTmpdir()
    {
        return new DirWrapper( System.getProperty( "java.io.tmpdir" ) );
    }
}
