package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;

public
abstract
class MingleValueRestriction
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public
    abstract
    void
    validate( MingleValue mv,
              ObjectPath< MingleIdentifier > path );

    public
    abstract
    int
    hashCode();

    public
    abstract
    boolean
    equals( Object other );

    abstract
    void
    appendExternalForm( StringBuilder sb );

    @Override
    public
    final
    String
    toString()
    {
        StringBuilder sb = new StringBuilder();
        appendExternalForm( sb );

        return sb.toString();
    }

    public final CharSequence getExternalForm() { return toString(); }
}
