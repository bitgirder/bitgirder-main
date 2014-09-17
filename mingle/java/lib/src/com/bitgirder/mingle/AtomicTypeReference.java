package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class AtomicTypeReference
extends MingleTypeReference
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final QualifiedTypeName name;
    private final MingleValueRestriction restriction;

    AtomicTypeReference( QualifiedTypeName name,
                         MingleValueRestriction restriction )
    {
        this.name = inputs.notNull( name, "name" );
        this.restriction = restriction;
    }

    public QualifiedTypeName getName() { return name; }

    public MingleValueRestriction getRestriction() { return restriction; }

    public AtomicTypeReference asUnrestrictedType() { return create( name ); }

    public int hashCode() { return name.hashCode(); }

    public 
    boolean
    equals( Object other )
    {
        if ( other == this ) return true;

        if ( other instanceof AtomicTypeReference )
        {
            AtomicTypeReference o = (AtomicTypeReference) other;

            if ( o.name.equals( name ) )
            {
                if ( restriction == null ) return o.restriction == null;
                else return restriction.equals( o.restriction );
            }
            else return false;
        }
        else return false;
    }

    public 
    String
    getExternalForm() 
    { 
        String nmStr = name.getExternalForm();

        if ( restriction == null ) return nmStr;
            
        StringBuilder sb =
            new StringBuilder().
                append( nmStr ).
                append( "~" );
        
        restriction.appendExternalForm( sb );
        return sb.toString();
    }

    public
    static
    AtomicTypeReference
    create( QualifiedTypeName nm )
    {
        return new AtomicTypeReference( nm, null );
    }

    public
    static
    AtomicTypeReference
    create( QualifiedTypeName nm,
            MingleValueRestriction restriction )
    {
        inputs.notNull( restriction, "restriction" );
        return new AtomicTypeReference( nm, restriction );
    }
}
