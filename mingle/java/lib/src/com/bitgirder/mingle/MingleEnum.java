package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class MingleEnum
implements MingleValue
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final QualifiedTypeName type;
    private final MingleIdentifier value;

    MingleEnum( QualifiedTypeName type,
                MingleIdentifier value )
    {
        this.type = type;
        this.value = value;
    }

    public QualifiedTypeName getType() { return type; }
    public MingleIdentifier getValue() { return value; }

    public int hashCode() { return type.hashCode() | value.hashCode(); }

    public
    boolean
    equals( Object o )
    {
        if ( o == this ) return true;

        if ( o instanceof MingleEnum )
        {
            MingleEnum e = (MingleEnum) o;

            return type.equals( e.type ) && value.equals( e.value );
        }

        return false;
    }

    public
    static
    MingleEnum
    create( QualifiedTypeName name,
            MingleIdentifier value )
    {
        inputs.notNull( name, "name" );
        inputs.notNull( value, "value" );

        return new MingleEnum( name, value );
    }
}
