package com.bitgirder.lang.path;

public
interface ObjectPathFormatter< E >
{
    public 
    void 
    formatPathStart( StringBuilder sb );

    public 
    void 
    formatSeparator( StringBuilder sb );

    public
    void
    formatDictionaryKey( StringBuilder sb,
                         E key );

    public
    void
    formatListIndex( StringBuilder sb,
                     int indx );
}
