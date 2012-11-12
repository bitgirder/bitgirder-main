package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Inspector;
import com.bitgirder.lang.Inspectable;

public
final
class UnionMatch< T >
implements ProductionMatch< T >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final ProductionMatch< T > match;
    private final int alt;

    private
    UnionMatch( ProductionMatch< T > match,
                int alt )
    {
        this.match = match;
        this.alt = alt;
    }

    public ProductionMatch< T > getMatch() { return match; }
    public int getAlternative() { return alt; }

    public
    Inspector
    accept( Inspector i )
    {
        return i.add( "match", match ).
                 add( "alt", alt );
    }

    static
    < T >
    UnionMatch< T >
    create( ProductionMatch< T > match,
            int alt )
    {
        inputs.notNull( match, "match" );
        inputs.nonnegativeI( alt, "alt" );

        return new UnionMatch< T >( match, alt );
    }
}
