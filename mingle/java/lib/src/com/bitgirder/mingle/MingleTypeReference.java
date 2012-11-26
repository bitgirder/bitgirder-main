package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

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
        inputs.notNull( str, "str" );

        return 
            MingleParser.createTypeReference( str, Mingle.CORE_NAME_RESOLVER );
    }

    public
    static
    MingleTypeReference
    parse( CharSequence str )
        throws MingleSyntaxException
    {
        inputs.notNull( str, "str" );

        return 
            MingleParser.parseTypeReference( str, Mingle.CORE_NAME_RESOLVER );
    }
}
