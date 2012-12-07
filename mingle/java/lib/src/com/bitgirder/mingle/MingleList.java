package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;
import java.util.Iterator;

public
final
class MingleList
implements Iterable< MingleValue >,
           MingleValue
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static MingleList EMPTY =
        new MingleList( Lang.< MingleValue >emptyList() );

    private final Iterable< MingleValue > vals;

    private MingleList( Iterable< MingleValue > vals ) { this.vals = vals; }

    public Iterator< MingleValue > iterator() { return vals.iterator(); }

    static
    MingleList
    createLive( Iterable< MingleValue > vals )
    {
        return new MingleList( state.notNull( vals, "vals" ) );
    }

    public
    static
    MingleList
    asList( List< MingleValue > vals )
    {
        return new MingleList( inputs.noneNull( vals, "vals" ) );
    }

    public
    static
    MingleList
    asList( MingleValue... vals )
    {
        inputs.notNull( vals, "vals" );

        return asList( Lang.< MingleValue >asList( vals ) );
    }

    public static MingleList empty() { return EMPTY; }
}
