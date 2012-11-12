package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.io.FileNotFoundException;

public
final
class IoExceptionFactories
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static IoExceptionFactory INSTANCE =
        new IoExceptionFactory()
        {
            public
            Exception
            createFilePermissionException( FileWrapper f,
                                           FileOpenMode mode )
            {
                return 
                    new FilePermissionException( 
                        "Permission denied attempting to open for " + mode +
                        ": " + f
                    );
            }

            public
            Exception
            createNoSuchFileException( FileWrapper f )
            {
                return new FileNotFoundException( f.toString() );
            }
        };

    private IoExceptionFactories() {}

    public static IoExceptionFactory getDefaultFactory() { return INSTANCE; }
}
