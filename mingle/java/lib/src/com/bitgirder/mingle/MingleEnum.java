package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class MingleEnum
extends TypedMingleValue
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleIdentifier value;

    MingleEnum( AtomicTypeReference typ,
                MingleIdentifier value )
    {
        super( typ ); 
        this.value = value;
    }

    public MingleIdentifier getValue() { return value; }

    public
    int
    hashCode()
    {
        return getType().hashCode() | value.hashCode();
    }

    public
    boolean
    equals( Object o )
    {
        if ( o == this ) return true;

        if ( o instanceof MingleEnum )
        {
            MingleEnum e = (MingleEnum) o;

            return getType().equals( e.getType() ) && value.equals( e.value );
        }

        return false;
    }

    public
    static
    MingleEnum
    create( AtomicTypeReference typ,
            MingleIdentifier value )
    {
        inputs.notNull( typ, "typ" );
        inputs.notNull( value, "value" );

        return new MingleEnum( typ, value );
    }
}
