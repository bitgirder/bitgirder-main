package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class AtomicTypeReference
extends MingleTypeReference
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Name name;
    private final MingleValueRestriction restriction;

    private
    AtomicTypeReference( Name name,
                         MingleValueRestriction restriction )
    {
        this.name = inputs.notNull( name, "name" );
        this.restriction = restriction;
    }

    public Name getName() { return name; }

    public MingleValueRestriction getRestriction() { return restriction; }

    public int hashCode() { return name.hashCode(); }

    public 
    boolean
    equals( Object other )
    {
        if ( other == this ) return true;
        else if ( other instanceof AtomicTypeReference )
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
    CharSequence 
    getExternalForm() 
    { 
        CharSequence nmStr = name.getExternalForm();

        if ( restriction == null ) return nmStr;
        else
        {
            StringBuilder sb =
                new StringBuilder().
                    append( nmStr ).
                    append( "~" );
            
            restriction.appendExternalForm( sb );
            return sb;
        }
    }

    public
    static
    AtomicTypeReference
    create( Name nm )
    {
        return new AtomicTypeReference( nm, null );
    }

    public
    static
    AtomicTypeReference
    create( Name nm,
            MingleValueRestriction restriction )
    {
        inputs.notNull( restriction, "restriction" );
        return new AtomicTypeReference( nm, restriction );
    }

    public
    static
    interface Name
    {
        public
        CharSequence 
        getExternalForm();
    }
}
