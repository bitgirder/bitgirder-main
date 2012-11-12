package com.bitgirder.io;

public
interface IoExceptionFactory
{
    public
    Exception
    createFilePermissionException( FileWrapper fw,
                                   FileOpenMode mode );
    
    public
    Exception
    createNoSuchFileException( FileWrapper fw );
}
