package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Inspector;

import java.util.List;
import java.util.Iterator;

abstract
class AbstractListMatch< T >
implements ProductionMatch< T >,
           Iterable< ProductionMatch< T > >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final List< ProductionMatch< T > > matches;

    // param checking done here on behalf of subclasses, on the assumption that
    // they also name their parameters 'matches'
    AbstractListMatch( List< ProductionMatch< T > > matches )
    {
        this.matches =
            Lang.unmodifiableCopy( inputs.noneNull( matches, "matches" ) );
    }

    public
    final
    ProductionMatch< T >
    get( int indx )
    {
        return matches.get( indx ); 
    }

    public final int size() { return matches.size(); }

    public
    final
    Inspector
    accept( Inspector i )
    {
        return i.add( "matches", matches );
    }

    public
    final
    Iterator< ProductionMatch< T > >
    iterator()
    {
        return matches.iterator();
    }
}
