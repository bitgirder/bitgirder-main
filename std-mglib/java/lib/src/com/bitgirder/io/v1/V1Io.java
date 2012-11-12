package com.bitgirder.io.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.IoExceptionFactory;
import com.bitgirder.io.IoProcessors;
import com.bitgirder.io.FileWrapper;
import com.bitgirder.io.DataSize;

import com.bitgirder.mingle.model.MingleTimestamp;

import java.io.File;

public
final
class V1Io
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static IoExceptionFactory EXCPT_FACT =
        new IoExceptionFactory()
        {
            public
            Exception
            createFilePermissionException( FileWrapper f,
                                           IoProcessors.FileOpenMode mode )
            {
                return PathPermissionException.create( f.toString() );
            }

            public
            Exception
            createNoSuchFileException( FileWrapper f )
            {
                return NoSuchPathException.create( f.toString() );
            }
        };

    private V1Io() {}

    public
    static
    IoExceptionFactory
    getIoExceptionFactory()
    {
        return EXCPT_FACT;
    }

    public
    static
    BasicFileInfo
    basicFileInfoFor( File f )
    {
        inputs.notNull( f, "f" );

        inputs.isTrue( f.isFile(), "File.isFile() is false for", f );

        return
            BasicFileInfo.create(
                DataSize.ofBytes( f.length() ),
                MingleTimestamp.fromMillis( f.lastModified() )
            );
    }

    public
    static
    BasicFileInfo
    basicFileInfoFor( FileWrapper f )
    {
        inputs.notNull( f, "f" );
        return basicFileInfoFor( f.getFile() );
    }
}
