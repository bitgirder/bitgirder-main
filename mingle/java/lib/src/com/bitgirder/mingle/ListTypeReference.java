package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class ListTypeReference
extends MingleTypeReference
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleTypeReference ref;
    private final boolean allowsEmpty;

    ListTypeReference( MingleTypeReference ref,
                       boolean allowsEmpty )
    {
        this.ref = inputs.notNull( ref, "ref" );
        this.allowsEmpty = allowsEmpty;
    }

    public MingleTypeReference getElementType() { return ref; }
    public boolean allowsEmpty() { return allowsEmpty; }

    public
    int
    hashCode()
    {
        int res = ref.hashCode();
        if ( allowsEmpty ) res = res ^ Boolean.TRUE.hashCode();

        return res;
    }

    public
    boolean
    equals( Object o )
    {
        if ( o == this ) return true;
        else if ( o instanceof ListTypeReference )
        {
            ListTypeReference l = (ListTypeReference) o;
            return l.allowsEmpty == allowsEmpty && l.ref.equals( ref );
        }
        else return false;
    }

    public
    String
    getExternalForm()
    {
        return ref.getExternalForm() + ( allowsEmpty ? "*" : "+" );
    }

    public
    static
    ListTypeReference
    create( MingleTypeReference ref,
            boolean allowsEmpty )
    {
        inputs.notNull( ref, "ref" );
        return new ListTypeReference( ref, allowsEmpty );
    }
}
