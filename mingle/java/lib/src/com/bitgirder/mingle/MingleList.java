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

    private final ListTypeReference type;
    private final Iterable< MingleValue > vals;

    private 
    MingleList( ListTypeReference type,
                Iterable< MingleValue > vals )
    {   
        this.type = type;
        this.vals = vals; 
    }

    public Iterator< MingleValue > iterator() { return vals.iterator(); }

    public ListTypeReference type() { return type; }

    public int hashCode() { return type.hashCode() ^ vals.hashCode(); }

    public
    boolean
    equals( Object o )
    {
        if ( o == this ) return true;
        if ( ! ( o instanceof MingleList ) ) return false;

        MingleList o2 = (MingleList) o;

        return type.equals( o2.type ) && vals.equals( o2.vals );
    }

    static
    MingleList
    createLive( ListTypeReference type,
                Iterable< MingleValue > vals )
    {
        inputs.notNull( type, "type" );
        inputs.notNull( vals, "vals" );

        return new MingleList( type, vals );
    }

    public
    static
    MingleList
    asList( ListTypeReference type,
            List< MingleValue > vals )
    {
        inputs.notNull( type, "type" );
        inputs.noneNull( vals, "vals" );

        return new MingleList( type, vals );
    }

    public
    static
    MingleList
    asList( ListTypeReference type,
            MingleValue... vals )
    {
        inputs.notNull( vals, "vals" );

        return asList( type, Lang.< MingleValue >asList( vals ) );
    }

    // class impl is as-if it were public, but keeping package-only for the
    // moment.
    final
    static
    class Builder
    {
        private final List< MingleValue > l;
        private ListTypeReference type;

        public Builder() { l = Lang.newList(); }

        public
        Builder
        addUnsafe( MingleValue mv )
        {
            inputs.notNull( mv, "mv" );
            l.add( mv );

            return this;
        }

        public
        Builder
        setType( ListTypeReference type )
        {
            this.type = inputs.notNull( type, "type" );
            return this;
        }

        public
        MingleList
        buildLive()
        {
            return MingleList.createLive( type, l );
        }
    }
}
