package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class NullableTypeReference
extends MingleTypeReference
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleTypeReference ref;

    NullableTypeReference( MingleTypeReference ref )
    {
        this.ref = inputs.notNull( ref, "ref" );
    }

    public MingleTypeReference getTypeReference() { return ref; }

    public
    int
    hashCode()
    {
        int base = ref.hashCode();

        int left = base & 0xff00;
        int right = base & 0x00ff;

        return ( ( left ^ right ) << 15 ) | ( right ^ left );
    }

    public
    boolean
    equals( Object o )
    {
        return
            o == this ||
            ( ( o instanceof NullableTypeReference ) &&
              ( (NullableTypeReference) o ).ref.equals( ref ) );
    }

    public 
    CharSequence 
    getExternalForm() 
    { 
        return ref.getExternalForm() + "?"; 
    }

    public
    static
    NullableTypeReference
    create( MingleTypeReference ref )
    {
        return new NullableTypeReference( inputs.notNull( ref, "ref" ) );
    }
}
