package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.nio.charset.Charset;

public
final
class Charsets
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private Charsets() {}

    public final static CharsetHelper UTF_8 = 
        new CharsetHelper( Charset.forName( "utf-8" ) );
    
    public final static CharsetHelper UTF_16LE =
        new CharsetHelper( Charset.forName( "utf-16le" ) );

    public final static CharsetHelper UTF_16BE =
        new CharsetHelper( Charset.forName( "utf-16be" ) );

    public final static CharsetHelper UTF_32LE =
        new CharsetHelper( Charset.forName( "utf-32le" ) );

    public final static CharsetHelper UTF_32BE =
        new CharsetHelper( Charset.forName( "utf-32be" ) );

    public final static CharsetHelper ISO_8859_1 = 
        new CharsetHelper( Charset.forName( "iso-8859-1" ) );

    public final static CharsetHelper US_ASCII =
        new CharsetHelper( Charset.forName( "us-ascii" ) );
}
