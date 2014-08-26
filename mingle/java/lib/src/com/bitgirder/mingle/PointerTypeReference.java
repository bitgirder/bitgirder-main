package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class PointerTypeReference
extends MingleTypeReference
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleTypeReference type;

    PointerTypeReference( MingleTypeReference type )
    {
        this.type = inputs.notNull( type, "type" );
    }

    public MingleTypeReference getType() { return type; }

    public int hashCode() { return type.hashCode(); }

    public
    boolean
    equals( Object o )
    {
        return
            o == this ||
            ( ( o instanceof PointerTypeReference ) &&
              ( (PointerTypeReference) o ).type.equals( type ) );
    }

    public 
    CharSequence 
    getExternalForm() 
    { 
        return "&(" + type.getExternalForm() + ")";
    }

    public
    static
    PointerTypeReference
    create( MingleTypeReference type )
    {
        return new PointerTypeReference( inputs.notNull( type, "type" ) );
    }
}
