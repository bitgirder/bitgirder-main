package com.bitgirder.mingle;

public
enum MingleIdentifierFormat
{
    LC_HYPHENATED( "lc", "hyphenated" ),
    LC_UNDERSCORE( "lc", "underscore" ),
    LC_CAMEL_CAPPED( "lc", "camel", "capped" );

    private final MingleIdentifier id;

    private 
    MingleIdentifierFormat( String... parts )
    {
        this.id = new MingleIdentifier( parts );
    }
    
    public MingleIdentifier getIdentifier() { return id; }
}
