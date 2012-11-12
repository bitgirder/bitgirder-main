package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.util.List;

public
final
class QuantifierMatch< T >
extends AbstractListMatch< T >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private QuantifierMatch( List< ProductionMatch< T > > matches )
    {
        super( matches );
    }

    static
    < T >
    QuantifierMatch< T >
    create( List< ProductionMatch< T > > matches )
    {
        return new QuantifierMatch< T >( matches );
    }
}
