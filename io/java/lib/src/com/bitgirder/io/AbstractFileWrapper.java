package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.io.IOException;
import java.io.File;
import java.io.FileNotFoundException;

/**
 * Base class providing common utilities and helper methods on top of {@link
 * java.io.File}.
 */
public
abstract
class AbstractFileWrapper< F extends AbstractFileWrapper >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final File file;

    AbstractFileWrapper( File f )
    {
        this.file = inputs.notNull( f, "f" );
    }

    AbstractFileWrapper( CharSequence cs )
    {
        this( new File( inputs.notNull( cs, "cs" ).toString() ) );
    }

    AbstractFileWrapper( DirWrapper parent,
                         CharSequence child )
    {
        this( 
            new File( 
                inputs.notNull( parent, "parent" ).getFile(),
                inputs.notNull( child, "child" ).toString()
            )
        );
    }

    /**
     * Gets the underlying java.io.File object.
     */
    public
    File
    getFile()
    {
        return file;
    }

    /**
     * Returns a DirWrapper representing the parent of this AbstractFile
     * object.
     *
     * @return The parent of this AbstractFileWrapper, or null if this
     * AbstractFileWrapper has none.
     */
    public
    DirWrapper
    dirName()
    {
        String parent = file.getParent();
        return parent == null ? null : new DirWrapper( parent );
    }

    public
    String
    baseName()
    {
        return file.getName();
    }

    /**
     * Returns true if and only if this AbstractFileWrapper exists in the
     * filesystem.
     */
    public boolean exists() { return file.exists(); }

    /**
     * Convenience method that throws a runtime exception if this
     * AbstractFileWrapper does not exist.
     *
     *  @throws IllegalStateException If this file or directory does not exist.
     */
    public
    F
    assertExists()
    {
        state.isTrue( 
            exists(), "File or directory " + this + " does not exist" );

        // Since we control all subclasses we're okay that this is a sane cast.
        @SuppressWarnings( "unchecked" )
        F res = (F) this;

        return res;
    }

    public
    void
    assertNotExists()
    {
        state.isFalse( 
            exists(), "File or directory " + this + " already exists" );
    }

    /**
     * Calls {@link File#delete} on the underlying object, but throws an
     * IOException instead of silently returning false.
     *
     * @throws IOException If the underlying call to delete() returns false.
     */
    public
    void
    delete()
        throws IOException
    {
        if ( ! file.delete() ) 
        {
            throw new IOException( "Couldn't delete " + this );
        }
    }
    
    public
    void
    renameTo( AbstractFileWrapper other )
        throws IOException
    {
        inputs.notNull( other, "other" );

        if ( ! getFile().renameTo( other.getFile() ) )
        {
            throw new IOException( 
                "File.renameTo() returned false renaming " + this + " to " +
                other );
        }
    }

    /**
     * Equivalent to <code>getFile().toString()</code>.
     */
    public
    String
    toString()
    {
        return file.toString();
    }

    /**
     * Utility method to wrap a File object into the appropriate instance of
     * AbstractFileWrapper (either a FileWrapper or DirWrapper).
     */
    public
    static
    AbstractFileWrapper
    wrap( File f )
    {
        inputs.notNull( f, "f" );

        return f.isDirectory() ? new DirWrapper( f ) : new FileWrapper( f );
    }
}
