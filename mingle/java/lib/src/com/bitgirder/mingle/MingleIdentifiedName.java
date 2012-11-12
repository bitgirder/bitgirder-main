package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;
import java.util.Arrays;

public
final
class MingleIdentifiedName
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private MingleNamespace ns;
    private MingleIdentifier[] names;

    MingleIdentifiedName( MingleNamespace ns,
                          MingleIdentifier[] names )
    {
        this.ns = inputs.notNull( ns, "ns" );
        this.names = inputs.noneNull( names, "names" );
    }

    public MingleNamespace getNamespace() { return ns; }

    public
    List< MingleIdentifier >
    getNames()
    {
        return Lang.unmodifiableList( Lang.asList( names ) );
    }

    public
    CharSequence
    getExternalForm()
    {
        StringBuilder res = 
            new StringBuilder().
                append( ns.getExternalForm() );
        
        for ( MingleIdentifier id : names )
        {
            CharSequence idStr =
                id.format( MingleIdentifierFormat.LC_CAMEL_CAPPED );

            res.append( '/' ).
                append( idStr );
        }

        return res;
    }

    @Override public String toString() { return getExternalForm().toString(); }

    public int hashCode() { return ns.hashCode() | Arrays.hashCode( names ); }

    public
    boolean
    equals( Object o )
    {
        if ( o == this ) return true;
        else if ( o instanceof MingleIdentifiedName )
        {
            MingleIdentifiedName n = (MingleIdentifiedName) o;
            return ns.equals( n.ns ) && Arrays.equals( names, n.names );
        }
        else return false;
    }

    public
    static
    MingleIdentifiedName
    create( CharSequence str )
    {
        throw new UnsupportedOperationException( "Unimplemented" );
    }

    public
    static
    MingleIdentifiedName
    parse( CharSequence str )
        throws MingleSyntaxException
    {
        throw new UnsupportedOperationException( "Unimplemented" );
    }
}
