package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;

public
final
class MingleTypeCastException
extends MingleValueException
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleTypeReference expected;
    private final MingleTypeReference actual;

    MingleTypeCastException( MingleTypeReference expected,
                             MingleTypeReference actual,
                             ObjectPath< MingleIdentifier > loc )
    {
        super(
            "Expected value of type " +
                inputs.notNull( expected, "expected" ).getExternalForm() +
                " but found " +
                inputs.notNull( actual, "actual" ).getExternalForm(),
            loc
        );

        this.expected = expected;
        this.actual = actual;
    }

    public MingleTypeReference expected() { return expected; }
    public MingleTypeReference actual() { return actual; }
}
