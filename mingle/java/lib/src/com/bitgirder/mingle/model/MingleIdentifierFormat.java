package com.bitgirder.mingle.model;

public
enum MingleIdentifierFormat
{
    LC_HYPHENATED,
    LC_UNDERSCORE,
    LC_CAMEL_CAPPED;

    // Regression not: holding these values in the IDS array was arrived at
    // after first attempting to associate with each enum instance an id field
    // of type MingleIdentifier. However, we can't call
    // MingleIdentifier.create() during instance construction, since it calls
    // MingleParsers.createIdentifier(), which itself refers to
    // MingleIdentifierFormat to do its parsing. We could just hand create the
    // String[] array for the MingleIdentifier constructor, but then run the
    // risk of a typo in string literals that don't match the enum constant. So,
    // we do what we're doing below.
    private final static MingleIdentifier[] IDS;

    public MingleIdentifier getIdentifier() { return IDS[ ordinal() ]; }

    static
    {
        MingleIdentifierFormat[] fmts =
            MingleIdentifierFormat.class.getEnumConstants();

        IDS = new MingleIdentifier[ fmts.length ];

        for ( int i = 0, e = fmts.length; i < e; ++i )
        {
            IDS[ i ] = 
                MingleIdentifier.create( fmts[ i ].name().toLowerCase() );
        }
    }
}
