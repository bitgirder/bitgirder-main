package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Inspector;
import com.bitgirder.lang.Inspectable;

public
final
class DerivationMatch< N, T >
implements ProductionMatch< T >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final N head;
    private final ProductionMatch< T > match;

    private
    DerivationMatch( N head,
                     ProductionMatch< T > match )
    {
        this.head = head;
        this.match = match;
    }

    public N getHead() { return head; }
    public ProductionMatch< T > getMatch() { return match; }

    public
    Inspector
    accept( Inspector i )
    {
        return i.add( "head", head ).
                 add( "match", match );
    }

    static
    < N, T >
    DerivationMatch< N, T >
    create( N head,
            ProductionMatch< T > match )
    {
        inputs.notNull( head, "head" );
        inputs.notNull( match, "match" );

        return new DerivationMatch< N, T >( head, match );
    }
}
