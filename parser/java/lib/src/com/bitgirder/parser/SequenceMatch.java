package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.util.List;

public
final
class SequenceMatch< T >
extends AbstractListMatch< T >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private
    SequenceMatch( List< ProductionMatch< T > > matches )
    {
        super( matches );
    }

    static
    < T >
    SequenceMatch< T >
    create( List< ProductionMatch< T > > matches )
    {
        return new SequenceMatch< T >( matches );
    }
}
