package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;
import java.util.Iterator;

public
final
class MingleList
implements MingleValue,
           Iterable< MingleValue >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final List< MingleValue > list;

    private
    MingleList( Builder b )
    {
        this.list = Lang.unmodifiableCopy( b.list );
    }

    public Iterator< MingleValue > iterator() { return list.iterator(); }

    // can provide a more efficient impl as needed later
    public
    static
    MingleList
    create( MingleValue... vals )
    {
        inputs.notNull( vals, "vals" );
        
        Builder b = new Builder();
        for ( MingleValue mv : vals ) b.add( mv );

        return b.build();
    }

    public
    final
    static
    class Builder
    {
        private final List< MingleValue > list;

        private Builder( List< MingleValue > list ) { this.list = list; }

        public Builder() { this( Lang.< MingleValue >newList() ); }

        public
        Builder( int initialCapacity )
        {
            this(
                Lang.< MingleValue >newList(
                    inputs.positiveI( initialCapacity, "initialCapacity" ) ) );
        }

        public
        Builder
        add( MingleValue v )
        {
            list.add( inputs.notNull( v, "v" ) );
            return this;
        }

        public MingleList build() { return new MingleList( this ); }
    }
}
