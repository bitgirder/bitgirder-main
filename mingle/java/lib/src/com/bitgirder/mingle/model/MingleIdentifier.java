package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Strings;
import com.bitgirder.lang.Lang;

import com.bitgirder.parser.SyntaxException;

import com.bitgirder.mingle.parser.MingleParsers;

import java.util.Arrays;
import java.util.List;

public
final
class MingleIdentifier
implements Comparable< MingleIdentifier >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final String[] parts; // in canonical, lowercase form

    MingleIdentifier( String[] parts ) 
    { 
        this.parts = inputs.noneNull( parts, "parts" ); 
    }

    public int hashCode() { return Arrays.hashCode( parts ); }

    // meant only for package-level read-only access
    String[] getPartsArray() { return parts; }

    public 
    List< String > 
    getParts() 
    { 
        return Lang.unmodifiableList( Arrays.asList( parts ) ); 
    }

    public
    boolean
    equals( Object other )
    {
        return
            other == this ||
            ( other instanceof MingleIdentifier &&
              Arrays.equals( parts, ( (MingleIdentifier) other ).parts ) );
    }

    public CharSequence getExternalForm() { return Strings.join( "-", parts ); }

    @Override 
    public final String toString() { return getExternalForm().toString(); }

    // we'll optimize this down the road when needed (obvious choice would be to
    // proceed lexicographically through the parts array of each)
    public
    int
    compareTo( MingleIdentifier other )
    {
        if ( other == null ) throw new NullPointerException();
        else return toString().compareTo( other.toString() );
    }

    public
    static
    MingleIdentifier
    create( CharSequence str )
    {
        return MingleParsers.createIdentifier( inputs.notNull( str, "str" ) );
    }

    public
    static
    MingleIdentifier
    parse( CharSequence str )
        throws SyntaxException
    {
        return MingleParsers.parseIdentifier( inputs.notNull( str, "str" ) );
    }

    public
    static
    MingleIdentifier
    create( Enum< ? > e )
    {
        return create( inputs.notNull( e, "e" ).name().toLowerCase() );
    }

    public
    static
    MingleIdentifier
    parse( Enum< ? > e )
        throws SyntaxException
    {
        return parse( inputs.notNull( e, "e" ).name().toLowerCase() );
    }
}
