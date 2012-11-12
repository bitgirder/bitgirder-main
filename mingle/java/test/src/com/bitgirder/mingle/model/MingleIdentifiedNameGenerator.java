package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.util.concurrent.atomic.AtomicInteger;

public
final
class MingleIdentifiedNameGenerator
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final String prefix;

    private final AtomicInteger idNum = new AtomicInteger();

    private
    MingleIdentifiedNameGenerator( CharSequence prefix )
    {
        this.prefix = prefix.toString();
    }

    public
    MingleIdentifiedName
    next()
    {
        return MingleIdentifiedName.create( prefix + idNum.getAndIncrement() );
    }

    public
    static
    MingleIdentifiedNameGenerator
    forPrefix( CharSequence prefix )
    {
        inputs.notNull( prefix, "prefix" );
        return new MingleIdentifiedNameGenerator( prefix );
    }
}
