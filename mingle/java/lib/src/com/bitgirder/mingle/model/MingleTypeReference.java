package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.parser.MingleParsers;

import com.bitgirder.parser.SyntaxException;

public
abstract
class MingleTypeReference
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    MingleTypeReference() {}

    public
    abstract
    CharSequence
    getExternalForm();

    public
    abstract
    int
    hashCode();

    public
    abstract
    boolean
    equals( Object o );

    @Override
    public
    final
    String
    toString()
    {
        return getExternalForm().toString();
    }

    public
    static
    MingleTypeReference
    create( CharSequence str )
    {
        return 
            MingleParsers.createTypeReference( inputs.notNull( str, "str" ) );
    }

    public
    static
    MingleTypeReference
    create( CharSequence str,
            MingleIdentifier scopedVer )
    {
        inputs.notNull( str, "str" );
        inputs.notNull( scopedVer, "scopedVer" );

        return MingleParsers.createTypeReference( str, scopedVer );
    }

    public
    static
    MingleTypeReference
    parse( CharSequence str )
        throws SyntaxException
    {
        return MingleParsers.parseTypeReference( inputs.notNull( str, "str" ) );
    }

    public
    static
    MingleTypeReference
    parse( CharSequence str,
           MingleIdentifier scopedVer )
        throws SyntaxException
    {
        inputs.notNull( str, "str" );
        inputs.notNull( scopedVer, "scopedVer" );

        return MingleParsers.parseTypeReference( str, scopedVer );
    }
}
